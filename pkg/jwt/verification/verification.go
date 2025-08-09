package verification

import (
	"encoding/base64"
	"fmt"
	motmedelCryptoErrors "github.com/Motmedel/utils_go/pkg/crypto/errors"
	motmedelCryptoInterfaces "github.com/Motmedel/utils_go/pkg/crypto/interfaces"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	jwtErrors "github.com/Motmedel/utils_go/pkg/jwt/errors"
	"github.com/Motmedel/utils_go/pkg/utils"
	"strings"
)

func Verify(header string, payload string, signature []byte, verifier motmedelCryptoInterfaces.Verifier) error {
	if utils.IsNil(verifier) {
		return motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrNilVerifier)
	}

	err := verifier.Verify([]byte(strings.Join([]string{header, payload}, ".")), signature)
	if err != nil {
		return fmt.Errorf("%w: verifier verify: %w", motmedelErrors.ErrVerificationError, err)
	}

	return nil
}

func VerifyTokenString(tokenString string, verifier motmedelCryptoInterfaces.Verifier) error {
	if utils.IsNil(verifier) {
		return motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrNilVerifier)
	}

	if len(tokenString) == 0 {
		return motmedelErrors.NewWithTrace(
			fmt.Errorf("%w: %w", motmedelErrors.ErrParseError, jwtErrors.ErrEmptyTokenString),
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
