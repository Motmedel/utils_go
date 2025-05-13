package utils

import (
	"crypto/x509"
	motmedelIter "github.com/Motmedel/utils_go/pkg/iter"
)

func ExtractAlternativeNames(certificate *x509.Certificate) []string {
	if certificate == nil {
		return nil
	}

	var names []string

	names = append(names, certificate.DNSNames...)
	for _, ip := range certificate.IPAddresses {
		names = append(names, ip.String())
	}

	names = append(names, certificate.EmailAddresses...)
	for _, u := range certificate.URIs {
		names = append(names, u.String())
	}

	return motmedelIter.Set(names)
}
