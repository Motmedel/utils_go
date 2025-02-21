package response_error

import (
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	muxTypesResponse "github.com/Motmedel/utils_go/pkg/http/mux/types/response"
	"github.com/Motmedel/utils_go/pkg/http/problem_detail"
	"strings"
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

func (responseError *ResponseError) MakeResponse() (*muxTypesResponse.Response, error) {
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

	responseBody, err := problemDetail.Bytes()
	if err != nil {
		return nil, motmedelErrors.MakeError(fmt.Errorf("problem detail bytes: %w", err), problemDetail)
	}

	responseHeaders := responseError.Headers

	if responseHeaders == nil {
		responseHeaders = []*muxTypesResponse.HeaderEntry{{Name: "Content-Type", Value: "application/problem+json"}}
	} else {
		for i, header := range responseHeaders {
			if strings.ToLower(header.Name) == "content-type" {
				responseHeaders[i] = nil
			}
		}
		responseHeaders = append(
			responseHeaders,
			&muxTypesResponse.HeaderEntry{Name: "Content-Type", Value: "application/problem+json"},
		)
	}

	return &muxTypesResponse.Response{StatusCode: statusCode, Body: responseBody, Headers: responseHeaders}, nil
}
