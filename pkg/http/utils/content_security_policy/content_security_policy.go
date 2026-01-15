package content_security_policy

import (
	"net/url"
	"slices"
	"strings"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	csp "github.com/Motmedel/utils_go/pkg/http/types/content_security_policy"
)

func buildHostSources(hostUrls ...*url.URL) []csp.SourceI {
	var hostSources []csp.SourceI
	for _, hostUrl := range hostUrls {
		if hostUrl == nil {
			continue
		}
		if hostSource := csp.HostSourceFromUrl(hostUrl); hostSource != nil {
			hostSources = append(hostSources, hostSource)
		}
	}
	return hostSources
}

func PatchCspConnectSrcWithHostSrc(contentSecurityPolicy *csp.ContentSecurityPolicy, hostUrls ...*url.URL) {
	if contentSecurityPolicy == nil {
		return
	}

	hostSources := buildHostSources(hostUrls...)

	if len(hostSources) == 0 {
		return
	}

	connectSrcDirective := &csp.ConnectSrcDirective{
		SourceDirective: csp.SourceDirective{
			Sources: slices.Concat(
				[]csp.SourceI{
					&csp.KeywordSource{Keyword: "self"},
				},
				hostSources,
			),
		},
	}

	if existingConnectSrcDirective := contentSecurityPolicy.GetConnectSrc(); existingConnectSrcDirective != nil {
		sourceMap := make(map[string]struct{})
		for _, source := range existingConnectSrcDirective.Sources {
			sourceMap[source.String()] = struct{}{}
		}

		for _, hostSource := range hostSources {
			if _, found := sourceMap[hostSource.String()]; !found {
				existingConnectSrcDirective.Sources = append(existingConnectSrcDirective.Sources, hostSource)
			}
		}
	} else {
		contentSecurityPolicy.Directives = append(contentSecurityPolicy.Directives, connectSrcDirective)
	}
}

func PatchCspFrameSrcWithHostSrc(contentSecurityPolicy *csp.ContentSecurityPolicy, hostUrls ...*url.URL) {
	if contentSecurityPolicy == nil {
		return
	}

	hostSources := buildHostSources(hostUrls...)

	if len(hostSources) == 0 {
		return
	}

	frameSrcDirective := &csp.FrameSrcDirective{
		SourceDirective: csp.SourceDirective{
			Sources: slices.Concat(
				[]csp.SourceI{
					&csp.KeywordSource{Keyword: "self"},
				},
				hostSources,
			),
		},
	}

	if existingFrameSrcDirective := contentSecurityPolicy.GetFrameSrc(); existingFrameSrcDirective != nil {
		sourceMap := make(map[string]struct{})
		for _, source := range existingFrameSrcDirective.Sources {
			sourceMap[source.String()] = struct{}{}
		}

		for _, hostSource := range hostSources {
			if _, found := sourceMap[hostSource.String()]; !found {
				existingFrameSrcDirective.Sources = append(existingFrameSrcDirective.Sources, hostSource)
			}
		}
	} else {
		contentSecurityPolicy.Directives = append(contentSecurityPolicy.Directives, frameSrcDirective)
	}
}

func PatchCspStyleSrcWithNonce(contentSecurityPolicy *csp.ContentSecurityPolicy, nonces ...string) {
	if contentSecurityPolicy == nil {
		return
	}

	var nonceSources []csp.SourceI
	for _, nonce := range nonces {
		if nonce == "" {
			continue
		}
		nonceSources = append(nonceSources, &csp.NonceSource{
			Base64Value: nonce,
		})
	}

	if len(nonceSources) == 0 {
		return
	}

	if existingStyleSrcDirective := contentSecurityPolicy.GetStyleSrc(); existingStyleSrcDirective != nil {
		sourceMap := make(map[string]struct{})
		for _, source := range existingStyleSrcDirective.Sources {
			sourceMap[source.String()] = struct{}{}
		}

		for _, nonceSource := range nonceSources {
			if _, found := sourceMap[nonceSource.String()]; !found {
				existingStyleSrcDirective.Sources = append(existingStyleSrcDirective.Sources, nonceSource)
			}
		}
	} else {
		styleSrcDirective := &csp.StyleSrcDirective{
			SourceDirective: csp.SourceDirective{
				Sources: nonceSources,
			},
		}
		contentSecurityPolicy.Directives = append(contentSecurityPolicy.Directives, styleSrcDirective)
	}
}

func PatchCspStyleSrcWithHash(contentSecurityPolicy *csp.ContentSecurityPolicy, values ...string) error {
	if contentSecurityPolicy == nil {
		return nil
	}

	var hashSources []csp.SourceI
	for _, value := range values {
		if value == "" {
			continue
		}

		hashAlgorithm, hash, found := strings.Cut(value, "-")
		if !found {
			return motmedelErrors.NewWithTrace(
				motmedelErrors.ErrBadSplit,
				value,
			)
		}

		hashSources = append(hashSources, &csp.HashSource{HashAlgorithm: hashAlgorithm, Base64Value: hash})
	}

	if len(hashSources) == 0 {
		return nil
	}

	if existingStyleSrcDirective := contentSecurityPolicy.GetStyleSrc(); existingStyleSrcDirective != nil {
		sourceMap := make(map[string]struct{})
		for _, source := range existingStyleSrcDirective.Sources {
			sourceMap[source.String()] = struct{}{}
		}

		for _, hashSource := range hashSources {
			if _, found := sourceMap[hashSource.String()]; !found {
				existingStyleSrcDirective.Sources = append(existingStyleSrcDirective.Sources, hashSource)
			}
		}
	} else {
		styleSrcDirective := &csp.StyleSrcDirective{
			SourceDirective: csp.SourceDirective{
				Sources: hashSources,
			},
		}
		contentSecurityPolicy.Directives = append(contentSecurityPolicy.Directives, styleSrcDirective)
	}

	return nil
}
