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

func getFullType(typeValue string, subtypeValue string, normalize bool) string {
	if typeValue == "" {
		typeValue = "*"
	}
	if subtypeValue == "*" {
		subtypeValue = "*"
	}

	fullType := fmt.Sprintf("%s/%s", typeValue, subtypeValue)
	if normalize {
		return strings.ToLower(fullType)
	}

	return fullType
}

func getParameterMap(parameters [][2]string, normalize bool) map[string]string {
	if len(parameters) == 0 {
		return nil
	}

	parameterMap := make(map[string]string)

	for _, parameter := range parameters {
		key := parameter[0]
		if normalize {
			key = strings.ToLower(key)
		}
		value := parameter[1]

		if _, ok := parameterMap[key]; !ok {
			parameterMap[key] = value
		}
	}

	return parameterMap
}

type MediaRange struct {
	Type       string
	Subtype    string
	Parameters [][2]string
	Weight     float32
}

func (mediaRange *MediaRange) GetFullType(normalize bool) string {
	return getFullType(mediaRange.Type, mediaRange.Subtype, normalize)
}

func (mediaRange *MediaRange) GetParameterMap(normalize bool) map[string]string {
	if len(mediaRange.Parameters) == 0 {
		return nil
	}

	return getParameterMap(mediaRange.Parameters, normalize)
}

type Accept struct {
	MediaRanges []*MediaRange
	Raw         string
}

func (accept *Accept) GetPriorityOrderedEncodings() []*MediaRange {
	mediaRanges := make([]*MediaRange, len(accept.MediaRanges))
	copy(mediaRanges, accept.MediaRanges)

	sort.SliceStable(mediaRanges, func(i, j int) bool {
		return mediaRanges[i].Weight > mediaRanges[j].Weight
	})

	return mediaRanges
}

// TODO: I should parse those `a/b+c` types somehow too. Is that some official format?

type MediaType struct {
	Type       string
	Subtype    string
	Parameters [][2]string
}

func (mediaType *MediaType) GetFullType(normalize bool) string {
	return getFullType(mediaType.Type, mediaType.Subtype, normalize)
}

func (mediaType *MediaType) GetParametersMap(normalize bool) map[string]string {
	if len(mediaType.Parameters) == 0 {
		return nil
	}

	return getParameterMap(mediaType.Parameters, normalize)
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
