package content_security_policy

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Motmedel/utils_go/pkg/utils"
)

type SourceI interface {
	GetRaw() string
	String() string
}

type Source struct {
	Raw string `json:"raw,omitempty"`
}

func (source *Source) GetRaw() string {
	return source.Raw
}

type NoneSource struct {
	Source
}

func (noneSource *NoneSource) String() string {
	return "'none'"
}

type SchemeSource struct {
	Source
	Scheme string `json:"scheme,omitempty"`
}

func (schemeSource *SchemeSource) String() string {
	if scheme := schemeSource.Scheme; scheme != "" {
		return scheme + ":"
	}
	return ""
}

type HostSource struct {
	Source
	Scheme     string `json:"scheme,omitempty"`
	Host       string `json:"host,omitempty"`
	PortString string `json:"port_string,omitempty"`
	Path       string `json:"path,omitempty"`
}

func (hostSource *HostSource) String() string {
	scheme := hostSource.Scheme
	host := hostSource.Host
	portString := hostSource.PortString
	path := hostSource.Path

	if scheme == "" && host == "" && portString == "" && path == "" {
		return ""
	}

	var builder strings.Builder

	if scheme != "" {
		builder.WriteString(scheme)
		builder.WriteString("://")
	}

	builder.WriteString(host)
	if portString != "" {
		builder.WriteString(":")
		builder.WriteString(portString)
	}
	if path != "" {
		builder.WriteString(path)
	}

	return builder.String()
}

type KeywordSource struct {
	Source
	Keyword string `json:"keyword,omitempty"`
}

func (keywordSource *KeywordSource) String() string {
	if keyword := keywordSource.Keyword; keyword != "" {
		return fmt.Sprintf("'%s'", keyword)
	}
	return ""
}

type NonceSource struct {
	Source
	Base64Value string `json:"base64_value,omitempty"`
}

func (nonceSource *NonceSource) String() string {
	if nonceSource.Base64Value == "" {
		return ""
	}
	return "'nonce-" + nonceSource.Base64Value + "'"
}

type HashSource struct {
	Source
	HashAlgorithm string `json:"hash_algorithm,omitempty"`
	Base64Value   string `json:"base64_value,omitempty"`
}

func (hashSource *HashSource) String() string {
	hashAlgorithm := hashSource.HashAlgorithm
	base64Value := hashSource.Base64Value

	if hashAlgorithm == "" || base64Value == "" {
		return ""
	}

	return fmt.Sprintf("'%s-%s'", hashAlgorithm, base64Value)
}

type DirectiveI interface {
	GetName() string
	GetRawName() string
	GetRawValue() string
	String() string
}

type Directive struct {
	Name     string `json:"name,omitempty"`
	RawName  string `json:"raw_name,omitempty"`
	RawValue string `json:"raw_value,omitempty"`
}

func (directive *Directive) GetName() string {
	return directive.Name
}

func (directive *Directive) GetRawName() string {
	return directive.RawName
}

func (directive *Directive) GetRawValue() string {
	return directive.RawValue
}

func (directive *Directive) String() string {
	value := directive.GetRawValue()
	if value == "" {
		return ""
	}

	return directive.RawName + " " + value
}

type SourceDirectiveI interface {
	GetSources() []SourceI
}

type SourceDirective struct {
	Directive
	Sources []SourceI `json:"sources,omitempty"`
}

func (sourceDirective *SourceDirective) String() string {
	var sourceStrings []string

	for _, source := range sourceDirective.Sources {
		if source == nil {
			continue
		}
		if sourceString := source.String(); sourceString != "" {
			sourceStrings = append(sourceStrings, sourceString)
		}
	}

	if len(sourceStrings) == 0 {
		return ""
	}

	return fmt.Sprintf("%s %s", sourceDirective.GetName(), strings.Join(sourceStrings, " "))
}

func (sourceDirective *SourceDirective) GetSources() []SourceI {
	return sourceDirective.Sources
}

type BaseUriDirective struct {
	SourceDirective
}

type ChildSrcDirective struct {
	SourceDirective
}

type ConnectSrcDirective struct {
	SourceDirective
}

type DefaultSrcDirective struct {
	SourceDirective
}

type FontSrcDirective struct {
	SourceDirective
}

type FormActionDirective struct {
	SourceDirective
}

type FrameSrcDirective struct {
	SourceDirective
}

type ImgSrcDirective struct {
	SourceDirective
}

type ManifestSrcDirective struct {
	SourceDirective
}

type MediaSrcDirective struct {
	SourceDirective
}

type ObjectSrcDirective struct {
	SourceDirective
}

type ScriptSrcAttrDirective struct {
	SourceDirective
}

type ScriptSrcDirective struct {
	SourceDirective
}

type ScriptSrcElemDirective struct {
	SourceDirective
}

type StyleSrcAttrDirective struct {
	SourceDirective
}

type StyleSrcDirective struct {
	SourceDirective
}

type StyleSrcElemDirective struct {
	SourceDirective
}

