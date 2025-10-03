package content_security_policy

import (
	"reflect"
	"testing"
)

func TestContentSecurityPolicy_String_Directives(t *testing.T) {
	tests := []struct {
		name      string
		directive DirectiveI
		want      string
	}{
		{
			name: "default-src with keyword, scheme and host",
			directive: &DefaultSrcDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "default-src", RawName: "default-src"}, Sources: []SourceI{
				&KeywordSource{Keyword: "self"},
				&SchemeSource{Scheme: "https"},
				&HostSource{Host: "cdn.example.com"},
			}}},
			want: "default-src 'self' https: cdn.example.com",
		},
		{
			name: "script-src with keywords, schemes, nonce and hash",
			directive: &ScriptSrcDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "script-src", RawName: "script-src"}, Sources: []SourceI{
				&KeywordSource{Keyword: "unsafe-inline"},
				&KeywordSource{Keyword: "unsafe-eval"},
				&KeywordSource{Keyword: "strict-dynamic"},
				&SchemeSource{Scheme: "https"},
				&SchemeSource{Scheme: "http"},
				&NonceSource{Base64Value: "dGVzdA=="},
				&HashSource{HashAlgorithm: "sha256", Base64Value: "AbCd012+/_-=="},
			}}},
			want: "script-src 'unsafe-inline' 'unsafe-eval' 'strict-dynamic' https: http: 'nonce-dGVzdA==' 'sha256-AbCd012+/_-=='",
		},
		{
			name: "style-src with report-sample and host",
			directive: &StyleSrcDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "style-src", RawName: "style-src"}, Sources: []SourceI{
				&KeywordSource{Keyword: "report-sample"},
				&HostSource{Scheme: "https", Host: "styles.example.com"},
			}}},
			want: "style-src 'report-sample' https://styles.example.com",
		},
		{
			name: "img-src with host and data scheme",
			directive: &ImgSrcDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "img-src", RawName: "img-src"}, Sources: []SourceI{
				&HostSource{Scheme: "https", Host: "example.com", PortString: "443", Path: "/path"},
				&SchemeSource{Scheme: "data"},
			}}},
			want: "img-src https://example.com:443/path data:",
		},
		{
			name: "connect-src with wildcard port",
			directive: &ConnectSrcDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "connect-src", RawName: "connect-src"}, Sources: []SourceI{
				&HostSource{Host: "*.example.com", PortString: "*"},
			}}},
			want: "connect-src *.example.com:*",
		},
		{
			name: "frame-ancestors with self and https parent",
			directive: &FrameAncestorsDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "frame-ancestors", RawName: "frame-ancestors"}, Sources: []SourceI{
				&KeywordSource{Keyword: "self"},
				&HostSource{Scheme: "https", Host: "parent.example.com"},
			}}},
			want: "frame-ancestors 'self' https://parent.example.com",
		},
		{
			name:      "sandbox with tokens",
			directive: &SandboxDirective{Directive: Directive{Name: "sandbox", RawName: "sandbox"}, Tokens: []string{"allow-same-origin", "allow-scripts"}},
			want:      "sandbox allow-same-origin allow-scripts",
		},
		{
			name:      "sandbox without tokens",
			directive: &SandboxDirective{Directive: Directive{Name: "sandbox", RawName: "sandbox"}},
			want:      "sandbox",
		},
		{
			name:      "report-uri with multiple endpoints",
			directive: &ReportUriDirective{Directive: Directive{Name: "report-uri", RawName: "report-uri"}, UriReferences: []string{"/csp", "/csp2", "https://report.example.com/endpoint"}},
			want:      "report-uri /csp /csp2 https://report.example.com/endpoint",
		},
		{
			name:      "report-to with token",
			directive: &ReportToDirective{Directive: Directive{Name: "report-to", RawName: "report-to"}, Token: "csp-endpoint"},
			want:      "report-to csp-endpoint",
		},
		{
			name:      "require-sri-for script",
			directive: &RequireSriForDirective{Directive: Directive{Name: "require-sri-for", RawName: "require-sri-for"}, ResourceTypes: []string{"script"}},
			want:      "require-sri-for script",
		},
		{
			name: "trusted-types with expressions",
			directive: &TrustedTypesDirective{Directive: Directive{Name: "trusted-types", RawName: "trusted-types"}, Expressions: []TrustedTypeExpression{
				{Kind: "policy-name", Value: "default"},
				{Kind: "policy-name", Value: "policy1"},
				{Kind: "keyword", Value: "allow-duplicates"},
				{Kind: "keyword", Value: "none"},
			}},
			want: "trusted-types default policy1 'allow-duplicates' 'none'",
		},
		{
			name:      "require-trusted-types-for script",
			directive: &RequireTrustedTypesForDirective{Directive: Directive{Name: "require-trusted-types-for", RawName: "require-trusted-types-for"}, SinkGroups: []string{"script"}},
			want:      "require-trusted-types-for 'script'",
		},
		{
			name:      "upgrade-insecure-request",
			directive: &UpgradeInsecureRequestDirective{Directive: Directive{Name: "upgrade-insecure-request", RawName: "upgrade-insecure-request"}},
			want:      "upgrade-insecure-request",
		},
		{
			name:      "webrtc allow",
			directive: &WebrtcDirective{Directive: Directive{Name: "webrtc", RawName: "webrtc"}, Value: "allow"},
			want:      "webrtc 'allow'",
		},
		{
			name: "base-uri self",
			directive: &BaseUriDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "base-uri", RawName: "base-uri"}, Sources: []SourceI{
				&KeywordSource{Keyword: "self"},
			}}},
			want: "base-uri 'self'",
		},
		{
			name: "object-src none",
			directive: &ObjectSrcDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "object-src", RawName: "object-src"}, Sources: []SourceI{
				&NoneSource{},
			}}},
			want: "object-src 'none'",
		},
		{
			name: "worker-src data and blob",
			directive: &WorkerSrcDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "worker-src", RawName: "worker-src"}, Sources: []SourceI{
				&SchemeSource{Scheme: "data"},
				&SchemeSource{Scheme: "blob"},
			}}},
			want: "worker-src data: blob:",
		},
		{
			name: "child-src https host",
			directive: &ChildSrcDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "child-src", RawName: "child-src"}, Sources: []SourceI{
				&HostSource{Scheme: "https", Host: "child.example.com"},
			}}},
			want: "child-src https://child.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csp := &ContentSecurityPolicy{Directives: []DirectiveI{tt.directive}}
			if got := csp.String(); got != tt.want {
				t.Errorf("String() got = %q, want %q", got, tt.want)
			}
		})
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

