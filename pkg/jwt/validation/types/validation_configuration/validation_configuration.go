package validation_configuration

import (
	"github.com/Motmedel/utils_go/pkg/interfaces/validator"
	"github.com/Motmedel/utils_go/pkg/jwt/types/parsed_claims"
)

type ValidationConfiguration struct {
	HeaderValidator  validator.Validator[map[string]any]
	PayloadValidator validator.Validator[parsed_claims.ParsedClaims]
}
