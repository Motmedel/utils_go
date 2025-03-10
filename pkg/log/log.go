package log

import (
	"context"
	"fmt"
	motmedelContext "github.com/Motmedel/utils_go/pkg/context"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelStrings "github.com/Motmedel/utils_go/pkg/strings"
	"log/slog"
	"os/exec"
	"reflect"
	"strconv"
)

type ContextExtractor interface {
	Handle(context.Context, *slog.Record) error
}

type ContextExtractorFunction func(context.Context, *slog.Record) error

func (cef ContextExtractorFunction) Handle(ctx context.Context, record *slog.Record) error {
	return cef(ctx, record)
}

type ContextHandler struct {
	Next       slog.Handler
	Extractors []ContextExtractor
}

func (contextHandler *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	contextHandler.Next = contextHandler.Next.WithAttrs(attrs)
	return contextHandler
}

func (contextHandler *ContextHandler) WithGroup(name string) slog.Handler {
	contextHandler.Next = contextHandler.Next.WithGroup(name)
	return contextHandler
}

func (contextHandler *ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return contextHandler.Next.Enabled(ctx, level)
}

func (contextHandler *ContextHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, extractor := range contextHandler.Extractors {
		if extractor != nil {
			if err := extractor.Handle(ctx, &record); err != nil {
				return fmt.Errorf("extractor handle: %w", err)
			}
		}
	}

	return contextHandler.Next.Handle(ctx, record)
}

type ErrorContextExtractor struct {
	SkipCause      bool
	SkipInput      bool
	SkipStackTrace bool
	SkipOutput     bool
}

func (extractor *ErrorContextExtractor) MakeErrorAttrs(err error) []any {
	if err == nil {
		return nil
	}

	errorMessage := err.Error()
	errType := reflect.TypeOf(err).String()

	var attrs []any

	switch err.(type) {
	case *motmedelErrors.Error:
		break
	case *motmedelErrors.ExtendedError:
		break
	default:
		switch errType {
		case "*errors.errorString", "*fmt.wrapError":
			break
		default:
			attrs = append(attrs, slog.String("type", errType))
		}
	}

	if inputError, ok := err.(motmedelErrors.InputErrorI); ok && !extractor.SkipInput {
		if input := inputError.GetInput(); input != nil {
			var inputSlice []any
			var typeName string
			switch typedInput := input.(type) {
			case []any:
				inputSlice = typedInput
			default:
				inputSlice = []any{typedInput}
				if t := reflect.TypeOf(input); t != nil {
					typeName = t.String()
				}
			}

			var textualRepresentations []string

			for _, inputElement := range inputSlice {
				textualRepresentation, err := motmedelStrings.MakeTextualRepresentation(inputElement)
				if err != nil {
					slog.Error(
						fmt.Sprintf(
							"An error occurred when making a textual representation of error input: %v",
							fmt.Errorf("make textual representation: %w", err),
						),
					)
					continue
				}

				textualRepresentations = append(textualRepresentations, textualRepresentation)
			}

			if len(textualRepresentations) != 0 {
				var logValue any = textualRepresentations
				if len(textualRepresentations) == 1 {
					logValue = textualRepresentations[0]
				}

				logArgs := []any{slog.Any("value", logValue)}
				if typeName != "" {
					logArgs = append(logArgs, slog.String("type", typeName))
				}

				attrs = append(attrs, slog.Group("input", logArgs...))
			}
		}
	}

	if !extractor.SkipCause {
		wrappedErrors := motmedelErrors.CollectWrappedErrors(err)
		var lastWrappedErrorAttrs []any

		for i := len(wrappedErrors) - 1; i >= 0; i-- {
			wrappedError := wrappedErrors[i]
			if wrappedError == nil {
				continue
			}

			switch reflect.TypeOf(wrappedError).String() {
			case "*errors.joinError", "*fmt.wrapError":
				continue
			}

			wrappedErrorAttrs := extractor.MakeErrorAttrs(wrappedError)

			if lastWrappedErrorAttrs != nil {
				wrappedErrorAttrs = append(
					wrappedErrorAttrs,
					slog.Group("cause", lastWrappedErrorAttrs...),
				)
			}

			lastWrappedErrorAttrs = wrappedErrorAttrs
		}

		if lastWrappedErrorAttrs != nil {
			if errType == "*errors.joinError" {
				return lastWrappedErrorAttrs
			}
			attrs = append(attrs, slog.Group("cause", lastWrappedErrorAttrs...))
		}
	}

	if codeError, ok := err.(motmedelErrors.CodeErrorI); ok {
		if code := codeError.GetCode(); code != "" {
			attrs = append(attrs, slog.String("code", code))
		}
	}

	if idError, ok := err.(motmedelErrors.IdErrorI); ok {
		if id := idError.GetId(); id != "" {
			attrs = append(attrs, slog.String("id", id))
		}
	}

	if stackTraceError, ok := err.(motmedelErrors.StackTraceErrorI); ok && !extractor.SkipStackTrace {
		if stackTrace := stackTraceError.GetStackTrace(); stackTrace != "" {
			attrs = append(attrs, slog.String("stack_trace", stackTrace))
		}
	}

	if execExitError, ok := err.(*exec.ExitError); ok {
		exitCode := execExitError.ExitCode()
		if exitCode != 0 {
			attrs = append(attrs, slog.String("code", strconv.Itoa(exitCode)))
		}

		if stderr := execExitError.Stderr; len(stderr) != 0 && !extractor.SkipOutput {
			attrs = append(
				attrs,
				slog.Group(
					"output",
					slog.String("stderr", string(stderr)),
					slog.String("type", "stderr"),
				),
			)
		}
	}

	if errorMessage != "" {
		attrs = append(attrs, slog.String("message", errorMessage))
	}

	return attrs
}

func (extractor *ErrorContextExtractor) Handle(ctx context.Context, record *slog.Record) error {
	if record == nil {
		return nil
	}

	if logErr, ok := ctx.Value(motmedelContext.ErrorContextKey).(error); ok {
		record.Add(slog.Group("error", extractor.MakeErrorAttrs(logErr)...))
	}

	return nil
}

func AttrsFromMap(m map[string]any) []any {
	var attrs []any
	for key, value := range m {
		if stringAnyMap, ok := value.(map[string]any); ok {
			attrs = append(attrs, slog.Group(key, AttrsFromMap(stringAnyMap)...))
		} else {
			attrs = append(attrs, slog.Any(key, value))
		}
	}
	return attrs
}
