package content_security_policy

import (
	"fmt"
	"strings"
)

// extractDirectiveValue finds the full directive string from the original policy
// that matches the effective directive name. This produces output like
// "script-src 'self' 'nonce-abc'" rather than just "script-src-elem".
func extractDirectiveValue(effectiveDirective string, originalPolicy string) string {
	if originalPolicy == "" {
		return effectiveDirective
	}

	for part := range strings.SplitSeq(originalPolicy, ";") {
		trimmed := strings.TrimSpace(part)
		if trimmed == effectiveDirective || strings.HasPrefix(trimmed, effectiveDirective+" ") {
			return trimmed
		}
	}

	for part := range strings.SplitSeq(originalPolicy, ";") {
		trimmed := strings.TrimSpace(part)
		if trimmed == DirectiveNameDefaultSrc || strings.HasPrefix(trimmed, DirectiveNameDefaultSrc+" ") {
			return trimmed
		}
	}

	return effectiveDirective
}

func isStyleDirective(directive string) bool {
	switch directive {
	case DirectiveNameStyleSrc, DirectiveNameStyleSrcElem, DirectiveNameStyleSrcAttr:
		return true
	}
	return false
}

func isScriptDirective(directive string) bool {
	switch directive {
	case DirectiveNameScriptSrc, DirectiveNameScriptSrcElem:
		return true
	}
	return false
}

func cspDirectiveResourceDescription(directive string) string {
	switch directive {
	case DirectiveNameImgSrc:
		return "an image"
	case DirectiveNameFontSrc:
		return "a font"
	case DirectiveNameConnectSrc:
		return "a connection"
	case DirectiveNameMediaSrc:
		return "media"
	case DirectiveNameObjectSrc:
		return "an object"
	case DirectiveNameFrameSrc:
		return "a frame"
	case DirectiveNameChildSrc:
		return "a child resource"
	case DirectiveNameManifestSrc:
		return "a manifest"
	case DirectiveNameBaseUri:
		return "a base URI"
	case DirectiveNameFormAction:
		return "a form action"
	default:
		return "a resource"
	}
}

func cspViolationMessage(effectiveDirective, originalPolicy, blockedURL string, reportOnly bool) string {
	directiveValue := extractDirectiveValue(effectiveDirective, originalPolicy)

	var prefix, blocked string
	if reportOnly {
		prefix = "(Report-Only policy) "
		blocked = "would block"
	} else {
		blocked = "blocked"
	}

	violates := fmt.Sprintf(`because it violates the following directive: "%s"`, directiveValue)

	switch {
	// Trusted Types sink assignment (require-trusted-types-for)
	case effectiveDirective == DirectiveNameRequireTrustedTypesFor:
		return fmt.Sprintf(
			`%sThe page's settings %s assigning to an injection sink because it violates the following directive: "require-trusted-types-for 'script'"`,
			prefix, blocked,
		)

	// Trusted Types policy creation
	case effectiveDirective == DirectiveNameTrustedTypes:
		return fmt.Sprintf(
			"%sThe page's settings %s creating a Trusted Types policy %s",
			prefix, blocked, violates,
		)

	// Inline violations
	case blockedURL == "inline":
		switch {
		case isStyleDirective(effectiveDirective):
			return fmt.Sprintf(
				"%sThe page's settings %s an inline style from being applied %s",
				prefix, blocked, violates,
			)
		case effectiveDirective == DirectiveNameScriptSrcAttr:
			return fmt.Sprintf(
				"%sThe page's settings %s an event handler from being executed %s",
				prefix, blocked, violates,
			)
		default:
			return fmt.Sprintf(
				"%sThe page's settings %s an inline script from being executed %s",
				prefix, blocked, violates,
			)
		}

	// JavaScript eval
	case blockedURL == "eval":
		return fmt.Sprintf(
			"%sThe page's settings %s a JavaScript eval from being executed %s (Missing 'unsafe-eval')",
			prefix, blocked, violates,
		)

	// WebAssembly
	case blockedURL == "wasm-eval":
		return fmt.Sprintf(
			"%sThe page's settings %s WebAssembly from being executed %s (Missing 'wasm-unsafe-eval' or 'unsafe-eval')",
			prefix, blocked, violates,
		)

	// External resource with a URL
	case blockedURL != "":
		switch {
		case isStyleDirective(effectiveDirective):
			return fmt.Sprintf(
				"%sThe page's settings %s a style at %s from being applied %s",
				prefix, blocked, blockedURL, violates,
			)
		case isScriptDirective(effectiveDirective) || effectiveDirective == DirectiveNameScriptSrcAttr:
			return fmt.Sprintf(
				"%sThe page's settings %s a script at %s from being executed %s",
				prefix, blocked, blockedURL, violates,
			)
		case effectiveDirective == DirectiveNameWorkerSrc:
			return fmt.Sprintf(
				"%sThe page's settings %s a worker script at %s from being executed %s",
				prefix, blocked, blockedURL, violates,
			)
		default:
			description := cspDirectiveResourceDescription(effectiveDirective)
			return fmt.Sprintf(
				"%sThe page's settings %s the loading of %s at %s %s",
				prefix, blocked, description, blockedURL, violates,
			)
		}

	// No blocked URL - fallback to generic messages
	default:
		if reportOnly {
			return fmt.Sprintf(
				`A violation occurred for a report-only CSP policy ("%s"). The behavior was allowed, and a CSP report was sent.`,
				directiveValue,
			)
		}
		return fmt.Sprintf(
			"The page's settings blocked the loading of a resource: %s",
			directiveValue,
		)
	}
}

