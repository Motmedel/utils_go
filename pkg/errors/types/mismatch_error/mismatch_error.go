package mismatch_error

type Error struct {
	Field  string
	Values []any
}

func (e *Error) Error() string {
	return e.Field + " mismatch"
}

func New(field string, values ...any) *Error {
	return &Error{Field: field, Values: values}
}
