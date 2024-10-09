package errors

import "errors"

var (
	ErrCouldNotBreakDownDomain = errors.New("the domain could not be broken down")
)

type CauseErrorI interface {
	Error() string
	GetCause() error
	Unwrap() error
}

type CauseError struct {
	Message string
	Cause   error
}

func (causeError *CauseError) Error() string {
	return causeError.Message
}

func (causeError *CauseError) GetCause() error {
	return causeError.Cause
}

func (causeError *CauseError) Is(target error) bool {
	_, ok := target.(*CauseError)
	if !ok {
		return false
	}
	return true
}

func (causeError *CauseError) Unwrap() error {
	return causeError.Cause
}

type InputErrorI interface {
	Error() string
	GetInput() any
}

type InputError struct {
	Message string
	Cause   error
	Input   any
}

func (inputError *InputError) Error() string {
	return inputError.Message
}

func (inputError *InputError) GetCause() error {
	return inputError.Cause
}

func (inputError *InputError) GetInput() any {
	return inputError.Input
}

func (inputError *InputError) Unwrap() error {
	return inputError.Cause
}

func (inputError *InputError) Is(target error) bool {
	_, ok := target.(*InputError)
	if !ok {
		return false
	}
	return true
}

type CouldNotBreakDownDomainError struct {
	Domain string
}

func (couldNotBreakDownDomainError *CouldNotBreakDownDomainError) Is(target error) bool {
	return target == ErrCouldNotBreakDownDomain
}

func (couldNotBreakDownDomainError *CouldNotBreakDownDomainError) Error() string {
	return ErrCouldNotBreakDownDomain.Error()
}

func (couldNotBreakDownDomainError *CouldNotBreakDownDomainError) GetInput() any {
	return couldNotBreakDownDomainError.Domain
}