// CspViolationMessage produces a Firefox-style console message from a CSP
// violation report body. The message format follows the strings defined in
// Firefox's dom/chrome/security/csp.properties.

// Report is the deprecated report body sent via the report-uri directive
// (CSP Level 3 section 5.3). JSON keys use hyphenated names.
type Report struct {
	DocumentURI        string  `json:"document-uri,omitempty"`
	Referrer           *string `json:"referrer,omitempty" jsonschema:"referrer,optional,minlength:0"`
	ViolatedDirective  string  `json:"violated-directive,omitempty"`
	EffectiveDirective string  `json:"effective-directive,omitempty"`
	OriginalPolicy     string  `json:"original-policy,omitempty"`
	BlockedUri         string  `json:"blocked-uri,omitempty" jsonschema:"blocked-uri,optional,minlength:0"`
	Disposition        string  `json:"disposition,omitempty"`
	StatusCode         int     `json:"status-code,omitempty"`
	Sample             *string `json:"sample,omitempty" jsonschema:"sample,optional,minlength:0"`
	SourceFile         *string `json:"source-file,omitempty"`
	LineNumber         *int    `json:"line-number,omitempty"`
	ColumnNumber       *int    `json:"column-number,omitempty"`
}

func (r *Report) Message() string {
	return cspViolationMessage(r.EffectiveDirective, r.OriginalPolicy, r.BlockedUri, r.Disposition == "report")
}

// ReportEnvelope is the outer JSON object sent by browsers via the report-uri
// directive (CSP Level 2 §4.4). It wraps the violation report in a "csp-report" key.
type ReportEnvelope struct {
	CspReport *Report `json:"csp-report"`
}

func (e *ReportEnvelope) Message() string {
	if e.CspReport == nil {
		return ""
	}
	return e.CspReport.Message()
}

// CSPViolationReportBody is the report body for "csp-violation" reports sent
// via the Reporting API (report-to directive). Defined in CSP Level 3 section 5.
type CSPViolationReportBody struct {
	// The body is written by the browser, which is free to say more about what it blocked than
	// what is specified today. What it adds is not worth refusing the report over.
	//nolint:revive // The blank field is what carries what holds for the object; it is read from the type.
	_ struct{} `jsonschema:",additionalProperties:true"`

	DocumentURL        string  `json:"documentURL,omitempty"`
	Referrer           *string `json:"referrer,omitempty" jsonschema:"referrer,optional,minlength:0"`
	BlockedURL         string  `json:"blockedURL,omitempty"`
	EffectiveDirective string  `json:"effectiveDirective,omitempty"`
	OriginalPolicy     string  `json:"originalPolicy,omitempty"`
	SourceFile         *string `json:"sourceFile,omitempty" jsonschema:"sourceFile,optional,minlength:0"`
	Sample             *string `json:"sample,omitempty" jsonschema:"sample,optional,minlength:0"`
	Disposition        string  `json:"disposition,omitempty"`
	StatusCode         int     `json:"statusCode,omitempty"`
	LineNumber         *int    `json:"lineNumber,omitempty"`
	ColumnNumber       *int    `json:"columnNumber,omitempty"`
}

func (body *CSPViolationReportBody) Message() string {
	return cspViolationMessage(body.EffectiveDirective, body.OriginalPolicy, body.BlockedURL, body.Disposition == "report")
}

// CSPHashReportBody is the report body for "csp-hash" reports sent via the
// Reporting API. Defined in CSP Level 3 section 4.1.4.
type CSPHashReportBody struct {
	DocumentURL    string `json:"documentURL,omitempty"`
	SubresourceURL string `json:"subresourceURL,omitempty"`
	Hash           string `json:"hash,omitempty"`
	Destination    string `json:"destination,omitempty"`
	Type           string `json:"type,omitempty"`
}
