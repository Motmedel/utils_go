package errors

import "github.com/Motmedel/utils_go/pkg/errors"

var (
	ErrSignatureMismatch    = errors.New("signature mismatch")
	ErrUnsupportedAlgorithm = errors.New("unsupported algorithm")
	ErrCurveMismatch        = errors.New("curve mismatch")
	ErrUnsupportedCurve     = errors.New("unsupported curve")
)
