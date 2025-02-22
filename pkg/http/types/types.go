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
	if subtypeValue == "" {
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

func getStructuredSyntaxName(subtype string, normalize bool) string {
	if subtype == "" {
		return ""
	}

	separator := "+"

	lastSeparatorIndex := strings.LastIndex(subtype, separator)
	if lastSeparatorIndex == -1 {
		return ""
	}

	structuredSyntaxName := subtype[lastSeparatorIndex+len(separator):]
	if normalize {
		structuredSyntaxName = strings.ToLower(structuredSyntaxName)
	}

	return structuredSyntaxName
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
	parameters := mediaRange.Parameters
	if len(parameters) == 0 {
		return nil
	}

	return getParameterMap(parameters, normalize)
}

func (mediaRange *MediaRange) GetStructuredSyntaxName(normalize bool) string {
	return getStructuredSyntaxName(mediaRange.Subtype, normalize)
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

type MediaType struct {
	Type       string
	Subtype    string
	Parameters [][2]string
}

func (mediaType *MediaType) GetFullType(normalize bool) string {
	return getFullType(mediaType.Type, mediaType.Subtype, normalize)
}

func (mediaType *MediaType) GetStructuredSyntaxName(normalize bool) string {
	return getStructuredSyntaxName(mediaType.Subtype, normalize)
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

type RetryAfter struct {
	// The time can be either a timestamp or a duration.
	WaitTime any
	Raw      string
}

type ContentDisposition struct {
	DispositionType     string
	Filename            string
	FilenameAsterisk    string
	ExtensionParameters map[string]string
}

type ContentNegotiation struct {
	Accept         *Accept
	AcceptEncoding *AcceptEncoding
	// TODO: Add more headers.
}
