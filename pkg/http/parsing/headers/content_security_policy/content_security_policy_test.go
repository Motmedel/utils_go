package content_security_policy

import (
	"sort"
	"strings"
	"testing"

	contentSecurityPolicyTypes "github.com/Motmedel/utils_go/pkg/http/types/content_security_policy"
	goabnf "github.com/pandatix/go-abnf"
)

// helper: find directive by name
func findDirectiveByName(directives []contentSecurityPolicyTypes.DirectiveI, name string) contentSecurityPolicyTypes.DirectiveI {
	for _, d := range directives {
		if d.GetName() == name {
			return d
		}
	}
	return nil
}

func TestParseContentSecurityPolicy_ComprehensiveValidPolicies(t *testing.T) {
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

	// Raw should be preserved
	if csp.Raw != policy {
		t.Errorf("raw policy mismatch")
	}

	// First default-src should be in Directives, duplicate in Ineffective
	firstDefault := findDirectiveByName(csp.Directives, "default-src")
	if firstDefault == nil {
		t.Fatalf("default-src (first) not found among directives")
	}
	dupDefault := findDirectiveByName(csp.IneffectiveDirectives, "default-src")
	if dupDefault == nil {
		t.Fatalf("duplicate default-src not found among ineffective directives")
	}

	// Check script-src sources variety
	sd := findDirectiveByName(csp.Directives, "script-src")
	if sd == nil {
		t.Fatalf("script-src not found")
	}
	ssd, ok := sd.(*contentSecurityPolicyTypes.ScriptSrcDirective)
	if !ok {
		t.Fatalf("script-src has wrong type: %T", sd)
	}
	// Expect at least: two schemes, three keywords (unsafe-inline, unsafe-eval, strict-dynamic), a nonce, a hash
	var (
		schemes, keywords, nonces, hashes, hosts int
	)
	for _, s := range ssd.Sources {
		switch s := s.(type) {
		case *contentSecurityPolicyTypes.SchemeSource:
			schemes++
			if s.Scheme != "https" && s.Scheme != "http" {
				t.Errorf("unexpected scheme in script-src: %q", s.Scheme)
			}
		case *contentSecurityPolicyTypes.KeywordSource:
			keywords++
		case *contentSecurityPolicyTypes.NonceSource:
			nonces++
			if s.Base64Value != "dGVzdA==" {
				t.Errorf("unexpected nonce value: %q", s.Base64Value)
			}
		case *contentSecurityPolicyTypes.HashSource:
			hashes++
			if s.HashAlgorithm != "sha256" {
				t.Errorf("unexpected hash alg: %q", s.HashAlgorithm)
			}
		case *contentSecurityPolicyTypes.HostSource:
			hosts++ // shouldn't appear here, but allow future-proofing
		}
	}
	if schemes < 2 || keywords < 3 || nonces != 1 || hashes != 1 {
		t.Errorf("unexpected script-src source counts: schemes=%d keywords=%d nonces=%d hashes=%d", schemes, keywords, nonces, hashes)
	}

	// style-src contains report-sample keyword and a host
	std := findDirectiveByName(csp.Directives, "style-src")
	if std == nil {
		t.Fatalf("style-src not found")
	}
	styles, ok := std.(*contentSecurityPolicyTypes.StyleSrcDirective)
	if !ok {
		t.Fatalf("style-src type mismatch: %T", std)
	}
	foundReportSample := false
	foundHost := false
	for _, s := range styles.Sources {
		switch ss := s.(type) {
		case *contentSecurityPolicyTypes.KeywordSource:
			if ss.Keyword == "'report-sample'" {
				foundReportSample = true
			}
		case *contentSecurityPolicyTypes.HostSource:
			foundHost = true
		}
	}
	if !foundReportSample || !foundHost {
		t.Errorf("style-src missing expected sources: report-sample=%v host=%v", foundReportSample, foundHost)
	}

	// img-src with absolute, relative, and data:
	imgd := findDirectiveByName(csp.Directives, "img-src")
	if imgd == nil {
		t.Fatalf("img-src not found")
	}
	img, ok := imgd.(*contentSecurityPolicyTypes.ImgSrcDirective)
	if !ok {
		t.Fatalf("img-src type mismatch: %T", imgd)
	}
	var haveAbsHost, haveDataScheme bool
	for _, s := range img.Sources {
		switch ss := s.(type) {
		case *contentSecurityPolicyTypes.HostSource:
			haveAbsHost = true
			if ss.Host != "example.com" || ss.PortString != "443" || ss.Path != "/path" {
				t.Errorf("unexpected host source parts: %+v", ss)
			}
		case *contentSecurityPolicyTypes.SchemeSource:
			if ss.Scheme == "data" {
				haveDataScheme = true
			}
		case *contentSecurityPolicyTypes.KeywordSource:
			// none
		case *contentSecurityPolicyTypes.Source:
			_ = ss
		}
	}
	// relative path becomes host-source with empty host? Not supported by grammar (img-src includes uri schemes and host sources). We don't assert relative.
	if !haveAbsHost || !haveDataScheme {
		t.Errorf("img-src missing expected sources: absHost=%v data=%v", haveAbsHost, haveDataScheme)
	}

	// connect-src wildcard host and port
	cond := findDirectiveByName(csp.Directives, "connect-src")
	if cond == nil {
		t.Fatalf("connect-src not found")
	}
	con, ok := cond.(*contentSecurityPolicyTypes.ConnectSrcDirective)
	if !ok {
		t.Fatalf("connect-src type mismatch: %T", cond)
	}
	foundWildcard := false
	for _, s := range con.Sources {
		if hs, ok := s.(*contentSecurityPolicyTypes.HostSource); ok {
			if hs.Host == "*.example.com" && hs.PortString == "*" {
				foundWildcard = true
			}
		}
	}
	if !foundWildcard {
		t.Errorf("connect-src missing wildcard host:port source")
	}

	// frame-ancestors with self and host
	fad := findDirectiveByName(csp.Directives, "frame-ancestors")
	if fad == nil {
		t.Fatalf("frame-ancestors not found")
	}
	fa, ok := fad.(*contentSecurityPolicyTypes.FrameAncestorsDirective)
	if !ok {
		t.Fatalf("frame-ancestors type mismatch: %T", fad)
	}
	var haveSelf, haveParent bool
	for _, s := range fa.Sources {
		switch ss := s.(type) {
		case *contentSecurityPolicyTypes.KeywordSource:
			if ss.Keyword == "'self'" {
				haveSelf = true
			}
		case *contentSecurityPolicyTypes.HostSource:
			if ss.Host == "parent.example.com" {
				haveParent = true
			}
		}
	}
	if !haveSelf || !haveParent {
		t.Errorf("frame-ancestors missing expected sources: self=%v parent=%v", haveSelf, haveParent)
	}

	// sandbox tokens
	sbd := findDirectiveByName(csp.Directives, "sandbox")
	if sbd == nil {
		t.Fatalf("sandbox not found")
	}
	sb, ok := sbd.(*contentSecurityPolicyTypes.SandboxDirective)
	if !ok {
		t.Fatalf("sandbox type mismatch: %T", sbd)
	}
	if len(sb.Tokens) != 2 {
		t.Fatalf("sandbox tokens length = %d, want 2", len(sb.Tokens))
	}
	sort.Strings(sb.Tokens)
	if sb.Tokens[0] != "allow-same-origin" || sb.Tokens[1] != "allow-scripts" {
		t.Errorf("unexpected sandbox tokens: %v", sb.Tokens)
	}

	// report-uri list
	rud := findDirectiveByName(csp.Directives, "report-uri")
	if rud == nil {
		t.Fatalf("report-uri not found")
	}
	ru, ok := rud.(*contentSecurityPolicyTypes.ReportUriDirective)
	if !ok {
		t.Fatalf("report-uri type mismatch: %T", rud)
	}
	if len(ru.UriReferences) != 3 || ru.UriReferences[0] != "/csp" {
		t.Errorf("unexpected report-uri values: %v", ru.UriReferences)
	}

	// report-to token
	rtd := findDirectiveByName(csp.Directives, "report-to")
	if rtd == nil {
		t.Fatalf("report-to not found")
	}
	rt, ok := rtd.(*contentSecurityPolicyTypes.ReportToDirective)
	if !ok {
		t.Fatalf("report-to type mismatch: %T", rtd)
	}
	if rt.Token == "" || rt.Token != "csp-endpoint" {
		t.Errorf("unexpected report-to token: %q", rt.Token)
	}

	// require-sri-for parsed and lowercased
	rsd := findDirectiveByName(csp.Directives, "require-sri-for")
	if rsd == nil {
		t.Fatalf("require-sri-for not found")
	}
	rs, ok := rsd.(*contentSecurityPolicyTypes.RequireSriForDirective)
	if !ok {
		t.Fatalf("require-sri-for type mismatch: %T", rsd)
	}
	if len(rs.ResourceTypes) != 2 || rs.ResourceTypes[0] != "script" || rs.ResourceTypes[1] != "style" {
		t.Errorf("unexpected require-sri-for resources: %v", rs.ResourceTypes)
	}

	// trusted-types should be parsed with expressions; keywords have no quotes
	ttdDir := findDirectiveByName(csp.Directives, "trusted-types")
	if ttdDir == nil {
		t.Fatalf("trusted-types not found among directives")
	}
	ttd, ok := ttdDir.(*contentSecurityPolicyTypes.TrustedTypesDirective)
	if !ok {
		t.Fatalf("trusted-types type mismatch: %T", ttdDir)
	}
	if ttdDir.GetRawValue() != "default policy1 'allow-duplicates' 'none'" {
		t.Errorf("unexpected trusted-types raw value: %q", ttdDir.GetRawValue())
	}
	var haveDefault, havePolicy1, haveAllowDup, haveNone bool
	for _, expr := range ttd.Expressions {
		switch expr.Kind {
		case "policy-name":
			if expr.Value == "default" {
				haveDefault = true
			}
			if expr.Value == "policy1" {
				havePolicy1 = true
			}
		case "keyword":
			if expr.Value == "allow-duplicates" {
				haveAllowDup = true
			}
			if expr.Value == "none" {
				haveNone = true
			}
		}
	}
	if !(haveDefault && havePolicy1 && haveAllowDup && haveNone) {
		t.Errorf("trusted-types expressions missing: default=%v policy1=%v allow-duplicates=%v none=%v", haveDefault, havePolicy1, haveAllowDup, haveNone)
	}

	// require-trusted-types-for parsed and preserved raw value
	rttfDir := findDirectiveByName(csp.Directives, "require-trusted-types-for")
	if rttfDir == nil {
		t.Fatalf("require-trusted-types-for not found among directives")
	}
	if rttfDir.GetRawValue() != "'script'" {
		t.Errorf("unexpected require-trusted-types-for raw value: %q", rttfDir.GetRawValue())
	}

	// upgrade-insecure-request
	uid := findDirectiveByName(csp.Directives, "upgrade-insecure-request")
	if uid == nil {
		t.Errorf("upgrade-insecure-request not found")
	}

	// webrtc allow
	wd := findDirectiveByName(csp.Directives, "webrtc")
	if wd == nil {
		t.Fatalf("webrtc not found")
	}
	_, ok = wd.(*contentSecurityPolicyTypes.WebrtcDirective)
	if !ok {
		t.Fatalf("webrtc type mismatch: %T", wd)
	}

	// object-src none
	od := findDirectiveByName(csp.Directives, "object-src")
	if od == nil {
		t.Fatalf("object-src not found")
	}
	osd, ok := od.(*contentSecurityPolicyTypes.ObjectSrcDirective)
	if !ok {
		t.Fatalf("object-src type mismatch: %T", od)
	}
	if len(osd.Sources) != 1 {
		t.Fatalf("object-src sources length = %d, want 1", len(osd.Sources))
	}
	if _, ok := osd.Sources[0].(*contentSecurityPolicyTypes.NoneSource); !ok {
		t.Errorf("object-src expected 'none' source, got %T", osd.Sources[0])
	}
}

