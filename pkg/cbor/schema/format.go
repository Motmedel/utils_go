package schema

import (
	"errors"
	"fmt"
	"net/mail"
)

// FormatValidator validates a text value against a named format.
type FormatValidator func(value string) error

// ErrNotBareAddress reports an email value with anything besides the address
// itself, e.g. a display name ("Name <user@example.com>").
var ErrNotBareAddress = errors.New("not a bare address")

var formatRegistry = map[string]FormatValidator{
	"email": validateEmail,
}

// RegisterFormat makes a format validator available to schemas referencing it by name.
func RegisterFormat(name string, formatValidator FormatValidator) {
	formatRegistry[name] = formatValidator
}

func validateEmail(value string) error {
	address, err := mail.ParseAddress(value)
	if err != nil {
		return fmt.Errorf("parse address: %w", err)
	}

	// Reject forms with a display name ("Name <user@example.com>").
	if address.Address != value {
		return ErrNotBareAddress
	}

	return nil
}
