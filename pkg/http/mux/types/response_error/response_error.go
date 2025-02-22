package response_error

import (
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	muxTypesResponse "github.com/Motmedel/utils_go/pkg/http/mux/types/response"
	"github.com/Motmedel/utils_go/pkg/http/parsing/headers/content_type"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
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

type ResponseError struct {
	ProblemDetail *problem_detail.ProblemDetail
	Headers       []*muxTypesResponse.HeaderEntry
	ClientError   error
	ServerError   error
	BodyMaker     func(*problem_detail.ProblemDetail, *motmedelHttpTypes.ContentNegotiation) ([]byte, string, error)
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
		return nil, motmedelErrors.MakeErrorWithStackTrace(muxErrors.ErrMultipleResponseErrorErrors)
	}

	if responseError.ServerError != nil {
		return problem_detail.MakeInternalServerErrorProblemDetail("", nil), nil
	}

	if responseError.ClientError != nil {
		return problem_detail.MakeBadRequestProblemDetail("", nil), nil
	}

	return nil, motmedelErrors.MakeErrorWithStackTrace(
		fmt.Errorf(
			"%w: %w, %w",
			muxErrors.ErrUnusableResponseError,
			muxErrors.ErrNilProblemDetail,
			muxErrors.ErrEmptyResponseErrorErrors,
		),
	)
}

func (responseError *ResponseError) MakeResponse(
	contentNegotiation *motmedelHttpTypes.ContentNegotiation,
) (*muxTypesResponse.Response, error) {
	problemDetail := responseError.ProblemDetail
	if problemDetail == nil {
		return nil, motmedelErrors.MakeErrorWithStackTrace(
			fmt.Errorf("%w: %w", muxErrors.ErrUnusableResponseError, muxErrors.ErrNilProblemDetail),
		)
	}

	statusCode := problemDetail.Status
	if statusCode == 0 {
		return nil, motmedelErrors.MakeErrorWithStackTrace(
			fmt.Errorf("%w: problem detail: %w", muxErrors.ErrUnusableResponseError, muxErrors.ErrEmptyStatus),
		)
	}

	headers := responseError.Headers
	if len(headers) != 0 {
		for i, header := range headers {
			if header == nil || header.Name == "" {
				continue
			}

			if http.CanonicalHeaderKey(header.Name) == "Content-Type" {
				headers[i] = nil
			}
		}
	}

	var body []byte
	supportsResponseBody := true

	if contentNegotiation != nil && contentNegotiation.AcceptEncoding != nil {
		supportsResponseBody = motmedelHttpUtils.GetMatchingContentEncoding(
			contentNegotiation.AcceptEncoding.GetPriorityOrderedEncodings(),
			[]string{motmedelHttpUtils.AcceptContentIdentity},
		) == motmedelHttpUtils.AcceptContentIdentity
	}

	if supportsResponseBody {
		var contentType string
		var err error

		if bodyMaker := responseError.BodyMaker; bodyMaker != nil {
			body, contentType, err = bodyMaker(problemDetail, contentNegotiation)
			if err != nil {
				return nil, motmedelErrors.MakeError(
					fmt.Errorf("body maker: %w", err),
					problemDetail, contentNegotiation,
				)
			}

			contentTypeData := []byte(contentType)
			if _, err := content_type.ParseContentType(contentTypeData); err != nil {
				return nil, motmedelErrors.MakeError(
					fmt.Errorf("parse content type (body maker): %w", err),
					contentTypeData,
				)
			}
		} else {
			body, err = problemDetail.Bytes()
			if err != nil {
				return nil, motmedelErrors.MakeError(fmt.Errorf("problem detail bytes: %w", err), problemDetail)
			}
			contentType = "application/problem+json"
		}

		if len(body) != 0 {
			if contentType == "" {
				return nil, motmedelErrors.MakeErrorWithStackTrace(muxErrors.ErrEmptyResponseErrorContentType)
			}

			headers = append(
				headers,
				&muxTypesResponse.HeaderEntry{Name: "Content-Type", Value: contentType},
			)
		}
	}

	return &muxTypesResponse.Response{StatusCode: statusCode, Body: body, Headers: headers}, nil
}
