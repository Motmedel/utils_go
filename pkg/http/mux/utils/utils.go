package utils

import (
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	motmedelHttpTypes "github.com/Motmedel/utils_go/pkg/http/types"
	"strings"
	"time"
)

const AcceptContentIdentityIdentifier = "identity"

func parseLastModifiedTimestamp(timestamp string) (time.Time, error) {
	if t, err := time.Parse(time.RFC1123, timestamp); err != nil {
		return time.Time{}, err
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
		return false, &muxErrors.BadIfModifiedSinceTimestamp{
			InputError: motmedelErrors.InputError{
				Message: "An error occurred when parsing a If-Modified-Since timestamp.",
				Cause:   err,
				Input:   ifModifiedSinceValue,
			},
		}
	}

	lastModifiedTimestamp, err := parseLastModifiedTimestamp(lastModifiedValue)
	if err != nil {
		return false, &motmedelErrors.InputError{
			Message: "An error occurred when parsing a Last-Modified timestamp.",
			Cause:   err,
			Input:   lastModifiedValue,
		}
	}

	return ifModifiedSinceTimestamp.Equal(lastModifiedTimestamp) || lastModifiedTimestamp.Before(ifModifiedSinceTimestamp), nil
}

func GetMatchingContentEncoding(
	acceptableEncodings []*motmedelHttpTypes.Encoding,
	supportedEncodings []string,
) string {
	if len(acceptableEncodings) == 0 {
		return AcceptContentIdentityIdentifier
	}

	disallowIdentity := false

	for _, acceptableEncoding := range acceptableEncodings {
		coding := strings.ToLower(acceptableEncoding.Coding)
		qualityValue := acceptableEncoding.QualityValue

		if coding == "*" {
			if qualityValue == 0 {
				disallowIdentity = true
			} else {
				if len(supportedEncodings) != 0 {
					return supportedEncodings[0]
				} else {
					if !disallowIdentity {
						return AcceptContentIdentityIdentifier
					}
				}
			}
		}

		if coding == AcceptContentIdentityIdentifier {
			if qualityValue == 0 {
				disallowIdentity = true
			} else {
				return AcceptContentIdentityIdentifier
			}
		}

		if qualityValue == 0 {
			continue
		}

		for _, supportedEncoding := range supportedEncodings {
			if acceptableEncoding.Coding == supportedEncoding {
				return supportedEncoding
			}
		}
	}

	if !disallowIdentity {
		return AcceptContentIdentityIdentifier
	} else {
		return ""
	}
}
