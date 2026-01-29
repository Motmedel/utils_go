package jwt

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	motmedelCryptoErrors "github.com/Motmedel/utils_go/pkg/crypto/errors"
	"github.com/Motmedel/utils_go/pkg/crypto/interfaces"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/errors/types/empty_error"
	motmedelJwtErrors "github.com/Motmedel/utils_go/pkg/json/jose/jwt/errors"
	"github.com/Motmedel/utils_go/pkg/utils"
)

func Verify(header string, payload string, signature []byte, verifier interfaces.Verifier) error {
	if utils.IsNil(verifier) {
		return motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrNilVerifier)
	}

	err := verifier.Verify([]byte(strings.Join([]string{header, payload}, ".")), signature)
	if err != nil {
		return fmt.Errorf("%w: verifier verify: %w", motmedelErrors.ErrVerificationError, err)
	}

	return nil
}

func VerifyTokenString(tokenString string, verifier interfaces.Verifier) error {
	if utils.IsNil(verifier) {
		return motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrNilVerifier)
	}

	if tokenString == "" {
		return motmedelErrors.NewWithTrace(
			fmt.Errorf("%w: %w", motmedelErrors.ErrParseError, empty_error.New("token")),
		)
	}

	rawSplit := strings.Split(tokenString, ".")
	if len(rawSplit) != 3 {
		return motmedelErrors.NewWithTrace(
			fmt.Errorf("%w: %w", motmedelErrors.ErrParseError, motmedelErrors.ErrBadSplit),
		)
	}

	header := rawSplit[0]
	payload := rawSplit[1]

	signature, err := base64.RawURLEncoding.DecodeString(rawSplit[2])
	if err != nil {
		return motmedelErrors.NewWithTrace(
			fmt.Errorf("%w: %w", motmedelErrors.ErrParseError, motmedelErrors.ErrBadSplit),
		)
	}

	if err := Verify(header, payload, signature, verifier); err != nil {
		return motmedelErrors.New(fmt.Errorf("verifier verify: %w", err), header, payload, signature)
	}

	return nil
}

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
		wrappedErr := fmt.Errorf("split token: %w", err)
		if errors.Is(err, motmedelErrors.ErrBadSplit) {
			return nil, nil, nil, fmt.Errorf("%w: %w", motmedelErrors.ErrParseError, wrappedErr)
		}

		return nil, nil, nil, wrappedErr
	}

	var decodedParts [3][]byte

	for i := range parts {
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
			return nil, nil, nil,
				motmedelErrors.NewWithTrace(
					fmt.Errorf(
						"%w: base64 raw url encoding decode string%s: %w",
						motmedelErrors.ErrParseError, partName, err,
					),
				)
		}
	}

	return decodedParts[0], decodedParts[1], decodedParts[2], nil
}

func ValidateExpiresAt(expiresAt time.Time, cmp time.Time) error {
	if cmp.After(expiresAt) {
		return motmedelJwtErrors.ErrExpExpired
	}
	return nil
}

func ValidateNotBefore(notBefore time.Time, cmp time.Time) error {
	if cmp.Before(notBefore) {
		return motmedelJwtErrors.ErrNbfBefore
	}
	return nil
}

func ValidateIssuedAt(issuedAt time.Time, cmp time.Time) error {
	if cmp.Before(issuedAt) {
		return motmedelJwtErrors.ErrIatBefore
	}
	return nil
}
