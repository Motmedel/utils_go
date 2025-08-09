package parsed_claims

import (
	"fmt"
	"github.com/Motmedel/utils_go/pkg/errors"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	jwtErrors "github.com/Motmedel/utils_go/pkg/jwt/errors"
	"github.com/Motmedel/utils_go/pkg/jwt/parsing/types/claims_strings"
	"github.com/Motmedel/utils_go/pkg/jwt/parsing/types/numeric_date"
	"maps"
)

type ParsedClaims map[string]any

func FromMap(claimsMap map[string]any) (ParsedClaims, error) {
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
