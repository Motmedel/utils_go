package utils_go

import (
	"encoding/json"
	"io"
)

func DecodeJson[T any](reader io.Reader) (T, error) {
	var obj T

	data, err := io.ReadAll(reader)
	if err != nil {
		return obj, &CauseError{Message: "An error occurred when reading the data.", Cause: err}
	}

	if err := json.Unmarshal(data, &obj); err != nil {
		return obj, &InputError{
			Message: "An error occurred when unmarshalling the data.",
			Cause:   err,
			Input:   data,
		}
	}

	return obj, err
}

func JsonObjectToMap[T any, S comparable](value T) (map[S]any, error) {
	var valueBytes []byte
	var err error

	valueBytes, err = json.Marshal(value)
	if err != nil {
		return nil, err
	}

	var valueMap map[S]any
	err = json.Unmarshal(valueBytes, &valueMap)

	return valueMap, err
}
