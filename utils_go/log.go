package utils_go

import (
	"log/slog"
	"os"
	"reflect"
)

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

func MakeErrorGroup(err error) *slog.Attr {
	args := []any{
		slog.String("message", err.Error()),
		slog.String("type", reflect.TypeOf(err).String()),
	}

	if inputError, ok := err.(InputErrorI); ok {
		textualRepresentation, err := MakeTextualRepresentation(inputError.GetInput())
		if err != nil {
			go func() {
				LogError(
					"An error occurred when making a textual representation of error input.",
					err,
					LOG,
				)
			}()
		} else {
			args = append(args, slog.String("input", textualRepresentation))
		}
	}

	if causeError, ok := err.(CauseErrorI); ok {
		args = append(args, slog.Group("cause", *MakeErrorGroup(causeError.GetCause())))
	}

	group := slog.Group(
		"error",
		args...,
	)
	return &group
}

func LogError(message string, err error, logger *slog.Logger) {
	logger.Error(message, *MakeErrorGroup(err))
}

func LogWarning(message string, err error, logger *slog.Logger) {
	logger.Warn(message, *MakeErrorGroup(err))
}

func LogFatal(message string, err error, logger *slog.Logger, exitCode int) {
	logger.Error(message, *MakeErrorGroup(err))
	os.Exit(exitCode)
}