func Test_makeSourcesFromPaths(t *testing.T) {
	t.Parallel()

	value := "https: http: 'self' 'unsafe-inline' 'nonce-dGVzdA==' 'sha384-AAAABBBBCCCCDDDD' example.com:8443/path https://sub.example.com"
	paths, err := goabnf.Parse([]byte(value), ContentSecurityPolicyGrammar, "serialized-source-list")
	if err != nil {
		t.Fatalf("goabnf.Parse error: %v", err)
	}
	if len(paths) == 0 {
		t.Fatalf("no paths returned")
	}

	sources, err := makeSourcesFromPaths([]byte(value), paths, "source-expression")
	if err != nil {
		t.Fatalf("makeSourcesFromPaths error: %v", err)
	}
	if len(sources) == 0 {
		t.Fatalf("no sources parsed")
	}

	var haveHTTP, haveHTTPS, haveSelf, haveUnsafeInline, haveNonce, haveHash, haveHost1, haveHost2 bool
	for _, s := range sources {
		switch ss := s.(type) {
		case *contentSecurityPolicyTypes.SchemeSource:
			if ss.Scheme == "http" {
				haveHTTP = true
			} else if ss.Scheme == "https" {
				haveHTTPS = true
			}
		case *contentSecurityPolicyTypes.KeywordSource:
			if ss.Keyword == "'self'" {
				haveSelf = true
			}
			if ss.Keyword == "'unsafe-inline'" {
				haveUnsafeInline = true
			}
		case *contentSecurityPolicyTypes.NonceSource:
			haveNonce = ss.Base64Value == "dGVzdA=="
		case *contentSecurityPolicyTypes.HashSource:
			haveHash = ss.HashAlgorithm == "sha384"
		case *contentSecurityPolicyTypes.HostSource:
			if ss.Host == "example.com" && ss.PortString == "8443" && ss.Path == "/path" {
				haveHost1 = true
			}
			if ss.Scheme == "https" && ss.Host == "sub.example.com" {
				haveHost2 = true
			}
		}
	}
	if !(haveHTTP && haveHTTPS && haveSelf && haveUnsafeInline && haveNonce && haveHash && haveHost1 && haveHost2) {
		t.Errorf("missing some sources: http=%v https=%v self=%v inline=%v nonce=%v hash=%v host1=%v host2=%v", haveHTTP, haveHTTPS, haveSelf, haveUnsafeInline, haveNonce, haveHash, haveHost1, haveHost2)
	}
}

