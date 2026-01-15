package security

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"strings"
	"time"

	motmedelTlsSecurity "github.com/Motmedel/types_go/pkg/types/tls_security"
	"github.com/Motmedel/utils_go/pkg/tls/security/rule_id"
	"github.com/Motmedel/utils_go/pkg/tls/security/rule_id_mappings"
)

// MakeRuleIdProblem creates a TlsProblem from a rule ID using the mappings.
func MakeRuleIdProblem(ruleId string) *motmedelTlsSecurity.TlsProblem {
	return &motmedelTlsSecurity.TlsProblem{
		Id:          ruleId,
		Title:       rule_id_mappings.RuleIdToTitle[ruleId],
		Description: rule_id_mappings.RuleIdToDescription[ruleId],
		Severity:    rule_id_mappings.RuleIdToSeverity[ruleId],
	}
}

// GetTlsVersionName returns the human-readable name for a TLS version.
func GetTlsVersionName(version uint16) string {
	switch version {
	case tls.VersionSSL30:
		return "SSL 3.0"
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return "Unknown"
	}
}

// GetCipherSuiteName returns the name for a cipher suite ID.
func GetCipherSuiteName(suite uint16) string {
	if cs := tls.CipherSuiteName(suite); cs != "" {
		return cs
	}
	return "Unknown"
}

// getPublicKeyBitLength returns the bit length of a public key.
func getPublicKeyBitLength(pubKey any) int {
	switch key := pubKey.(type) {
	case *rsa.PublicKey:
		return key.N.BitLen()
	case *ecdsa.PublicKey:
		return key.Curve.Params().BitSize
	case ed25519.PublicKey:
		return 256 // Ed25519 keys are always 256 bits
	default:
		return 0
	}
}

// getPublicKeyAlgorithm returns the algorithm name for a public key.
func getPublicKeyAlgorithm(pubKey any) string {
	switch pubKey.(type) {
	case *rsa.PublicKey:
		return "RSA"
	case *ecdsa.PublicKey:
		return "ECDSA"
	case ed25519.PublicKey:
		return "Ed25519"
	default:
		return "Unknown"
	}
}

// ExtractCertificateInfo extracts certificate information into a CertificateInfo struct.
func ExtractCertificateInfo(cert *x509.Certificate) *motmedelTlsSecurity.CertificateInfo {
	if cert == nil {
		return nil
	}

	var ipAddresses []string
	for _, ip := range cert.IPAddresses {
		ipAddresses = append(ipAddresses, ip.String())
	}

	return &motmedelTlsSecurity.CertificateInfo{
		Subject:            cert.Subject.String(),
		Issuer:             cert.Issuer.String(),
		NotBefore:          cert.NotBefore.Format(time.RFC3339),
		NotAfter:           cert.NotAfter.Format(time.RFC3339),
		SerialNumber:       cert.SerialNumber.String(),
		SignatureAlgorithm: cert.SignatureAlgorithm.String(),
		PublicKeyAlgorithm: getPublicKeyAlgorithm(cert.PublicKey),
		PublicKeyBitLength: getPublicKeyBitLength(cert.PublicKey),
		DNSNames:           cert.DNSNames,
		IPAddresses:        ipAddresses,
		IsSelfSigned:       cert.Issuer.String() == cert.Subject.String(),
	}
}

