package errors

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

type errorContextType struct{}

var ErrorContextKey errorContextType

func ErrCtx(err error) context.Context {
	return context.WithValue(context.Background(), ErrorContextKey, err)
}

var (
	ErrSyntaxError   = errors.New("syntax error")
	ErrSemanticError = errors.New("semantic error")
)

func CollectWrappedErrors(err error) []error {
	var results []error

	queue := []error{err}

	for len(queue) > 0 {
		poppedErr := queue[0]
		queue = queue[1:]

		if poppedErr == nil {
			continue
		}

		if poppedErr != err {
			results = append(results, poppedErr)
		}

		switch typedErr := poppedErr.(type) {
		case interface{ Unwrap() error }:
			unwrappedErr := typedErr.Unwrap()
			if unwrappedErr == nil {
				continue
			}

			queue = append(queue, unwrappedErr)
		case interface{ Unwrap() []error }:
			for _, unwrappedErr := range typedErr.Unwrap() {
				if unwrappedErr == nil {
					continue
				}

				queue = append(queue, unwrappedErr)
			}
		}
	}

	return results
}

func removeFunctionFromStackTrace(stackTrace, funcName string) string {
	lines := strings.Split(stackTrace, "\n")
	filtered := make([]string, 0, len(lines))

	for i := 0; i < len(lines); i++ {
		// Check if the line matches the function signature (e.g., "main.funcName()")
		if strings.HasPrefix(lines[i], funcName+"(") {
			// Skip this line and the next line (file/line info)
			i++
		} else {
			filtered = append(filtered, lines[i])
		}
	}
	return strings.Join(filtered, "\n")
}

func getFunctionName(f any) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func CaptureStackTrace() string {
	buf := make([]byte, 64<<10)
	return strings.TrimSpace(
		removeFunctionFromStackTrace(string(buf[:runtime.Stack(buf, false)]), getFunctionName(CaptureStackTrace)),
	)
}

type CodeErrorI interface {
	Error() string
	GetCode() string
}

type IdErrorI interface {
	Error() string
	GetId() string
}

type StackTraceErrorI interface {
	Error() string
	GetStackTrace() string
}

type CauseErrorI interface {
	Error() string
	GetCause() error
	Unwrap() error
}

type InputErrorI interface {
	Error() string
	GetInput() any
}

type Error struct {
	Message    string
	Cause      error
	Input      any
	Code       string
	Id         string
	StackTrace string
}

func (err *Error) Error() string {
	return err.Message
}

func (err *Error) GetCause() error {
	return err.Cause
}

func (err *Error) GetInput() any {
	return err.Input
}

func (err *Error) GetCode() string {
	return err.Code
}

func (err *Error) GetId() string {
	return err.Id
}

func (err *Error) GetStackTrace() string {
	return err.StackTrace
}

func (err *Error) Unwrap() error {
	return err.Cause
}

func (err *Error) Is(target error) bool {
	_, ok := target.(*Error)
	return ok
}

type ExtendedError struct {
	error
	Input      any
	Code       string
	Id         string
	StackTrace string
}

func (err *ExtendedError) Error() string {
	if err.error == nil {
		return ""
	}
	return err.error.Error()
}

func (err *ExtendedError) GetInput() any {
	if input := err.Input; input != nil {
		return input
	}

	includedErr := err.error
	if includedErr == nil {
		return nil
	}

	if inputError, ok := includedErr.(InputErrorI); ok {
		return inputError.GetInput()
	}

	return nil
}

func (err *ExtendedError) GetCode() string {
	if code := err.Code; code != "" {
		return err.Code
	}

	includedErr := err.error
	if includedErr == nil {
		return ""
	}

	if codeError, ok := includedErr.(CodeErrorI); ok {
		return codeError.GetCode()
	}

	return ""
}

func (err *ExtendedError) GetId() string {
	if id := err.Id; id != "" {
		return err.Id
	}

	includedErr := err.error
	if includedErr == nil {
		return ""
	}

	if idError, ok := includedErr.(IdErrorI); ok {
		return idError.GetId()
	}

	return ""
}

func (err *ExtendedError) GetStackTrace() string {
	if stackTrace := err.StackTrace; stackTrace != "" {
		return stackTrace
	}

	includedErr := err.error
	if includedErr == nil {
		return ""
	}

	if stackTraceError, ok := includedErr.(StackTraceErrorI); ok {
		return stackTraceError.GetStackTrace()
	}

	return ""
}

func (err *ExtendedError) Unwrap() []error {
	switch typedErr := err.error.(type) {
	case interface{ Unwrap() error }:
		return []error{typedErr.Unwrap()}
	case interface{ Unwrap() []error }:
		return typedErr.Unwrap()
	}

	return nil
}

func (err *ExtendedError) Is(target error) bool {
	return errors.Is(err.error, target)
}

func MakeError(e any, input ...any) *ExtendedError {
	var err error

	// Expecting `e` to be an `error` or a string. If not, make it a string.
	switch typedE := e.(type) {
	case error:
		err = typedE
	case string:
		err = errors.New(typedE)
	case nil:
		break
	default:
		err = errors.New(fmt.Sprintf("%v", typedE))
	}

	var errInput any = input
	if len(input) == 0 {
		errInput = nil
	}
	if len(input) == 1 {
		errInput = input[0]
	}

	return &ExtendedError{error: err, Input: errInput}
}

func MakeErrorWithStackTrace(e any, input ...any) *ExtendedError {
	extendedErr := MakeError(e, input...)
	extendedErr.StackTrace = removeFunctionFromStackTrace(
		CaptureStackTrace(),
		getFunctionName(MakeErrorWithStackTrace),
	)

	return extendedErr
}

func IsAny(err error, targets ...error) bool {
	for _, target := range targets {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

func IsAll(err error, targets ...error) bool {
	for _, target := range targets {
		if !errors.Is(err, target) {
			return false
		}
	}
	return true
}
