package errors

import "errors"

var (
	ErrNotStruct   = errors.New("not a struct")
	ErrNotTypeName = errors.New("not a type name")
	ErrNotNamed    = errors.New("not a named")
)
