package parsing

import (
	"encoding/base64"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"strings"
)

const (
	TokenDelimiter = "."
)

func SplitToken(token string) ([3]string, error) {
	var parts [3]string

	splitParts := strings.SplitN(token, TokenDelimiter, 3)
	if len(splitParts) != 3 {
		return parts, motmedelErrors.NewWithTrace(motmedelErrors.ErrBadSplit)
	}

	parts[0] = splitParts[0]
	parts[1] = splitParts[1]
	parts[2] = splitParts[2]

	return parts, nil
}

func Parse(token string) ([]byte, []byte, []byte, error) {
	parts, err := SplitToken(token)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("split token: %w", err)
	}

	var decodedParts [3][]byte

	for i := 0; i < len(parts); i++ {
		decodedParts[i], err = base64.RawURLEncoding.DecodeString(parts[i])
		if err != nil {
			var partName string
			switch i {
			case 0:
				partName = " (header part)"
			case 1:
				partName = " (payload part)"
			case 2:
				partName = " (signature part)"

			}
			return nil, nil, nil, motmedelErrors.NewWithTrace(
				fmt.Errorf("base64 raw url encoding decode string%s: %w", partName, err),
			)
		}
	}

	return decodedParts[0], decodedParts[1], decodedParts[2], nil
}
