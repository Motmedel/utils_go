package raw_token

import (
	"fmt"
	"strings"

	motmedelCryptoErrors "github.com/Motmedel/utils_go/pkg/crypto/errors"
	motmedelCryptoInterfaces "github.com/Motmedel/utils_go/pkg/crypto/interfaces"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwt"
	"github.com/Motmedel/utils_go/pkg/utils"
)

type Token struct {
	Header    []byte
	Payload   []byte
	Signature []byte
	Raw       string
}

func (rawToken *Token) Verify(verifier motmedelCryptoInterfaces.Verifier) error {
	if utils.IsNil(verifier) {
		return motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrNilVerifier)
	}

	rawSplit := strings.Split(rawToken.Raw, ".")
	if len(rawSplit) != 3 {
		return motmedelErrors.NewWithTrace(motmedelErrors.ErrBadSplit, rawToken.Raw)
	}

	header := rawSplit[0]
	payload := rawSplit[1]
	if err := jwt.Verify(header, payload, rawToken.Signature, verifier); err != nil {
		return motmedelErrors.New(fmt.Errorf("verifier verify: %w", err), header, payload, rawToken.Signature)
	}

	return nil
}

func New(token string) (*Token, error) {
	header, payload, signature, err := jwt.Parse(token)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	return &Token{Header: header, Payload: payload, Signature: signature, Raw: token}, nil
}
