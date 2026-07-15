package cose

import (
	"crypto/ecdh"
	"fmt"
)

// Elliptic curve identifiers from the IANA COSE Elliptic Curves registry.
const (
	CurveP256 int64 = 1
	CurveP384 int64 = 2
	CurveP521 int64 = 3
)

const (
	keyTypeEc2 int64 = 2

	keyParameterKty int64 = 1
	keyParameterCrv int64 = -1
	keyParameterX   int64 = -2
	keyParameterY   int64 = -3
)

type curveParameters struct {
	curve          ecdh.Curve
	coordinateSize int
}

var curveRegistry = map[int64]*curveParameters{
	CurveP256: {curve: ecdh.P256(), coordinateSize: 32},
	CurveP384: {curve: ecdh.P384(), coordinateSize: 48},
	CurveP521: {curve: ecdh.P521(), coordinateSize: 66},
}

func curveId(curve ecdh.Curve) (int64, *curveParameters, bool) {
	for id, parameters := range curveRegistry {
		if parameters.curve == curve {
			return id, parameters, true
		}
	}

	return 0, nil, false
}

// ec2KeyFromPublicKey converts an ECDH public key into a COSE_Key map (EC2 key type).
func ec2KeyFromPublicKey(publicKey *ecdh.PublicKey) (map[int64]any, error) {
	id, parameters, ok := curveId(publicKey.Curve())
	if !ok {
		return nil, fmt.Errorf("%w: unsupported curve", ErrUnsupportedAlgorithm)
	}

	// Uncompressed point: 0x04 || X || Y
	raw := publicKey.Bytes()
	coordinateSize := parameters.coordinateSize
	if len(raw) != 1+2*coordinateSize {
		return nil, fmt.Errorf("%w: unexpected public key length %d", ErrMalformedKey, len(raw))
	}

	return map[int64]any{
		keyParameterKty: keyTypeEc2,
		keyParameterCrv: id,
		keyParameterX:   raw[1 : 1+coordinateSize],
		keyParameterY:   raw[1+coordinateSize:],
	}, nil
}

func ec2KeyParameter(keyMap map[any]any, label int64) (any, bool) {
	return headerValue(keyMap, label)
}

// publicKeyFromEc2Key converts a COSE_Key map (EC2 key type) into an ECDH public key.
func publicKeyFromEc2Key(keyMap map[any]any) (*ecdh.PublicKey, error) {
	ktyValue, ok := ec2KeyParameter(keyMap, keyParameterKty)
	if !ok {
		return nil, fmt.Errorf("%w: missing kty", ErrMalformedKey)
	}
	if kty, ok := toInt64(ktyValue); !ok || kty != keyTypeEc2 {
		return nil, fmt.Errorf("%w: unsupported key type %v", ErrUnsupportedAlgorithm, ktyValue)
	}

	crvValue, ok := ec2KeyParameter(keyMap, keyParameterCrv)
	if !ok {
		return nil, fmt.Errorf("%w: missing crv", ErrMalformedKey)
	}
	crv, ok := toInt64(crvValue)
	if !ok {
		return nil, fmt.Errorf("%w: malformed crv", ErrMalformedKey)
	}
	parameters, ok := curveRegistry[crv]
	if !ok {
		return nil, fmt.Errorf("%w: unsupported curve %d", ErrUnsupportedAlgorithm, crv)
	}

	xValue, xOk := ec2KeyParameter(keyMap, keyParameterX)
	yValue, yOk := ec2KeyParameter(keyMap, keyParameterY)
	if !xOk || !yOk {
		return nil, fmt.Errorf("%w: missing coordinate", ErrMalformedKey)
	}

	x, xOk := xValue.([]byte)
	y, yOk := yValue.([]byte)
	if !xOk || !yOk {
		return nil, fmt.Errorf("%w: malformed coordinate", ErrMalformedKey)
	}

	coordinateSize := parameters.coordinateSize
	if len(x) > coordinateSize || len(y) > coordinateSize {
		return nil, fmt.Errorf("%w: oversized coordinate", ErrMalformedKey)
	}

	raw := make([]byte, 1+2*coordinateSize)
	raw[0] = 4
	copy(raw[1+coordinateSize-len(x):], x)
	copy(raw[1+2*coordinateSize-len(y):], y)

	publicKey, err := parameters.curve.NewPublicKey(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: new public key: %w", ErrMalformedKey, err)
	}

	return publicKey, nil
}