func TestParseContentSecurityPolicy_AdditionalValidCoverage(t *testing.T) {
	t.Parallel()

	policy := strings.Join([]string{
		"script-src-attr 'unsafe-hashes' 'nonce-QUJD' 'sha384-AAAA' https:",
		"script-src-elem https: 'strict-dynamic'",
		"font-src https://fonts.example.com",
		"form-action 'self' https://forms.example.com",
		"frame-src https://frames.example.com",
		"media-src 'none'",
		"manifest-src https://app.example.com",
		"base-uri https://base.example.com",
		"worker-src blob: data:",
		"child-src *.cdn.example.com:* https://child.example.com/path%20with%20spaces",
		"frame-ancestors 'none'",
		"sandbox allow-popups",
		"report-uri /a https://example.com/r",
		"report-to reporting-endpoint",
	}, "; ")

	csp, err := ParseContentSecurityPolicy([]byte(policy))
	if err != nil {
		t.Fatalf("ParseContentSecurityPolicy error: %v", err)
	}
	if csp == nil {
		t.Fatalf("expected csp, got nil")
	}

	// script-src-attr
	a := findDirectiveByName(csp.Directives, "script-src-attr")
	if a == nil {
		t.Fatalf("script-src-attr not found")
	}
	attr, ok := a.(*contentSecurityPolicyTypes.ScriptSrcAttrDirective)
	if !ok {
		t.Fatalf("script-src-attr type mismatch: %T", a)
	}
	var haveUH, haveNonce, haveHash, haveHTTPS bool
	for _, s := range attr.Sources {
		switch ss := s.(type) {
		case *contentSecurityPolicyTypes.KeywordSource:
			if ss.Keyword == "'unsafe-hashes'" {
				haveUH = true
			}
		case *contentSecurityPolicyTypes.NonceSource:
			haveNonce = ss.Base64Value == "QUJD" // "ABC"
		case *contentSecurityPolicyTypes.HashSource:
			haveHash = ss.HashAlgorithm == "sha384"
		case *contentSecurityPolicyTypes.SchemeSource:
			if ss.Scheme == "https" {
				haveHTTPS = true
			}
		}
	}
	if !(haveUH && haveNonce && haveHash && haveHTTPS) {
		t.Errorf("script-src-attr missing expected sources: unsafe-hashes=%v nonce=%v hash=%v https=%v", haveUH, haveNonce, haveHash, haveHTTPS)
	}

	// script-src-elem
	e := findDirectiveByName(csp.Directives, "script-src-elem")
	if e == nil {
		t.Fatalf("script-src-elem not found")
	}
	elem := e.(*contentSecurityPolicyTypes.ScriptSrcElemDirective)
	var elemHTTPS, elemStrictDynamic bool
	for _, s := range elem.Sources {
		switch ss := s.(type) {
		case *contentSecurityPolicyTypes.SchemeSource:
			if ss.Scheme == "https" {
				elemHTTPS = true
			}
		case *contentSecurityPolicyTypes.KeywordSource:
			if ss.Keyword == "'strict-dynamic'" {
				elemStrictDynamic = true
			}
		}
	}
	if !(elemHTTPS && elemStrictDynamic) {
		t.Errorf("script-src-elem missing expected sources: https=%v strict-dynamic=%v", elemHTTPS, elemStrictDynamic)
	}

	// font-src host
	fd := findDirectiveByName(csp.Directives, "font-src")
	if fd == nil {
		t.Fatalf("font-src not found")
	}
	if _, ok := fd.(*contentSecurityPolicyTypes.FontSrcDirective); !ok {
		t.Fatalf("font-src type mismatch: %T", fd)
	}

	// form-action with self and absolute host
	fad := findDirectiveByName(csp.Directives, "form-action")
	if fad == nil {
		t.Fatalf("form-action not found")
	}
	fa := fad.(*contentSecurityPolicyTypes.FormActionDirective)
	var haveFAself, haveFAhost bool
	for _, s := range fa.Sources {
		switch ss := s.(type) {
		case *contentSecurityPolicyTypes.KeywordSource:
			if ss.Keyword == "'self'" {
				haveFAself = true
			}
		case *contentSecurityPolicyTypes.HostSource:
			if ss.Scheme == "https" && ss.Host == "forms.example.com" {
				haveFAhost = true
			}
		}
	}
	if !(haveFAself && haveFAhost) {
		t.Errorf("form-action missing expected sources: self=%v host=%v", haveFAself, haveFAhost)
	}

	// frame-src host
	frs := findDirectiveByName(csp.Directives, "frame-src")
	if frs == nil {
		t.Fatalf("frame-src not found")
	}
	if _, ok := frs.(*contentSecurityPolicyTypes.FrameSrcDirective); !ok {
		t.Fatalf("frame-src type mismatch: %T", frs)
	}

	// media-src 'none'
	md := findDirectiveByName(csp.Directives, "media-src")
	if md == nil {
		t.Fatalf("media-src not found")
	}
	ms := md.(*contentSecurityPolicyTypes.MediaSrcDirective)
	if len(ms.Sources) != 1 {
		t.Fatalf("media-src expected 1 source got %d", len(ms.Sources))
	}
	if _, ok := ms.Sources[0].(*contentSecurityPolicyTypes.NoneSource); !ok {
		t.Errorf("media-src expected none source, got %T", ms.Sources[0])
	}

	// manifest-src host
	mfd := findDirectiveByName(csp.Directives, "manifest-src")
	if mfd == nil {
		t.Fatalf("manifest-src not found")
	}
	if _, ok := mfd.(*contentSecurityPolicyTypes.ManifestSrcDirective); !ok {
		t.Fatalf("manifest-src type mismatch: %T", mfd)
	}

	// base-uri absolute host
	bud := findDirectiveByName(csp.Directives, "base-uri")
	if bud == nil {
		t.Fatalf("base-uri not found")
	}
	bu := bud.(*contentSecurityPolicyTypes.BaseUriDirective)
	foundBase := false
	for _, s := range bu.Sources {
		if hs, ok := s.(*contentSecurityPolicyTypes.HostSource); ok {
			if hs.Scheme == "https" && hs.Host == "base.example.com" {
				foundBase = true
			}
		}
	}
	if !foundBase {
		t.Errorf("base-uri missing expected https host")
	}

	// worker-src data and blob schemes
	wsd := findDirectiveByName(csp.Directives, "worker-src")
	if wsd == nil {
		t.Fatalf("worker-src not found")
	}
	ws := wsd.(*contentSecurityPolicyTypes.WorkerSrcDirective)
	var haveData, haveBlob bool
	for _, s := range ws.Sources {
		if sc, ok := s.(*contentSecurityPolicyTypes.SchemeSource); ok {
			if sc.Scheme == "data" {
				haveData = true
			}
			if sc.Scheme == "blob" {
				haveBlob = true
			}
		}
	}
	if !(haveData && haveBlob) {
		t.Errorf("worker-src missing data/blob schemes: data=%v blob=%v", haveData, haveBlob)
	}

	// child-src wildcard host with wildcard port and percent-encoded path
	csd := findDirectiveByName(csp.Directives, "child-src")
	if csd == nil {
		t.Fatalf("child-src not found")
	}
	cs := csd.(*contentSecurityPolicyTypes.ChildSrcDirective)
	var haveWild, haveEncoded bool
	for _, s := range cs.Sources {
		if hs, ok := s.(*contentSecurityPolicyTypes.HostSource); ok {
			if hs.Host == "*.cdn.example.com" && hs.PortString == "*" {
				haveWild = true
			}
			if hs.Scheme == "https" && hs.Host == "child.example.com" && hs.Path == "/path%20with%20spaces" {
				haveEncoded = true
			}
		}
	}
	if !(haveWild && haveEncoded) {
		t.Errorf("child-src missing wildcard or encoded path: wild=%v encoded=%v", haveWild, haveEncoded)
	}

	// frame-ancestors 'none'
	fand := findDirectiveByName(csp.Directives, "frame-ancestors")
	if fand == nil {
		t.Fatalf("frame-ancestors not found")
	}
	fa2 := fand.(*contentSecurityPolicyTypes.FrameAncestorsDirective)
	if len(fa2.Sources) != 1 {
		t.Fatalf("frame-ancestors expected 1 source got %d", len(fa2.Sources))
	}
	if _, ok := fa2.Sources[0].(*contentSecurityPolicyTypes.NoneSource); !ok {
		t.Errorf("frame-ancestors expected none source, got %T", fa2.Sources[0])
	}
}

