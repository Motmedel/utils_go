package json

import (
	"encoding/json"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"io"
)

func DecodeJson[T any](reader io.Reader) (T, error) {
	var obj T

	data, err := io.ReadAll(reader)
	if err != nil {
		return obj, motmedelErrors.New(fmt.Errorf("io read all: %w", err))
	}

	if err := json.Unmarshal(data, &obj); err != nil {
		return obj, motmedelErrors.New(fmt.Errorf("json unmarshal: %w", err), data)
	}

	return obj, err
}

func ObjectToMap(object any) (map[string]any, error) {
	data, err := json.Marshal(object)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("json marshal: %w", err), object)
	}

	var objectMap map[string]any
	if err = json.Unmarshal(data, &objectMap); err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("json unmarshal: %w", err), data)
	}

	return objectMap, nil
}

func ObjectToBytes(object any) ([]byte, error) {
	objectMap, err := ObjectToMap(object)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("object to map: %w", err), object)
	}

	data, err := json.Marshal(objectMap)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("json marshal: %w", err), objectMap)
	}

	return data, nil
}
