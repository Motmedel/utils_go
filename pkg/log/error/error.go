package error

import (
	"context"
	"fmt"
	motmedelContext "github.com/Motmedel/utils_go/pkg/context"
	"log/slog"
	"os"
)

func LogError(message string, err error, logger *slog.Logger, args ...any) {
	if logger == nil {
		logger = slog.Default()
	}
	logger.ErrorContext(
		motmedelContext.WithErrorContextValue(context.Background(), err),
		message,
		args...,
	)
}

func LogWarning(message string, err error, logger *slog.Logger, args ...any) {
	if logger == nil {
		logger = slog.Default()
	}
	logger.WarnContext(
		motmedelContext.WithErrorContextValue(context.Background(), err),
		message,
		args...,
	)
}

func LogDebug(message string, err error, logger *slog.Logger, args ...any) {
	if logger == nil {
		logger = slog.Default()
	}
	logger.DebugContext(
		motmedelContext.WithErrorContextValue(context.Background(), err),
		message,
		args...,
	)
}

func LogFatalWithExitCode(message string, err error, logger *slog.Logger, exitCode int, args ...any) {
	if logger == nil {
		logger = slog.Default()
	}
	logger.ErrorContext(
		motmedelContext.WithErrorContextValue(context.Background(), err),
		message,
		args...,
	)
	os.Exit(exitCode)
}

func LogFatalWithExitingMessage(message string, err error, logger *slog.Logger, args ...any) {
	LogFatalWithExitCode(fmt.Sprintf("%s Exiting.", message), err, logger, 1, args...)
}

func LogFatal(message string, err error, logger *slog.Logger, args ...any) {
	LogFatalWithExitCode(message, err, logger, 1, args...)
}
