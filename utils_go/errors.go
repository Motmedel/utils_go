package utils_go

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
