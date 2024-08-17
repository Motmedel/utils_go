package utils_go

import (
	"context"
	"log/slog"
	"os"
	"reflect"
)

// NOTE: Apparently, according to convention, you should not use string keys, and you should not use context as a kind
// storage for optional parameters. But it seems to me a logger is a special case, and I see no better, idiomatic way.

const LoggerCtxKey = "logger"

func GetLoggerFromCtx(ctx context.Context) *slog.Logger {
	logger, _ := ctx.Value(LoggerCtxKey).(*slog.Logger)
	return logger
}

func GetLoggerFromCtxWithDefault(ctx context.Context, defaultLogger *slog.Logger) *slog.Logger {
	logger := GetLoggerFromCtx(ctx)
	if logger == nil {
		if defaultLogger != nil {
			logger = defaultLogger
		} else {
			logger = slog.Default()
		}
	}
	return logger
}

func CtxWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, LoggerCtxKey, logger)
}

func AttrsFromMap(m map[string]any) []any {
	var attrs []any

	for key, value := range m {
		switch typedValue := value.(type) {
		case string:
			attrs = append(attrs, slog.String(key, typedValue))
		case int:
			attrs = append(attrs, slog.Int(key, typedValue))
		case float64:
			attrs = append(attrs, slog.Float64(key, typedValue))
		case bool:
			attrs = append(attrs, slog.Bool(key, typedValue))
		case map[string]any:
			attrs = append(attrs, slog.Group(key, AttrsFromMap(typedValue)...))
		default:
			attrs = append(attrs, slog.Any(key, value))
		}
	}

	return attrs
}

func makeErrorAttrs(err error) []any {
	if err == nil {
		return nil
	}

	attrs := []any{
		slog.String("message", err.Error()),
		slog.String("type", reflect.TypeOf(err).String()),
	}

	if inputError, ok := err.(InputErrorI); ok {
		input := inputError.GetInput()
		inputTextualRepresentation, err := MakeTextualRepresentation(input)
		if err != nil {
			go func() {
				LogError(
					"An error occurred when making a textual representation of error input.",
					err,
					LOG,
				)
			}()
		} else {
			attrs = append(
				attrs,
				slog.Group(
					"input",
					slog.String("value", inputTextualRepresentation),
					slog.String("type", reflect.TypeOf(input).String()),
				),
			)
		}
	}

	if causeError, ok := err.(CauseErrorI); ok {
		attrs = append(attrs, slog.Group("cause", makeErrorAttrs(causeError.GetCause())...))
	}

	return attrs
}

func MakeErrorGroup(err error) *slog.Attr {
	group := slog.Group("error", makeErrorAttrs(err)...)
	return &group
}

func LogError(message string, err error, logger *slog.Logger) {
	if errorGroup := MakeErrorGroup(err); errorGroup != nil {
		logger.Error(message, *errorGroup)
	} else {
		logger.Error(message)
	}
}

func LogWarning(message string, err error, logger *slog.Logger) {
	if errorGroup := MakeErrorGroup(err); errorGroup != nil {
		logger.Warn(message, *errorGroup)
	} else {
		logger.Warn(message)
	}
}

func LogDebug(message string, err error, logger *slog.Logger) {
	if errorGroup := MakeErrorGroup(err); errorGroup != nil {
		logger.Debug(message, *errorGroup)
	} else {
		logger.Debug(message)
	}
}

func LogFatal(message string, err error, logger *slog.Logger, exitCode int) {
	if errorGroup := MakeErrorGroup(err); errorGroup != nil {
		logger.Error(message, *errorGroup)
	} else {
		logger.Error(message)
	}
	os.Exit(exitCode)
}
