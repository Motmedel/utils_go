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
)

var CoseAlgNames = map[int]string{
	CoseAlgEs256: "Es256",
	CoseAlgEs384: "Es384",
	CoseAlgEs512: "Es512",

	CoseAlgRs256: "Rs256",
	CoseAlgRs384: "Rs384",
	CoseAlgRs512: "Rs512",

	CoseAlgPs256: "Ps256",
	CoseAlgPs384: "Ps384",
	CoseAlgPs512: "Ps512",
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
