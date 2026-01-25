package problem_detail

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/errors/types/nil_error"
	"github.com/Motmedel/utils_go/pkg/http/types/problem_detail/problem_detail_config"
)

type Detail struct {
	Type      string         `json:"type,omitempty"`
	Title     string         `json:"title,omitempty"`
	Status    int            `json:"status,omitempty"`
	Detail    string         `json:"detail,omitempty"`
	Instance  string         `json:"instance,omitempty"`
	Extension map[string]any `json:"extension,omitempty"`
}

// MarshalJSON flattens the Extension map into the top-level JSON object,
// instead of nesting it under the "extension" key.
func (d *Detail) MarshalJSON() ([]byte, error) {
	if d == nil {
		return []byte("null"), nil
	}

	m := make(map[string]any, 5)

	if d.Type != "" {
		m["type"] = d.Type
	}
	if d.Title != "" {
		m["title"] = d.Title
	}
	if d.Status != 0 {
		m["status"] = d.Status
	}
	if d.Detail != "" {
		m["detail"] = d.Detail
	}
	if d.Instance != "" {
		m["instance"] = d.Instance
	}

	if ext := d.Extension; ext != nil {
		for k, v := range ext {
			if k == "" {
				continue
			}
			m[k] = v
		}
	}

	b, err := json.Marshal(m)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("json marshal (detail): %w", err), m)
	}
	return b, nil
}

func (d *Detail) MarshalXML(encoder *xml.Encoder, start xml.StartElement) error {
	if encoder == nil {
		return motmedelErrors.NewWithTrace(nil_error.New("xml encoder"))
	}

	start.Name.Local = "problem"
	start.Name.Space = "urn:ietf:rfc:7807"

	if err := encoder.EncodeToken(start); err != nil {
		return motmedelErrors.NewWithTrace(fmt.Errorf("encode token (start): %w", err), start)
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

	if err := encode("type", d.Type); err != nil {
		return err
	}

	if err := encode("title", d.Title); err != nil {
		return err
	}

	if err := encode("status", d.Status); err != nil {
		return err
	}

	if err := encode("detail", d.Detail); err != nil {
		return err
	}

	if err := encode("instance", d.Instance); err != nil {
		return err
	}

	if extension := d.Extension; extension != nil {
		extensionBytes, err := json.Marshal(d.Extension)
		if err != nil {
			return motmedelErrors.NewWithTrace(
				fmt.Errorf("json marshal (extension): %w", err),
				extension,
			)
		}

		var extensionMap map[string]any
		if err := json.Unmarshal(extensionBytes, &extensionMap); err != nil {
			return motmedelErrors.NewWithTrace(
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
		return motmedelErrors.NewWithTrace(fmt.Errorf("encode token (end): %w", err), start)
	}

	return nil
}

func (d *Detail) String() (string, error) {
	var text string

	if status := d.Status; status != 0 {
		text = strconv.Itoa(status)
		if title := d.Title; title != "" {
			text += fmt.Sprintf(" %s", title)
		}
	} else if title := d.Type; title != "" {
		text = title
	}

	for _, s := range []string{d.Detail, d.Type, d.Instance} {
		if s == "" {
			continue
		}

		if text != "" {
			text += "\n\n"
		}
		text += s
	}

	var extensionText string
	for k, v := range d.Extension {
		if extensionText != "" {
			extensionText += "\n"
		}

		extensionText += fmt.Sprintf("%s:%v", k, v)
	}
	if extensionText != "" {
		if text != "" {
			text += "\n\n"
		}
		text += extensionText
	}

	return text, nil
}

func New(code int, options ...problem_detail_config.Option) *Detail {
	config := problem_detail_config.New(options...)
	return &Detail{
		Type:      config.Type,
		Title:     http.StatusText(code),
		Status:    code,
		Detail:    config.Detail,
		Instance:  config.Instance,
		Extension: config.Extension,
	}
}
