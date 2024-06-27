package utils_go

import (
	"encoding/json"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"reflect"
)

func MakeStatusCodeProblemDetail(code int) *ProblemDetail {
	return &ProblemDetail{
		Type:     "about:blank",
		Title:    http.StatusText(code),
		Status:   code,
		Instance: uuid.New().String(),
	}
}

func MakeInternalServerErrorProblemDetail() *ProblemDetail {
	return MakeStatusCodeProblemDetail(http.StatusInternalServerError)
}

func MakeBadRequestProblemDetail() *ProblemDetail {
	return MakeStatusCodeProblemDetail(http.StatusBadRequest)
}

type ProblemDetail struct {
	Type      string `json:"type,omitempty"`
	Title     string `json:"title,omitempty"`
	Status    int    `json:"status,omitempty"`
	Detail    string `json:"detail,omitempty"`
	Instance  string `json:"instance,omitempty"`
	Extension any    `json:"extension,omitempty"`
}

func (problemDetail *ProblemDetail) makeOutputMap() (map[string]any, error) {
	var outputMap map[string]any

	problemDetailData, err := json.Marshal(problemDetail)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(problemDetailData, &outputMap)
	if err != nil {
		return nil, err
	}

	if extension, ok := outputMap["extension"]; ok {
		delete(outputMap, "extension")

		var extensionMap map[string]any
		extensionData, err := json.Marshal(extension)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(extensionData, &extensionMap)
		if err != nil {
			return nil, err
		}

		for key, value := range extensionMap {
			outputMap[key] = value
		}
	}

	return outputMap, nil
}

func (problemDetail *ProblemDetail) String() (string, error) {
	outputMap, err := problemDetail.makeOutputMap()
	if err != nil {
		return "", err
	}

	if outputData, err := json.Marshal(outputMap); err != nil {
		return "", err
	} else {
		return string(outputData), nil
	}
}

func makeSlogAttrs(sourceMap map[string]any) []slog.Attr {
	slogAttrs := make([]slog.Attr, 0)

	for key, value := range sourceMap {
		switch typedValue := value.(type) {
		case int:
			slogAttrs = append(slogAttrs, slog.Int(key, typedValue))
		case string:
			slogAttrs = append(slogAttrs, slog.String(key, typedValue))
		default:
			if reflect.TypeOf(value).Kind() == reflect.Map {
				slogAttrs = append(slogAttrs, slog.Attr{Key: key, Value: slog.GroupValue(makeSlogAttrs(value.(map[string]any))...)})
			} else {
				slogAttrs = append(slogAttrs, slog.Any(key, value))
			}
		}
	}

	return slogAttrs
}

func (problemDetail *ProblemDetail) LogValue() slog.Value {
	outputMap, err := problemDetail.makeOutputMap()
	if err != nil {
		return slog.GroupValue()
	}
	return slog.GroupValue(makeSlogAttrs(outputMap)...)
}
