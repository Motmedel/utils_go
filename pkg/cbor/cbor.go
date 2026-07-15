// Package cbor implements the CBOR (RFC 8949) subset used by COSE structures: integers, byte
// strings, text strings, arrays, maps, tags, booleans, null, and undefined. Encoding is
// deterministic (definite lengths, minimal integer encoding, bytewise-sorted map keys). Decoding
// is strict: indefinite lengths, floating-point values, extended simple values, duplicate map
// keys, non-integer and non-string map keys, excessive nesting, and trailing data are rejected.
package cbor

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"sort"
	"unicode/utf8"
)

var (
	ErrUnsupportedValue = errors.New("unsupported value")
	ErrMalformed        = errors.New("malformed data")
)

// maxDepth bounds the nesting depth of decoded data. COSE structures nest a handful of levels.
const maxDepth = 32

// Undefined is the CBOR "undefined" simple value.
type Undefined struct{}

// Tag is a CBOR tagged data item.
type Tag struct {
	Number  uint64
	Content any
}

func writeTypeAndArgument(buffer *bytes.Buffer, majorType byte, argument uint64) {
	switch {
	case argument < 24:
		buffer.WriteByte(majorType<<5 | byte(argument))
	case argument <= math.MaxUint8:
		buffer.WriteByte(majorType<<5 | 24)
		buffer.WriteByte(byte(argument))
	case argument <= math.MaxUint16:
		buffer.WriteByte(majorType<<5 | 25)
		buffer.WriteByte(byte(argument >> 8))
		buffer.WriteByte(byte(argument))
	case argument <= math.MaxUint32:
		buffer.WriteByte(majorType<<5 | 26)
		for shift := 24; shift >= 0; shift -= 8 {
			buffer.WriteByte(byte(argument >> shift))
		}
	default:
		buffer.WriteByte(majorType<<5 | 27)
		for shift := 56; shift >= 0; shift -= 8 {
			buffer.WriteByte(byte(argument >> shift))
		}
	}
}

func encodeInt64(buffer *bytes.Buffer, value int64) {
	if value >= 0 {
		writeTypeAndArgument(buffer, 0, uint64(value))
	} else {
		writeTypeAndArgument(buffer, 1, uint64(^value))
	}
}

func encodeValue(buffer *bytes.Buffer, value any) error {
	switch typedValue := value.(type) {
	case nil:
		buffer.WriteByte(0xf6)
	case Undefined:
		buffer.WriteByte(0xf7)
	case bool:
		if typedValue {
			buffer.WriteByte(0xf5)
		} else {
			buffer.WriteByte(0xf4)
		}
	case int:
		encodeInt64(buffer, int64(typedValue))
	case int64:
		encodeInt64(buffer, typedValue)
	case uint64:
		writeTypeAndArgument(buffer, 0, typedValue)
	case []byte:
		writeTypeAndArgument(buffer, 2, uint64(len(typedValue)))
		buffer.Write(typedValue)
	case string:
		writeTypeAndArgument(buffer, 3, uint64(len(typedValue)))
		buffer.WriteString(typedValue)
	case []any:
		writeTypeAndArgument(buffer, 4, uint64(len(typedValue)))
		for _, item := range typedValue {
			if err := encodeValue(buffer, item); err != nil {
				return err
			}
		}
	case map[int64]any:
		entries := make(map[any]any, len(typedValue))
		for key, item := range typedValue {
			entries[key] = item
		}
		return encodeMap(buffer, entries)
	case map[any]any:
		return encodeMap(buffer, typedValue)
	case Tag:
		writeTypeAndArgument(buffer, 6, typedValue.Number)
		return encodeValue(buffer, typedValue.Content)
	default:
		return fmt.Errorf("%w: %T", ErrUnsupportedValue, value)
	}

	return nil
}

func encodeMap(buffer *bytes.Buffer, entries map[any]any) error {
	type encodedEntry struct {
		key  []byte
		item []byte
	}

	encodedEntries := make([]encodedEntry, 0, len(entries))
	for key, item := range entries {
		keyData, err := Encode(key)
		if err != nil {
			return fmt.Errorf("encode map key: %w", err)
		}

		itemData, err := Encode(item)
		if err != nil {
			return fmt.Errorf("encode map value: %w", err)
		}

		encodedEntries = append(encodedEntries, encodedEntry{key: keyData, item: itemData})
	}

	sort.Slice(encodedEntries, func(i, j int) bool {
		return bytes.Compare(encodedEntries[i].key, encodedEntries[j].key) < 0
	})

	writeTypeAndArgument(buffer, 5, uint64(len(encodedEntries)))
	for _, entry := range encodedEntries {
		buffer.Write(entry.key)
		buffer.Write(entry.item)
	}

	return nil
}