// AnalyzeCertificate checks a single certificate for security issues.
func AnalyzeCertificate(cert *x509.Certificate, referenceTime time.Time) []*motmedelTlsSecurity.TlsProblem {
	var problems []*motmedelTlsSecurity.TlsProblem

	if cert == nil {
		return problems
	}

	// Check if certificate has expired
	if referenceTime.After(cert.NotAfter) {
		problems = append(problems, MakeRuleIdProblem(rule_id.CertificateExpired))
	} else if referenceTime.AddDate(0, 0, 30).After(cert.NotAfter) {
		// Check if certificate is expiring soon (within 30 days)
		problems = append(problems, MakeRuleIdProblem(rule_id.CertificateExpiringSoon))
	}

	// Check if certificate is not yet valid
	if referenceTime.Before(cert.NotBefore) {
		problems = append(problems, MakeRuleIdProblem(rule_id.CertificateNotYetValid))
	}

	// Check for weak RSA key length
	if rsaKey, ok := cert.PublicKey.(*rsa.PublicKey); ok {
		if rsaKey.N.BitLen() < 2048 {
			problems = append(problems, MakeRuleIdProblem(rule_id.WeakRsaKeyLength))
		}
	}

	// Check for weak signature algorithms
	switch cert.SignatureAlgorithm {
	case x509.MD2WithRSA, x509.MD5WithRSA, x509.SHA1WithRSA, x509.DSAWithSHA1, x509.ECDSAWithSHA1:
		problems = append(problems, MakeRuleIdProblem(rule_id.WeakSignatureAlgorithm))
	}

	return problems
}

// AnalyzeCertificateChain validates the certificate chain and checks for chain-related issues.
func AnalyzeCertificateChain(certs []*x509.Certificate, hostname string, referenceTime time.Time) []*motmedelTlsSecurity.TlsProblem {
	var problems []*motmedelTlsSecurity.TlsProblem

	if len(certs) == 0 {
		return problems
	}

	leafCert := certs[0]

	// Check self-signed
	if leafCert.Issuer.String() == leafCert.Subject.String() {
		problems = append(problems, MakeRuleIdProblem(rule_id.CertificateSelfSigned))
	}

	// Full chain validation using system trust store
	opts := x509.VerifyOptions{
		DNSName:       hostname,
		Intermediates: x509.NewCertPool(),
		CurrentTime:   referenceTime,
	}

	// Add intermediate certs
	for _, cert := range certs[1:] {
		opts.Intermediates.AddCert(cert)
	}

	if _, err := leafCert.Verify(opts); err != nil {
		errStr := err.Error()
		switch {
		case strings.Contains(errStr, "certificate has expired") || strings.Contains(errStr, "x509: certificate has expired"):
			// Already handled by AnalyzeCertificate
		case strings.Contains(errStr, "unknown authority") || strings.Contains(errStr, "signed by unknown authority"):
			problems = append(problems, MakeRuleIdProblem(rule_id.CertificateChainUntrusted))
		case strings.Contains(errStr, "doesn't contain any IP SANs") ||
			strings.Contains(errStr, "certificate is not valid for any names") ||
			strings.Contains(errStr, "certificate is valid for") ||
			strings.Contains(errStr, "cannot validate certificate"):
			problems = append(problems, MakeRuleIdProblem(rule_id.CertificateHostnameMismatch))
		default:
			problem := MakeRuleIdProblem(rule_id.CertificateChainIncomplete)
			problem.Details = errStr
			problems = append(problems, problem)
		}
	}

	return problems
}

// AnalyzeTlsVersion checks for deprecated TLS versions.
func AnalyzeTlsVersion(version uint16) []*motmedelTlsSecurity.TlsProblem {
	var problems []*motmedelTlsSecurity.TlsProblem

	switch version {
	case tls.VersionSSL30:
		problems = append(problems, MakeRuleIdProblem(rule_id.TlsVersionDeprecated))
	case tls.VersionTLS10:
		problems = append(problems, MakeRuleIdProblem(rule_id.Tls10Enabled))
	case tls.VersionTLS11:
		problems = append(problems, MakeRuleIdProblem(rule_id.Tls11Enabled))
	}

	return problems
}

// isRc4CipherSuite checks if a cipher suite uses RC4.
func isRc4CipherSuite(suite uint16) bool {
	rc4Suites := []uint16{
		tls.TLS_RSA_WITH_RC4_128_SHA,
		tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
		tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
	}
	for _, s := range rc4Suites {
		if suite == s {
			return true
		}
	}
	return false
}

