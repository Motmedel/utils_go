package content_security_policy

import (
	"reflect"
	"strings"
	"testing"
)

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
			&RequireSriForDirective{Directive: Directive{Name: "require-sri-for", RawName: "require-sri-for"}, ResourceTypes: []string{"script"}},
			&TrustedTypesDirective{Directive: Directive{Name: "trusted-types", RawName: "trusted-types", RawValue: "default policy1 'allow-duplicates' 'none'"}},
			&RequireTrustedTypesForDirective{Directive: Directive{Name: "require-trusted-types-for", RawName: "require-trusted-types-for", RawValue: "'script'"}},
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
		"require-sri-for script",
		"trusted-types default policy1 'allow-duplicates' 'none'",
		"require-trusted-types-for 'script'",
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

func TestContentSecurityPolicy_GetDirective(t *testing.T) {
	type fields struct {
		Directives            []DirectiveI
		OtherDirectives       []DirectiveI
		IneffectiveDirectives []DirectiveI
		Raw                   string
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   DirectiveI
		want1  bool
	}{
		{
			name: "found in Directives",
			fields: fields{
				Directives: []DirectiveI{
					&DefaultSrcDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "default-src", RawName: "default-src"}}},
				},
			},
			args:  args{name: "default-src"},
			want:  &DefaultSrcDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "default-src", RawName: "default-src"}}},
			want1: true,
		},
		{
			name: "found in OtherDirectives",
			fields: fields{
				OtherDirectives: []DirectiveI{
					&Directive{Name: "foo", RawName: "foo", RawValue: "bar"},
				},
			},
			args:  args{name: "foo"},
			want:  &Directive{Name: "foo", RawName: "foo", RawValue: "bar"},
			want1: true,
		},
		{
			name:   "not found",
			fields: fields{},
			args:   args{name: "nope"},
			want:   nil,
			want1:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csp := &ContentSecurityPolicy{
				Directives:            tt.fields.Directives,
				OtherDirectives:       tt.fields.OtherDirectives,
				IneffectiveDirectives: tt.fields.IneffectiveDirectives,
				Raw:                   tt.fields.Raw,
			}
			got, got1 := csp.GetDirective(tt.args.name)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDirective() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetDirective() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
