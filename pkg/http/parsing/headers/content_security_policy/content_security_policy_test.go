package content_security_policy

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	contentSecurityPolicyTypes "github.com/Motmedel/utils_go/pkg/http/types/content_security_policy"
)

func findDirectiveByName(directives []contentSecurityPolicyTypes.DirectiveI, name string) contentSecurityPolicyTypes.DirectiveI {
	for _, d := range directives {
		if d.GetName() == name {
			return d
		}
	}
	return nil
}

func TestParseContentSecurityPolicy_Directives_TableDriven(t *testing.T) {
	t.Parallel()

	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		check   func(t *testing.T, csp *contentSecurityPolicyTypes.ContentSecurityPolicy)
		wantErr bool
	}{
		{
			name: "default-src 'self'",
			args: args{data: []byte("default-src 'self'")},
			check: func(t *testing.T, csp *contentSecurityPolicyTypes.ContentSecurityPolicy) {
				d := findDirectiveByName(csp.Directives, "default-src")
				if d == nil {
					t.Fatalf("default-src not found")
				}
				sd := d.(*contentSecurityPolicyTypes.DefaultSrcDirective)
				if len(sd.Sources) != 1 {
					t.Fatalf("expected 1 source, got %d", len(sd.Sources))
				}
				if _, ok := sd.Sources[0].(*contentSecurityPolicyTypes.KeywordSource); !ok {
					t.Fatalf("expected KeywordSource for 'self'")
				}
			},
		},
		{
			name: "script-src schemes nonce hash",
			args: args{data: []byte("script-src https: http: 'nonce-dGVzdA==' 'sha256-AbCd012+/_-==' 'unsafe-inline' 'unsafe-eval' 'strict-dynamic'")},
			check: func(t *testing.T, csp *contentSecurityPolicyTypes.ContentSecurityPolicy) {
				d := findDirectiveByName(csp.Directives, "script-src")
				if d == nil {
					t.Fatalf("script-src not found")
				}
				sd := d.(*contentSecurityPolicyTypes.ScriptSrcDirective)
				var schemes, nonces, hashes, keywords int
				for _, s := range sd.Sources {
					switch s.(type) {
					case *contentSecurityPolicyTypes.SchemeSource:
						schemes++
					case *contentSecurityPolicyTypes.NonceSource:
						nonces++
					case *contentSecurityPolicyTypes.HashSource:
						hashes++
					case *contentSecurityPolicyTypes.KeywordSource:
						keywords++
					}
				}
				if schemes != 2 || nonces != 1 || hashes != 1 || keywords < 1 {
					t.Fatalf("unexpected counts: schemes=%d nonces=%d hashes=%d keywords=%d", schemes, nonces, hashes, keywords)
				}
			},
		},
		{
			name: "frame-ancestors 'self' host",
			args: args{data: []byte("frame-ancestors 'self' https://parent.example.com")},
			check: func(t *testing.T, csp *contentSecurityPolicyTypes.ContentSecurityPolicy) {
				d := findDirectiveByName(csp.Directives, "frame-ancestors")
				if d == nil {
					t.Fatalf("frame-ancestors not found")
				}
				sd := d.(*contentSecurityPolicyTypes.FrameAncestorsDirective)
				if len(sd.Sources) != 2 {
					t.Fatalf("expected 2 sources, got %d", len(sd.Sources))
				}
			},
		},
		{
			name: "sandbox tokens",
			args: args{data: []byte("sandbox allow-same-origin allow-scripts")},
			check: func(t *testing.T, csp *contentSecurityPolicyTypes.ContentSecurityPolicy) {
				d := findDirectiveByName(csp.Directives, "sandbox")
				if d == nil {
					t.Fatalf("sandbox not found")
				}
				sd := d.(*contentSecurityPolicyTypes.SandboxDirective)
				if got, want := strings.Join(sd.Tokens, ","), "allow-same-origin,allow-scripts"; got != want {
					t.Fatalf("tokens = %s, want %s", got, want)
				}
			},
		},
		{
			name: "report-uri multiple",
			args: args{data: []byte("report-uri /csp /csp2 https://report.example.com/endpoint")},
			check: func(t *testing.T, csp *contentSecurityPolicyTypes.ContentSecurityPolicy) {
				d := findDirectiveByName(csp.Directives, "report-uri")
				if d == nil {
					t.Fatalf("report-uri not found")
				}
				sd := d.(*contentSecurityPolicyTypes.ReportUriDirective)
				if len(sd.UriReferences) != 3 {
					t.Fatalf("expected 3 uris, got %d", len(sd.UriReferences))
				}
			},
		},
		{
			name: "report-to token",
			args: args{data: []byte("report-to csp-endpoint")},
			check: func(t *testing.T, csp *contentSecurityPolicyTypes.ContentSecurityPolicy) {
				d := findDirectiveByName(csp.Directives, "report-to")
				if d == nil {
					t.Fatalf("report-to not found")
				}
				sd := d.(*contentSecurityPolicyTypes.ReportToDirective)
				if sd.Token != "csp-endpoint" {
					t.Fatalf("unexpected token: %q", sd.Token)
				}
			},
		},
		{
			name: "require-sri-for list",
			args: args{data: []byte("require-sri-for script style")},
			check: func(t *testing.T, csp *contentSecurityPolicyTypes.ContentSecurityPolicy) {
				d := findDirectiveByName(csp.Directives, "require-sri-for")
				if d == nil {
					t.Fatalf("require-sri-for not found")
				}
				sd := d.(*contentSecurityPolicyTypes.RequireSriForDirective)
				if !reflect.DeepEqual(sd.ResourceTypes, []string{"script", "style"}) {
					t.Fatalf("unexpected resource types: %#v", sd.ResourceTypes)
				}
			},
		},
		{
			name: "trusted-types expressions",
			args: args{data: []byte("trusted-types default policy1 'allow-duplicates' 'none'")},
			check: func(t *testing.T, csp *contentSecurityPolicyTypes.ContentSecurityPolicy) {
				d := findDirectiveByName(csp.Directives, "trusted-types")
				if d == nil {
					t.Fatalf("trusted-types not found")
				}
				sd := d.(*contentSecurityPolicyTypes.TrustedTypesDirective)
				if len(sd.Expressions) != 4 {
					t.Fatalf("expected 4 expressions, got %d", len(sd.Expressions))
				}
			},
		},
		{
			name: "require-trusted-types-for 'script'",
			args: args{data: []byte("require-trusted-types-for 'script'")},
			check: func(t *testing.T, csp *contentSecurityPolicyTypes.ContentSecurityPolicy) {
				d := findDirectiveByName(csp.Directives, "require-trusted-types-for")
				if d == nil {
					t.Fatalf("require-trusted-types-for not found")
				}
				sd := d.(*contentSecurityPolicyTypes.RequireTrustedTypesForDirective)
				if len(sd.SinkGroups) != 1 || sd.SinkGroups[0] != "script" {
					t.Fatalf("unexpected sink groups: %#v", sd.SinkGroups)
				}
			},
		},
		{
			name: "upgrade-insecure-requests",
			args: args{data: []byte("upgrade-insecure-requests")},
			check: func(t *testing.T, csp *contentSecurityPolicyTypes.ContentSecurityPolicy) {
				if findDirectiveByName(csp.Directives, "upgrade-insecure-requests") == nil {
					t.Fatalf("upgrade-insecure-requests not found")
				}
			},
		},
		{
			name: "webrtc allow",
			args: args{data: []byte("webrtc allow")},
			check: func(t *testing.T, csp *contentSecurityPolicyTypes.ContentSecurityPolicy) {
				if findDirectiveByName(csp.Directives, "webrtc") == nil {
					t.Fatalf("webrtc not found")
				}
			},
		},
		{
			name: "object-src 'none'",
			args: args{data: []byte("object-src 'none'")},
			check: func(t *testing.T, csp *contentSecurityPolicyTypes.ContentSecurityPolicy) {
				d := findDirectiveByName(csp.Directives, "object-src")
				sd := d.(*contentSecurityPolicyTypes.ObjectSrcDirective)
				if len(sd.Sources) != 1 {
					t.Fatalf("expected 1 source, got %d", len(sd.Sources))
				}
				if _, ok := sd.Sources[0].(*contentSecurityPolicyTypes.NoneSource); !ok {
					t.Fatalf("expected NoneSource")
				}
			},
		},
		{
			name: "worker-src data blob",
			args: args{data: []byte("worker-src data: blob:")},
			check: func(t *testing.T, csp *contentSecurityPolicyTypes.ContentSecurityPolicy) {
				d := findDirectiveByName(csp.Directives, "worker-src")
				sd := d.(*contentSecurityPolicyTypes.WorkerSrcDirective)
				if len(sd.Sources) != 2 {
					t.Fatalf("expected 2 sources, got %d", len(sd.Sources))
				}
			},
		},
		{
			name: "child-src host",
			args: args{data: []byte("child-src https://child.example.com")},
			check: func(t *testing.T, csp *contentSecurityPolicyTypes.ContentSecurityPolicy) {
				if findDirectiveByName(csp.Directives, "child-src") == nil {
					t.Fatalf("child-src not found")
				}
			},
		},
		{
			name: "duplicate directive ineffective",
			args: args{data: []byte("default-src 'self'; default-src https:")},
			check: func(t *testing.T, csp *contentSecurityPolicyTypes.ContentSecurityPolicy) {
				if findDirectiveByName(csp.IneffectiveDirectives, "default-src") == nil {
					t.Fatalf("expected duplicate in ineffective directives")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csp, err := ParseContentSecurityPolicy(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseContentSecurityPolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if csp == nil {
				t.Fatalf("nil csp")
			}
			if tt.check != nil {
				// keep deterministic order for any order-dependent checks
				sort.SliceStable(csp.Directives, func(i, j int) bool { return csp.Directives[i].GetName() < csp.Directives[j].GetName() })
				sort.SliceStable(csp.IneffectiveDirectives, func(i, j int) bool {
					return csp.IneffectiveDirectives[i].GetName() < csp.IneffectiveDirectives[j].GetName()
				})
				tt.check(t, csp)
			}
		})
	}
}

func TestParseContentSecurityPolicy_FullCSP_TableDriven(t *testing.T) {
	t.Parallel()
	policy := strings.Join([]string{
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
		"trusted-types default policy1 'allow-duplicates' 'none'",
		"require-trusted-types-for 'script'",
		"upgrade-insecure-request",
		"webrtc allow",
		"base-uri 'self'",
		"object-src 'none'",
		"worker-src data: blob:",
		"child-src https://child.example.com",
		// duplicate to check ineffective behavior
		"default-src https://other.example.com",
	}, "; ")

	csp, err := ParseContentSecurityPolicy([]byte(policy))
	if err != nil {
		t.Fatalf("ParseContentSecurityPolicy error: %v", err)
	}
	if csp == nil {
		t.Fatalf("expected csp, got nil")
	}
	if csp.Raw != policy {
		t.Fatalf("raw policy mismatch")
	}
	if findDirectiveByName(csp.IneffectiveDirectives, "default-src") == nil {
		t.Fatalf("expected duplicate default-src ineffective")
	}
}
