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
	slog.Handler
	Extractors []ContextExtractor
}

func (contextHandler *ContextHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, extractor := range contextHandler.Extractors {
		if extractor != nil {
			if err := extractor.Handle(ctx, &record); err != nil {
				return fmt.Errorf("extractor handle: %w", err)
			}
		}
	}
	return contextHandler.Handler.Handle(ctx, record)
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
			inputTextualRepresentation, err := motmedelStrings.MakeTextualRepresentation(input)
			if err != nil {
				go func() {
					slog.Error(
						fmt.Sprintf(
							"An error occurred when making a textual representation of error input: %v",
							err,
						),
					)
				}()
			} else {
				var typeName string
				if t := reflect.TypeOf(input); t != nil {
					typeName = t.String()
				}

				attrs = append(
					attrs,
					slog.Group(
						"input",
						slog.String("value", inputTextualRepresentation),
						slog.String("type", typeName),
					),
				)
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
