package content_security_policy

import (
	"strings"
	"testing"

)

func TestContentSecurityPolicy_String_UsesRaw(t *testing.T) {
	policy := "default-src 'self' https: cdn.example.com; upgrade-insecure-request"

	csp := &ContentSecurityPolicy{Raw: policy}
	if got := csp.String(); got != policy {
		t.Fatalf("String() mismatch.\nexpected: %q\n     got: %q", policy, got)
	}
}

func TestContentSecurityPolicy_String_Constructed(t *testing.T) {
	csp := &ContentSecurityPolicy{
		Directives: []DirectiveI{
			&DefaultSrcDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "default-src", RawName: "default-src"}, Sources: []SourceI{
				&KeywordSource{Keyword: "'self'"},
				&SchemeSource{Scheme: "https"},
				&HostSource{Host: "cdn.example.com"},
			}}},
			&UpgradeInsecureRequestDirective{Directive: Directive{Name: "upgrade-insecure-request", RawName: "upgrade-insecure-request"}},
		},
	}

	expected := "default-src 'self' https: cdn.example.com; upgrade-insecure-request"
	if got := csp.String(); got != expected {
		t.Fatalf("constructed String() mismatch.\nexpected: %q\n     got: %q", expected, got)
	}
}

func TestContentSecurityPolicy_String_IncludesOtherButNotIneffective(t *testing.T) {
	csp := &ContentSecurityPolicy{
		Directives: []DirectiveI{
			&DefaultSrcDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "default-src", RawName: "default-src"}, Sources: []SourceI{
				&KeywordSource{Keyword: "'self'"},
			}}},
		},
		OtherDirectives: []DirectiveI{
			&Directive{Name: "foo", RawName: "foo", RawValue: "bar"},
		},
		IneffectiveDirectives: []DirectiveI{
			&DefaultSrcDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "default-src", RawName: "default-src", RawValue: "https://other.example.com"}}},
		},
	}

	expected := "default-src 'self'; foo bar"
	if got := csp.String(); got != expected {
		t.Fatalf("String() should include other but not ineffective directives. expected %q, got %q", expected, got)
	}
}


func TestContentSecurityPolicy_String_Constructed_Comprehensive(t *testing.T) {
	csp := &ContentSecurityPolicy{
		Directives: []DirectiveI{
			&DefaultSrcDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "default-src", RawName: "default-src"}, Sources: []SourceI{
				&KeywordSource{Keyword: "'self'"},
				&HostSource{Scheme: "https", Host: "cdn.example.com"},
			}}},
			&ScriptSrcDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "script-src", RawName: "script-src"}, Sources: []SourceI{
				&KeywordSource{Keyword: "'unsafe-inline'"},
				&KeywordSource{Keyword: "'unsafe-eval'"},
				&KeywordSource{Keyword: "'strict-dynamic'"},
				&SchemeSource{Scheme: "https"},
				&SchemeSource{Scheme: "http"},
				&NonceSource{Base64Value: "dGVzdA=="},
				&HashSource{HashAlgorithm: "sha256", Base64Value: "AbCd012+/_-=="},
			}}},
			&StyleSrcDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "style-src", RawName: "style-src"}, Sources: []SourceI{
				&KeywordSource{Keyword: "'report-sample'"},
				&HostSource{Scheme: "https", Host: "styles.example.com"},
			}}},
			&ImgSrcDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "img-src", RawName: "img-src"}, Sources: []SourceI{
				&HostSource{Scheme: "https", Host: "example.com", PortString: "443", Path: "/path"},
				&SchemeSource{Scheme: "data"},
			}}},
			&ConnectSrcDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "connect-src", RawName: "connect-src"}, Sources: []SourceI{
				&HostSource{Host: "*.example.com", PortString: "*"},
			}}},
			&FrameAncestorsDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "frame-ancestors", RawName: "frame-ancestors"}, Sources: []SourceI{
				&KeywordSource{Keyword: "'self'"},
				&HostSource{Scheme: "https", Host: "parent.example.com"},
			}}},
			&SandboxDirective{Directive: Directive{Name: "sandbox", RawName: "sandbox"}, Tokens: []string{"allow-same-origin", "allow-scripts"}},
			&ReportUriDirective{Directive: Directive{Name: "report-uri", RawName: "report-uri"}, UriReferences: []string{"/csp", "/csp2", "https://report.example.com/endpoint"}},
			&ReportToDirective{Directive: Directive{Name: "report-to", RawName: "report-to"}, Token: "csp-endpoint"},
			&RequireSriForDirective{Directive: Directive{Name: "require-sri-for", RawName: "require-sri-for"}, ResourceTypes: []string{"script", "style"}},
			&UpgradeInsecureRequestDirective{Directive: Directive{Name: "upgrade-insecure-request", RawName: "upgrade-insecure-request"}},
			&WebrtcDirective{Directive: Directive{Name: "webrtc", RawName: "webrtc", RawValue: "allow"}},
			&BaseUriDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "base-uri", RawName: "base-uri"}, Sources: []SourceI{
				&KeywordSource{Keyword: "'self'"},
			}}},
			&ObjectSrcDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "object-src", RawName: "object-src"}, Sources: []SourceI{
				&NoneSource{},
			}}},
			&WorkerSrcDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "worker-src", RawName: "worker-src"}, Sources: []SourceI{
				&SchemeSource{Scheme: "data"},
				&SchemeSource{Scheme: "blob"},
			}}},
			&ChildSrcDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "child-src", RawName: "child-src"}, Sources: []SourceI{
				&HostSource{Scheme: "https", Host: "child.example.com"},
			}}},
		},
	}

	expected := strings.Join([]string{
		"default-src 'self' https://cdn.example.com",
		"script-src 'unsafe-inline' 'unsafe-eval' 'strict-dynamic' https: http: 'nonce-dGVzdA==' 'sha256-AbCd012+/_-=='",
		"style-src 'report-sample' https://styles.example.com",
		"img-src https://example.com:443/path data:",
		"connect-src *.example.com:*",
		"frame-ancestors 'self' https://parent.example.com",
		"sandbox allow-same-origin allow-scripts",
		"report-uri /csp /csp2 https://report.example.com/endpoint",
		"report-to csp-endpoint",
		"require-sri-for script style",
		"upgrade-insecure-request",
		"webrtc allow",
		"base-uri 'self'",
		"object-src 'none'",
		"worker-src data: blob:",
		"child-src https://child.example.com",
	}, "; ")

	if got := csp.String(); got != expected {
		t.Fatalf("constructed comprehensive String() mismatch.\nexpected: %q\n     got: %q", expected, got)
	}
}
