package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	"github.com/Motmedel/utils_go/pkg/tls/security/rule_id"
)

func generateTestCertificate(opts struct {
	notBefore          time.Time
	notAfter           time.Time
	keyBits            int
	signatureAlgorithm x509.SignatureAlgorithm
	isSelfSigned       bool
	dnsNames           []string
	subject            pkix.Name
	issuer             pkix.Name
}) (*x509.Certificate, error) {
	if opts.keyBits == 0 {
		opts.keyBits = 2048
	}
	if opts.signatureAlgorithm == 0 {
		opts.signatureAlgorithm = x509.SHA256WithRSA
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, opts.keyBits)
	if err != nil {
		return nil, err
	}

	if opts.subject.CommonName == "" {
		opts.subject = pkix.Name{CommonName: "test.example.com"}
	}
	if opts.isSelfSigned {
		opts.issuer = opts.subject
	} else if opts.issuer.CommonName == "" {
		opts.issuer = pkix.Name{CommonName: "Test CA"}
	}

	template := &x509.Certificate{
		SerialNumber:       big.NewInt(1),
		Subject:            opts.subject,
		Issuer:             opts.issuer,
		NotBefore:          opts.notBefore,
		NotAfter:           opts.notAfter,
		SignatureAlgorithm: opts.signatureAlgorithm,
		DNSNames:           opts.dnsNames,
		PublicKey:          &privateKey.PublicKey,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, err
	}

	return x509.ParseCertificate(certBytes)
}

