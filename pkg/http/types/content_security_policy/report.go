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

	for _, part := range strings.Split(originalPolicy, ";") {
		trimmed := strings.TrimSpace(part)
		if trimmed == effectiveDirective || strings.HasPrefix(trimmed, effectiveDirective+" ") {
			return trimmed
		}
	}

	for _, part := range strings.Split(originalPolicy, ";") {
		trimmed := strings.TrimSpace(part)
		if trimmed == "default-src" || strings.HasPrefix(trimmed, "default-src ") {
			return trimmed
		}
	}

	return effectiveDirective
}

func isStyleDirective(directive string) bool {
	switch directive {
	case "style-src", "style-src-elem", "style-src-attr":
		return true
	}
	return false
}

func isScriptDirective(directive string) bool {
	switch directive {
	case "script-src", "script-src-elem":
		return true
	}
	return false
}

func cspDirectiveResourceDescription(directive string) string {
	switch directive {
	case "img-src":
		return "an image"
	case "font-src":
		return "a font"
	case "connect-src":
		return "a connection"
	case "media-src":
		return "media"
	case "object-src":
		return "an object"
	case "frame-src":
		return "a frame"
	case "child-src":
		return "a child resource"
	case "manifest-src":
		return "a manifest"
	case "base-uri":
		return "a base URI"
	case "form-action":
		return "a form action"
	default:
		return "a resource"
	}
}

// CspViolationMessage produces a Firefox-style console message from a CSP
// violation report body. The message format follows the strings defined in
// Firefox's dom/chrome/security/csp.properties.

// Report is the deprecated report body sent via the report-uri directive
// (CSP Level 3 section 5.3). JSON keys use hyphenated names.
type Report struct {
	DocumentURI        string `json:"document-uri,omitempty" required:"true"`
	Referrer           string `json:"referrer,omitempty" required:"true"`
	ViolatedDirective  string `json:"violated-directive,omitempty" required:"true"`
	EffectiveDirective string `json:"effective-directive,omitempty"`
	OriginalPolicy     string `json:"original-policy,omitempty" required:"true"`
	BlockedUri         string `json:"blocked-uri,omitempty" required:"true"`
	LineNumber         int    `json:"line-number,omitempty"`
	ColumnNumber       int    `json:"column-number,omitempty"`
	SourceFile         string `json:"source-file,omitempty"`
}

// CSPViolationReportBody is the report body for "csp-violation" reports sent
// via the Reporting API (report-to directive). Defined in CSP Level 3 section 5.
type CSPViolationReportBody struct {
	DocumentURL        string `json:"documentURL,omitempty"`
	Referrer           string `json:"referrer,omitempty"`
	BlockedURL         string `json:"blockedURL,omitempty"`
	EffectiveDirective string `json:"effectiveDirective,omitempty"`
	OriginalPolicy     string `json:"originalPolicy,omitempty"`
	SourceFile         string `json:"sourceFile,omitempty"`
	Sample             string `json:"sample,omitempty"`
	Disposition        string `json:"disposition,omitempty"`
	StatusCode         int    `json:"statusCode,omitempty"`
	LineNumber         *int   `json:"lineNumber,omitempty"`
	ColumnNumber       *int   `json:"columnNumber,omitempty"`
}

func (body *CSPViolationReportBody) Message() string {
	effectiveDirective := body.EffectiveDirective
	directiveValue := extractDirectiveValue(effectiveDirective, body.OriginalPolicy)
	blockedURL := body.BlockedURL
	reportOnly := body.Disposition == "report"

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
	case effectiveDirective == "require-trusted-types-for":
		return fmt.Sprintf(
			`%sThe page's settings %s assigning to an injection sink because it violates the following directive: "require-trusted-types-for 'script'"`,
			prefix, blocked,
		)

	// Trusted Types policy creation
	case effectiveDirective == "trusted-types":
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
		case effectiveDirective == "script-src-attr":
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
		case isScriptDirective(effectiveDirective) || effectiveDirective == "script-src-attr":
			return fmt.Sprintf(
				"%sThe page's settings %s a script at %s from being executed %s",
				prefix, blocked, blockedURL, violates,
			)
		case effectiveDirective == "worker-src":
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

// CSPHashReportBody is the report body for "csp-hash" reports sent via the
// Reporting API. Defined in CSP Level 3 section 4.1.4.
type CSPHashReportBody struct {
	DocumentURL    string `json:"documentURL,omitempty"`
	SubresourceURL string `json:"subresourceURL,omitempty"`
	Hash           string `json:"hash,omitempty"`
	Destination    string `json:"destination,omitempty"`
	Type           string `json:"type,omitempty"`
}
