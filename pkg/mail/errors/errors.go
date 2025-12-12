package errors

import "errors"

var (
	ErrEmptyFrom        = errors.New("empty from")
	ErrEmptyTo          = errors.New("empty to")
	ErrEmptySubject     = errors.New("empty subject")
	ErrEmptyContentType = errors.New("empty content type")
	ErrBadFromAddress   = errors.New("bad from address")
	ErrEmptyDomain      = errors.New("empty domain")
	ErrNilMessage       = errors.New("nil message")
)
