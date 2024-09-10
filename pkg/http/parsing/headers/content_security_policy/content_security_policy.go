package content_security_policy

import (
	"bytes"
	"context"
	"github.com/Motmedel/parsing_utils/parsing_utils"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	contentSecurityPolicyTypes "github.com/Motmedel/utils_go/pkg/http/types/content_security_policy"
	"github.com/Motmedel/utils_go/pkg/log"
	motmedelLog "github.com/Motmedel/utils_go/pkg/log"
	goabnf "github.com/pandatix/go-abnf"
	"strings"
)

var sourceListDirectiveNames = map[string]struct{}{
	"base-uri":        {},
	"child-src":       {},
	"connect-src":     {},
	"default-src":     {},
	"font-src":        {},
	"form-action":     {},
	"frame-src":       {},
	"img-src":         {},
	"manifest-src":    {},
	"media-src":       {},
	"object-src":      {},
	"script-src":      {},
	"script-src-attr": {},
	"script-src-elem": {},
	"style-src":       {},
	"style-src-attr":  {},
	"style-src-elem":  {},
	"worker-src":      {},
}

var ContentSecurityPolicyGrammar *goabnf.Grammar

func makeSourcesFromPaths(
	ctx context.Context,
	data []byte,
	paths []*goabnf.Path,
	ruleName string,
) []contentSecurityPolicyTypes.SourceI {
	sourceExpressionPaths := parsing_utils.SearchPath(paths[0], []string{ruleName}, 2, false)
	if len(sourceExpressionPaths) == 0 {
		return nil
	}

	var sources []contentSecurityPolicyTypes.SourceI

	for _, sourceExpressionPath := range sourceExpressionPaths {
		concreteSourcePath := sourceExpressionPath.Subpaths[0]

		innerSource := contentSecurityPolicyTypes.Source{Raw: string(parsing_utils.ExtractPathValue(data, concreteSourcePath))}

		matchRuleName := concreteSourcePath.MatchRule
		switch matchRuleName {
		case "scheme-source":
			sources = append(
				sources,
				&contentSecurityPolicyTypes.SchemeSource{
					Source: innerSource,
					Scheme: string(parsing_utils.ExtractPathValue(data, concreteSourcePath.Subpaths[0])),
				},
			)
		case "host-source":
			hostSource := &contentSecurityPolicyTypes.HostSource{Source: innerSource}

			schemePartPath := parsing_utils.SearchPathSingleName(
				concreteSourcePath,
				"scheme-part",
				1,
				false,
			)
			if schemePartPath != nil {
				hostSource.Scheme = string(parsing_utils.ExtractPathValue(data, schemePartPath))
			}

			hostPartPath := parsing_utils.SearchPathSingleName(
				concreteSourcePath,
				"host-part",
				1,
				false,
			)
			if hostPartPath == nil {
				panic("missing")
			}

			hostSource.Host = string(parsing_utils.ExtractPathValue(data, hostPartPath))

			portPartPath := parsing_utils.SearchPathSingleName(
				concreteSourcePath,
				"port-part",
				1,
				false,
			)
			if portPartPath != nil {
				hostSource.PortString = string(parsing_utils.ExtractPathValue(data, portPartPath))
			}

			pathPartPath := parsing_utils.SearchPathSingleName(
				concreteSourcePath,
				"path-part",
				1,
				false,
			)
			if pathPartPath != nil {
				hostSource.Path = string(parsing_utils.ExtractPathValue(data, pathPartPath))
			}

			sources = append(sources, hostSource)
		case "keyword-source":
			sources = append(
				sources,
				&contentSecurityPolicyTypes.KeywordSource{
					Source:  innerSource,
					Keyword: string(parsing_utils.ExtractPathValue(data, concreteSourcePath.Subpaths[0])),
				},
			)
		case "nonce-source":
			sources = append(
				sources,
				&contentSecurityPolicyTypes.NonceSource{
					Source:      innerSource,
					Base64Value: string(parsing_utils.ExtractPathValue(data, concreteSourcePath.Subpaths[1])),
				},
			)
		case "hash-source":
			sources = append(
				sources,
				&contentSecurityPolicyTypes.HashSource{
					Source:        innerSource,
					HashAlgorithm: string(parsing_utils.ExtractPathValue(data, concreteSourcePath.Subpaths[1])),
					Base64Value:   string(parsing_utils.ExtractPathValue(data, concreteSourcePath.Subpaths[3])),
				},
			)
		default:
			motmedelLog.LogWarning(
				"An unexpected source rule was observed.",
				&motmedelErrors.InputError{
					Message: "An unexpected source rule was observed.",
					Input:   matchRuleName,
				},
				log.GetLoggerFromCtxWithDefault(ctx, nil),
			)
		}
	}

	return sources
}

