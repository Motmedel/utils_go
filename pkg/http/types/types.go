package types

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/schema"
	motmedelTlsTypes "github.com/Motmedel/utils_go/pkg/tls/types"
)

type HttpContext struct {
	Request      *http.Request
	RequestBody  []byte
	Response     *http.Response
	ResponseBody []byte
	Reporting    *schema.HttpReporting
	TlsContext   *motmedelTlsTypes.TlsContext
	User         *schema.User
	Extra        []*HttpContext
	LocalAddr    net.Addr
	RemoteAddr   net.Addr
}

func getFullType(typeValue string, subtypeValue string, normalize bool) string {
	if typeValue == "" {
		typeValue = "*"
	}
	if subtypeValue == "" {
		subtypeValue = "*"
	}

	fullType := fmt.Sprintf("%s/%s", typeValue, subtypeValue)
	if normalize {
		return strings.ToLower(fullType)
	}

	return fullType
}

func getParameterMap(parameters [][2]string, normalize bool) map[string]string {
	if len(parameters) == 0 {
		return nil
	}

	parameterMap := make(map[string]string)

	for _, parameter := range parameters {
		key := parameter[0]
		if normalize {
			key = strings.ToLower(key)
		}
		value := parameter[1]

		if _, ok := parameterMap[key]; !ok {
			parameterMap[key] = value
		}
	}

	return parameterMap
}

func getStructuredSyntaxName(subtype string, normalize bool) string {
	if subtype == "" {
		return ""
	}

	separator := "+"

	lastSeparatorIndex := strings.LastIndex(subtype, separator)
	if lastSeparatorIndex == -1 {
		return ""
	}

	structuredSyntaxName := subtype[lastSeparatorIndex+len(separator):]
	if normalize {
		structuredSyntaxName = strings.ToLower(structuredSyntaxName)
	}

	return structuredSyntaxName
}

type MediaRange struct {
	Type       string
	Subtype    string
	Parameters [][2]string
	Weight     float32
}

func (mediaRange *MediaRange) GetFullType(normalize bool) string {
	return getFullType(mediaRange.Type, mediaRange.Subtype, normalize)
}

func (mediaRange *MediaRange) GetParameterMap(normalize bool) map[string]string {
	parameters := mediaRange.Parameters
	if len(parameters) == 0 {
		return nil
	}

	return getParameterMap(parameters, normalize)
}

func (mediaRange *MediaRange) GetStructuredSyntaxName(normalize bool) string {
	return getStructuredSyntaxName(mediaRange.Subtype, normalize)
}

type ServerMediaRange struct {
	Type    string
	Subtype string
}

func (serverMediaRange *ServerMediaRange) GetFullType(normalize bool) string {
	return getFullType(serverMediaRange.Type, serverMediaRange.Subtype, normalize)
}

func (serverMediaRange *ServerMediaRange) GetStructuredSyntaxName(normalize bool) string {
	return getStructuredSyntaxName(serverMediaRange.Subtype, normalize)
}

type Accept struct {
	MediaRanges []*MediaRange
	Raw         string
}

func (accept *Accept) GetPriorityOrderedEncodings() []*MediaRange {
	mediaRanges := make([]*MediaRange, len(accept.MediaRanges))
	copy(mediaRanges, accept.MediaRanges)

	sort.SliceStable(mediaRanges, func(i, j int) bool {
		return mediaRanges[i].Weight > mediaRanges[j].Weight
	})

	return mediaRanges
}

type Authorization struct {
	Scheme  string
	Token68 string
	Params  map[string]string
}

func isHttpTokenRune(c byte) bool {
	if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' {
		return true
	}
	switch c {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	}
	return false
}

func isHttpToken(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if !isHttpTokenRune(s[i]) {
			return false
		}
	}
	return true
}

func quoteHttpString(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' || c == '\\' {
			b.WriteByte('\\')
		}
		b.WriteByte(c)
	}
	b.WriteByte('"')
	return b.String()
}

func (authorization *Authorization) String() string {
	if authorization.Scheme == "" {
		return ""
	}

	if authorization.Token68 != "" {
		return authorization.Scheme + " " + authorization.Token68
	}

	if len(authorization.Params) == 0 {
		return authorization.Scheme
	}

	keys := make([]string, 0, len(authorization.Params))
	for k := range authorization.Params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	params := make([]string, 0, len(keys))
	for _, k := range keys {
		v := authorization.Params[k]
		if !isHttpToken(v) {
			v = quoteHttpString(v)
		}
		params = append(params, k+"="+v)
	}

	return authorization.Scheme + " " + strings.Join(params, ", ")
}

type MediaType struct {
	Type       string
	Subtype    string
	Parameters [][2]string
}

