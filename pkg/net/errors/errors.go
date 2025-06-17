package errors

import "errors"

var (
	ErrNilDomainBreakdown    = errors.New("nil domain breakdown")
	ErrIpVersionMismatch     = errors.New("IP address version mismatch")
	ErrNotOnSubnetBoundaries = errors.New("the start and end IP addresses are not on the exact subnet boundaries")
	ErrStartAfterEnd         = errors.New("the start IP address does not come before the end IP address")
	ErrUnexpectedIpVersion   = errors.New("unexpected IP version")
	ErrNilConn               = errors.New("nil conn")
	ErrEmptyPort             = errors.New("empty port")
)