func TestContentSecurityPolicy_String_SpecificPolicy(t *testing.T) {
	expected := "default-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'; require-trusted-types-for 'script'; trusted-types lit-html"

	csp := &ContentSecurityPolicy{
		Directives: []DirectiveI{
			&DefaultSrcDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "default-src", RawName: "default-src"}, Sources: []SourceI{
				&KeywordSource{Keyword: "self"},
			}}},
			&FrameAncestorsDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "frame-ancestors", RawName: "frame-ancestors"}, Sources: []SourceI{
				&NoneSource{},
			}}},
			&BaseUriDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "base-uri", RawName: "base-uri"}, Sources: []SourceI{
				&NoneSource{},
			}}},
			&FormActionDirective{SourceDirective: SourceDirective{Directive: Directive{Name: "form-action", RawName: "form-action"}, Sources: []SourceI{
				&NoneSource{},
			}}},
			&RequireTrustedTypesForDirective{Directive: Directive{Name: "require-trusted-types-for", RawName: "require-trusted-types-for"}, SinkGroups: []string{"script"}},
			&TrustedTypesDirective{Directive: Directive{Name: "trusted-types", RawName: "trusted-types"}, Expressions: []TrustedTypeExpression{
				{Kind: "policy-name", Value: "lit-html"},
			}},
		},
	}

	if got := csp.String(); got != expected {
		to := got
		want := expected
		t.Errorf("String() got = %q, want %q", to, want)
	}
}
