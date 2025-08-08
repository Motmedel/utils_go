package numeric_date

import (
	"encoding/json"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"math"
	"strconv"
	"time"
)

var TimePrecision = time.Second

// NumericDate represents a JSON numeric date value, as referenced at
// https://datatracker.ietf.org/doc/html/rfc7519#section-2.
type NumericDate struct {
	time.Time
}

// MarshalJSON is an implementation of the json.RawMessage interface and serializes the UNIX epoch
// represented in NumericDate to a byte array, using the precision specified in TimePrecision.
func (date NumericDate) MarshalJSON() (b []byte, err error) {
	var prec int
	if TimePrecision < time.Second {
		prec = int(math.Log10(float64(time.Second) / float64(TimePrecision)))
	}
	truncatedDate := date.Truncate(TimePrecision)

	// For very large timestamps, UnixNano would overflow an int64, but this
	// function requires nanosecond level precision, so we have to use the
	// following technique to get round the issue:
	//
	// 1. Take the normal unix timestamp to form the whole number part of the
	//    output,
	// 2. Take the result of the Nanosecond function, which returns the offset
	//    within the second of the particular unix time instance, to form the
	//    decimal part of the output
	// 3. Concatenate them to produce the final result
	seconds := strconv.FormatInt(truncatedDate.Unix(), 10)
	nanosecondsOffset := strconv.FormatFloat(float64(truncatedDate.Nanosecond())/float64(time.Second), 'f', prec, 64)

	output := append([]byte(seconds), []byte(nanosecondsOffset)[1:]...)

	return output, nil
}

// UnmarshalJSON is an implementation of the json.RawMessage interface and
// deserializes a [NumericDate] from a JSON representation, i.e. a
// [json.Number]. This number represents a UNIX epoch with either integer or
// non-integer seconds.
func (date *NumericDate) UnmarshalJSON(b []byte) (err error) {
	var (
		number json.Number
		f      float64
	)

	if err = json.Unmarshal(b, &number); err != nil {
		return motmedelErrors.NewWithTrace(fmt.Errorf("json unmarshal: %w", err), b)
	}

	if f, err = number.Float64(); err != nil {
		return motmedelErrors.NewWithTrace(fmt.Errorf("json number float64: %w", err), number)
	}

	n := FromSeconds(f)
	*date = *n

	return nil
}

func FromSeconds(f float64) *NumericDate {
	round, frac := math.Modf(f)
	return New(time.Unix(int64(round), int64(frac*1e9)))
}

func New(t time.Time) *NumericDate {
	return &NumericDate{t.Truncate(TimePrecision)}
}

func Convert(value any) (*NumericDate, error) {
	switch typedValue := value.(type) {
	case float64:
		if typedValue == 0 {
			return nil, nil
		}

		return FromSeconds(typedValue), nil
	case json.Number:
		typedFloatValue, err := typedValue.Float64()
		if err != nil {
			return nil, motmedelErrors.NewWithTrace(fmt.Errorf("json number float64: %w", err), typedValue)
		}

		return FromSeconds(typedFloatValue), nil
	default:
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("%w: %T", motmedelErrors.ErrUnexpectedType, typedValue),
			typedValue,
		)
	}
}