func TestParseContentSecurityPolicy_UnknownAndDuplicates(t *testing.T) {
	t.Parallel()

	policy := strings.Join([]string{
		"unknown-directive foo",
		"default-src 'self'",
		"unknown-directive bar", // duplicate should be ineffective
	}, "; ")

	csp, err := ParseContentSecurityPolicy([]byte(policy))
	if err != nil {
		t.Fatalf("ParseContentSecurityPolicy error: %v", err)
	}

	// unknown-directive should appear once in Other and once in Ineffective
	var otherCount, ineffectiveCount int
	for _, d := range csp.OtherDirectives {
		if d.GetName() == "unknown-directive" {
			otherCount++
		}
	}
	for _, d := range csp.IneffectiveDirectives {
		if d.GetName() == "unknown-directive" {
			ineffectiveCount++
		}
	}
	if otherCount != 1 || ineffectiveCount != 1 {
		t.Errorf("unexpected counts for unknown-directive: other=%d ineffective=%d", otherCount, ineffectiveCount)
	}

	// default-src should be in main directives and normalized to lowercase name
	d := findDirectiveByName(csp.Directives, "default-src")
	if d == nil {
		t.Fatalf("default-src not found in directives")
	}
	if d.GetRawName() != "default-src" {
		t.Errorf("raw name mismatch: %q", d.GetRawName())
	}
}

func Test_makeSourcesFromPaths_MoreSchemes(t *testing.T) {
	value := "ws: wss: blob: data:"
	paths, err := goabnf.Parse([]byte(value), ContentSecurityPolicyGrammar, "serialized-source-list")
	if err != nil {
		t.Fatalf("goabnf.Parse error: %v", err)
	}
	sources, err := makeSourcesFromPaths([]byte(value), paths, "source-expression")
	if err != nil {
		t.Fatalf("makeSourcesFromPaths error: %v", err)
	}
	want := map[string]bool{"ws": false, "wss": false, "blob": false, "data": false}
	for _, s := range sources {
		if sc, ok := s.(*contentSecurityPolicyTypes.SchemeSource); ok {
			if _, ok := want[sc.Scheme]; ok {
				want[sc.Scheme] = true
			}
		}
	}
	for k, v := range want {
		if !v {
			t.Errorf("missing scheme %s", k)
		}
	}
}