func ParseContentSecurityPolicy(ctx context.Context, data []byte) (*contentSecurityPolicyTypes.ContentSecurityPolicy, error) {
	paths, err := goabnf.Parse(data, ContentSecurityPolicyGrammar, "root")
	if err != nil {
		return nil, &motmedelErrors.InputError{
			Message: "An error occurred when parsing data as a content security policy.",
			Cause:   err,
			Input:   data,
		}
	}
	if len(paths) == 0 {
		return nil, nil
	}

	directiveNameSet := make(map[string]struct{})

	contentSecurityPolicy := &contentSecurityPolicyTypes.ContentSecurityPolicy{}
	contentSecurityPolicy.Raw = string(data)

	interestingPaths := parsing_utils.SearchPath(paths[0], []string{"serialized-directive"}, 2, false)
	for _, interestingPath := range interestingPaths {
		directiveNamePath := parsing_utils.SearchPathSingleName(
			interestingPath,
			"directive-name",
			1,
			false,
		)
		if directiveNamePath == nil {
			return nil, &motmedelErrors.InputError{
				Message: "No directive name could be found in an interesting path.",
				Input:   interestingPath,
			}
		}
		directiveName := string(parsing_utils.ExtractPathValue(data, directiveNamePath))

		var directiveValue []byte
		directiveValuePath := parsing_utils.SearchPathSingleName(
			interestingPath,
			"directive-value",
			1,
			false,
		)
		if directiveValuePath != nil {
			directiveValue = parsing_utils.ExtractPathValue(data, directiveValuePath)
		}

		isOtherDirective := false
		isIneffectiveDirective := false

		lowercaseDirectiveName := strings.ToLower(directiveName)
		if _, ok := directiveNameSet[lowercaseDirectiveName]; ok {
			isIneffectiveDirective = true
		}
		directiveNameSet[lowercaseDirectiveName] = struct{}{}

		var directive contentSecurityPolicyTypes.DirectiveI
		var sources []contentSecurityPolicyTypes.SourceI

		if _, ok := sourceListDirectiveNames[lowercaseDirectiveName]; ok {
			if bytes.Equal(directiveValue, []byte("'none'")) {
				sources = []contentSecurityPolicyTypes.SourceI{
					&contentSecurityPolicyTypes.NoneSource{
						Source: contentSecurityPolicyTypes.Source{Raw: string(parsing_utils.ExtractPathValue(data, directiveValuePath))},
					},
				}
			} else {
				serializedSourceListPaths, err := goabnf.Parse(directiveValue, ContentSecurityPolicyGrammar, "serialized-source-list")
				if err != nil {
					return nil, &motmedelErrors.InputError{
						Message: "An error occurred when parsing a serialized-source-list.",
						Cause:   err,
						Input:   directiveValue,
					}
				}
				if len(serializedSourceListPaths) == 0 {
					return nil, nil
				}

				sources = makeSourcesFromPaths(
					ctx,
					directiveValue,
					serializedSourceListPaths,
					"source-expression",
				)
				if sources == nil {
					return nil, nil
				}
			}
		}

		innerDirective := contentSecurityPolicyTypes.Directive{
			Name:     lowercaseDirectiveName,
			RawName:  directiveName,
			RawValue: string(directiveValue),
		}
		sourceDirective := contentSecurityPolicyTypes.SourceDirective{Directive: innerDirective, Sources: sources}

		switch lowercaseDirectiveName {
		case "base-uri":
			directive = &contentSecurityPolicyTypes.BaseUriDirective{SourceDirective: sourceDirective}
		case "child-src":
			directive = &contentSecurityPolicyTypes.ChildSrcDirective{SourceDirective: sourceDirective}
		case "connect-src":
			directive = &contentSecurityPolicyTypes.ConnectSrcDirective{SourceDirective: sourceDirective}
		case "default-src":
			directive = &contentSecurityPolicyTypes.DefaultSrcDirective{SourceDirective: sourceDirective}
		case "font-src":
			directive = &contentSecurityPolicyTypes.FontSrcDirective{SourceDirective: sourceDirective}
		case "form-action":
			directive = &contentSecurityPolicyTypes.FormActionDirective{SourceDirective: sourceDirective}
		case "frame-src":
			directive = &contentSecurityPolicyTypes.FrameSrcDirective{SourceDirective: sourceDirective}
		case "img-src":
			directive = &contentSecurityPolicyTypes.ImgSrcDirective{SourceDirective: sourceDirective}
		case "manifest-src":
			directive = &contentSecurityPolicyTypes.ManifestSrcDirective{SourceDirective: sourceDirective}
		case "media-src":
			directive = &contentSecurityPolicyTypes.MediaSrcDirective{SourceDirective: sourceDirective}
		case "object-src":
			directive = &contentSecurityPolicyTypes.ObjectSrcDirective{SourceDirective: sourceDirective}
		case "script-src":
			directive = &contentSecurityPolicyTypes.ScriptSrcDirective{SourceDirective: sourceDirective}
		case "script-src-attr":
			directive = &contentSecurityPolicyTypes.ScriptSrcAttrDirective{SourceDirective: sourceDirective}
		case "script-src-elem":
			directive = &contentSecurityPolicyTypes.ScriptSrcElemDirective{SourceDirective: sourceDirective}
		case "style-src":
			directive = &contentSecurityPolicyTypes.StyleSrcDirective{SourceDirective: sourceDirective}
		case "style-src-attr":
			directive = &contentSecurityPolicyTypes.StyleSrcAttrDirective{SourceDirective: sourceDirective}
		case "style-src-elem":
			directive = &contentSecurityPolicyTypes.StyleSrcElemDirective{SourceDirective: sourceDirective}
		case "worker-src":
			directive = &contentSecurityPolicyTypes.WorkerSrcDirective{SourceDirective: sourceDirective}
		case "sandbox":
			sandboxDirective := &contentSecurityPolicyTypes.SandboxDirective{Directive: innerDirective}

			sandboxDirectiveValuePaths, err := goabnf.Parse(
				directiveValue,
				ContentSecurityPolicyGrammar,
				"sandbox-directive-value-root",
			)
			if err != nil {
				return nil, &motmedelErrors.InputError{
					Message: "An error occurred when parsing data as a sandbox directive value.",
					Cause:   err,
					Input:   directiveValue,
				}
			}
			if len(sandboxDirectiveValuePaths) == 0 {
				return nil, nil
			}

			tokenPaths := parsing_utils.SearchPath(directiveValuePath, []string{"token"}, 2, false)
			for _, tokenPath := range tokenPaths {
				sandboxDirective.Tokens = append(
					sandboxDirective.Tokens,
					string(parsing_utils.ExtractPathValue(data, tokenPath)),
				)
			}
			directive = sandboxDirective
		case "webrtc":
			if innerDirective.RawValue != "allow" && innerDirective.RawValue != "block" {
				return nil, nil
			}
			webrtcDirective := &contentSecurityPolicyTypes.WebrtcDirective{Directive: innerDirective}
			directive = webrtcDirective
		case "report-uri":
			reportUriDirective := &contentSecurityPolicyTypes.ReportUriDirective{Directive: innerDirective}

			reportUriDirectivePaths, err := goabnf.Parse(directiveValue, ContentSecurityPolicyGrammar, "report-uri-directive-value-root")
			if err != nil {
				return nil, &motmedelErrors.InputError{
					Message: "An error occurred when parsing a report-uri directive value.",
					Cause:   err,
					Input:   directiveValue,
				}
			}
			if len(reportUriDirectivePaths) == 0 {
				return nil, nil
			}

			reportUriDirectivePath := reportUriDirectivePaths[0]
			uriReferencePaths := parsing_utils.SearchPath(
				reportUriDirectivePath,
				[]string{"uri-reference"},
				1,
				false,
			)
			if len(uriReferencePaths) == 0 {
				return nil, nil
			}

			for _, uriReferencePath := range uriReferencePaths {
				reportUriDirective.UriReferences = append(
					reportUriDirective.UriReferences,
					string(parsing_utils.ExtractPathValue(directiveValue, uriReferencePath)),
				)
			}
			directive = reportUriDirective
		case "frame-ancestors":
			frameAncestorsDirective := &contentSecurityPolicyTypes.FrameAncestorsDirective{SourceDirective: sourceDirective}
			if bytes.Equal(directiveValue, []byte("'none'")) {
				sources = []contentSecurityPolicyTypes.SourceI{
					&contentSecurityPolicyTypes.NoneSource{
						Source: contentSecurityPolicyTypes.Source{Raw: string(parsing_utils.ExtractPathValue(data, directiveValuePath))},
					},
				}
			} else if bytes.Equal(directiveValue, []byte("'self'")) {
				sources = []contentSecurityPolicyTypes.SourceI{
					&contentSecurityPolicyTypes.KeywordSource{
						Source:  contentSecurityPolicyTypes.Source{Raw: string(parsing_utils.ExtractPathValue(data, directiveValuePath))},
						Keyword: "'self'",
					},
				}
			} else {
				ancestorSourceListPaths, err := goabnf.Parse(directiveValue, ContentSecurityPolicyGrammar, "ancestor-source-list-root")
				if err != nil {
					return nil, &motmedelErrors.InputError{
						Message: "An error occurred when parsing a ancestor-source-list root.",
						Cause:   err,
						Input:   directiveValue,
					}
				}
				if len(ancestorSourceListPaths) == 0 {
					return nil, nil
				}

				sources = makeSourcesFromPaths(ctx, directiveValue, ancestorSourceListPaths, "ancestor-source")
				if sources == nil {
					return nil, nil
				}
			}

			frameAncestorsDirective.Sources = sources
			directive = frameAncestorsDirective
		case "report-to":
			reportToDirective := &contentSecurityPolicyTypes.ReportToDirective{Directive: innerDirective, Token: innerDirective.RawValue}
			directive = reportToDirective
		default:
			directive = &innerDirective
			isOtherDirective = true
		}

		if isIneffectiveDirective {
			contentSecurityPolicy.IneffectiveDirectives = append(contentSecurityPolicy.IneffectiveDirectives, directive)
		} else if isOtherDirective {
			contentSecurityPolicy.OtherDirectives = append(contentSecurityPolicy.OtherDirectives, directive)
		} else {
			contentSecurityPolicy.Directives = append(contentSecurityPolicy.Directives, directive)
		}
	}

	return contentSecurityPolicy, nil
}