func (mediaType *MediaType) GetFullType(normalize bool) string {
	return getFullType(mediaType.Type, mediaType.Subtype, normalize)
}

func (mediaType *MediaType) GetStructuredSyntaxName(normalize bool) string {
	return getStructuredSyntaxName(mediaType.Subtype, normalize)
}

func (mediaType *MediaType) GetParametersMap(normalize bool) map[string]string {
	if len(mediaType.Parameters) == 0 {
		return nil
	}

	return getParameterMap(mediaType.Parameters, normalize)
}

type ContentType struct {
	MediaType
}

type Encoding struct {
	Coding       string
	QualityValue float32
}

type AcceptEncoding struct {
	Encodings []*Encoding
	Raw       string
}

func (acceptEncoding *AcceptEncoding) GetPriorityOrderedEncodings() []*Encoding {
	encodings := make([]*Encoding, len(acceptEncoding.Encodings))
	copy(encodings, acceptEncoding.Encodings)

	sort.SliceStable(encodings, func(i, j int) bool {
		return encodings[i].QualityValue > encodings[j].QualityValue
	})

	return encodings
}

type LanguageTag struct {
	PrimarySubtag string
	Subtag        string
}

type LanguageQ struct {
	Tag *LanguageTag
	Q   float32
}
type AcceptLanguage struct {
	LanguageQs []*LanguageQ
	Raw        string
}

func (acceptLanguage *AcceptLanguage) GetPriorityOrderedLanguages() []*LanguageQ {
	languages := make([]*LanguageQ, len(acceptLanguage.LanguageQs))
	copy(languages, acceptLanguage.LanguageQs)

	sort.SliceStable(languages, func(i, j int) bool {
		return languages[i].Q > languages[j].Q
	})

	return languages
}

type StrictTransportSecurityPolicy struct {
	MaxAga            int
	IncludeSubdomains bool
	Raw               string
}

type RetryAfter struct {
	// The time can be either a timestamp or a duration.
	WaitTime any
	Raw      string
}

type ContentDisposition struct {
	DispositionType     string
	Filename            string
	FilenameAsterisk    string
	ExtensionParameters map[string]string
}

type ContentNegotiation struct {
	Accept         *Accept
	AcceptEncoding *AcceptEncoding
	AcceptLanguage *AcceptLanguage

	NegotiatedAccept         string
	NegotiatedAcceptEncoding string
}

type RobotsTxt struct {
	Groups []*RobotsTxtGroup
}

func (robotsTxt *RobotsTxt) String() string {
	var nonEmptyGroupStrings []string

	for _, group := range robotsTxt.Groups {
		if group == nil {
			continue
		}

		if groupString := group.String(); groupString != "" {
			nonEmptyGroupStrings = append(nonEmptyGroupStrings, groupString)
		}
	}

	return strings.Join(nonEmptyGroupStrings, "\n\n")
}

func makeLine(label string, value string, allowEmpty bool) string {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" && !allowEmpty {
		return ""
	}

	return fmt.Sprintf("%s: %s", label, trimmedValue)
}

func makePart(values []string, label string, allowEmpty bool) string {
	var parts []string
	for _, value := range values {
		if line := makeLine(label, value, allowEmpty); line != "" {
			parts = append(parts, line)
		}
	}

	return strings.Join(parts, "\n")
}

type RobotsTxtGroup struct {
	UserAgents   []string
	Disallowed   []string
	Allowed      []string
	OtherRecords [][2]string
}

func (robotsTxtGroup *RobotsTxtGroup) String() string {
	if len(robotsTxtGroup.UserAgents) == 0 {
		return ""
	}

	userAgentPart := makePart(robotsTxtGroup.UserAgents, "User-Agent", false)
	if userAgentPart == "" {
		return ""
	}

	parts := []string{userAgentPart}

	if disallowedPart := makePart(robotsTxtGroup.Disallowed, "Disallow", true); disallowedPart != "" {
		parts = append(parts, disallowedPart)
	}

	if allowedPart := makePart(robotsTxtGroup.Allowed, "Allow", false); allowedPart != "" {
		parts = append(parts, allowedPart)
	}

	for _, otherRecord := range robotsTxtGroup.OtherRecords {
		if line := makeLine(otherRecord[0], otherRecord[1], false); line != "" {
			parts = append(parts, line)
		}
	}

	return strings.Join(parts, "\n")
}

type SecurityTxt struct {
}

type CorsConfiguration struct {
	Origin        string
	Methods       []string
	Headers       []string
	Credentials   bool
	MaxAge        int
	ExposeHeaders []string
}

