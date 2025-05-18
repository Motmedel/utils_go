package env

import (
	"context"
	"fmt"
	motmedelContext "github.com/Motmedel/utils_go/pkg/context"
	motmedelEnvErrors "github.com/Motmedel/utils_go/pkg/env/errors"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"log/slog"
	"os"
)

func GetEnvWithDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func ReadEnvFatal(ctx context.Context, name string) string {
	value, found := os.LookupEnv(name)

	var err error
	if !found {
		err = motmedelErrors.NewWithTrace(fmt.Errorf("%w: %q", motmedelEnvErrors.ErrNotPresent, name), name)
	} else if value == "" {
		err = motmedelErrors.NewWithTrace(fmt.Errorf("%w: %q", motmedelEnvErrors.ErrEmpty, name), name)
	}

	if err != nil {
		slog.ErrorContext(
			motmedelContext.WithErrorContextValue(ctx, err),
			"An environment variable could not be read.",
		)
		os.Exit(1)
	}

	return value
}
