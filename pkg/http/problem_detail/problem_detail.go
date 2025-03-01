package problem_detail

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	problemDetailErrors "github.com/Motmedel/utils_go/pkg/http/problem_detail/errors"
	"github.com/google/uuid"
	"net/http"
	"reflect"
	"strconv"
)

type ProblemDetail struct {
	Type      string `json:"type,omitempty"`
	Title     string `json:"title,omitempty"`
	Status    int    `json:"status,omitempty"`
	Detail    string `json:"detail,omitempty"`
	Instance  string `json:"instance,omitempty"`
	Extension any    `json:"extension,omitempty"`
}

func (problemDetail *ProblemDetail) ExtensionMap() (map[string]any, error) {
	extension := problemDetail.Extension
	if extension == nil {
		return nil, nil
	}

	extensionBytes, err := json.Marshal(extension)
	if err != nil {
		return nil, motmedelErrors.MakeErrorWithStackTrace(
			fmt.Errorf("json marshal (extension): %w", err),
			extension,
		)
	}

	var extensionMap map[string]any
	if err := json.Unmarshal(extensionBytes, &extensionMap); err != nil {
		return nil, motmedelErrors.MakeErrorWithStackTrace(
			fmt.Errorf("json unmarshal (extension map): %w", err),
			extensionMap,
		)
	}

	return extensionMap, nil
}

func (problemDetail *ProblemDetail) Map() (map[string]any, error) {
	base := map[string]any{
		"type":     problemDetail.Type,
		"title":    problemDetail.Title,
		"status":   problemDetail.Status,
		"detail":   problemDetail.Detail,
		"instance": problemDetail.Instance,
	}

	extensionMap, err := problemDetail.ExtensionMap()
	if err != nil {
		return nil, fmt.Errorf("extension map: %w", err)
	}

	for k, v := range extensionMap {
		base[k] = v
	}

	return base, nil
}

func (problemDetail *ProblemDetail) MarshalJSON() ([]byte, error) {
	base, err := problemDetail.Map()
	if err != nil {
		return nil, fmt.Errorf("map: %w", err)
	}

	data, err := json.Marshal(base)
	if err != nil {
		return nil, motmedelErrors.MakeErrorWithStackTrace(fmt.Errorf("json marshal: %w", err), base)
	}

	return data, nil
}

func (problemDetail *ProblemDetail) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	if encoder == nil {
		return motmedelErrors.NewWithTrace(problemDetailErrors.ErrNilEncoder)
	}

	start.Name.Local = "problem"
	start.Name.Space = "urn:ietf:rfc:7807"

	if err := encoder.EncodeToken(start); err != nil {
		return motmedelErrors.MakeErrorWithStackTrace(fmt.Errorf("encode token (start): %w", err), start)
	}

	encode := func(localName string, value any) error {
		if reflect.ValueOf(value).IsZero() {
			return nil
		}

		if err := encoder.EncodeElement(value, xml.StartElement{Name: xml.Name{Local: localName}}); err != nil {
			return motmedelErrors.NewWithTrace(fmt.Errorf("encode element: %w", err), localName, value)
		}

		return nil
	}

	problemDetailType := problemDetail.Type
	if problemDetailType == "" {
		problemDetailType = "about:blank"
	}

	if err := encode("type", problemDetailType); err != nil {
		return err
	}

	if err := encode("title", problemDetail.Title); err != nil {
		return err
	}

	if err := encode("status", problemDetail.Status); err != nil {
		return err
	}

	if err := encode("detail", problemDetail.Detail); err != nil {
		return err
	}

	if err := encode("instance", problemDetail.Instance); err != nil {
		return err
	}

	if extension := problemDetail.Extension; extension != nil {
		extensionBytes, err := json.Marshal(problemDetail.Extension)
		if err != nil {
			return motmedelErrors.MakeErrorWithStackTrace(
				fmt.Errorf("json marshal (extension): %w", err),
				extension,
			)
		}

		var extensionMap map[string]any
		if err := json.Unmarshal(extensionBytes, &extensionMap); err != nil {
			return motmedelErrors.MakeErrorWithStackTrace(
				fmt.Errorf("json unmarshal (extension map): %w", err),
				extensionMap,
			)
		}

		for k, v := range extensionMap {
			if err := encode(k, v); err != nil {
				return err
			}
		}
	}

	if err := encoder.EncodeToken(xml.EndElement{Name: start.Name}); err != nil {
		return motmedelErrors.MakeErrorWithStackTrace(fmt.Errorf("encode token (end): %w", err), start)
	}

	return nil
}

func (problemDetail *ProblemDetail) String() (string, error) {
	var text string

	if status := problemDetail.Status; status != 0 {
		text = strconv.Itoa(status)
		if title := problemDetail.Title; title != "" {
			text += fmt.Sprintf(" %s", title)
		}
	} else if title := problemDetail.Type; title != "" {
		text = title
	}

	problemDetailType := problemDetail.Type
	if problemDetailType == "about:blank" {
		problemDetailType = ""
	}

	for _, s := range []string{problemDetail.Detail, problemDetailType, problemDetail.Instance} {
		if s == "" {
			continue
		}

		if text != "" {
			text += "\n\n"
		}
		text += s
	}

	extensionMap, err := problemDetail.ExtensionMap()
	if err != nil {
		return "", fmt.Errorf("extension map: %w", err)
	}

	var extensionText string
	for k, v := range extensionMap {
		if extensionText != "" {
			extensionText += "\n"
		}

		text += fmt.Sprintf("%s:%v", k, v)
	}
	if extensionText != "" {
		if text != "" {
			text += "\n\n"
		}
		text += extensionText
	}

	return text, nil
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
