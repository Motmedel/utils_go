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

// NOTE: Apparently, according to convention, you should not use string keys, and you should not use context as a kind
// storage for optional parameters. But it seems to me a logger is a special case, and I see no better, idiomatic way.

// TODO: Change type

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
		attrs = append(attrs, slog.Any(key, value))
	}
	return attrs
}

func makeErrorAttrs(err error) []any {
	if err == nil {
		return nil
	}

	errorMessage := err.Error()
	errType := reflect.TypeOf(err).String()

	var attrs []any

	switch errType {
	case "*errors.errorString", "*fmt.wrapError":
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

		wrappedErrorAttrs := makeErrorAttrs(wrappedError)

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

		errorMessage = string(execExitError.Stderr)
	}

	attrs = append(attrs, slog.String("message", errorMessage))

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
