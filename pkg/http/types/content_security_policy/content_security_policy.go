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
	Directive
	Sources []SourceI `json:"sources,omitempty"`
}

type ChildSrcDirective struct {
	Directive
	Sources []SourceI `json:"sources,omitempty"`
}

type ConnectSrcDirective struct {
	Directive
	Sources []SourceI `json:"sources,omitempty"`
}

type DefaultSrcDirective struct {
	Directive
	Sources []SourceI `json:"sources,omitempty"`
}

type FontSrcDirective struct {
	Directive
	Sources []SourceI `json:"sources,omitempty"`
}

type FormActionDirective struct {
	Directive
	Sources []SourceI `json:"sources,omitempty"`
}

type FrameSrcDirective struct {
	Directive
	Sources []SourceI `json:"sources,omitempty"`
}

type ImgSrcDirective struct {
	Directive
	Sources []SourceI `json:"sources,omitempty"`
}

type ManifestSrcDirective struct {
	Directive
	Sources []SourceI `json:"sources,omitempty"`
}

type MediaSrcDirective struct {
	Directive
	Sources []SourceI `json:"sources,omitempty"`
}

type ObjectSrcDirective struct {
	Directive
	Sources []SourceI `json:"sources,omitempty"`
}

type ScriptSrcAttrDirective struct {
	Directive
	Sources []SourceI `json:"sources,omitempty"`
}

type ScriptSrcDirective struct {
	Directive
	Sources []SourceI `json:"sources,omitempty"`
}

type ScriptSrcElemDirective struct {
	Directive
	Sources []SourceI `json:"sources,omitempty"`
}

type StyleSrcAttrDirective struct {
	Directive
	Sources []SourceI `json:"sources,omitempty"`
}

type StyleSrcDirective struct {
	Directive
	Sources []SourceI `json:"sources,omitempty"`
}

type StyleSrcElemDirective struct {
	Directive
	Sources []SourceI `json:"sources,omitempty"`
}

type WorkerSrcDirective struct {
	Directive
	Sources []SourceI `json:"sources,omitempty"`
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
	Directive
	Sources []SourceI `json:"sources,omitempty"`
}

type ContentSecurityPolicy struct {
	Directives      []DirectiveI `json:"directives"`
	OtherDirectives []DirectiveI `json:"other_directives"`
	Raw             string       `json:"raw,omitempty"`
}