// is3DesCipherSuite checks if a cipher suite uses 3DES.
func is3DesCipherSuite(suite uint16) bool {
	tripleDesSuites := []uint16{
		tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
		tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
	}
	for _, s := range tripleDesSuites {
		if suite == s {
			return true
		}
	}
	return false
}

// hasForwardSecrecy checks if a cipher suite provides forward secrecy.
func hasForwardSecrecy(suite uint16) bool {
	name := tls.CipherSuiteName(suite)
	return strings.Contains(name, "ECDHE") || strings.Contains(name, "DHE")
}

// AnalyzeCipherSuite checks for weak cipher suites.
func AnalyzeCipherSuite(cipherSuite uint16) []*motmedelTlsSecurity.TlsProblem {
	var problems []*motmedelTlsSecurity.TlsProblem

	name := tls.CipherSuiteName(cipherSuite)

	// Check for NULL cipher
	if strings.Contains(name, "NULL") {
		problems = append(problems, MakeRuleIdProblem(rule_id.CipherSuiteNull))
		return problems
	}

	// Check for export cipher
	if strings.Contains(name, "EXPORT") {
		problems = append(problems, MakeRuleIdProblem(rule_id.CipherSuiteExport))
		return problems
	}

	// Check for RC4
	if isRc4CipherSuite(cipherSuite) {
		problems = append(problems, MakeRuleIdProblem(rule_id.CipherSuiteRc4))
	}

	// Check for 3DES
	if is3DesCipherSuite(cipherSuite) {
		problems = append(problems, MakeRuleIdProblem(rule_id.CipherSuite3des))
	}

	// Check for lack of forward secrecy
	if !hasForwardSecrecy(cipherSuite) {
		problems = append(problems, MakeRuleIdProblem(rule_id.CipherSuiteNoForwardSecrecy))
	}

	return problems
}

// AnalyzeConnectionState is the main entry point that combines all security checks.
func AnalyzeConnectionState(state *tls.ConnectionState, hostname string, referenceTime time.Time) []*motmedelTlsSecurity.TlsProblem {
	var problems []*motmedelTlsSecurity.TlsProblem

	if state == nil {
		return problems
	}

	// Analyze TLS version
	problems = append(problems, AnalyzeTlsVersion(state.Version)...)

	// Analyze cipher suite
	problems = append(problems, AnalyzeCipherSuite(state.CipherSuite)...)

	// Analyze certificate chain
	if len(state.PeerCertificates) > 0 {
		// Analyze the leaf certificate
		problems = append(problems, AnalyzeCertificate(state.PeerCertificates[0], referenceTime)...)

		// Analyze the full chain
		problems = append(problems, AnalyzeCertificateChain(state.PeerCertificates, hostname, referenceTime)...)
	}

	return problems
}

// BuildTlsSecurityData creates a TlsSecurityData struct from a TLS connection state.
func BuildTlsSecurityData(state *tls.ConnectionState, hostname string, referenceTime time.Time) *motmedelTlsSecurity.TlsSecurityData {
	if state == nil {
		return nil
	}

	data := &motmedelTlsSecurity.TlsSecurityData{
		ObservedAt:              referenceTime.Format(time.RFC3339),
		ObservedAtNanoTimestamp: referenceTime.UnixNano(),
		Version:                 state.Version,
		VersionName:             GetTlsVersionName(state.Version),
		CipherSuite:             state.CipherSuite,
		CipherSuiteName:         GetCipherSuiteName(state.CipherSuite),
		ServerName:              state.ServerName,
	}

	// Extract certificate chain info
	for _, cert := range state.PeerCertificates {
		data.CertificateChain = append(data.CertificateChain, ExtractCertificateInfo(cert))
	}

	// Run security analysis
	data.Problems = AnalyzeConnectionState(state, hostname, referenceTime)

	return data
}
