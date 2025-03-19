package content_security_policy

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
