package problem_detail

import (
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
)

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
