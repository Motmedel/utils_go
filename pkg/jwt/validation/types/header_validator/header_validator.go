package header_validator

import (
	"errors"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelJwtErrors "github.com/Motmedel/utils_go/pkg/jwt/errors"
	setting "github.com/Motmedel/utils_go/pkg/jwt/validation/types/setting"
	"github.com/Motmedel/utils_go/pkg/utils"
)

type ExpectedFields struct {
	Alg string
	Typ string
}

type HeaderValidator struct {
	Settings map[string]setting.Setting
	Expected *ExpectedFields
}

func (validator *HeaderValidator) Validate(fields map[string]any) error {
	if fields == nil {
		return nil
	}

	expected := validator.Expected
	if expected == nil {
		expected = &ExpectedFields{}
	}

	var errs []error

	for key, value := range validator.Settings {
		if _, ok := fields[key]; value == setting.SettingRequired && !ok {
			errs = append(
				errs,
				&motmedelJwtErrors.MissingRequiredFieldError{Name: key},
			)
		}
	}

	for key, value := range fields {
		if fieldSetting := validator.Settings[key]; fieldSetting == setting.SettingSkip {
			continue
		}

		switch key {
		case "alg":
			if expectedAlg := expected.Alg; expectedAlg != "" {
				alg, err := utils.Convert[string](value)
				if err != nil {
					return motmedelErrors.New(fmt.Errorf("convert (%s): %w", key, err), value)
				}

				if alg != expectedAlg {
					errs = append(
						errs,
						motmedelErrors.New(motmedelJwtErrors.ErrAlgMismatch, expectedAlg, alg),
					)
				}
			}
		case "typ":
			if expectedTyp := expected.Typ; expectedTyp != "" {
				typ, err := utils.Convert[string](value)
				if err != nil {
					return motmedelErrors.New(fmt.Errorf("convert (%s): %w", key, err), value)
				}

				if typ != expectedTyp {
					errs = append(
						errs,
						motmedelErrors.New(motmedelJwtErrors.ErrTypMismatch, expectedTyp, typ),
					)
				}
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, errors.Join(errs...))
	}

	return nil
}
