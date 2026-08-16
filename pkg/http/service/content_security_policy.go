package service

import (
	"fmt"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"github.com/Motmedel/utils_go/pkg/errors/types/nil_error"
	motmedelMux "github.com/Motmedel/utils_go/pkg/http/mux"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_writer"
	csp "github.com/Motmedel/utils_go/pkg/http/types/content_security_policy"
	cspUtils "github.com/Motmedel/utils_go/pkg/http/utils/content_security_policy"
)

const contentSecurityPolicyHeaderName = "Content-Security-Policy"

// ApiContentSecurityPolicy is what a service that serves no documents answers with. A policy is
// worth as much to it as to one that does -- what differs is which policy.
//
//   - default-src 'none' permits no subresource of any kind, there being no document to load one.
//   - base-uri and form-action do not fall back to default-src, so they are said in full.
//   - frame-ancestors 'none' keeps the response out of a frame.
//   - sandbox, with no value, applies every restriction there is: a response a browser is made to
//     render as a document -- by a content type it sniffed past, or by being navigated to directly
//     -- gets an opaque origin and runs no script.
const ApiContentSecurityPolicy = "default-src 'none'; base-uri 'none'; form-action 'none'; " +
	"frame-ancestors 'none'; sandbox"

// contentSecurityPolicyHeaders are the headers the service's policy is answered in: the ones every
// response carries, where the service answers with the policy for one that serves no documents, and
// the ones only a document carries otherwise. Which it is settles before anything patches the
// policy, so that what is patched is the policy that is actually answered with.
func contentSecurityPolicyHeaders(mux *motmedelMux.Mux) map[string]string {
	if mux == nil {
		return nil
	}

	if _, found := mux.DefaultHeaders[contentSecurityPolicyHeaderName]; found {
		return mux.DefaultHeaders
	}

	return mux.DefaultDocumentHeaders
}

// patchContentSecurityPolicy hands the policy the service answers with to patch, and writes back
// what it made of it.
func patchContentSecurityPolicy(
	mux *motmedelMux.Mux,
	patch func(*csp.ContentSecurityPolicy) error,
) error {
	if mux == nil {
		return motmedelErrors.NewWithTrace(nil_error.New("mux"))
	}

	if patch == nil {
		return motmedelErrors.NewWithTrace(nil_error.New("patch"))
	}

	headers := contentSecurityPolicyHeaders(mux)
	if headers == nil {
		return motmedelErrors.NewWithTrace(nil_error.NewWithInstance("map", "content security policy headers"))
	}

	contentSecurityPolicyString := headers[contentSecurityPolicyHeaderName]
	if contentSecurityPolicyString == "" {
		contentSecurityPolicyString = response_writer.DefaultContentSecurityPolicyString
	}

	contentSecurityPolicy, err := csp.Parse([]byte(contentSecurityPolicyString))
	if err != nil {
		return motmedelErrors.New(
			fmt.Errorf("content security policy parse: %w", err),
			contentSecurityPolicyString,
		)
	}
	if contentSecurityPolicy == nil {
		return motmedelErrors.NewWithTrace(nil_error.New("content security policy"))
	}

	if err := patch(contentSecurityPolicy); err != nil {
		return fmt.Errorf("patch: %w", err)
	}

	headers[contentSecurityPolicyHeaderName] = contentSecurityPolicy.String()

	return nil
}

// patchViewerStyleHashes permits the styles a browser's own viewer applies to a response it renders
// as a document -- Chrome's for XML, Edge's for PDF. Without them the viewer's styling is blocked
// by the policy, and what it renders comes out unstyled.
//
// This is not a matter for the services that serve documents alone: a service that serves none
// still answers a request it will not serve with a problem detail, which a browser asks for as XML
// and renders through the same viewer.
//
// What the two viewers take differs. Chrome's styles the document tree through style elements,
// whose bodies a hash source matches as it is. Edge's styles through style attributes, which a hash
// source reaches only where 'unsafe-hashes' is permitted with it -- so that is permitted for Edge's
// viewer alone, rather than for every service on the chance that it serves a PDF.
func patchViewerStyleHashes(mux *motmedelMux.Mux, chromeXmlViewer bool, edgePdfViewer bool) error {
	if !chromeXmlViewer && !edgePdfViewer {
		return nil
	}

	err := patchContentSecurityPolicy(
		mux,
		func(contentSecurityPolicy *csp.ContentSecurityPolicy) error {
			// 'self' is permitted first, for the stylesheets a viewer links rather than inlines.
			cspUtils.PatchCspStyleSrcWithKeyword(contentSecurityPolicy, "self")

			if edgePdfViewer {
				cspUtils.PatchCspStyleSrcWithKeyword(contentSecurityPolicy, "unsafe-hashes")
			}

			if chromeXmlViewer {
				err := cspUtils.PatchCspStyleSrcWithHash(
					contentSecurityPolicy,
					cspUtils.ChromeXmlViewerStyleHashes...,
				)
				if err != nil {
					return fmt.Errorf("patch csp style src with hash (chrome xml viewer): %w", err)
				}
			}

			if edgePdfViewer {
				err := cspUtils.PatchCspStyleSrcWithHash(
					contentSecurityPolicy,
					cspUtils.EdgePdfViewerStyleHashes...,
				)
				if err != nil {
					return fmt.Errorf("patch csp style src with hash (edge pdf viewer): %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("patch content security policy: %w", err)
	}

	return nil
}

// patchTrustedTypes requires the named trusted types policies of the scripts the documents run.
func patchTrustedTypes(mux *motmedelMux.Mux, policies ...string) error {
	if len(policies) == 0 {
		return nil
	}

	err := patchContentSecurityPolicy(
		mux,
		func(contentSecurityPolicy *csp.ContentSecurityPolicy) error {
			cspUtils.PatchCspTrustedTypes(contentSecurityPolicy, policies...)
			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("patch content security policy: %w", err)
	}

	return nil
}

// patchApiContentSecurityPolicy answers every response with the policy for a service that serves no
// documents, rather than answering the documents it does not serve with the policy for one.
//
// It settles which policy the service answers with, so it runs before anything that patches one:
// what follows patches this policy rather than the one for documents, which is dropped here.
func patchApiContentSecurityPolicy(mux *motmedelMux.Mux) error {
	if mux == nil {
		return motmedelErrors.NewWithTrace(nil_error.New("mux"))
	}

	defaultHeaders := mux.DefaultHeaders
	if defaultHeaders == nil {
		return motmedelErrors.NewWithTrace(nil_error.NewWithInstance("map", "default headers"))
	}

	defaultHeaders[contentSecurityPolicyHeaderName] = ApiContentSecurityPolicy

	// The document policy is dropped rather than left. Both header sets are written to a response
	// that is a document, so a policy in each would be two Content-Security-Policy headers, and a
	// browser enforces every policy it is sent -- the two together permitting only what both do.
	delete(mux.DefaultDocumentHeaders, contentSecurityPolicyHeaderName)

	return nil
}
