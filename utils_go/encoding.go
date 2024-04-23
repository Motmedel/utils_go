package utils_go

import (
	"encoding/json"
	"io"
)

func DecodeJson[T any](reader io.Reader) (T, error) {
	var obj T
	err := json.NewDecoder(reader).Decode(&obj)
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