type WorkerSrcDirective struct {
	SourceDirective
}

type SandboxDirective struct {
	Directive
	Tokens []string `json:"tokens,omitempty"`
}

func (sandboxDirective *SandboxDirective) String() string {
	name := sandboxDirective.GetName()

	if len(sandboxDirective.Tokens) == 0 {
		return name
	}

	return fmt.Sprintf("%s %s", name, strings.Join(sandboxDirective.Tokens, " "))
}

type WebrtcDirective struct {
	Directive
	Value string `json:"value,omitempty"`
}

func (webrtcDirective *WebrtcDirective) String() string {
	value := webrtcDirective.Value
	if value == "" {
		return ""
	}

	return webrtcDirective.GetName() + " " + fmt.Sprintf("'%s'", value)
}

type ReportUriDirective struct {
	Directive
	UriReferences []string `json:"uri_references,omitempty"`
}

func (reportUriDirective *ReportUriDirective) String() string {
	if len(reportUriDirective.UriReferences) == 0 {
		return ""
	}

	return fmt.Sprintf(
		"%s %s",
		reportUriDirective.GetName(),
		strings.Join(reportUriDirective.UriReferences, " "),
	)
}

type ReportToDirective struct {
	Directive
	Token string `json:"token,omitempty"`
}

func (reportToDirective *ReportToDirective) String() string {
	token := reportToDirective.Token
	if token == "" {
		return ""
	}

	return reportToDirective.GetName() + " " + token
}

type FrameAncestorsDirective struct {
	SourceDirective
}

type UpgradeInsecureRequestDirective struct {
	Directive
}

func (upgradeInsecureRequestDirective *UpgradeInsecureRequestDirective) String() string {
	return upgradeInsecureRequestDirective.GetName()
}

type RequireSriForDirective struct {
	Directive
	ResourceTypes []string `json:"resource_types,omitempty"`
}

func (requireSriForDirective *RequireSriForDirective) String() string {
	resourcesTypes := requireSriForDirective.ResourceTypes
	if len(resourcesTypes) == 0 {
		return ""
	}

	return fmt.Sprintf("%s %s", requireSriForDirective.GetName(), strings.Join(resourcesTypes, " "))
}

type TrustedTypeExpression struct {
	Kind  string `json:"kind,omitempty"`
	Value string `json:"value,omitempty"`
}

type TrustedTypesDirective struct {
	Directive
	Expressions []TrustedTypeExpression `json:"expressions,omitempty"`
}

func (trustedTypesDirective *TrustedTypesDirective) String() string {
	expressions := trustedTypesDirective.Expressions
	if len(expressions) == 0 {
		return ""
	}

	var expressionStrings []string
	for _, expression := range expressions {
		kind := expression.Kind
		value := expression.Value
		if kind == "" || value == "" {
			continue
		}

		if kind == "keyword" {
			value = fmt.Sprintf("'%s'", value)
		}

		expressionStrings = append(expressionStrings, value)
	}

	return fmt.Sprintf("%s %s", trustedTypesDirective.GetName(), strings.Join(expressionStrings, " "))
}

type RequireTrustedTypesForDirective struct {
	Directive
	SinkGroups []string `json:"sink_groups,omitempty"`
}

func (requireTrustedTypesForDirective *RequireTrustedTypesForDirective) String() string {
	sinkGroups := requireTrustedTypesForDirective.SinkGroups
	if len(sinkGroups) == 0 {
		return ""
	}

	var sinkGroupStrings []string
	for _, sinkGroup := range sinkGroups {
		if sinkGroup == "" {
			continue
		}
		sinkGroupStrings = append(sinkGroupStrings, fmt.Sprintf("'%s'", sinkGroup))
	}

	return fmt.Sprintf("%s %s", requireTrustedTypesForDirective.GetName(), strings.Join(sinkGroupStrings, " "))
}

type ContentSecurityPolicy struct {
	Directives            []DirectiveI `json:"directives"`
	OtherDirectives       []DirectiveI `json:"other_directives"`
	IneffectiveDirectives []DirectiveI `json:"ineffective_directives"`
	Raw                   string       `json:"raw,omitempty"`
}

func (csp *ContentSecurityPolicy) GetDirective(name string) (DirectiveI, bool) {
	for _, directive := range csp.Directives {
		if directive.GetName() == name {
			return directive, true
		}
	}

	for _, directive := range csp.OtherDirectives {
		if directive.GetName() == name {
			return directive, true
		}
	}

	return nil, false
}

func (csp *ContentSecurityPolicy) String() string {
	directives := csp.Directives
	otherDirectives := csp.OtherDirectives

	if len(directives) == 0 && len(otherDirectives) == 0 {
		return ""
	}

	var policies []string
	for _, directive := range slices.Concat(directives, otherDirectives) {
		if utils.IsNil(directive) {
			continue
		}

		if policyString := directive.String(); policyString != "" {
			policies = append(policies, policyString)
		}
	}

	if len(policies) == 0 {
		return ""
	}

	return strings.Join(policies, "; ")
}
