package types

import "crypto/tls"

type TlsContext struct {
	ConnectionState *tls.ConnectionState
}