// Encode serializes a value deterministically (RFC 8949, Section 4.2.1). Supported types: nil,
// Undefined, bool, int, int64, uint64, []byte, string, []any, map[int64]any, map[any]any, and Tag.
func Encode(value any) ([]byte, error) {
	var buffer bytes.Buffer
	if err := encodeValue(&buffer, value); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

type decoder struct {
	data   []byte
	offset int
	noCopy bool
}

func (d *decoder) readTypeAndArgument() (byte, uint64, byte, error) {
	if d.offset >= len(d.data) {
		return 0, 0, 0, fmt.Errorf("%w: unexpected end of data", ErrMalformed)
	}

	initialByte := d.data[d.offset]
	d.offset++

	majorType := initialByte >> 5
	additionalInformation := initialByte & 0x1f

	if additionalInformation < 24 {
		return majorType, uint64(additionalInformation), additionalInformation, nil
	}

	if additionalInformation == 31 {
		return 0, 0, 0, fmt.Errorf("%w: indefinite length", ErrMalformed)
	}

	if additionalInformation > 27 {
		return 0, 0, 0, fmt.Errorf(
			"%w: additional information %d",
			ErrMalformed,
			additionalInformation,
		)
	}

	argumentSize := 1 << (additionalInformation - 24)
	if d.offset+argumentSize > len(d.data) {
		return 0, 0, 0, fmt.Errorf("%w: unexpected end of data", ErrMalformed)
	}

	var argument uint64
	for _, argumentByte := range d.data[d.offset : d.offset+argumentSize] {
		argument = argument<<8 | uint64(argumentByte)
	}
	d.offset += argumentSize

	return majorType, argument, additionalInformation, nil
}

// readSlice returns the next length bytes of the input without copying.
func (d *decoder) readSlice(length uint64) ([]byte, error) {
	if length > uint64(len(d.data)-d.offset) {
		return nil, fmt.Errorf("%w: unexpected end of data", ErrMalformed)
	}

	end := d.offset + int(length)
	data := d.data[d.offset:end:end]
	d.offset = end

	return data, nil
}

func (d *decoder) readBytes(length uint64) ([]byte, error) {
	slicedData, err := d.readSlice(length)
	if err != nil {
		return nil, err
	}

	// The capacity of the returned slice is capped so that appending to it cannot overwrite input
	// bytes following the byte string.
	if d.noCopy {
		return slicedData, nil
	}

	return bytes.Clone(slicedData), nil
}

func (d *decoder) decodeValue(depth int) (any, error) {
	if depth > maxDepth {
		return nil, fmt.Errorf("%w: excessive nesting", ErrMalformed)
	}

	majorType, argument, additionalInformation, err := d.readTypeAndArgument()
	if err != nil {
		return nil, err
	}

	switch majorType {
	case 0:
		if argument > math.MaxInt64 {
			return nil, fmt.Errorf("%w: unsupported integer %d", ErrMalformed, argument)
		}
		return int64(argument), nil
	case 1:
		if argument > math.MaxInt64 {
			return nil, fmt.Errorf("%w: unsupported integer -%d", ErrMalformed, argument)
		}
		return -1 - int64(argument), nil
	case 2:
		return d.readBytes(argument)
	case 3:
		data, err := d.readSlice(argument)
		if err != nil {
			return nil, err
		}
		if !utf8.Valid(data) {
			return nil, fmt.Errorf("%w: invalid utf-8 in text string", ErrMalformed)
		}
		// The string conversion is the single copy; strings are immutable and never alias the
		// input.
		return string(data), nil
	case 4:
		// Each element occupies at least one byte.
		if argument > uint64(len(d.data)-d.offset) {
			return nil, fmt.Errorf("%w: array length %d exceeds remaining data", ErrMalformed, argument)
		}

		array := make([]any, 0, argument)
		for i := uint64(0); i < argument; i++ {
			item, err := d.decodeValue(depth + 1)
			if err != nil {
				return nil, err
			}
			array = append(array, item)
		}
		return array, nil
	case 5:
		// Each entry occupies at least two bytes.
		if argument > uint64(len(d.data)-d.offset)/2 {
			return nil, fmt.Errorf("%w: map length %d exceeds remaining data", ErrMalformed, argument)
		}

		entries := make(map[any]any, argument)
		for i := uint64(0); i < argument; i++ {
			key, err := d.decodeValue(depth + 1)
			if err != nil {
				return nil, err
			}

			switch key.(type) {
			case int64, string:
			default:
				return nil, fmt.Errorf("%w: unsupported map key type %T", ErrMalformed, key)
			}

			if _, ok := entries[key]; ok {
				return nil, fmt.Errorf("%w: duplicate map key %v", ErrMalformed, key)
			}

			item, err := d.decodeValue(depth + 1)
			if err != nil {
				return nil, err
			}

			entries[key] = item
		}
		return entries, nil
	case 6:
		content, err := d.decodeValue(depth + 1)
		if err != nil {
			return nil, err
		}
		return Tag{Number: argument, Content: content}, nil
	default:
		if additionalInformation >= 24 {
			return nil, fmt.Errorf("%w: floats and extended simple values are not supported", ErrMalformed)
		}
		switch argument {
		case 20:
			return false, nil
		case 21:
			return true, nil
		case 22:
			return nil, nil
		case 23:
			return Undefined{}, nil
		default:
			return nil, fmt.Errorf("%w: unsupported simple value %d", ErrMalformed, argument)
		}
	}
}

func decode(parser *decoder) (any, error) {
	value, err := parser.decodeValue(0)
	if err != nil {
		return nil, err
	}

	if parser.offset != len(parser.data) {
		return nil, fmt.Errorf("%w: trailing data", ErrMalformed)
	}

	return value, nil
}

// Decode deserializes a single value, rejecting trailing data. Integers are returned as int64,
// byte strings as []byte, text strings as string, arrays as []any, maps as map[any]any (with
// int64 or string keys), and tagged data items as Tag.
func Decode(data []byte) (any, error) {
	return decode(&decoder{data: data})
}

// DecodeNoCopy is Decode, except decoded byte strings alias data instead of being copied. The
// caller must keep data alive and unmodified for as long as the decoded byte strings are in use.
// The aliasing slices have their capacity capped, so appending to them does not modify data.
func DecodeNoCopy(data []byte) (any, error) {
	return decode(&decoder{data: data, noCopy: true})
}
