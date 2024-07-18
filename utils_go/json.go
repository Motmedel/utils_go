package utils_go

import "encoding/json"

func JsonUnmarshal(data []byte, v any) error {
	err := json.Unmarshal(data, &v)
	if typedErr, ok := err.(*json.SyntaxError); ok {
		return JsonSyntaxError{
			SyntaxError: typedErr,
			InputError: &InputError{
				err,
				data,
			},
		}
	}
	return err
}
