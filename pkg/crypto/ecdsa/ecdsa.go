package ecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/asn1"
	"fmt"
	motmedelCryptoErrors "github.com/Motmedel/utils_go/pkg/crypto/errors"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/utils"
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
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
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

	r, s, err := ecdsa.Sign(rand.Reader, m.PrivateKey, digest)
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
		return motmedelErrors.NewWithTrace(fmt.Errorf("hash: %w", err))
	}

	if !ecdsa.Verify(pub, digest, r, s) {
		return motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrSignatureMismatch)
	}

	return nil
}

func (m *Method) GetName() string {
	return m.Name
}

func FromPublicKey(publicKey *ecdsa.PublicKey) (*Method, error) {
	return New(nil, publicKey)
}

func FromPrivateKey(privateKey *ecdsa.PrivateKey) (*Method, error) {
	return New(privateKey, nil)
}

// deriveAlgFromCurve picks the JOSE alg name and hash function based on the curve.
func deriveAlgFromCurveParams(curveParams *elliptic.CurveParams) (name string, hashFunc func() hash.Hash, err error) {
	if curveParams == nil {
		return "", nil, nil
	}

	switch curveName := curveParams.Name; curveName {
	case "P-256":
		return "ES256", sha256.New, nil
	case "P-384":
		return "ES384", sha512.New384, nil
	case "P-521":
		return "ES512", sha512.New, nil
	default:
		return "", nil, motmedelErrors.NewWithTrace(
			motmedelCryptoErrors.ErrUnsupportedCurve,
			curveName,
		)
	}
}

func getCurveParams(curve elliptic.Curve) *elliptic.CurveParams {
	if utils.IsNil(curve) {
		return nil
	}

	return curve.Params()
}

func New(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey) (*Method, error) {
	if privateKey == nil && publicKey == nil {
		return nil, nil
	}

	var curve elliptic.Curve
	var privateKeyCurveParams *elliptic.CurveParams
	var publicKeyCurveParams *elliptic.CurveParams

	if privateKey != nil {
		privateKeyCurveParams = getCurveParams(privateKey.Curve)

		curve = privateKey.Curve
	}

	if publicKey != nil {
		publicKeyCurve := publicKey.Curve
		publicKeyCurveParams = getCurveParams(publicKeyCurve)

		if utils.IsNil(curve) {
			curve = publicKeyCurve
		}
	}

	if utils.IsNil(curve) {
		return nil, motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrNilCurve)
	}

	if privateKeyCurveParams != nil && publicKeyCurveParams != nil && privateKeyCurveParams.Name != publicKeyCurveParams.Name {
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("%w (private/public)", motmedelCryptoErrors.ErrCurveMismatch),
			privateKeyCurveParams.Name, publicKeyCurveParams.Name,
		)
	}

	name, hashFunc, err := deriveAlgFromCurveParams(curve.Params())
	if err != nil {
		return nil, err
	}

	size := (curve.Params().BitSize + 7) / 8

	return &Method{PrivateKey: privateKey,
		PublicKey: publicKey,
		HashFunc:  hashFunc,
		Name:      name,
		curve:     curve,
		size:      size,
	}, nil
}

type Asn1DerEncodedMethod struct {
	Method
}

func (m *Asn1DerEncodedMethod) Sign(message []byte) ([]byte, error) {
	raw, err := m.Method.Sign(message)
	if err != nil {
		return nil, fmt.Errorf("method sign: %w", err)
	}

	if len(raw) != 2*m.size {
		return nil, fmt.Errorf("invalid signature length: %d", len(raw))
	}

	r := new(big.Int).SetBytes(raw[:m.size])
	s := new(big.Int).SetBytes(raw[m.size:])

	data, err := asn1.Marshal(struct{ R, S *big.Int }{R: r, S: s})
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("asn1 marshal: %w", err), r, s)
	}

	return data, nil
}

func (m *Asn1DerEncodedMethod) Verify(message []byte, signature []byte) error {
	publicKey := m.PublicKey
	if publicKey == nil && m.PrivateKey != nil {
		publicKey = &m.PrivateKey.PublicKey
	}
	if publicKey == nil {
		return motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrEmptyPublicKey)
	}

	var decodedSignature struct {
		R, S *big.Int
	}

	if _, err := asn1.Unmarshal(signature, &decodedSignature); err != nil {
		return motmedelErrors.NewWithTrace(fmt.Errorf("asn1 unmarshal: %w", err))
	}

	digest, err := m.hash(message)
	if err != nil {
		return motmedelErrors.NewWithTrace(fmt.Errorf("hash: %w", err))
	}

	if !ecdsa.Verify(publicKey, digest, decodedSignature.R, decodedSignature.S) {
		return motmedelErrors.NewWithTrace(motmedelCryptoErrors.ErrSignatureMismatch)
	}

	return nil
}
