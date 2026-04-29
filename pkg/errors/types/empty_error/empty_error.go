package empty_error

import "fmt"

type Error struct {
	Field    string
	Instance string
}

func (e *Error) Error() string {
	msg := fmt.Sprintf("empty %s", e.Field)
	if e.Instance != "" {
		msg += fmt.Sprintf("(%s)", e.Instance)
	}

	return msg
}

func New(field string) *Error {
	return &Error{Field: field}
}

func NewWithInstance(field, instance string) *Error {
	return &Error{Field: field, Instance: instance}
}
