package rsa

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	motmedelCryptoErrors "github.com/Motmedel/utils_go/pkg/crypto/errors"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"hash"
)

// NOTE: Not tested (AI-generated...)

type Method struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	HashFunc   func() hash.Hash
	Hash       crypto.Hash
	Name       string

	pss bool // true => RSASSA-PSS, false => RSASSA-PKCS1-v1_5
}

func (m *Method) hash(message []byte) ([]byte, error) {
	h := m.HashFunc()
	if _, err := h.Write(message); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

func (m *Method) Sign(message []byte) ([]byte, error) {
	if m.PrivateKey == nil {
		return nil, motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrEmptySecret)
	}

	digest, err := m.hash(message)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(err)
	}

	if m.pss {
		sig, err := rsa.SignPSS(rand.Reader, m.PrivateKey, m.Hash, digest, &rsa.PSSOptions{
			SaltLength: m.Hash.Size(), // per RFC 7518: salt length == hash length
			Hash:       m.Hash,
		})
		if err != nil {
			return nil, motmedelErrors.NewWithTrace(err)
		}
		return sig, nil
	}

	sig, err := rsa.SignPKCS1v15(rand.Reader, m.PrivateKey, m.Hash, digest)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(err)
	}
	return sig, nil
}

func (m *Method) Verify(message []byte, signature []byte) error {
	pub := m.PublicKey
	if pub == nil && m.PrivateKey != nil {
		pub = &m.PrivateKey.PublicKey
	}
	if pub == nil {
		return motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrEmptyPublicKey)
	}

	digest, err := m.hash(message)
	if err != nil {
		return motmedelErrors.NewWithTrace(err)
	}

	if m.pss {
		if err := rsa.VerifyPSS(pub, m.Hash, digest, signature, &rsa.PSSOptions{
			SaltLength: m.Hash.Size(),
			Hash:       m.Hash,
		}); err != nil {
			return motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrSignatureMismatch)
		}
		return nil
	}

	if err := rsa.VerifyPKCS1v15(pub, m.Hash, digest, signature); err != nil {
		return motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrSignatureMismatch)
	}
	return nil
}

func (m *Method) GetName() string {
	return m.Name
}

// TODO: Derive from the keys

func New(algorithm string, privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) (*Method, error) {
	var (
		hashFunc func() hash.Hash
		hash     crypto.Hash
		pss      bool
		name     string
	)

	switch algorithm {
	case "RS256":
		hashFunc = sha256.New
		hash = crypto.SHA256
		pss = false
		name = "RS256"
	case "RS384":
		hashFunc = sha512.New384
		hash = crypto.SHA384
		pss = false
		name = "RS384"
	case "RS512":
		hashFunc = sha512.New
		hash = crypto.SHA512
		pss = false
		name = "RS512"
	case "PS256":
		hashFunc = sha256.New
		hash = crypto.SHA256
		pss = true
		name = "PS256"
	case "PS384":
		hashFunc = sha512.New384
		hash = crypto.SHA384
		pss = true
		name = "PS384"
	case "PS512":
		hashFunc = sha512.New
		hash = crypto.SHA512
		pss = true
		name = "PS512"
	default:
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("%w: %q", motmedelCryptoErrors.ErrUnsupportedAlgorithm, algorithm),
		)
	}

	return &Method{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		HashFunc:   hashFunc,
		Hash:       hash,
		Name:       name,
		pss:        pss,
	}, nil
}
