package errors

import "errors"

var (
	ErrIpVersionMismatch          = errors.New("IP address version mismatch")
	ErrNotOnSubnetBoundaries      = errors.New("the start and end IP addresses are not on the exact subnet boundaries")
	ErrStartAfterEnd              = errors.New("the start IP address does not come before the end IP address")
	ErrUnexpectedIpVersion        = errors.New("unexpected IP version")
	ErrUndeterminableIpVersion    = errors.New("undeterminable ip version")
	ErrUndeterminableTargetFormat = errors.New("undeterminable target format")
)
