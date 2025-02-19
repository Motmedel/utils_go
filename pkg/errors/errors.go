package errors

import "errors"

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
