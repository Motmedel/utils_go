package key_set

import "github.com/Motmedel/utils_go/pkg/json/jose/jwk/types/key"

type KeySet struct {
	Keys []*key.Key `json:"keys,omitempty"`
}
