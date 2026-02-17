package uuid

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"io"
	"time"
)

// UUID is a 128-bit universally unique identifier.
type UUID [16]byte

// String returns the string representation of the UUID in the format
// xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.
func (uuid UUID) String() string {
	var buf [36]byte
	encodeHex(buf[:], uuid)
	return string(buf[:])
}

func encodeHex(dst []byte, uuid UUID) {
	hex.Encode(dst, uuid[:4])
	dst[8] = '-'
	hex.Encode(dst[9:13], uuid[4:6])
	dst[13] = '-'
	hex.Encode(dst[14:18], uuid[6:8])
	dst[18] = '-'
	hex.Encode(dst[19:23], uuid[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:], uuid[10:])
}

// New generates a new random UUID v4.
func New() UUID {
	uuid, err := NewRandom()
	if err != nil {
		// Fall back to a nil UUID if random generation fails.
		return UUID{}
	}
	return uuid
}

// NewRandom generates a new random UUID v4 and returns an error if the random
// source fails.
func NewRandom() (UUID, error) {
	var uuid UUID
	_, err := io.ReadFull(rand.Reader, uuid[:])
	if err != nil {
		return UUID{}, err
	}
	// Set version (4) and variant (RFC 4122).
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10
	return uuid, nil
}

// NewV7 generates a new UUID v7 based on the current timestamp.
// UUID v7 uses a Unix timestamp in milliseconds for time-ordering.
func NewV7() (UUID, error) {
	var uuid UUID

	// Get current time in milliseconds since Unix epoch.
	now := time.Now()
	ms := now.UnixMilli()

	// Encode the 48-bit timestamp (big-endian) in the first 6 bytes.
	uuid[0] = byte(ms >> 40)
	uuid[1] = byte(ms >> 32)
	uuid[2] = byte(ms >> 24)
	uuid[3] = byte(ms >> 16)
	uuid[4] = byte(ms >> 8)
	uuid[5] = byte(ms)

	// Fill the remaining bytes with random data.
	_, err := io.ReadFull(rand.Reader, uuid[6:])
	if err != nil {
		return UUID{}, err
	}

	// Set version (7) and variant (RFC 4122).
	uuid[6] = (uuid[6] & 0x0f) | 0x70 // Version 7
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10

	return uuid, nil
}

// NewString generates a new random UUID v4 and returns it as a string.
func NewString() string {
	return New().String()
}

// Timestamp returns the timestamp from a UUID v7.
// For non-v7 UUIDs, the result is undefined.
func (uuid UUID) Timestamp() time.Time {
	ms := int64(uuid[0])<<40 |
		int64(uuid[1])<<32 |
		int64(uuid[2])<<24 |
		int64(uuid[3])<<16 |
		int64(uuid[4])<<8 |
		int64(uuid[5])
	return time.UnixMilli(ms)
}

// Version returns the version of the UUID.
func (uuid UUID) Version() int {
	return int(uuid[6] >> 4)
}

// Variant returns the variant of the UUID.
func (uuid UUID) Variant() int {
	switch {
	case (uuid[8] & 0x80) == 0x00:
		return 0 // Reserved, NCS backward compatibility.
	case (uuid[8] & 0xc0) == 0x80:
		return 1 // RFC 4122
	case (uuid[8] & 0xe0) == 0xc0:
		return 2 // Reserved, Microsoft Corporation backward compatibility.
	default:
		return 3 // Reserved for future definition.
	}
}

// Bytes returns the raw bytes of the UUID.
func (uuid UUID) Bytes() []byte {
	return uuid[:]
}

