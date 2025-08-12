package crypto

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
)

const (
	CoseAlgEs256 = -7
	CoseAlgEs384 = -35
	CoseAlgEs512 = -36

	CoseAlgRs256 = -257
	CoseAlgRs384 = -258
	CoseAlgRs512 = -259

	CoseAlgPs256 = -37
	CoseAlgPs384 = -38
	CoseAlgPs512 = -39

	AlgEs256 = "ES256"
	AlgEs384 = "ES384"
	AlgEs512 = "ES512"

	AlgRs256 = "RS256"
	AlgRs384 = "RS384"
	AlgRs512 = "RS512"

	AlgPs256 = "PS256"
	AlgPs384 = "PS384"
	AlgPs512 = "PS512"
)

var CoseAlgNames = map[int]string{
	CoseAlgEs256: AlgEs256,
	CoseAlgEs384: AlgEs384,
	CoseAlgEs512: AlgEs512,

	CoseAlgRs256: AlgRs256,
	CoseAlgRs384: AlgRs384,
	CoseAlgRs512: AlgRs512,

	CoseAlgPs256: AlgPs256,
	CoseAlgPs384: AlgPs384,
	CoseAlgPs512: AlgPs512,
}

func MakeRawDerCertificateChain(certificates []*x509.Certificate) [][]byte {
	var certificateChain [][]byte

	for _, certificate := range certificates {
		if certificate == nil {
			continue
		}
		if raw := certificate.Raw; len(raw) != 0 {
			certificateChain = append(certificateChain, raw)
		}
	}

	return certificateChain
}

func MakeTlsCertificateFromX509Certificates(certificates []*x509.Certificate, key crypto.PrivateKey) *tls.Certificate {
	if len(certificates) == 0 {
		return nil
	}

	return &tls.Certificate{
		Certificate: MakeRawDerCertificateChain(certificates),
		PrivateKey:  key,
		Leaf:        certificates[0],
	}
}
