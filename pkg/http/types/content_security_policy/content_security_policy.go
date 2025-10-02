package content_security_policy

import (
	"fmt"
	"strings"
)

type SourceI interface {
	GetRaw() string
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

type SchemeSource struct {
	Source
	Scheme string `json:"scheme,omitempty"`
}

type HostSource struct {
	Source
	Scheme     string `json:"scheme,omitempty"`
	Host       string `json:"host,omitempty"`
	PortString string `json:"port_string,omitempty"`
	Path       string `json:"path,omitempty"`
}

type KeywordSource struct {
	Source
	Keyword string `json:"keyword,omitempty"`
}

type NonceSource struct {
	Source
	Base64Value string `json:"base64_value,omitempty"`
}

type HashSource struct {
	Source
	HashAlgorithm string `json:"hash_algorithm,omitempty"`
	Base64Value   string `json:"base64_value,omitempty"`
}

type DirectiveI interface {
	GetName() string
	GetRawName() string
	GetRawValue() string
}

type Directive struct {
	Name     string `json:"name,omitempty"`
	RawName  string `json:"raw_name,omitempty"`
	RawValue string `json:"raw_value,omitempty"`
}

type SourceDirectiveI interface {
	GetSources() []SourceI
}

type SourceDirective struct {
	Directive
	Sources []SourceI `json:"sources,omitempty"`
}

func (sourceDirective *SourceDirective) GetSources() []SourceI {
	return sourceDirective.Sources
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

type WebrtcDirective struct {
	Directive
}

type ReportUriDirective struct {
	Directive
	UriReferences []string `json:"uri_references,omitempty"`
}

type ReportToDirective struct {
	Directive
	Token string `json:"token,omitempty"`
}

type FrameAncestorsDirective struct {
	SourceDirective
}

type UpgradeInsecureRequestDirective struct {
	Directive
}

type RequireSriForDirective struct {
	Directive
	ResourceTypes []string `json:"resource_types,omitempty"`
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

// String returns a serialized Content-Security-Policy header built from the instance.
// Serialize directives in the following order: Directives, OtherDirectives (exclude IneffectiveDirectives).
func (csp *ContentSecurityPolicy) String() string {
	var parts []string
	appendDirectives := func(list []DirectiveI) {
		for _, d := range list {
			if d == nil {
				continue
			}
			if s := directiveToString(d); s != "" {
				parts = append(parts, s)
			}
		}
	}
	appendDirectives(csp.Directives)
	appendDirectives(csp.OtherDirectives)
	return strings.Join(parts, "; ")
}

func directiveToString(d DirectiveI) string {
	if d == nil {
		return ""
	}
	name := d.GetRawName()
	if strings.TrimSpace(name) == "" {
		name = d.GetName()
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}

	// If we have a raw value, prefer it (preserves exact spacing/quoting)
	if v := strings.TrimSpace(d.GetRawValue()); v != "" {
		return fmt.Sprintf("%s %s", name, v)
	}

	// Try to serialize based on concrete directive types
	var value string

	// Directives with sources
	if sd, ok := d.(SourceDirectiveI); ok {
		sources := sd.GetSources()
		if len(sources) == 1 {
			if _, isNone := sources[0].(*NoneSource); isNone {
				value = "'none'"
			}
		}
		if value == "" {
			var srcParts []string
			for _, s := range sources {
				if s == nil {
					continue
				}
				if ss := sourceToString(s); ss != "" {
					srcParts = append(srcParts, ss)
				}
			}
			value = strings.Join(srcParts, " ")
		}
	}

	switch v := d.(type) {
	case *SandboxDirective:
		value = strings.Join(v.Tokens, " ")
	case *ReportUriDirective:
		value = strings.Join(v.UriReferences, " ")
	case *ReportToDirective:
		value = strings.TrimSpace(v.Token)
	case *RequireSriForDirective:
		value = strings.Join(v.ResourceTypes, " ")
	case *UpgradeInsecureRequestDirective:
		// no value
	case *WebrtcDirective:
		// if no raw value, cannot infer; keep empty
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return name
	}
	return fmt.Sprintf("%s %s", name, value)
}

func sourceToString(s SourceI) string {
	if s == nil {
		return ""
	}
	if raw := strings.TrimSpace(s.GetRaw()); raw != "" {
		return raw
	}
	switch v := s.(type) {
	case *NoneSource:
		return "'none'"
	case *SchemeSource:
		if v.Scheme == "" {
			return ""
		}
		return v.Scheme + ":"
	case *HostSource:
		if v.Host == "" && v.Scheme == "" && v.PortString == "" && v.Path == "" {
			return ""
		}
		b := strings.Builder{}
		if v.Scheme != "" {
			b.WriteString(v.Scheme)
			b.WriteString("://")
		}
		b.WriteString(v.Host)
		if v.PortString != "" {
			b.WriteString(":")
			b.WriteString(v.PortString)
		}
		if v.Path != "" {
			b.WriteString(v.Path)
		}
		return b.String()
	case *KeywordSource:
		return v.Keyword
	case *NonceSource:
		if v.Base64Value == "" {
			return ""
		}
		return "'nonce-" + v.Base64Value + "'"
	case *HashSource:
		if v.HashAlgorithm == "" || v.Base64Value == "" {
			return ""
		}
		return "'" + v.HashAlgorithm + "-" + v.Base64Value + "'"
	default:
		return ""
	}
}
