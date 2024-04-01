package utils_go

import (
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
)

func MakeInternalServerErrorProblemDetail() *ProblemDetail {
	return &ProblemDetail{
		Type:     "about:blank",
		Title:    "Internal Server Error",
		Status:   http.StatusInternalServerError,
		Instance: uuid.New().String(),
	}
}

type ProblemDetail struct {
	Type      string `json:"type,omitempty"`
	Title     string `json:"title,omitempty"`
	Status    int    `json:"status,omitempty"`
	Detail    string `json:"detail,omitempty"`
	Instance  string `json:"instance,omitempty"`
	Errors    []any  `json:"errors,omitempty"`
	Extension any    `json:"extension,omitempty"`
}

func (problemDetail *ProblemDetail) String() (string, error) {

	var outputMap map[string]interface{}

	problemDetailData, err := json.Marshal(problemDetail)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(problemDetailData, &outputMap)
	if err != nil {
		return "", err
	}

	if extension, ok := outputMap["extension"]; ok {
		delete(outputMap, "extension")

		var extensionMap map[string]interface{}
		extensionData, err := json.Marshal(extension)
		if err != nil {
			return "", err
		}

		err = json.Unmarshal(extensionData, &extensionMap)
		if err != nil {
			return "", err
		}

		for key, value := range extensionMap {
			outputMap[key] = value
		}
	}

	if outputData, err := json.Marshal(outputMap); err != nil {
		return "", err
	} else {
		return string(outputData), nil
	}
}
