package claims_strings

import (
	"encoding/json"
	"fmt"

	"github.com/Motmedel/utils_go/pkg/errors"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/utils"
)

var MarshalSingleStringAsArray = true

type ClaimStrings []string

func (s *ClaimStrings) UnmarshalJSON(data []byte) (err error) {
	var value any

	if err = json.Unmarshal(data, &value); err != nil {
		return err
	}

	var aud []string

	switch v := value.(type) {
	case string:
		aud = append(aud, v)
	case []string:
		aud = v
	case []any:
		for _, vv := range v {
			vs, err := utils.Convert[string](vv)
			if err != nil {
				return errors.NewWithTrace(fmt.Errorf("convert: %w", err), vv)
			}
			aud = append(aud, vs)
		}
	case nil:
		return nil
	default:
		return errors.NewWithTrace(fmt.Errorf("%w: %T", motmedelErrors.ErrUnexpectedType, v), v)
	}

	*s = aud

	return
}

func (s ClaimStrings) MarshalJSON() (b []byte, err error) {
	// This handles a special case in the JWT RFC. If the string array, e.g.
	// used by the "aud" field, only contains one element, it MAY be serialized
	// as a single string.
	if len(s) == 1 && !MarshalSingleStringAsArray {
		return json.Marshal(s[0])
	}

	return json.Marshal([]string(s))
}

func Convert(value any) (ClaimStrings, error) {
	var claimsString []string

	switch typedValue := value.(type) {
	case string:
		claimsString = append(claimsString, typedValue)
	case []string:
		claimsString = typedValue
	case []any:
		for _, a := range typedValue {
			vs, err := utils.Convert[string](a)
			if err != nil {
				return nil, errors.NewWithTrace(fmt.Errorf("convert: %w", err), a)
			}
			claimsString = append(claimsString, vs)
		}
	default:
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("%w: %T", motmedelErrors.ErrUnexpectedType, typedValue),
			typedValue,
		)
	}

	return claimsString, nil
}
