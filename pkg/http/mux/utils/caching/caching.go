package caching

import (
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	"time"
)

func parseLastModifiedTimestamp(timestamp string) (time.Time, error) {
	if t, err := time.Parse(time.RFC1123, timestamp); err != nil {
		return time.Time{}, motmedelErrors.MakeErrorWithStackTrace(
			fmt.Errorf(
				"%w: time parse rfc1123: %w",
				muxErrors.ErrBadIfModifiedSinceTimestamp,
				err,
			),
			timestamp,
		)
	} else {
		return t, nil
	}
}

func IfNoneMatchCacheHit(ifNoneMatchValue string, etag string) bool {
	if ifNoneMatchValue == "" || etag == "" {
		return false
	}

	return ifNoneMatchValue == etag
}

func IfModifiedSinceCacheHit(ifModifiedSinceValue string, lastModifiedValue string) (bool, error) {
	if ifModifiedSinceValue == "" || lastModifiedValue == "" {
		return false, nil
	}

	ifModifiedSinceTimestamp, err := parseLastModifiedTimestamp(ifModifiedSinceValue)
	if err != nil {
		return false, motmedelErrors.MakeError(
			fmt.Errorf("parse last modified timestamp (If-Modified-Since): %w", err),
			ifModifiedSinceValue,
		)
	}

	lastModifiedTimestamp, err := parseLastModifiedTimestamp(lastModifiedValue)
	if err != nil {
		return false, motmedelErrors.MakeError(
			fmt.Errorf("parse last modified timestamp (Last-Modified): %w", err),
			lastModifiedValue,
		)
	}

	return ifModifiedSinceTimestamp.Equal(lastModifiedTimestamp) || lastModifiedTimestamp.Before(ifModifiedSinceTimestamp), nil
}
