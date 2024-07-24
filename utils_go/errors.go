package utils_go

type InputErrorI interface {
	Error() string
	GetInput() any
}

type InputError struct {
	Err   error
	Input any
}

func (inputError *InputError) Error() string {
	return inputError.Err.Error()
}

func (inputError *InputError) GetInput() any {
	return inputError.Input
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
