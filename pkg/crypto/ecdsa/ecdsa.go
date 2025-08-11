package ecdsa

import (
	stdECDSA "crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	motmedelCryptoErrors "github.com/Motmedel/utils_go/pkg/crypto/errors"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"hash"
	"math/big"
)

func copyWithLeftPad(dst []byte, x *big.Int, size int) {
	b := x.Bytes()
	pad := size - len(b)
	for i := 0; i < pad; i++ {
		dst[i] = 0
	}
	copy(dst[pad:], b)
}

func canonicalizeS(s, n *big.Int) *big.Int {
	halfOrder := new(big.Int).Rsh(n, 1)
	if s.Cmp(halfOrder) == 1 {
		return new(big.Int).Sub(n, s)
	}
	return s
}

type Method struct {
	PrivateKey *stdECDSA.PrivateKey
	PublicKey  *stdECDSA.PublicKey
	HashFunc   func() hash.Hash
	Name       string

	curve elliptic.Curve
	size  int // byte length for R or S
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

	r, s, err := stdECDSA.Sign(rand.Reader, m.PrivateKey, digest)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(err)
	}

	// Canonicalize S to low-S form for interoperability
	s = canonicalizeS(s, m.curve.Params().N)

	// Serialize as fixed-length R||S
	sig := make([]byte, m.size*2)
	copyWithLeftPad(sig[:m.size], r, m.size)
	copyWithLeftPad(sig[m.size:], s, m.size)

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

	// Expect R||S with fixed lengths
	if len(signature) != 2*m.size {
		// Use signature mismatch to avoid leaking details about expected sizes
		return motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrSignatureMismatch)
	}

	r := new(big.Int).SetBytes(signature[:m.size])
	s := new(big.Int).SetBytes(signature[m.size:])

	digest, err := m.hash(message)
	if err != nil {
		return motmedelErrors.NewWithTrace(err)
	}

	if stdECDSA.Verify(pub, digest, r, s) {
		return nil
	}
	return motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrSignatureMismatch)
}

func (m *Method) GetName() string {
	return m.Name
}

func New(algorithm string, privateKey *stdECDSA.PrivateKey, publicKey *stdECDSA.PublicKey) (*Method, error) {
	var (
		curve    elliptic.Curve
		hashFunc func() hash.Hash
		name     string
	)

	switch algorithm {
	case "ES256":
		curve = elliptic.P256()
		hashFunc = sha256.New
		name = "ES256"
	case "ES384":
		curve = elliptic.P384()
		hashFunc = sha512.New384
		name = "ES384"
	case "ES512":
		curve = elliptic.P521()
		hashFunc = sha512.New
		name = "ES512"
	default:
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("%w: %q", motmedelCryptoErrors.ErrUnsupportedAlgorithm, algorithm),
		)
	}

	if privateKey != nil && privateKey.Curve != curve {
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("%w (private)", motmedelCryptoErrors.ErrCurveMismatch),
			privateKey.Curve,
			curve,
		)
	}

	if publicKey != nil && publicKey.Curve != curve {
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("%w (public)", motmedelCryptoErrors.ErrCurveMismatch),
			publicKey.Curve,
			curve,
		)
	}

	size := (curve.Params().BitSize + 7) / 8

	return &Method{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		HashFunc:   hashFunc,
		Name:       name,
		curve:      curve,
		size:       size,
	}, nil
}
