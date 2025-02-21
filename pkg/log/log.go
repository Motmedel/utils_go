package log

import (
	"context"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelStrings "github.com/Motmedel/utils_go/pkg/strings"
	"log/slog"
	"os"
	"os/exec"
	"reflect"
	"strconv"
)

type ContextHandler struct {
	slog.Handler
}

func (contextHandler *ContextHandler) Handle(ctx context.Context, record slog.Record) error {
	if logErr, ok := ctx.Value(motmedelErrors.ErrorContextKey).(error); ok {
		record.Add(slog.Group("error", MakeErrorAttrs(logErr)...))
	}

	return contextHandler.Handler.Handle(ctx, record)
}

func AttrsFromMap(m map[string]any) []any {
	var attrs []any
	for key, value := range m {
		attrs = append(attrs, slog.Any(key, value))
	}
	return attrs
}

func MakeErrorAttrs(err error) []any {
	if err == nil {
		return nil
	}

	errorMessage := err.Error()
	errType := reflect.TypeOf(err).String()

	var attrs []any

	switch errType {
	case "*errors.errorString", "*fmt.wrapError", "*errors.Error", "*errors.ExtendedError":
		break
	default:
		attrs = append(attrs, slog.String("type", errType))
	}

	if inputError, ok := err.(motmedelErrors.InputErrorI); ok {
		if input := inputError.GetInput(); input != nil {
			inputTextualRepresentation, err := motmedelStrings.MakeTextualRepresentation(input)
			if err != nil {
				go func() {
					LogError(
						"An error occurred when making a textual representation of error input.",
						err,
						slog.Default(),
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

		wrappedErrorAttrs := MakeErrorAttrs(wrappedError)

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

	if stackTraceError, ok := err.(motmedelErrors.StackTraceErrorI); ok {
		if stackTrace := stackTraceError.GetStackTrace(); stackTrace != "" {
			attrs = append(attrs, slog.String("stack_trace", stackTrace))
		}
	}

	if execExitError, ok := err.(*exec.ExitError); ok {
		exitCode := execExitError.ExitCode()
		if exitCode != 0 {
			attrs = append(attrs, slog.String("code", strconv.Itoa(exitCode)))
		}

		if stderr := execExitError.Stderr; len(stderr) != 0 {
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

func MakeErrorGroup(err error) *slog.Attr {
	group := slog.Group("error", MakeErrorAttrs(err)...)
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

func LogFatalWithExitCode(message string, err error, logger *slog.Logger, exitCode int) {
	if errorGroup := MakeErrorGroup(err); errorGroup != nil {
		logger.Error(message, *errorGroup)
	} else {
		logger.Error(message)
	}
	os.Exit(exitCode)
}

func LogFatalWithExitingMessage(message string, err error, logger *slog.Logger) {
	LogFatalWithExitCode(fmt.Sprintf("%s Exiting.", message), err, logger, 1)
}

func LogFatal(message string, err error, logger *slog.Logger) {
	LogFatalWithExitCode(message, err, logger, 1)
}

type Logger struct {
	*slog.Logger
}

func (logger *Logger) Error(message string, err error) {
	LogError(message, err, logger.Logger)
}

func (logger *Logger) Warning(message string, err error) {
	LogError(message, err, logger.Logger)
}

func (logger *Logger) Fatal(message string, err error) {
	LogFatalWithExitCode(message, err, logger.Logger, 1)
}

func (logger *Logger) FatalWithExitingMessage(message string, err error) {
	LogFatalWithExitCode(fmt.Sprintf("%s Exiting.", message), err, logger.Logger, 1)
}
