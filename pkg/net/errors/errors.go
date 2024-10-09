package errors

import "errors"

var (
	ErrCouldNotBreakDownDomain = errors.New("the domain could not be broken down")
	ErrIpVersionMismatch       = errors.New("IP address version mismatch")
	ErrNotOnSubnetBoundaries   = errors.New("the start and end IP addresses are not on the exact subnet boundaries")
	ErrStartAfterEnd           = errors.New("the start IP address does not come before the end IP address")
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
