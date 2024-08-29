package content_security_policy


type SourceI interface {
	GetRaw() string
}

type Source struct {
	Raw string
}

func (source *Source) GetRaw() string {
	return source.Raw
}

type NoneSource struct {
	Source
}

type SchemeSource struct {
	Source
	Scheme string
}

type HostSource struct {
	Source
	Scheme     string
	Host       string
	PortString string
	Path       string
}

type KeywordSource struct {
	Source
	Keyword string
}

type NonceSource struct {
	Source
	Base64Value string
}

type HashSource struct {
	Source
	HashAlgorithm string
	Base64Value   string
}

type DirectiveI interface {
	GetName() string
	GetRawName() string
	GetRawValue() string
}

type Directive struct {
	Name     string
	RawName  string
	RawValue string
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
	Sources []SourceI
}

type ChildSrcDirective struct {
	Directive
	Sources []SourceI
}

type ConnectSrcDirective struct {
	Directive
	Sources []SourceI
}

type DefaultSrcDirective struct {
	Directive
	Sources []SourceI
}

type FontSrcDirective struct {
	Directive
	Sources []SourceI
}

type FormActionDirective struct {
	Directive
	Sources []SourceI
}

type FrameSrcDirective struct {
	Directive
	Sources []SourceI
}

type ImgSrcDirective struct {
	Directive
	Sources []SourceI
}

type ManifestSrcDirective struct {
	Directive
	Sources []SourceI
}

type MediaSrcDirective struct {
	Directive
	Sources []SourceI
}

type ObjectSrcDirective struct {
	Directive
	Sources []SourceI
}

type ScriptSrcAttrDirective struct {
	Directive
	Sources []SourceI
}

type ScriptSrcDirective struct {
	Directive
	Sources []SourceI
}

type ScriptSrcElemDirective struct {
	Directive
	Sources []SourceI
}

type StyleSrcAttrDirective struct {
	Directive
	Sources []SourceI
}

type StyleSrcDirective struct {
	Directive
	Sources []SourceI
}

type StyleSrcElemDirective struct {
	Directive
	Sources []SourceI
}

type WorkerSrcDirective struct {
	Directive
	Sources []SourceI
}

type SandboxDirective struct {
	Directive
	Tokens []string
}

type WebrtcDirective struct {
	Directive
}

type ReportUriDirective struct {
	Directive
	UriReferences []string
}

type ReportToDirective struct {
	Directive
	Token string
}

type FrameAncestorsDirective struct {
	Directive
	Sources []SourceI
}

type ContentSecurityPolicy struct {
	Directives      []DirectiveI
	OtherDirectives []DirectiveI
	Raw             string
}
