package crypto

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
)

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
