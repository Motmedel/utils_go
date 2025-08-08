package raw_token

import (
	"fmt"
	motmedelCryptoErrors "github.com/Motmedel/utils_go/pkg/crypto/errors"
	motmedelCryptoInterfaces "github.com/Motmedel/utils_go/pkg/crypto/interfaces"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/jwt/parsing"
	"github.com/Motmedel/utils_go/pkg/utils"
	"strings"
)

type RawToken struct {
	Header    []byte
	Payload   []byte
	Signature []byte
	Raw       string
}

func (rawToken *RawToken) Verify(verifier motmedelCryptoInterfaces.Verifier) error {
	if utils.IsNil(verifier) {
		return motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrNilVerifier)
	}

	rawSplit := strings.Split(rawToken.Raw, ".")
	if len(rawSplit) != 3 {
		return motmedelErrors.NewWithTrace(motmedelErrors.ErrBadSplit, rawToken.Raw)
	}

	err := verifier.Verify(
		[]byte(strings.Join(rawSplit[:2], ".")),
		rawToken.Signature,
	)
	if err != nil {
		return fmt.Errorf("verifier verify: %w", err)
	}

	return nil
}

func New(token string) (*RawToken, error) {
	header, payload, signature, err := parsing.Parse(token)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	return &RawToken{Header: header, Payload: payload, Signature: signature, Raw: token}, nil
}