/*
root = serialized-policy
optional-ascii-whitespace = *( %x09 / %x0A / %x0C / %x0D / %x20 )
required-ascii-whitespace = 1*( %x09 / %x0A / %x0C / %x0D / %x20 )
serialized-policy = serialized-directive *( optional-ascii-whitespace ";" [ optional-ascii-whitespace serialized-directive ] )
serialized-directive = directive-name [ required-ascii-whitespace directive-value ]
directive-name       = 1*( ALPHA / DIGIT / "-" )
directive-value      = *( required-ascii-whitespace / ( %x21-2B / %x2D-3A / %x3C-7E ) )
serialized-source-list = ( source-expression *( required-ascii-whitespace source-expression ) ) / "'none'"
source-expression      = scheme-source / host-source / keyword-source / nonce-source / hash-source
scheme-source = scheme-part ":"
host-source = [ scheme-part "://" ] host-part [ ":" port-part ] [ path-part ]
scheme-part = scheme
scheme      = ALPHA *( ALPHA / DIGIT / "+" / "-" / "." )
host-part   = "*" / [ "*." ] 1*host-char *( "." 1*host-char ) [ "." ]
host-char   = ALPHA / DIGIT / "-"
port-part   = 1*DIGIT / "*"
path-part   = path-absolute
path-absolute = "/" [ segment-nz *( "/" segment ) ]
segment-nz    = 1*pchar
segment       = *pchar
pchar         = unreserved / pct-encoded / sub-delims / ":" / "@"
unreserved    = ALPHA / DIGIT / "-" / "." / "_" / "~"
pct-encoded   = "%" HEXDIG HEXDIG
sub-delims  = "!" / "$" / "&" / "'" / "(" / ")" / "*" / "+" / "="
keyword-source = "'self'" / "'unsafe-inline'" / "'unsafe-eval'" / "'strict-dynamic'" / "'unsafe-hashes'" / "'report-sample'" / "'unsafe-allow-redirects'" / "'wasm-unsafe-eval'"
nonce-source  = "'nonce-" base64-value "'"
base64-value  = 1*( ALPHA / DIGIT / "+" / "/" / "-" / "_" ) *2"="
hash-source    = "'" hash-algorithm "-" base64-value "'"
hash-algorithm = "sha256" / "sha384" / "sha512"
token = 1*tchar
tchar = "!" / "#" / "$" / "%" / "&" / "'" / "*" / "+" / "-" / "." / "^" / "_" / "`" / "|" / "~" / DIGIT / ALPHA
ancestor-source-list-root = ancestor-source-list
ancestor-source-list = ( ancestor-source *( required-ascii-whitespace ancestor-source) ) / "'none'"
ancestor-source = scheme-source / host-source / "'self'"
gen-delims = ":" / "/" / "?" / "#" / "[" / "]" / "@"
reserved = gen-delims / sub-delims
fragment = *( pchar / "/" / "?" )
query = *( pchar / "/" / "?" )
segment-nz-nc = 1*( unreserved / pct-encoded / sub-delims / "@" )
path-empty = ""
path-rootless = segment-nz *( "/" segment )
path-noscheme = segment-nz-nc *( "/" segment )
path-abempty = *( "/" segment )
path = path-abempty / path-absolute / path-noscheme / path-rootless / path-empty
reg-name = *( unreserved / pct-encoded / sub-delims)
dec-octet = DIGIT / %x31-39 DIGIT / "1" 2DIGIT / "2" %x30-34 DIGIT / "25" %x30-35
IPv4address = dec-octet "." dec-octet "." dec-octet "." dec-octet
h16 = 1*4HEXDIG
ls32 = ( h16 ":" h16 ) / IPv4address
IPv6address = 6( h16 ":" ) ls32 / "::" 5( h16 ":" ) ls32 / [ h16 ] "::" 4( h16 ":" ) ls32 / [ *1( h16 ":" ) h16 ] "::" 3( h16 ":" ) ls32 / [ *2( h16 ":" ) h16 ] "::" 2( h16 ":" ) ls32 / [ *3( h16 ":" ) h16 ] "::" h16 ":" ls32 / [ *4( h16 ":" ) h16 ] "::" ls32 / [ *5( h16 ":" ) h16 ] "::" h16 / [ *6( h16 ":" ) h16 ] "::"
IPvFuture = "v" 1*HEXDIG "." 1*( unreserved / sub-delims / ":" )
IP-literal = "[" ( IPv6address / IPvFuture ) "]"
port = *DIGIT
host = IP-literal / IPv4address / reg-name
userinfo = *( unreserved / pct-encoded / sub-delims / ":" )
authority = [ userinfo "@" ] host [ ":" port ]
relative-part = "//" authority path-abempty / path-absolute / path-noscheme / path-empty
relative-ref = relative-part [ "?" query ] [ "#" fragment ]
hier-part = "//" authority path-abempty / path-absolute / path-rootless / path-empty
absolute-URI = scheme ":" hier-part [ "?" query ]
URI = scheme ":" hier-part [ "?" query ] [ "#" fragment ]
uri-reference = URI / relative-ref
sandbox-directive-value-root = sandbox-directive-value
sandbox-directive-value = "" / token *( required-ascii-whitespace token )
report-uri-directive-value-root = report-uri-directive-value
report-uri-directive-value = uri-reference *( required-ascii-whitespace uri-reference )
*/

