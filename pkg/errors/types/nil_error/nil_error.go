package nil_error

type Error struct {
	Field string
}

func (e *Error) Error() string {
	return "nil " + e.Field
}

func New(field string) *Error {
	return &Error{Field: field}
}

// TODO: Add "instance" field.
