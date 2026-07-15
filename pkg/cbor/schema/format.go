package schema

import (
	"errors"
	"fmt"
	"net/mail"
)

// FormatValidator validates a text value against a named format.
type FormatValidator func(value string) error

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
		return errors.New("not a bare address")
	}

	return nil
}
