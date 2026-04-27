package signer

import "context"

// Signer signs the V4 string-to-sign for a Google Cloud Storage signed URL using
// RSA-SHA256 on behalf of a specific service account.
//
// Email returns the service-account email — embedded in the X-Goog-Credential parameter.
// Sign returns the raw signature bytes; callers hex-encode for the URL.
type Signer interface {
	Email() string
	Sign(ctx context.Context, payload []byte) ([]byte, error)
}
