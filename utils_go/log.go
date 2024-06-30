package utils_go

import (
	"io"
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
	group := slog.Group(
		"error",
		slog.String("message", err.Error()),
		slog.String("type", reflect.TypeOf(err).String()),
	)
	return &group
}

func GCPReplaceAttr(groups []string, attr slog.Attr) slog.Attr {
	if len(groups) > 0 {
		return attr
	}

	switch attr.Key {
	case slog.TimeKey:
		attr.Key = "time"
	case slog.LevelKey:
		attr.Key = "severity"
	case slog.MessageKey:
		attr.Key = "message"
	case slog.SourceKey:
		if source, ok := attr.Value.Any().(*slog.Source); ok {
			return slog.Group(
				"logging.googleapis.com/sourceLocation",
				"file", source.File,
				"line", source.Line,
				"function", source.Function,
			)
		}
	}

	return attr
}

func MakeGCPLogger(level slog.Leveler, writer io.Writer) *slog.Logger {
	return slog.New(
		slog.NewJSONHandler(
			writer,
			&slog.HandlerOptions{
				AddSource:   true,
				Level:       level,
				ReplaceAttr: GCPReplaceAttr,
			},
		),
	)
}