// Parse parses a UUID string in the format xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.
func Parse(s string) (UUID, error) {
	var uuid UUID
	if len(s) != 36 {
		return uuid, ErrInvalidUUIDLength
	}
	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return uuid, ErrInvalidUUIDFormat
	}

	// Decode each section.
	if _, err := hex.Decode(uuid[0:4], []byte(s[0:8])); err != nil {
		return uuid, ErrInvalidUUIDFormat
	}
	if _, err := hex.Decode(uuid[4:6], []byte(s[9:13])); err != nil {
		return uuid, ErrInvalidUUIDFormat
	}
	if _, err := hex.Decode(uuid[6:8], []byte(s[14:18])); err != nil {
		return uuid, ErrInvalidUUIDFormat
	}
	if _, err := hex.Decode(uuid[8:10], []byte(s[19:23])); err != nil {
		return uuid, ErrInvalidUUIDFormat
	}
	if _, err := hex.Decode(uuid[10:16], []byte(s[24:36])); err != nil {
		return uuid, ErrInvalidUUIDFormat
	}

	return uuid, nil
}

// MustParse parses a UUID string and panics if it fails.
func MustParse(s string) UUID {
	uuid, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return uuid
}

// FromBytes creates a UUID from a byte slice.
func FromBytes(b []byte) (UUID, error) {
	var uuid UUID
	if len(b) != 16 {
		return uuid, ErrInvalidUUIDLength
	}
	copy(uuid[:], b)
	return uuid, nil
}

// URN returns the UUID as a URN (urn:uuid:...).
func (uuid UUID) URN() string {
	return "urn:uuid:" + uuid.String()
}

// IsNil returns true if the UUID is the nil UUID (all zeros).
func (uuid UUID) IsNil() bool {
	return uuid == UUID{}
}

// SetVersion sets the version bits of the UUID.
func (uuid *UUID) SetVersion(v byte) {
	uuid[6] = (uuid[6] & 0x0f) | (v << 4)
}

// SetVariant sets the variant bits to RFC 4122.
func (uuid *UUID) SetVariant() {
	uuid[8] = (uuid[8] & 0x3f) | 0x80
}

// Equal returns true if two UUIDs are equal.
func (uuid UUID) Equal(other UUID) bool {
	return uuid == other
}

// MarshalText implements encoding.TextMarshaler.
func (uuid UUID) MarshalText() ([]byte, error) {
	var buf [36]byte
	encodeHex(buf[:], uuid)
	return buf[:], nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (uuid *UUID) UnmarshalText(data []byte) error {
	parsed, err := Parse(string(data))
	if err != nil {
		return err
	}
	*uuid = parsed
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler.
func (uuid UUID) MarshalBinary() ([]byte, error) {
	return uuid[:], nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler.
func (uuid *UUID) UnmarshalBinary(data []byte) error {
	if len(data) != 16 {
		return ErrInvalidUUIDLength
	}
	copy(uuid[:], data)
	return nil
}

// NewV7FromTime generates a UUID v7 from a specific timestamp.
func NewV7FromTime(t time.Time) (UUID, error) {
	var uuid UUID

	ms := t.UnixMilli()

	uuid[0] = byte(ms >> 40)
	uuid[1] = byte(ms >> 32)
	uuid[2] = byte(ms >> 24)
	uuid[3] = byte(ms >> 16)
	uuid[4] = byte(ms >> 8)
	uuid[5] = byte(ms)

	_, err := io.ReadFull(rand.Reader, uuid[6:])
	if err != nil {
		return UUID{}, err
	}

	uuid[6] = (uuid[6] & 0x0f) | 0x70
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return uuid, nil
}

// ToUint64Pair returns the UUID as two uint64 values (high and low).
func (uuid UUID) ToUint64Pair() (uint64, uint64) {
	high := binary.BigEndian.Uint64(uuid[0:8])
	low := binary.BigEndian.Uint64(uuid[8:16])
	return high, low
}

// FromUint64Pair creates a UUID from two uint64 values.
func FromUint64Pair(high, low uint64) UUID {
	var uuid UUID
	binary.BigEndian.PutUint64(uuid[0:8], high)
	binary.BigEndian.PutUint64(uuid[8:16], low)
	return uuid
}
