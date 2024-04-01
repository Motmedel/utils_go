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
