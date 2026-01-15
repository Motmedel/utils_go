package rule_id

const (
	// Certificate issues
	CertificateExpired          = "certificate_expired"
	CertificateExpiringSoon     = "certificate_expiring_soon"
	CertificateSelfSigned       = "certificate_self_signed"
	CertificateChainIncomplete  = "certificate_chain_incomplete"
	CertificateHostnameMismatch = "certificate_hostname_mismatch"
	CertificateNotYetValid      = "certificate_not_yet_valid"
	CertificateChainUntrusted   = "certificate_chain_untrusted"

	// Key strength
	WeakRsaKeyLength       = "weak_rsa_key_length"
	WeakSignatureAlgorithm = "weak_signature_algorithm"

	// TLS version
	TlsVersionDeprecated = "tls_version_deprecated"
	Tls10Enabled         = "tls_1_0_enabled"
	Tls11Enabled         = "tls_1_1_enabled"

	// Cipher suites
	WeakCipherSuite             = "weak_cipher_suite"
	CipherSuiteNoForwardSecrecy = "cipher_suite_no_forward_secrecy"
	CipherSuiteRc4              = "cipher_suite_rc4"
	CipherSuite3des             = "cipher_suite_3des"
	CipherSuiteExport           = "cipher_suite_export"
	CipherSuiteNull             = "cipher_suite_null"
)
