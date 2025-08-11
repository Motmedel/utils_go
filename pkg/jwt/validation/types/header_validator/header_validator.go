package header_validator

import (
	"errors"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/interfaces/comparer"
	motmedelJwtErrors "github.com/Motmedel/utils_go/pkg/jwt/errors"
	"github.com/Motmedel/utils_go/pkg/jwt/validation/types/setting"
	"github.com/Motmedel/utils_go/pkg/utils"
)

type ExpectedFields struct {
	Alg comparer.Comparer[string]
	Typ comparer.Comparer[string]
}

type HeaderValidator struct {
	Settings map[string]setting.Setting
	Expected *ExpectedFields
}

func (validator *HeaderValidator) Validate(fields map[string]any) error {
	if fields == nil {
		return fmt.Errorf("%w: %w", motmedelErrors.ErrValidationError, motmedelJwtErrors.ErrNilTokenHeader)
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
			alg, err := utils.Convert[string](value)
			if err != nil {
				return motmedelErrors.New(fmt.Errorf("convert (%s): %w", key, err), value)
			}

			if algComparer := expected.Alg; !utils.IsNil(algComparer) {
				ok, err := algComparer.Compare(alg)
				if err != nil {
					return motmedelErrors.New(fmt.Errorf("compare (%s): %w", key, err), alg)
				}
				if !ok {
					errs = append(
						errs,
						motmedelErrors.New(motmedelJwtErrors.ErrAlgMismatch, algComparer, alg),
					)
				}
			}
		case "typ":
			typ, err := utils.Convert[string](value)
			if err != nil {
				return motmedelErrors.New(fmt.Errorf("convert (%s): %w", key, err), value)
			}

			if typComparer := expected.Typ; !utils.IsNil(typComparer) {
				ok, err := typComparer.Compare(typ)
				if err != nil {
					return motmedelErrors.New(fmt.Errorf("compare (%s): %w", key, err), typ)
				}
				if !ok {
					errs = append(
						errs,
						motmedelErrors.New(motmedelJwtErrors.ErrTypMismatch, typComparer, typ),
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
