package content_security_policy

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"github.com/Motmedel/parsing_utils/pkg/parsing_utils"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	contentSecurityPolicyTypes "github.com/Motmedel/utils_go/pkg/http/types/content_security_policy"
	goabnf "github.com/pandatix/go-abnf"
	"strings"
)

//go:embed grammar.txt
var grammar []byte

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

var (
	ErrNilContentSecurityPolicy = errors.New("nil content security policy")
	ErrNilHostPartPath          = errors.New("nil host part path")
	ErrNilDirectiveNamePath     = errors.New("nil directive name path")
	ErrUnexpectedSourceRuleName = errors.New("unexpected source rule name")
)

// TODO: Update to use proper errors

func makeSourcesFromPaths(
	data []byte,
	paths []*goabnf.Path,
	ruleName string,
) ([]contentSecurityPolicyTypes.SourceI, error) {
	sourceExpressionPaths := parsing_utils.SearchPath(paths[0], []string{ruleName}, 2, false)
	if len(sourceExpressionPaths) == 0 {
		return nil, nil
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
				return nil, motmedelErrors.NewWithTrace(ErrNilHostPartPath, concreteSourcePath)
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
		case "ancestor-keyword-source":
			fallthrough
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
			return nil, motmedelErrors.NewWithTrace(
				fmt.Errorf("%w: %s", ErrUnexpectedSourceRuleName, matchRuleName),
				matchRuleName,
			)
		}
	}

	return sources, nil
}

func ParseContentSecurityPolicy(data []byte) (*contentSecurityPolicyTypes.ContentSecurityPolicy, error) {
	paths, err := parsing_utils.GetParsedDataPaths(ContentSecurityPolicyGrammar, data)
	if err != nil {
		return nil, motmedelErrors.New(fmt.Errorf("get parsed data paths: %w", err), data)
	}
	if len(paths) == 0 {
		return nil, motmedelErrors.NewWithTrace(motmedelErrors.ErrSyntaxError, data)
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
			return nil, motmedelErrors.NewWithTrace(ErrNilDirectiveNamePath, interestingPath)
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
					return nil, motmedelErrors.New(
						fmt.Errorf("goabnf parse (serialized source list): %w", err),
						directiveValue,
					)
				}
				if len(serializedSourceListPaths) == 0 {
					return nil, nil
				}

				sources, err = makeSourcesFromPaths(
					directiveValue,
					serializedSourceListPaths,
					"source-expression",
				)
				if err != nil {
					return nil, motmedelErrors.New(
						fmt.Errorf("make sources from paths (source expression): %w", err),
						directiveValue,
						serializedSourceListPaths,
					)
				}
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
		case "require-sri":
			directive = &contentSecurityPolicyTypes.RequireSriDirective{Directive: innerDirective}
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
		case "upgrade-insecure-request":
			directive = &contentSecurityPolicyTypes.UpgradeInsecureRequestDirective{Directive: innerDirective}
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
				return nil, motmedelErrors.New(
					fmt.Errorf("goabnf parse (sandbox directive value root): %w", err),
					directiveValue,
				)
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
				return nil, motmedelErrors.New(
					fmt.Errorf("goabnf parse (report uri directive value root): %w", err),
					directiveValue,
				)
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
			} else {
				ancestorSourceListPaths, err := goabnf.Parse(directiveValue, ContentSecurityPolicyGrammar, "ancestor-source-list-root")
				if err != nil {
					return nil, motmedelErrors.New(
						fmt.Errorf("goabnf parse (ancestor soruce list root): %w", err),
						directiveValue,
					)
				}
				if len(ancestorSourceListPaths) == 0 {
					return nil, nil
				}

				sources, err = makeSourcesFromPaths(directiveValue, ancestorSourceListPaths, "ancestor-source")
				if err != nil {
					return nil, motmedelErrors.New(
						fmt.Errorf("make sources from paths (ancestor source): %w", err),
						directiveValue,
					)
				}
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

func init() {
	var err error
	ContentSecurityPolicyGrammar, err = goabnf.ParseABNF(grammar)
	if err != nil {
		panic(fmt.Sprintf("colud not parse content security policy grammar: %v", err))
	}
}
