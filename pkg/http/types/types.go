package types

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
)

type HttpContext struct {
	Request      *http.Request
	RequestBody  []byte
	Response     *http.Response
	ResponseBody []byte
}

// TODO: I should parse those `a/b+c` types somehow too. Is that some official format?

type MediaType struct {
	Type       string
	Subtype    string
	Parameters [][2]string
}

func (mediaType *MediaType) GetFullType(normalize bool) string {
	typeValue := mediaType.Type
	if typeValue == "" {
		typeValue = "*"
	}
	subtypeValue := mediaType.Subtype
	if subtypeValue == "*" {
		subtypeValue = "*"
	}

	fullType := fmt.Sprintf("%s/%s", typeValue, subtypeValue)
	if normalize {
		return strings.ToLower(fullType)
	}
	return fullType
}

func (mediaType *MediaType) GetParametersMap(normalize bool) map[string]string {
	if len(mediaType.Parameters) == 0 {
		return nil
	}

	m := make(map[string]string)

	for _, parameter := range mediaType.Parameters {
		key := parameter[0]
		if normalize {
			key = strings.ToLower(key)
		}
		value := parameter[1]

		if _, ok := m[key]; !ok {
			m[key] = value
		}
	}

	return m
}

type ContentType struct {
	MediaType
}

type Encoding struct {
	Coding       string
	QualityValue float32
}

type AcceptEncoding struct {
	Encodings []*Encoding
	Raw       string
}

func (acceptEncoding *AcceptEncoding) GetPriorityOrderedEncodings() []*Encoding {
	encodings := make([]*Encoding, len(acceptEncoding.Encodings))
	copy(encodings, acceptEncoding.Encodings)

	sort.SliceStable(encodings, func(i, j int) bool {
		return encodings[i].QualityValue > encodings[j].QualityValue
	})

	return encodings
}

type StrictTransportSecurityPolicy struct {
	MaxAga            int
	IncludeSubdomains bool
	Raw               string
}
