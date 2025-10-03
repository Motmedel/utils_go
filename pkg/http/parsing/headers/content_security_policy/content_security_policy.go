package content_security_policy

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"strings"

	"github.com/Motmedel/parsing_utils/pkg/parsing_utils"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	contentSecurityPolicyTypes "github.com/Motmedel/utils_go/pkg/http/types/content_security_policy"
	goabnf "github.com/pandatix/go-abnf"
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
					Source: innerSource,
					Keyword: strings.Trim(
						string(parsing_utils.ExtractPathValue(data, concreteSourcePath.Subpaths[0])),
						"'",
					),
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
					return nil, motmedelErrors.New(
						fmt.Errorf("%w (serialized-source-list)", motmedelErrors.ErrSyntaxError),
						directiveValue,
					)
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
				if len(sources) == 0 {
					return nil, motmedelErrors.New(
						fmt.Errorf("%w (source-expression)", motmedelErrors.ErrSyntaxError),
						directiveValue,
						serializedSourceListPaths,
					)
				}
			}
		}

		parsedDirective := contentSecurityPolicyTypes.ParsedDirective{
			Name:    lowercaseDirectiveName,
			Value:   string(directiveValue),
			RawName: directiveName,
		}
		sourceDirective := contentSecurityPolicyTypes.SourceDirective{ParsedDirective: parsedDirective, Sources: sources}

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
		case "require-sri-for":
			resourceTypes := bytes.Split(directiveValue, []byte(" "))
			var trimmedResourceTypes []string
			for _, resourceType := range resourceTypes {
				trimmedResourceType := bytes.ToLower(bytes.TrimSpace(resourceType))
				if len(trimmedResourceType) == 0 {
					continue
				}
				trimmedResourceTypes = append(trimmedResourceTypes, string(trimmedResourceType))
			}

			directive = &contentSecurityPolicyTypes.RequireSriForDirective{
				ParsedDirective: parsedDirective,
				ResourceTypes:   trimmedResourceTypes,
			}
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
		case "upgrade-insecure-requests":
			directive = &contentSecurityPolicyTypes.UpgradeInsecureRequestsDirective{ParsedDirective: parsedDirective}
		case "worker-src":
			directive = &contentSecurityPolicyTypes.WorkerSrcDirective{SourceDirective: sourceDirective}
		case "sandbox":
			sandboxDirective := &contentSecurityPolicyTypes.SandboxDirective{ParsedDirective: parsedDirective}

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
				return nil, motmedelErrors.New(
					fmt.Errorf("%w (sandbox directive value root)", motmedelErrors.ErrSyntaxError),
					directiveValue,
				)
			}

			tokenPaths := parsing_utils.SearchPath(sandboxDirectiveValuePaths[0], []string{"token"}, 2, false)
			for _, tokenPath := range tokenPaths {
				sandboxDirective.Tokens = append(
					sandboxDirective.Tokens,
					string(parsing_utils.ExtractPathValue(directiveValue, tokenPath)),
				)
			}
			directive = sandboxDirective
		case "webrtc":
			rawValue := parsedDirective.Value
			if rawValue != "allow" && rawValue != "block" {
				return nil, motmedelErrors.New(
					fmt.Errorf("%w (webrtc directive)", motmedelErrors.ErrSyntaxError),
					rawValue,
				)
			}
			webrtcDirective := &contentSecurityPolicyTypes.WebrtcDirective{ParsedDirective: parsedDirective, Value: rawValue}
			directive = webrtcDirective
		case "report-uri":
			reportUriDirective := &contentSecurityPolicyTypes.ReportUriDirective{ParsedDirective: parsedDirective}

			reportUriDirectivePaths, err := goabnf.Parse(directiveValue, ContentSecurityPolicyGrammar, "report-uri-directive-value-root")
			if err != nil {
				return nil, motmedelErrors.New(
					fmt.Errorf("goabnf parse (report uri directive value root): %w", err),
					directiveValue,
				)
			}
			if len(reportUriDirectivePaths) == 0 {
				return nil, motmedelErrors.New(
					fmt.Errorf("%w (report uri directive value root)", motmedelErrors.ErrSyntaxError),
					directiveValue,
				)
			}

			reportUriDirectivePath := reportUriDirectivePaths[0]
			uriReferencePaths := parsing_utils.SearchPath(
				reportUriDirectivePath,
				[]string{"uri-reference"},
				1,
				false,
			)
			if len(uriReferencePaths) == 0 {
				return nil, motmedelErrors.New(
					fmt.Errorf("%w (uri-reference)", motmedelErrors.ErrSyntaxError),
					directiveValue,
				)
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
					return nil, motmedelErrors.New(
						fmt.Errorf("%w (ancestor source list root)", motmedelErrors.ErrSyntaxError),
						directiveValue,
					)
				}

				sources, err = makeSourcesFromPaths(directiveValue, ancestorSourceListPaths, "ancestor-source")
				if err != nil {
					return nil, motmedelErrors.New(
						fmt.Errorf("make sources from paths (ancestor source): %w", err),
						directiveValue,
					)
				}
				if sources == nil {
					return nil, motmedelErrors.New(
						fmt.Errorf("%w (ancestor source)", motmedelErrors.ErrSyntaxError),
						directiveValue,
						ancestorSourceListPaths,
					)
				}
			}

			frameAncestorsDirective.Sources = sources
			directive = frameAncestorsDirective
		case "report-to":
			reportToDirective := &contentSecurityPolicyTypes.ReportToDirective{ParsedDirective: parsedDirective, Token: parsedDirective.Value}
			directive = reportToDirective
		case "require-trusted-types-for":
			requireTrustedTypesForDirective := &contentSecurityPolicyTypes.RequireTrustedTypesForDirective{
				ParsedDirective: parsedDirective,
			}

			requireTrustedTypesForDirectiveValuePaths, err := goabnf.Parse(directiveValue, ContentSecurityPolicyGrammar, "require-trusted-types-for-directive-value-root")
			if err != nil {
				return nil, motmedelErrors.New(
					fmt.Errorf("goabnf parse (require trusted types for directive value root): %w", err),
					directiveValue,
				)
			}
			if len(requireTrustedTypesForDirectiveValuePaths) == 0 {
				return nil, motmedelErrors.New(
					fmt.Errorf("%w (require trusted types for directive value root)", motmedelErrors.ErrSyntaxError),
				)
			}

			sinkGroupPaths := parsing_utils.SearchPath(requireTrustedTypesForDirectiveValuePaths[0], []string{"trusted-types-sink-group"}, 2, false)
			for _, path := range sinkGroupPaths {
				requireTrustedTypesForDirective.SinkGroups = append(
					requireTrustedTypesForDirective.SinkGroups,
					strings.Trim(string(parsing_utils.ExtractPathValue(directiveValue, path)), "'"),
				)
			}
			directive = requireTrustedTypesForDirective
		case "trusted-types":
			trustedTypesDirective := &contentSecurityPolicyTypes.TrustedTypesDirective{ParsedDirective: parsedDirective}

			trustedTypesDirectiveValuePaths, err := goabnf.Parse(directiveValue, ContentSecurityPolicyGrammar, "trusted-types-directive-value-root")
			if err != nil {
				return nil, motmedelErrors.New(
					fmt.Errorf("goabnf parse (trusted types directive value root): %w", err),
					directiveValue,
				)
			}
			if len(trustedTypesDirectiveValuePaths) == 0 {
				return nil, motmedelErrors.New(
					fmt.Errorf("%w (trusted types directive value root)", motmedelErrors.ErrSyntaxError),
				)
			}

			ttExpressionPaths := parsing_utils.SearchPath(trustedTypesDirectiveValuePaths[0], []string{"tt-expression"}, 2, false)
			for _, ttExpressionPath := range ttExpressionPaths {
				concreteTTPath := ttExpressionPath.Subpaths[0]
				switch concreteTTPath.MatchRule {
				case "tt-policy-name":
					trustedTypesDirective.Expressions = append(
						trustedTypesDirective.Expressions,
						contentSecurityPolicyTypes.TrustedTypeExpression{
							Kind:  "policy-name",
							Value: string(parsing_utils.ExtractPathValue(directiveValue, concreteTTPath)),
						},
					)
				case "tt-keyword":
					val := string(parsing_utils.ExtractPathValue(directiveValue, concreteTTPath))
					trustedTypesDirective.Expressions = append(
						trustedTypesDirective.Expressions,
						contentSecurityPolicyTypes.TrustedTypeExpression{
							Kind:  "keyword",
							Value: strings.Trim(val, "'"),
						},
					)
				case "tt-wildcard":
					trustedTypesDirective.Expressions = append(
						trustedTypesDirective.Expressions,
						contentSecurityPolicyTypes.TrustedTypeExpression{
							Kind:  "wildcard",
							Value: string(parsing_utils.ExtractPathValue(directiveValue, concreteTTPath)),
						},
					)
				}
			}

			directive = trustedTypesDirective

		default:
			directive = &parsedDirective
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