func TestAnalyzeCertificate_Expired(t *testing.T) {
	now := time.Now()
	cert, err := generateTestCertificate(struct {
		notBefore          time.Time
		notAfter           time.Time
		keyBits            int
		signatureAlgorithm x509.SignatureAlgorithm
		isSelfSigned       bool
		dnsNames           []string
		subject            pkix.Name
		issuer             pkix.Name
	}{
		notBefore: now.AddDate(-1, 0, 0),
		notAfter:  now.AddDate(0, 0, -1), // Expired yesterday
	})
	if err != nil {
		t.Fatalf("failed to generate certificate: %v", err)
	}

	problems := AnalyzeCertificate(cert, now)

	found := false
	for _, p := range problems {
		if p.Id == rule_id.CertificateExpired {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected CertificateExpired problem for expired certificate")
	}
}

func TestAnalyzeCertificate_ExpiringSoon(t *testing.T) {
	now := time.Now()
	cert, err := generateTestCertificate(struct {
		notBefore          time.Time
		notAfter           time.Time
		keyBits            int
		signatureAlgorithm x509.SignatureAlgorithm
		isSelfSigned       bool
		dnsNames           []string
		subject            pkix.Name
		issuer             pkix.Name
	}{
		notBefore: now.AddDate(-1, 0, 0),
		notAfter:  now.AddDate(0, 0, 15), // Expires in 15 days
	})
	if err != nil {
		t.Fatalf("failed to generate certificate: %v", err)
	}

	problems := AnalyzeCertificate(cert, now)

	found := false
	for _, p := range problems {
		if p.Id == rule_id.CertificateExpiringSoon {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected CertificateExpiringSoon problem for certificate expiring within 30 days")
	}
}

func TestAnalyzeCertificate_NotYetValid(t *testing.T) {
	now := time.Now()
	cert, err := generateTestCertificate(struct {
		notBefore          time.Time
		notAfter           time.Time
		keyBits            int
		signatureAlgorithm x509.SignatureAlgorithm
		isSelfSigned       bool
		dnsNames           []string
		subject            pkix.Name
		issuer             pkix.Name
	}{
		notBefore: now.AddDate(0, 0, 1), // Starts tomorrow
		notAfter:  now.AddDate(1, 0, 0),
	})
	if err != nil {
		t.Fatalf("failed to generate certificate: %v", err)
	}

	problems := AnalyzeCertificate(cert, now)

	found := false
	for _, p := range problems {
		if p.Id == rule_id.CertificateNotYetValid {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected CertificateNotYetValid problem for certificate not yet valid")
	}
}

func TestAnalyzeCertificate_WeakRsaKey(t *testing.T) {
	now := time.Now()
	cert, err := generateTestCertificate(struct {
		notBefore          time.Time
		notAfter           time.Time
		keyBits            int
		signatureAlgorithm x509.SignatureAlgorithm
		isSelfSigned       bool
		dnsNames           []string
		subject            pkix.Name
		issuer             pkix.Name
	}{
		notBefore: now.AddDate(-1, 0, 0),
		notAfter:  now.AddDate(1, 0, 0),
		keyBits:   1024, // Weak key
	})
	if err != nil {
		t.Fatalf("failed to generate certificate: %v", err)
	}

	problems := AnalyzeCertificate(cert, now)

	found := false
	for _, p := range problems {
		if p.Id == rule_id.WeakRsaKeyLength {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected WeakRsaKeyLength problem for 1024-bit RSA key")
	}
}

func TestAnalyzeCertificate_WeakSignatureAlgorithm(t *testing.T) {
	now := time.Now()
	cert, err := generateTestCertificate(struct {
		notBefore          time.Time
		notAfter           time.Time
		keyBits            int
		signatureAlgorithm x509.SignatureAlgorithm
		isSelfSigned       bool
		dnsNames           []string
		subject            pkix.Name
		issuer             pkix.Name
	}{
		notBefore:          now.AddDate(-1, 0, 0),
		notAfter:           now.AddDate(1, 0, 0),
		signatureAlgorithm: x509.SHA1WithRSA, // Weak signature
	})
	if err != nil {
		t.Fatalf("failed to generate certificate: %v", err)
	}

	problems := AnalyzeCertificate(cert, now)

	found := false
	for _, p := range problems {
		if p.Id == rule_id.WeakSignatureAlgorithm {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected WeakSignatureAlgorithm problem for SHA1WithRSA")
	}
}

func TestAnalyzeCertificate_ValidCert(t *testing.T) {
	now := time.Now()
	cert, err := generateTestCertificate(struct {
		notBefore          time.Time
		notAfter           time.Time
		keyBits            int
		signatureAlgorithm x509.SignatureAlgorithm
		isSelfSigned       bool
		dnsNames           []string
		subject            pkix.Name
		issuer             pkix.Name
	}{
		notBefore:          now.AddDate(-1, 0, 0),
		notAfter:           now.AddDate(1, 0, 0), // Valid for another year
		keyBits:            2048,
		signatureAlgorithm: x509.SHA256WithRSA,
	})
	if err != nil {
		t.Fatalf("failed to generate certificate: %v", err)
	}

	problems := AnalyzeCertificate(cert, now)

	if len(problems) != 0 {
		t.Errorf("expected no problems for valid certificate, got %d: %v", len(problems), problems)
	}
}

func TestAnalyzeCertificateChain_SelfSigned(t *testing.T) {
	now := time.Now()
	cert, err := generateTestCertificate(struct {
		notBefore          time.Time
		notAfter           time.Time
		keyBits            int
		signatureAlgorithm x509.SignatureAlgorithm
		isSelfSigned       bool
		dnsNames           []string
		subject            pkix.Name
		issuer             pkix.Name
	}{
		notBefore:    now.AddDate(-1, 0, 0),
		notAfter:     now.AddDate(1, 0, 0),
		isSelfSigned: true,
		dnsNames:     []string{"test.example.com"},
		subject:      pkix.Name{CommonName: "test.example.com"},
	})
	if err != nil {
		t.Fatalf("failed to generate certificate: %v", err)
	}

	problems := AnalyzeCertificateChain([]*x509.Certificate{cert}, "test.example.com", now)

	found := false
	for _, p := range problems {
		if p.Id == rule_id.CertificateSelfSigned {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected CertificateSelfSigned problem for self-signed certificate")
	}
}

func TestAnalyzeTlsVersion_Tls10(t *testing.T) {
	problems := AnalyzeTlsVersion(tls.VersionTLS10)

	found := false
	for _, p := range problems {
		if p.Id == rule_id.Tls10Enabled {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Tls10Enabled problem for TLS 1.0")
	}
}

func TestAnalyzeTlsVersion_Tls11(t *testing.T) {
	problems := AnalyzeTlsVersion(tls.VersionTLS11)

	found := false
	for _, p := range problems {
		if p.Id == rule_id.Tls11Enabled {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected Tls11Enabled problem for TLS 1.1")
	}
}

func TestAnalyzeTlsVersion_Tls12(t *testing.T) {
	problems := AnalyzeTlsVersion(tls.VersionTLS12)

	if len(problems) != 0 {
		t.Errorf("expected no problems for TLS 1.2, got %d", len(problems))
	}
}

func TestAnalyzeTlsVersion_Tls13(t *testing.T) {
	problems := AnalyzeTlsVersion(tls.VersionTLS13)

	if len(problems) != 0 {
		t.Errorf("expected no problems for TLS 1.3, got %d", len(problems))
	}
}

func TestAnalyzeCipherSuite_Rc4(t *testing.T) {
	problems := AnalyzeCipherSuite(tls.TLS_RSA_WITH_RC4_128_SHA)

	found := false
	for _, p := range problems {
		if p.Id == rule_id.CipherSuiteRc4 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected CipherSuiteRc4 problem for RC4 cipher suite")
	}
}

func TestAnalyzeCipherSuite_3Des(t *testing.T) {
	problems := AnalyzeCipherSuite(tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA)

	found := false
	for _, p := range problems {
		if p.Id == rule_id.CipherSuite3des {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected CipherSuite3des problem for 3DES cipher suite")
	}
}

func TestAnalyzeCipherSuite_NoForwardSecrecy(t *testing.T) {
	problems := AnalyzeCipherSuite(tls.TLS_RSA_WITH_AES_128_CBC_SHA)

	found := false
	for _, p := range problems {
		if p.Id == rule_id.CipherSuiteNoForwardSecrecy {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected CipherSuiteNoForwardSecrecy problem for RSA key exchange")
	}
}

func TestAnalyzeCipherSuite_GoodCipherSuite(t *testing.T) {
	problems := AnalyzeCipherSuite(tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256)

	if len(problems) != 0 {
		t.Errorf("expected no problems for ECDHE cipher suite, got %d: %v", len(problems), problems)
	}
}

func TestGetTlsVersionName(t *testing.T) {
	tests := []struct {
		version  uint16
		expected string
	}{
		{tls.VersionSSL30, "SSL 3.0"},
		{tls.VersionTLS10, "TLS 1.0"},
		{tls.VersionTLS11, "TLS 1.1"},
		{tls.VersionTLS12, "TLS 1.2"},
		{tls.VersionTLS13, "TLS 1.3"},
		{0, "Unknown"},
	}

	for _, test := range tests {
		result := GetTlsVersionName(test.version)
		if result != test.expected {
			t.Errorf("GetTlsVersionName(%d) = %s, expected %s", test.version, result, test.expected)
		}
	}
}

func TestGetCipherSuiteName(t *testing.T) {
	name := GetCipherSuiteName(tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256)
	if name == "" || name == "Unknown" {
		t.Error("expected valid cipher suite name")
	}
}

func TestMakeRuleIdProblem(t *testing.T) {
	problem := MakeRuleIdProblem(rule_id.CertificateExpired)

	if problem.Id != rule_id.CertificateExpired {
		t.Errorf("expected Id %s, got %s", rule_id.CertificateExpired, problem.Id)
	}
	if problem.Title == "" {
		t.Error("expected non-empty Title")
	}
	if problem.Description == "" {
		t.Error("expected non-empty Description")
	}
	if problem.Severity == "" {
		t.Error("expected non-empty Severity")
	}
}

func TestExtractCertificateInfo(t *testing.T) {
	now := time.Now()
	cert, err := generateTestCertificate(struct {
		notBefore          time.Time
		notAfter           time.Time
		keyBits            int
		signatureAlgorithm x509.SignatureAlgorithm
		isSelfSigned       bool
		dnsNames           []string
		subject            pkix.Name
		issuer             pkix.Name
	}{
		notBefore:    now.AddDate(-1, 0, 0),
		notAfter:     now.AddDate(1, 0, 0),
		isSelfSigned: true,
		dnsNames:     []string{"test.example.com", "www.example.com"},
		subject:      pkix.Name{CommonName: "test.example.com"},
	})
	if err != nil {
		t.Fatalf("failed to generate certificate: %v", err)
	}

	info := ExtractCertificateInfo(cert)

	if info == nil {
		t.Fatal("expected non-nil CertificateInfo")
	}
	if info.Subject == "" {
		t.Error("expected non-empty Subject")
	}
	if info.Issuer == "" {
		t.Error("expected non-empty Issuer")
	}
	if info.NotBefore == "" {
		t.Error("expected non-empty NotBefore")
	}
	if info.NotAfter == "" {
		t.Error("expected non-empty NotAfter")
	}
	if info.SignatureAlgorithm == "" {
		t.Error("expected non-empty SignatureAlgorithm")
	}
	if info.PublicKeyAlgorithm != "RSA" {
		t.Errorf("expected RSA PublicKeyAlgorithm, got %s", info.PublicKeyAlgorithm)
	}
	if info.PublicKeyBitLength != 2048 {
		t.Errorf("expected 2048 bit key, got %d", info.PublicKeyBitLength)
	}
	if len(info.DNSNames) != 2 {
		t.Errorf("expected 2 DNS names, got %d", len(info.DNSNames))
	}
	if !info.IsSelfSigned {
		t.Error("expected IsSelfSigned to be true")
	}
}

func TestExtractCertificateInfo_Nil(t *testing.T) {
	info := ExtractCertificateInfo(nil)
	if info != nil {
		t.Error("expected nil for nil certificate")
	}
}

func TestAnalyzeCertificate_Nil(t *testing.T) {
	problems := AnalyzeCertificate(nil, time.Now())
	if len(problems) != 0 {
		t.Errorf("expected no problems for nil certificate, got %d", len(problems))
	}
}

func TestAnalyzeCertificateChain_Empty(t *testing.T) {
	problems := AnalyzeCertificateChain(nil, "example.com", time.Now())
	if len(problems) != 0 {
		t.Errorf("expected no problems for empty chain, got %d", len(problems))
	}
}

func TestBuildTlsSecurityData_Nil(t *testing.T) {
	data := BuildTlsSecurityData(nil, "example.com", time.Now())
	if data != nil {
		t.Error("expected nil for nil connection state")
	}
}
