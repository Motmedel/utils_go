package utils_go

import (
	"log/slog"
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
		args = append(args, slog.String("input", string(inputError.GetInput())))
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
