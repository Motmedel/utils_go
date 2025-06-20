package response_error

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	muxTypesResponse "github.com/Motmedel/utils_go/pkg/http/mux/types/response"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	problemDetailErrors "github.com/Motmedel/utils_go/pkg/http/problem_detail/errors"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	motmedelHttpUtils "github.com/Motmedel/utils_go/pkg/http/utils"
	"net/http"
)

type ResponseErrorType int

const (
	ResponseErrorType_Invalid ResponseErrorType = iota
	ResponseErrorType_ClientError
	ResponseErrorType_ServerError
)

type ProblemDetailConverter interface {
	Convert(*problem_detail.ProblemDetail, *motmedelHttpTypes.ContentNegotiation) ([]byte, string, error)
}

type ProblemDetailConverterFunction func(*problem_detail.ProblemDetail, *motmedelHttpTypes.ContentNegotiation) ([]byte, string, error)

func (f ProblemDetailConverterFunction) Convert(
	problemDetail *problem_detail.ProblemDetail,
	contentNegotiation *motmedelHttpTypes.ContentNegotiation,
) ([]byte, string, error) {
	return f(problemDetail, contentNegotiation)
}

var DefaultProblemDetailMediaRanges = []*motmedelHttpTypes.ServerMediaRange{
	{Type: "application", Subtype: "problem+json"},
	{Type: "application", Subtype: "json"},
	{Type: "application", Subtype: "problem+xml"},
	{Type: "application", Subtype: "xml"},
	{Type: "text", Subtype: "plain"},
}

func ConvertProblemDetail(
	detail *problem_detail.ProblemDetail,
	negotiation *motmedelHttpTypes.ContentNegotiation,
) ([]byte, string, error) {
	if detail == nil {
		return nil, "", nil
	}

	if negotiation != nil {
		if negotiation.NegotiatedAccept == "" && negotiation.Accept != nil {
			matchingServerMediaRange := motmedelHttpUtils.GetMatchingAccept(
				negotiation.Accept.GetPriorityOrderedEncodings(),
				DefaultProblemDetailMediaRanges,
			)
			if matchingServerMediaRange != nil {
				negotiation.NegotiatedAccept = matchingServerMediaRange.GetFullType(true)
			}
		}

		switch negotiatedAccept := negotiation.NegotiatedAccept; negotiatedAccept {
		case "application/problem+xml", "application/xml":
			data, err := xml.Marshal(detail)
			if err != nil {
				return nil, "", motmedelErrors.New(fmt.Errorf("xml marshal: %w", err), detail)
			}

			output := []byte(`<?xml version="1.0" encoding="UTF-8"?>`)
			output = append(output, data...)

			return output, "application/problem+xml", nil
		case "text/plain":
			text, err := detail.String()
			if err != nil {
				return nil, "", motmedelErrors.New(fmt.Errorf("problem detail string: %w", err), detail)
			}
			return []byte(text), negotiatedAccept, nil
		}
	}

	// Default to using JSON.
	data, err := json.Marshal(detail)
	if err != nil {
		return nil, "", motmedelErrors.New(fmt.Errorf("json marshal: %w", err), detail)
	}

	return data, "application/problem+json", nil
}

var DefaultProblemDetailConverter = ProblemDetailConverterFunction(ConvertProblemDetail)

type ResponseError struct {
	ProblemDetail          *problem_detail.ProblemDetail
	Headers                []*muxTypesResponse.HeaderEntry
	ClientError            error
	ServerError            error
	ProblemDetailConverter ProblemDetailConverter
}

func (responseError *ResponseError) Type() ResponseErrorType {
	if responseError.ServerError != nil {
		return ResponseErrorType_ServerError
	} else if responseError.ClientError != nil {
		return ResponseErrorType_ClientError
	} else if problemDetail := responseError.ProblemDetail; problemDetail != nil {
		statusCode := problemDetail.Status
		if statusCode >= 400 && statusCode < 500 {
			return ResponseErrorType_ClientError
		} else if statusCode >= 500 && statusCode < 600 {
			return ResponseErrorType_ServerError
		}
	}

	return ResponseErrorType_Invalid
}

func (responseError *ResponseError) GetEffectiveProblemDetail() (*problem_detail.ProblemDetail, error) {
	if problemDetail := responseError.ProblemDetail; problemDetail != nil {
		return problemDetail, nil
	}

	if responseError.ClientError != nil && responseError.ServerError != nil {
		return nil, motmedelErrors.NewWithTrace(muxErrors.ErrMultipleResponseErrorErrors)
	}

	if responseError.ServerError != nil {
		return problem_detail.MakeInternalServerErrorProblemDetail("", nil), nil
	}

	if responseError.ClientError != nil {
		return problem_detail.MakeBadRequestProblemDetail("", nil), nil
	}

	return nil, motmedelErrors.NewWithTrace(
		fmt.Errorf(
			"%w: %w, %w",
			muxErrors.ErrUnusableResponseError,
			problemDetailErrors.ErrNilProblemDetail,
			muxErrors.ErrEmptyResponseErrorErrors,
		),
	)
}

func (responseError *ResponseError) MakeResponse(
	negotiation *motmedelHttpTypes.ContentNegotiation,
) (*muxTypesResponse.Response, error) {
	problemDetail := responseError.ProblemDetail
	if problemDetail == nil {
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("%w: %w", muxErrors.ErrUnusableResponseError, problemDetailErrors.ErrNilProblemDetail),
		)
	}

	statusCode := problemDetail.Status
	if statusCode == 0 {
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("%w: problem detail: %w", muxErrors.ErrUnusableResponseError, muxErrors.ErrEmptyStatus),
		)
	}

	headers := responseError.Headers
	if len(headers) != 0 {
		for i, header := range headers {
			if header == nil || header.Name == "" {
				continue
			}

			// Clear any pre-existing Content-Type header.
			if http.CanonicalHeaderKey(header.Name) == "Content-Type" {
				headers[i] = nil
			}
		}
	}

	supportsResponseBody := true
	if negotiation != nil && negotiation.AcceptEncoding != nil {
		supportsResponseBody = motmedelHttpUtils.GetMatchingContentEncoding(
			negotiation.AcceptEncoding.GetPriorityOrderedEncodings(),
			[]string{motmedelHttpUtils.AcceptContentIdentity},
		) == motmedelHttpUtils.AcceptContentIdentity
	}

	var body []byte

	if supportsResponseBody {
		converter := responseError.ProblemDetailConverter
		if converter == nil {
			converter = DefaultProblemDetailConverter
		}

		var contentType string
		var err error
		body, contentType, err = converter.Convert(problemDetail, negotiation)
		if err != nil {
			return nil, motmedelErrors.New(
				fmt.Errorf("convert: %w", err),
				problemDetail, negotiation,
			)
		}

		if len(body) != 0 {
			if contentType == "" {
				return nil, motmedelErrors.NewWithTrace(muxErrors.ErrEmptyResponseErrorContentType)
			}

			headers = append(
				headers,
				&muxTypesResponse.HeaderEntry{Name: "Content-Type", Value: contentType},
			)
		}
	}

	return &muxTypesResponse.Response{StatusCode: statusCode, Body: body, Headers: headers}, nil
}
