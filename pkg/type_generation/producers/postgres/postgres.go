package postgres

import (
	"fmt"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/type_generation/producers/postgres/types"
	typeGenerationTypesContext "github.com/Motmedel/utils_go/pkg/type_generation/types/context"
)

func Convert(values ...any) (string, error) {
	postgresContext := types.Context{Context: typeGenerationTypesContext.New()}
	if err := postgresContext.Add(values...); err != nil {
		return "", fmt.Errorf("add: %w", err)
	}

	output, err := postgresContext.Render()
	if err != nil {
		return "", motmedelErrors.New(fmt.Errorf("render: %w", err), postgresContext)
	}

	return output, nil
}
