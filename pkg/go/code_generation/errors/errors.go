package errors

import "errors"

var (
	ErrInvalidValue                  = errors.New("invalid value")
	ErrUnsupportedFunctionFields     = errors.New("function fields are not supported")
	ErrUnsupportedChanField          = errors.New("chan fields are not supported")
	ErrUnsupportedUnsafePointerField = errors.New("unsafe pointer fields are not supported")
)
