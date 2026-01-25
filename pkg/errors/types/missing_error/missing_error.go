package missing_error

type Error struct {
	Field string
}

func (e *Error) Error() string {
	return "missing " + e.Field
}

func New(field string) *Error {
	return &Error{Field: field}
}