var grammar = []uint8{114, 111, 111, 116, 32, 61, 32, 115, 101, 114, 105, 97, 108, 105, 122, 101, 100, 45, 112, 111, 108, 105, 99, 121, 13, 10, 111, 112, 116, 105, 111, 110, 97, 108, 45, 97, 115, 99, 105, 105, 45, 119, 104, 105, 116, 101, 115, 112, 97, 99, 101, 32, 61, 32, 42, 40, 32, 37, 120, 48, 57, 32, 47, 32, 37, 120, 48, 65, 32, 47, 32, 37, 120, 48, 67, 32, 47, 32, 37, 120, 48, 68, 32, 47, 32, 37, 120, 50, 48, 32, 41, 13, 10, 114, 101, 113, 117, 105, 114, 101, 100, 45, 97, 115, 99, 105, 105, 45, 119, 104, 105, 116, 101, 115, 112, 97, 99, 101, 32, 61, 32, 49, 42, 40, 32, 37, 120, 48, 57, 32, 47, 32, 37, 120, 48, 65, 32, 47, 32, 37, 120, 48, 67, 32, 47, 32, 37, 120, 48, 68, 32, 47, 32, 37, 120, 50, 48, 32, 41, 13, 10, 115, 101, 114, 105, 97, 108, 105, 122, 101, 100, 45, 112, 111, 108, 105, 99, 121, 32, 61, 32, 115, 101, 114, 105, 97, 108, 105, 122, 101, 100, 45, 100, 105, 114, 101, 99, 116, 105, 118, 101, 32, 42, 40, 32, 111, 112, 116, 105, 111, 110, 97, 108, 45, 97, 115, 99, 105, 105, 45, 119, 104, 105, 116, 101, 115, 112, 97, 99, 101, 32, 34, 59, 34, 32, 91, 32, 111, 112, 116, 105, 111, 110, 97, 108, 45, 97, 115, 99, 105, 105, 45, 119, 104, 105, 116, 101, 115, 112, 97, 99, 101, 32, 115, 101, 114, 105, 97, 108, 105, 122, 101, 100, 45, 100, 105, 114, 101, 99, 116, 105, 118, 101, 32, 93, 32, 41, 13, 10, 115, 101, 114, 105, 97, 108, 105, 122, 101, 100, 45, 100, 105, 114, 101, 99, 116, 105, 118, 101, 32, 61, 32, 100, 105, 114, 101, 99, 116, 105, 118, 101, 45, 110, 97, 109, 101, 32, 91, 32, 114, 101, 113, 117, 105, 114, 101, 100, 45, 97, 115, 99, 105, 105, 45, 119, 104, 105, 116, 101, 115, 112, 97, 99, 101, 32, 100, 105, 114, 101, 99, 116, 105, 118, 101, 45, 118, 97, 108, 117, 101, 32, 93, 13, 10, 100, 105, 114, 101, 99, 116, 105, 118, 101, 45, 110, 97, 109, 101, 32, 32, 32, 32, 32, 32, 32, 61, 32, 49, 42, 40, 32, 65, 76, 80, 72, 65, 32, 47, 32, 68, 73, 71, 73, 84, 32, 47, 32, 34, 45, 34, 32, 41, 13, 10, 100, 105, 114, 101, 99, 116, 105, 118, 101, 45, 118, 97, 108, 117, 101, 32, 32, 32, 32, 32, 32, 61, 32, 42, 40, 32, 114, 101, 113, 117, 105, 114, 101, 100, 45, 97, 115, 99, 105, 105, 45, 119, 104, 105, 116, 101, 115, 112, 97, 99, 101, 32, 47, 32, 40, 32, 37, 120, 50, 49, 45, 50, 66, 32, 47, 32, 37, 120, 50, 68, 45, 51, 65, 32, 47, 32, 37, 120, 51, 67, 45, 55, 69, 32, 41, 32, 41, 13, 10, 115, 101, 114, 105, 97, 108, 105, 122, 101, 100, 45, 115, 111, 117, 114, 99, 101, 45, 108, 105, 115, 116, 32, 61, 32, 40, 32, 115, 111, 117, 114, 99, 101, 45, 101, 120, 112, 114, 101, 115, 115, 105, 111, 110, 32, 42, 40, 32, 114, 101, 113, 117, 105, 114, 101, 100, 45, 97, 115, 99, 105, 105, 45, 119, 104, 105, 116, 101, 115, 112, 97, 99, 101, 32, 115, 111, 117, 114, 99, 101, 45, 101, 120, 112, 114, 101, 115, 115, 105, 111, 110, 32, 41, 32, 41, 32, 47, 32, 34, 39, 110, 111, 110, 101, 39, 34, 13, 10, 115, 111, 117, 114, 99, 101, 45, 101, 120, 112, 114, 101, 115, 115, 105, 111, 110, 32, 32, 32, 32, 32, 32, 61, 32, 115, 99, 104, 101, 109, 101, 45, 115, 111, 117, 114, 99, 101, 32, 47, 32, 104, 111, 115, 116, 45, 115, 111, 117, 114, 99, 101, 32, 47, 32, 107, 101, 121, 119, 111, 114, 100, 45, 115, 111, 117, 114, 99, 101, 32, 47, 32, 110, 111, 110, 99, 101, 45, 115, 111, 117, 114, 99, 101, 32, 47, 32, 104, 97, 115, 104, 45, 115, 111, 117, 114, 99, 101, 13, 10, 115, 99, 104, 101, 109, 101, 45, 115, 111, 117, 114, 99, 101, 32, 61, 32, 115, 99, 104, 101, 109, 101, 45, 112, 97, 114, 116, 32, 34, 58, 34, 13, 10, 104, 111, 115, 116, 45, 115, 111, 117, 114, 99, 101, 32, 61, 32, 91, 32, 115, 99, 104, 101, 109, 101, 45, 112, 97, 114, 116, 32, 34, 58, 47, 47, 34, 32, 93, 32, 104, 111, 115, 116, 45, 112, 97, 114, 116, 32, 91, 32, 34, 58, 34, 32, 112, 111, 114, 116, 45, 112, 97, 114, 116, 32, 93, 32, 91, 32, 112, 97, 116, 104, 45, 112, 97, 114, 116, 32, 93, 13, 10, 115, 99, 104, 101, 109, 101, 45, 112, 97, 114, 116, 32, 61, 32, 115, 99, 104, 101, 109, 101, 13, 10, 115, 99, 104, 101, 109, 101, 32, 32, 32, 32, 32, 32, 61, 32, 65, 76, 80, 72, 65, 32, 42, 40, 32, 65, 76, 80, 72, 65, 32, 47, 32, 68, 73, 71, 73, 84, 32, 47, 32, 34, 43, 34, 32, 47, 32, 34, 45, 34, 32, 47, 32, 34, 46, 34, 32, 41, 13, 10, 104, 111, 115, 116, 45, 112, 97, 114, 116, 32, 32, 32, 61, 32, 34, 42, 34, 32, 47, 32, 91, 32, 34, 42, 46, 34, 32, 93, 32, 49, 42, 104, 111, 115, 116, 45, 99, 104, 97, 114, 32, 42, 40, 32, 34, 46, 34, 32, 49, 42, 104, 111, 115, 116, 45, 99, 104, 97, 114, 32, 41, 32, 91, 32, 34, 46, 34, 32, 93, 13, 10, 104, 111, 115, 116, 45, 99, 104, 97, 114, 32, 32, 32, 61, 32, 65, 76, 80, 72, 65, 32, 47, 32, 68, 73, 71, 73, 84, 32, 47, 32, 34, 45, 34, 13, 10, 112, 111, 114, 116, 45, 112, 97, 114, 116, 32, 32, 32, 61, 32, 49, 42, 68, 73, 71, 73, 84, 32, 47, 32, 34, 42, 34, 13, 10, 112, 97, 116, 104, 45, 112, 97, 114, 116, 32, 32, 32, 61, 32, 112, 97, 116, 104, 45, 97, 98, 115, 111, 108, 117, 116, 101, 13, 10, 112, 97, 116, 104, 45, 97, 98, 115, 111, 108, 117, 116, 101, 32, 61, 32, 34, 47, 34, 32, 91, 32, 115, 101, 103, 109, 101, 110, 116, 45, 110, 122, 32, 42, 40, 32, 34, 47, 34, 32, 115, 101, 103, 109, 101, 110, 116, 32, 41, 32, 93, 13, 10, 115, 101, 103, 109, 101, 110, 116, 45, 110, 122, 32, 32, 32, 32, 61, 32, 49, 42, 112, 99, 104, 97, 114, 13, 10, 115, 101, 103, 109, 101, 110, 116, 32, 32, 32, 32, 32, 32, 32, 61, 32, 42, 112, 99, 104, 97, 114, 13, 10, 112, 99, 104, 97, 114, 32, 32, 32, 32, 32, 32, 32, 32, 32, 61, 32, 117, 110, 114, 101, 115, 101, 114, 118, 101, 100, 32, 47, 32, 112, 99, 116, 45, 101, 110, 99, 111, 100, 101, 100, 32, 47, 32, 115, 117, 98, 45, 100, 101, 108, 105, 109, 115, 32, 47, 32, 34, 58, 34, 32, 47, 32, 34, 64, 34, 13, 10, 117, 110, 114, 101, 115, 101, 114, 118, 101, 100, 32, 32, 32, 32, 61, 32, 65, 76, 80, 72, 65, 32, 47, 32, 68, 73, 71, 73, 84, 32, 47, 32, 34, 45, 34, 32, 47, 32, 34, 46, 34, 32, 47, 32, 34, 95, 34, 32, 47, 32, 34, 126, 34, 13, 10, 112, 99, 116, 45, 101, 110, 99, 111, 100, 101, 100, 32, 32, 32, 61, 32, 34, 37, 34, 32, 72, 69, 88, 68, 73, 71, 32, 72, 69, 88, 68, 73, 71, 13, 10, 115, 117, 98, 45, 100, 101, 108, 105, 109, 115, 32, 32, 61, 32, 34, 33, 34, 32, 47, 32, 34, 36, 34, 32, 47, 32, 34, 38, 34, 32, 47, 32, 34, 39, 34, 32, 47, 32, 34, 40, 34, 32, 47, 32, 34, 41, 34, 32, 47, 32, 34, 42, 34, 32, 47, 32, 34, 43, 34, 32, 47, 32, 34, 61, 34, 13, 10, 107, 101, 121, 119, 111, 114, 100, 45, 115, 111, 117, 114, 99, 101, 32, 61, 32, 34, 39, 115, 101, 108, 102, 39, 34, 32, 47, 32, 34, 39, 117, 110, 115, 97, 102, 101, 45, 105, 110, 108, 105, 110, 101, 39, 34, 32, 47, 32, 34, 39, 117, 110, 115, 97, 102, 101, 45, 101, 118, 97, 108, 39, 34, 32, 47, 32, 34, 39, 115, 116, 114, 105, 99, 116, 45, 100, 121, 110, 97, 109, 105, 99, 39, 34, 32, 47, 32, 34, 39, 117, 110, 115, 97, 102, 101, 45, 104, 97, 115, 104, 101, 115, 39, 34, 32, 47, 32, 34, 39, 114, 101, 112, 111, 114, 116, 45, 115, 97, 109, 112, 108, 101, 39, 34, 32, 47, 32, 34, 39, 117, 110, 115, 97, 102, 101, 45, 97, 108, 108, 111, 119, 45, 114, 101, 100, 105, 114, 101, 99, 116, 115, 39, 34, 32, 47, 32, 34, 39, 119, 97, 115, 109, 45, 117, 110, 115, 97, 102, 101, 45, 101, 118, 97, 108, 39, 34, 13, 10, 110, 111, 110, 99, 101, 45, 115, 111, 117, 114, 99, 101, 32, 32, 61, 32, 34, 39, 110, 111, 110, 99, 101, 45, 34, 32, 98, 97, 115, 101, 54, 52, 45, 118, 97, 108, 117, 101, 32, 34, 39, 34, 13, 10, 98, 97, 115, 101, 54, 52, 45, 118, 97, 108, 117, 101, 32, 32, 61, 32, 49, 42, 40, 32, 65, 76, 80, 72, 65, 32, 47, 32, 68, 73, 71, 73, 84, 32, 47, 32, 34, 43, 34, 32, 47, 32, 34, 47, 34, 32, 47, 32, 34, 45, 34, 32, 47, 32, 34, 95, 34, 32, 41, 32, 42, 50, 34, 61, 34, 13, 10, 104, 97, 115, 104, 45, 115, 111, 117, 114, 99, 101, 32, 32, 32, 32, 61, 32, 34, 39, 34, 32, 104, 97, 115, 104, 45, 97, 108, 103, 111, 114, 105, 116, 104, 109, 32, 34, 45, 34, 32, 98, 97, 115, 101, 54, 52, 45, 118, 97, 108, 117, 101, 32, 34, 39, 34, 13, 10, 104, 97, 115, 104, 45, 97, 108, 103, 111, 114, 105, 116, 104, 109, 32, 61, 32, 34, 115, 104, 97, 50, 53, 54, 34, 32, 47, 32, 34, 115, 104, 97, 51, 56, 52, 34, 32, 47, 32, 34, 115, 104, 97, 53, 49, 50, 34, 13, 10, 116, 111, 107, 101, 110, 32, 61, 32, 49, 42, 116, 99, 104, 97, 114, 13, 10, 116, 99, 104, 97, 114, 32, 61, 32, 34, 33, 34, 32, 47, 32, 34, 35, 34, 32, 47, 32, 34, 36, 34, 32, 47, 32, 34, 37, 34, 32, 47, 32, 34, 38, 34, 32, 47, 32, 34, 39, 34, 32, 47, 32, 34, 42, 34, 32, 47, 32, 34, 43, 34, 32, 47, 32, 34, 45, 34, 32, 47, 32, 34, 46, 34, 32, 47, 32, 34, 94, 34, 32, 47, 32, 34, 95, 34, 32, 47, 32, 34, 96, 34, 32, 47, 32, 34, 124, 34, 32, 47, 32, 34, 126, 34, 32, 47, 32, 68, 73, 71, 73, 84, 32, 47, 32, 65, 76, 80, 72, 65, 13, 10, 97, 110, 99, 101, 115, 116, 111, 114, 45, 115, 111, 117, 114, 99, 101, 45, 108, 105, 115, 116, 45, 114, 111, 111, 116, 32, 61, 32, 97, 110, 99, 101, 115, 116, 111, 114, 45, 115, 111, 117, 114, 99, 101, 45, 108, 105, 115, 116, 13, 10, 97, 110, 99, 101, 115, 116, 111, 114, 45, 115, 111, 117, 114, 99, 101, 45, 108, 105, 115, 116, 32, 61, 32, 40, 32, 97, 110, 99, 101, 115, 116, 111, 114, 45, 115, 111, 117, 114, 99, 101, 32, 42, 40, 32, 114, 101, 113, 117, 105, 114, 101, 100, 45, 97, 115, 99, 105, 105, 45, 119, 104, 105, 116, 101, 115, 112, 97, 99, 101, 32, 97, 110, 99, 101, 115, 116, 111, 114, 45, 115, 111, 117, 114, 99, 101, 41, 32, 41, 32, 47, 32, 34, 39, 110, 111, 110, 101, 39, 34, 13, 10, 97, 110, 99, 101, 115, 116, 111, 114, 45, 115, 111, 117, 114, 99, 101, 32, 61, 32, 115, 99, 104, 101, 109, 101, 45, 115, 111, 117, 114, 99, 101, 32, 47, 32, 104, 111, 115, 116, 45, 115, 111, 117, 114, 99, 101, 32, 47, 32, 34, 39, 115, 101, 108, 102, 39, 34, 13, 10, 103, 101, 110, 45, 100, 101, 108, 105, 109, 115, 32, 61, 32, 34, 58, 34, 32, 47, 32, 34, 47, 34, 32, 47, 32, 34, 63, 34, 32, 47, 32, 34, 35, 34, 32, 47, 32, 34, 91, 34, 32, 47, 32, 34, 93, 34, 32, 47, 32, 34, 64, 34, 13, 10, 114, 101, 115, 101, 114, 118, 101, 100, 32, 61, 32, 103, 101, 110, 45, 100, 101, 108, 105, 109, 115, 32, 47, 32, 115, 117, 98, 45, 100, 101, 108, 105, 109, 115, 13, 10, 102, 114, 97, 103, 109, 101, 110, 116, 32, 61, 32, 42, 40, 32, 112, 99, 104, 97, 114, 32, 47, 32, 34, 47, 34, 32, 47, 32, 34, 63, 34, 32, 41, 13, 10, 113, 117, 101, 114, 121, 32, 61, 32, 42, 40, 32, 112, 99, 104, 97, 114, 32, 47, 32, 34, 47, 34, 32, 47, 32, 34, 63, 34, 32, 41, 13, 10, 115, 101, 103, 109, 101, 110, 116, 45, 110, 122, 45, 110, 99, 32, 61, 32, 49, 42, 40, 32, 117, 110, 114, 101, 115, 101, 114, 118, 101, 100, 32, 47, 32, 112, 99, 116, 45, 101, 110, 99, 111, 100, 101, 100, 32, 47, 32, 115, 117, 98, 45, 100, 101, 108, 105, 109, 115, 32, 47, 32, 34, 64, 34, 32, 41, 13, 10, 112, 97, 116, 104, 45, 101, 109, 112, 116, 121, 32, 61, 32, 34, 34, 13, 10, 112, 97, 116, 104, 45, 114, 111, 111, 116, 108, 101, 115, 115, 32, 61, 32, 115, 101, 103, 109, 101, 110, 116, 45, 110, 122, 32, 42, 40, 32, 34, 47, 34, 32, 115, 101, 103, 109, 101, 110, 116, 32, 41, 13, 10, 112, 97, 116, 104, 45, 110, 111, 115, 99, 104, 101, 109, 101, 32, 61, 32, 115, 101, 103, 109, 101, 110, 116, 45, 110, 122, 45, 110, 99, 32, 42, 40, 32, 34, 47, 34, 32, 115, 101, 103, 109, 101, 110, 116, 32, 41, 13, 10, 112, 97, 116, 104, 45, 97, 98, 101, 109, 112, 116, 121, 32, 61, 32, 42, 40, 32, 34, 47, 34, 32, 115, 101, 103, 109, 101, 110, 116, 32, 41, 13, 10, 112, 97, 116, 104, 32, 61, 32, 112, 97, 116, 104, 45, 97, 98, 101, 109, 112, 116, 121, 32, 47, 32, 112, 97, 116, 104, 45, 97, 98, 115, 111, 108, 117, 116, 101, 32, 47, 32, 112, 97, 116, 104, 45, 110, 111, 115, 99, 104, 101, 109, 101, 32, 47, 32, 112, 97, 116, 104, 45, 114, 111, 111, 116, 108, 101, 115, 115, 32, 47, 32, 112, 97, 116, 104, 45, 101, 109, 112, 116, 121, 13, 10, 114, 101, 103, 45, 110, 97, 109, 101, 32, 61, 32, 42, 40, 32, 117, 110, 114, 101, 115, 101, 114, 118, 101, 100, 32, 47, 32, 112, 99, 116, 45, 101, 110, 99, 111, 100, 101, 100, 32, 47, 32, 115, 117, 98, 45, 100, 101, 108, 105, 109, 115, 41, 13, 10, 100, 101, 99, 45, 111, 99, 116, 101, 116, 32, 61, 32, 68, 73, 71, 73, 84, 32, 47, 32, 37, 120, 51, 49, 45, 51, 57, 32, 68, 73, 71, 73, 84, 32, 47, 32, 34, 49, 34, 32, 50, 68, 73, 71, 73, 84, 32, 47, 32, 34, 50, 34, 32, 37, 120, 51, 48, 45, 51, 52, 32, 68, 73, 71, 73, 84, 32, 47, 32, 34, 50, 53, 34, 32, 37, 120, 51, 48, 45, 51, 53, 13, 10, 73, 80, 118, 52, 97, 100, 100, 114, 101, 115, 115, 32, 61, 32, 100, 101, 99, 45, 111, 99, 116, 101, 116, 32, 34, 46, 34, 32, 100, 101, 99, 45, 111, 99, 116, 101, 116, 32, 34, 46, 34, 32, 100, 101, 99, 45, 111, 99, 116, 101, 116, 32, 34, 46, 34, 32, 100, 101, 99, 45, 111, 99, 116, 101, 116, 13, 10, 104, 49, 54, 32, 61, 32, 49, 42, 52, 72, 69, 88, 68, 73, 71, 13, 10, 108, 115, 51, 50, 32, 61, 32, 40, 32, 104, 49, 54, 32, 34, 58, 34, 32, 104, 49, 54, 32, 41, 32, 47, 32, 73, 80, 118, 52, 97, 100, 100, 114, 101, 115, 115, 13, 10, 73, 80, 118, 54, 97, 100, 100, 114, 101, 115, 115, 32, 61, 32, 54, 40, 32, 104, 49, 54, 32, 34, 58, 34, 32, 41, 32, 108, 115, 51, 50, 32, 47, 32, 34, 58, 58, 34, 32, 53, 40, 32, 104, 49, 54, 32, 34, 58, 34, 32, 41, 32, 108, 115, 51, 50, 32, 47, 32, 91, 32, 104, 49, 54, 32, 93, 32, 34, 58, 58, 34, 32, 52, 40, 32, 104, 49, 54, 32, 34, 58, 34, 32, 41, 32, 108, 115, 51, 50, 32, 47, 32, 91, 32, 42, 49, 40, 32, 104, 49, 54, 32, 34, 58, 34, 32, 41, 32, 104, 49, 54, 32, 93, 32, 34, 58, 58, 34, 32, 51, 40, 32, 104, 49, 54, 32, 34, 58, 34, 32, 41, 32, 108, 115, 51, 50, 32, 47, 32, 91, 32, 42, 50, 40, 32, 104, 49, 54, 32, 34, 58, 34, 32, 41, 32, 104, 49, 54, 32, 93, 32, 34, 58, 58, 34, 32, 50, 40, 32, 104, 49, 54, 32, 34, 58, 34, 32, 41, 32, 108, 115, 51, 50, 32, 47, 32, 91, 32, 42, 51, 40, 32, 104, 49, 54, 32, 34, 58, 34, 32, 41, 32, 104, 49, 54, 32, 93, 32, 34, 58, 58, 34, 32, 104, 49, 54, 32, 34, 58, 34, 32, 108, 115, 51, 50, 32, 47, 32, 91, 32, 42, 52, 40, 32, 104, 49, 54, 32, 34, 58, 34, 32, 41, 32, 104, 49, 54, 32, 93, 32, 34, 58, 58, 34, 32, 108, 115, 51, 50, 32, 47, 32, 91, 32, 42, 53, 40, 32, 104, 49, 54, 32, 34, 58, 34, 32, 41, 32, 104, 49, 54, 32, 93, 32, 34, 58, 58, 34, 32, 104, 49, 54, 32, 47, 32, 91, 32, 42, 54, 40, 32, 104, 49, 54, 32, 34, 58, 34, 32, 41, 32, 104, 49, 54, 32, 93, 32, 34, 58, 58, 34, 13, 10, 73, 80, 118, 70, 117, 116, 117, 114, 101, 32, 61, 32, 34, 118, 34, 32, 49, 42, 72, 69, 88, 68, 73, 71, 32, 34, 46, 34, 32, 49, 42, 40, 32, 117, 110, 114, 101, 115, 101, 114, 118, 101, 100, 32, 47, 32, 115, 117, 98, 45, 100, 101, 108, 105, 109, 115, 32, 47, 32, 34, 58, 34, 32, 41, 13, 10, 73, 80, 45, 108, 105, 116, 101, 114, 97, 108, 32, 61, 32, 34, 91, 34, 32, 40, 32, 73, 80, 118, 54, 97, 100, 100, 114, 101, 115, 115, 32, 47, 32, 73, 80, 118, 70, 117, 116, 117, 114, 101, 32, 41, 32, 34, 93, 34, 13, 10, 112, 111, 114, 116, 32, 61, 32, 42, 68, 73, 71, 73, 84, 13, 10, 104, 111, 115, 116, 32, 61, 32, 73, 80, 45, 108, 105, 116, 101, 114, 97, 108, 32, 47, 32, 73, 80, 118, 52, 97, 100, 100, 114, 101, 115, 115, 32, 47, 32, 114, 101, 103, 45, 110, 97, 109, 101, 13, 10, 117, 115, 101, 114, 105, 110, 102, 111, 32, 61, 32, 42, 40, 32, 117, 110, 114, 101, 115, 101, 114, 118, 101, 100, 32, 47, 32, 112, 99, 116, 45, 101, 110, 99, 111, 100, 101, 100, 32, 47, 32, 115, 117, 98, 45, 100, 101, 108, 105, 109, 115, 32, 47, 32, 34, 58, 34, 32, 41, 13, 10, 97, 117, 116, 104, 111, 114, 105, 116, 121, 32, 61, 32, 91, 32, 117, 115, 101, 114, 105, 110, 102, 111, 32, 34, 64, 34, 32, 93, 32, 104, 111, 115, 116, 32, 91, 32, 34, 58, 34, 32, 112, 111, 114, 116, 32, 93, 13, 10, 114, 101, 108, 97, 116, 105, 118, 101, 45, 112, 97, 114, 116, 32, 61, 32, 34, 47, 47, 34, 32, 97, 117, 116, 104, 111, 114, 105, 116, 121, 32, 112, 97, 116, 104, 45, 97, 98, 101, 109, 112, 116, 121, 32, 47, 32, 112, 97, 116, 104, 45, 97, 98, 115, 111, 108, 117, 116, 101, 32, 47, 32, 112, 97, 116, 104, 45, 110, 111, 115, 99, 104, 101, 109, 101, 32, 47, 32, 112, 97, 116, 104, 45, 101, 109, 112, 116, 121, 13, 10, 114, 101, 108, 97, 116, 105, 118, 101, 45, 114, 101, 102, 32, 61, 32, 114, 101, 108, 97, 116, 105, 118, 101, 45, 112, 97, 114, 116, 32, 91, 32, 34, 63, 34, 32, 113, 117, 101, 114, 121, 32, 93, 32, 91, 32, 34, 35, 34, 32, 102, 114, 97, 103, 109, 101, 110, 116, 32, 93, 13, 10, 104, 105, 101, 114, 45, 112, 97, 114, 116, 32, 61, 32, 34, 47, 47, 34, 32, 97, 117, 116, 104, 111, 114, 105, 116, 121, 32, 112, 97, 116, 104, 45, 97, 98, 101, 109, 112, 116, 121, 32, 47, 32, 112, 97, 116, 104, 45, 97, 98, 115, 111, 108, 117, 116, 101, 32, 47, 32, 112, 97, 116, 104, 45, 114, 111, 111, 116, 108, 101, 115, 115, 32, 47, 32, 112, 97, 116, 104, 45, 101, 109, 112, 116, 121, 13, 10, 97, 98, 115, 111, 108, 117, 116, 101, 45, 85, 82, 73, 32, 61, 32, 115, 99, 104, 101, 109, 101, 32, 34, 58, 34, 32, 104, 105, 101, 114, 45, 112, 97, 114, 116, 32, 91, 32, 34, 63, 34, 32, 113, 117, 101, 114, 121, 32, 93, 13, 10, 85, 82, 73, 32, 61, 32, 115, 99, 104, 101, 109, 101, 32, 34, 58, 34, 32, 104, 105, 101, 114, 45, 112, 97, 114, 116, 32, 91, 32, 34, 63, 34, 32, 113, 117, 101, 114, 121, 32, 93, 32, 91, 32, 34, 35, 34, 32, 102, 114, 97, 103, 109, 101, 110, 116, 32, 93, 13, 10, 117, 114, 105, 45, 114, 101, 102, 101, 114, 101, 110, 99, 101, 32, 61, 32, 85, 82, 73, 32, 47, 32, 114, 101, 108, 97, 116, 105, 118, 101, 45, 114, 101, 102, 13, 10, 115, 97, 110, 100, 98, 111, 120, 45, 100, 105, 114, 101, 99, 116, 105, 118, 101, 45, 118, 97, 108, 117, 101, 45, 114, 111, 111, 116, 32, 61, 32, 115, 97, 110, 100, 98, 111, 120, 45, 100, 105, 114, 101, 99, 116, 105, 118, 101, 45, 118, 97, 108, 117, 101, 13, 10, 115, 97, 110, 100, 98, 111, 120, 45, 100, 105, 114, 101, 99, 116, 105, 118, 101, 45, 118, 97, 108, 117, 101, 32, 61, 32, 34, 34, 32, 47, 32, 116, 111, 107, 101, 110, 32, 42, 40, 32, 114, 101, 113, 117, 105, 114, 101, 100, 45, 97, 115, 99, 105, 105, 45, 119, 104, 105, 116, 101, 115, 112, 97, 99, 101, 32, 116, 111, 107, 101, 110, 32, 41, 13, 10, 114, 101, 112, 111, 114, 116, 45, 117, 114, 105, 45, 100, 105, 114, 101, 99, 116, 105, 118, 101, 45, 118, 97, 108, 117, 101, 45, 114, 111, 111, 116, 32, 61, 32, 114, 101, 112, 111, 114, 116, 45, 117, 114, 105, 45, 100, 105, 114, 101, 99, 116, 105, 118, 101, 45, 118, 97, 108, 117, 101, 13, 10, 114, 101, 112, 111, 114, 116, 45, 117, 114, 105, 45, 100, 105, 114, 101, 99, 116, 105, 118, 101, 45, 118, 97, 108, 117, 101, 32, 61, 32, 117, 114, 105, 45, 114, 101, 102, 101, 114, 101, 110, 99, 101, 32, 42, 40, 32, 114, 101, 113, 117, 105, 114, 101, 100, 45, 97, 115, 99, 105, 105, 45, 119, 104, 105, 116, 101, 115, 112, 97, 99, 101, 32, 117, 114, 105, 45, 114, 101, 102, 101, 114, 101, 110, 99, 101, 32, 41, 13, 10, 13, 10}

func init() {
	var err error
	ContentSecurityPolicyGrammar, err = goabnf.ParseABNF(grammar)
	if err != nil {
		panic(err)
	}
}
