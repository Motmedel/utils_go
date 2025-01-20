package utils

import (
	"context"
	motmedelHttpParsingAccept "github.com/Motmedel/utils_go/pkg/http/parsing/headers/accept"
	"net/http"
)

type acceptMediaRangesContextType struct{}

var AcceptMediaRangesContextKey acceptMediaRangesContextType

func AcceptMiddleware(request *http.Request) *http.Request {
	if request == nil {
		return nil
	}

	header := request.Header
	if header == nil {
		return request
	}

	acceptValue := header.Get("Accept")
	if acceptValue == "" {
		return request
	}

	accept, err := motmedelHttpParsingAccept.ParseAccept([]byte(acceptValue))
	if accept == nil || err != nil {
		return request
	}

	return request.WithContext(
		context.WithValue(
			request.Context(),
			AcceptMediaRangesContextKey,
			accept.GetPriorityOrderedEncodings(),
		),
	)
}