// ForwardedElement represents a single forwarded element containing multiple parameters.
// Standard parameters defined in RFC 7239 are:
//   - For: identifies the node making the request to the proxy
//   - By: identifies the interface where the request came in to the proxy
//   - Host: the original value of the Host request header
//   - Proto: indicates the protocol used to make the request (http or https)
type ForwardedElement struct {
	For   string
	By    string
	Host  string
	Proto string
	// Extensions contain any non-standard parameters
	Extensions map[string]string
}

// Forwarded represents the parsed Forwarded HTTP header as defined in RFC 7239.
// The header can contain multiple elements, each potentially originating from
// different proxies in the request chain.
type Forwarded struct {
	Elements []*ForwardedElement
}

type ETag struct {
	Weak bool
	Tag  string
}

func (etag *ETag) String() string {
	if etag == nil {
		return ""
	}

	var b strings.Builder
	b.Grow(len(etag.Tag) + 4)
	if etag.Weak {
		b.WriteString("W/")
	}
	b.WriteByte('"')
	b.WriteString(etag.Tag)
	b.WriteByte('"')
	return b.String()
}

type CacheControlDirective struct {
	Name  string
	Value string
}

type CacheControl struct {
	Directives []*CacheControlDirective
	Raw        string
}

func (cacheControl *CacheControl) findDirective(name string) *CacheControlDirective {
	for _, directive := range cacheControl.Directives {
		if directive.Name == name {
			return directive
		}
	}
	return nil
}

func (cacheControl *CacheControl) hasDirective(name string) bool {
	return cacheControl.findDirective(name) != nil
}

var ErrDirectiveNotPresent = errors.New("directive not present")

func (cacheControl *CacheControl) deltaSeconds(name string) (int, error) {
	directive := cacheControl.findDirective(name)
	if directive == nil {
		return 0, motmedelErrors.NewWithTrace(ErrDirectiveNotPresent)
	}

	value, err := strconv.Atoi(directive.Value)
	if err != nil {
		return 0, motmedelErrors.NewWithTrace(fmt.Errorf("strconv atoi: %w", err), directive.Value)
	}

	return value, nil
}

func splitFieldNames(value string) []string {
	parts := strings.Split(value, ",")
	var result []string
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// Request and response directives.

func (cacheControl *CacheControl) MaxAge() (int, error) {
	return cacheControl.deltaSeconds("max-age")
}

func (cacheControl *CacheControl) NoCache() bool {
	return cacheControl.hasDirective("no-cache")
}

func (cacheControl *CacheControl) NoCacheFieldNames() []string {
	directive := cacheControl.findDirective("no-cache")
	if directive == nil || directive.Value == "" {
		return nil
	}

	return splitFieldNames(directive.Value)
}

func (cacheControl *CacheControl) NoStore() bool {
	return cacheControl.hasDirective("no-store")
}

func (cacheControl *CacheControl) NoTransform() bool {
	return cacheControl.hasDirective("no-transform")
}

// Request-only directives.

func (cacheControl *CacheControl) MaxStale() (value int, hasValue bool, err error) {
	directive := cacheControl.findDirective("max-stale")
	if directive == nil {
		return 0, false, fmt.Errorf("max-stale: %w", ErrDirectiveNotPresent)
	}
	if directive.Value == "" {
		return 0, false, nil
	}

	v, err := strconv.Atoi(directive.Value)
	if err != nil {
		return 0, true, fmt.Errorf("invalid delta-seconds for max-stale: %w", err)
	}

	return v, true, nil
}

func (cacheControl *CacheControl) MinFresh() (int, error) {
	return cacheControl.deltaSeconds("min-fresh")
}

func (cacheControl *CacheControl) OnlyIfCached() bool {
	return cacheControl.hasDirective("only-if-cached")
}

// Response-only directives.

func (cacheControl *CacheControl) MustRevalidate() bool {
	return cacheControl.hasDirective("must-revalidate")
}

func (cacheControl *CacheControl) MustUnderstand() bool {
	return cacheControl.hasDirective("must-understand")
}

func (cacheControl *CacheControl) Private() bool {
	return cacheControl.hasDirective("private")
}

func (cacheControl *CacheControl) PrivateFieldNames() []string {
	directive := cacheControl.findDirective("private")
	if directive == nil || directive.Value == "" {
		return nil
	}
	return splitFieldNames(directive.Value)
}

func (cacheControl *CacheControl) ProxyRevalidate() bool {
	return cacheControl.hasDirective("proxy-revalidate")
}

func (cacheControl *CacheControl) Public() bool {
	return cacheControl.hasDirective("public")
}

func (cacheControl *CacheControl) SMaxAge() (int, error) {
	return cacheControl.deltaSeconds("s-maxage")
}
