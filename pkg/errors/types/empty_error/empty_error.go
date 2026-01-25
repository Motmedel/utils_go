package empty_error

type Error struct {
	Field string
}

func (e *Error) Error() string {
	return "empty " + e.Field
}

func New(field string) *Error {
	return &Error{Field: field}
}
