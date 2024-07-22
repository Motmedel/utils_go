package utils_go

import "encoding/json"

type InputErrorI interface {
	Error() string
	GetInput() []byte
}

type InputError struct {
	error
	Input []byte
}

func (inputError *InputError) GetInput() []byte {
	return inputError.Input
}

type JsonSyntaxError struct {
	*json.SyntaxError
	*InputError
}

type CauseErrorI interface {
	Error() string
	GetCause() error
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
