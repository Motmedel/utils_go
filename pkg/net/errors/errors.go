package errors

import "errors"

var (
	ErrCouldNotBreakDownDomain = errors.New("the domain could not be broken down")
)

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
