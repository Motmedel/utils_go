package parsed_claims

import (
	"fmt"
	"maps"

	"github.com/Motmedel/utils_go/pkg/errors"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	jwtErrors "github.com/Motmedel/utils_go/pkg/json/jose/jwt/errors"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwt/types/claims_strings"
	"github.com/Motmedel/utils_go/pkg/json/jose/jwt/types/numeric_date"
)

type Claims map[string]any

func New(claimsMap map[string]any) (Claims, error) {
	if claimsMap == nil {
		return nil, nil
	}

	clone := maps.Clone(claimsMap)
	if clone == nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("%w (claims map clone)", motmedelErrors.ErrNilMap))
	}

	for key, value := range claimsMap {
		switch key {
		case "exp", "nbf", "iat":
			numericDate, err := numeric_date.Convert(value)
			if err != nil {
				return nil, errors.New(fmt.Errorf("parse numeric date (%s): %w", key, err), value)
			}
			if numericDate == nil {
				return nil, motmedelErrors.NewWithTrace(jwtErrors.ErrNilNumericDate, value)
			}
			clone[key] = *numericDate
		case "aud":
			claimsString, err := claims_strings.Convert(value)
			if err != nil {
				return nil, errors.New(fmt.Errorf("parse claim string (%s): %w", key, err), value)
			}
			clone[key] = claimsString
		}
	}

	return clone, nil
}
