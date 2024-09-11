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

func (problemDetail *ProblemDetail) Bytes() ([]byte, error) {
	outputMap, err := problemDetail.makeOutputMap()
	if err != nil {
		return nil, err
	}
	if outputData, err := json.Marshal(outputMap); err != nil {
		return nil, err
	} else {
		return outputData, nil
	}
}

func (problemDetail *ProblemDetail) String() (string, error) {
	data, err := problemDetail.Bytes()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func MakeStatusCodeProblemDetail(code int, detail string, extension any) *ProblemDetail {
	return &ProblemDetail{
		Type:      "about:blank",
		Title:     http.StatusText(code),
		Status:    code,
		Detail:    detail,
		Instance:  uuid.New().String(),
		Extension: extension,
	}
}

func MakeInternalServerErrorProblemDetail(detail string, extension any) *ProblemDetail {
	return MakeStatusCodeProblemDetail(http.StatusInternalServerError, detail, extension)
}

func MakeBadRequestProblemDetail(detail string, extension any) *ProblemDetail {
	return MakeStatusCodeProblemDetail(http.StatusBadRequest, detail, extension)
}
