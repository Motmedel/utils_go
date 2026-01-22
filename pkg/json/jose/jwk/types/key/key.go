package key

type Key struct {
	Kty string   `json:"kty,omitempty"`
	Alg string   `json:"alg,omitempty"`
	Use string   `json:"use,omitempty"`
	Kid string   `json:"kid,omitempty"`
	N   string   `json:"n,omitempty"`
	E   string   `json:"e,omitempty"`
	X5c []string `json:"x5c,omitempty"`
}

type Keys struct {
	Keys []Key `json:"keys,omitempty"`
}
