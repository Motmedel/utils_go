package content_security_policy

import (
	"net/url"
	"slices"
	"strings"

	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	csp "github.com/Motmedel/utils_go/pkg/http/types/content_security_policy"
)

// sourceDirectivePointer is any concrete source-directive pointer (e.g. *csp.WorkerSrcDirective)
// whose source list can be read and replaced.
type sourceDirectivePointer[T any] interface {
	*T
	csp.DirectiveI
	GetSources() []csp.SourceI
	SetSources(sources []csp.SourceI)
}

// PatchCspSourceDirective merges sources into the source directive of type T, deduplicating by
// serialized value. If the policy has no directive of that type one is created and appended;
// otherwise the existing directive is extended in place. If a directive with the same name is
// present but is not of type T (e.g. an unparsed fallback), the policy is left untouched.
//
// T is the concrete source-directive type; the pointer type is inferred. For example, to allow a
// blob: Worker:
//
//	PatchCspSourceDirective[csp.WorkerSrcDirective](
//		policy,
//		&csp.KeywordSource{Keyword: "self"},
//		&csp.SchemeSource{Scheme: "blob"},
//	)
func PatchCspSourceDirective[T any, PT sourceDirectivePointer[T]](
	contentSecurityPolicy *csp.ContentSecurityPolicy,
	sources ...csp.SourceI,
) {
	if contentSecurityPolicy == nil || len(sources) == 0 {
		return
	}

	directive := PT(new(T))
	if existingDirective, found := contentSecurityPolicy.GetDirective(directive.GetName()); found {
		existingSourceDirective, ok := existingDirective.(PT)
		if !ok {
			return
		}
		directive = existingSourceDirective
	} else {
		contentSecurityPolicy.Directives = append(contentSecurityPolicy.Directives, directive)
	}

	mergedSources := directive.GetSources()
	sourceValues := make(map[string]struct{}, len(mergedSources))
	for _, source := range mergedSources {
		if source != nil {
			sourceValues[source.String()] = struct{}{}
		}
	}
	for _, source := range sources {
		if source == nil {
			continue
		}
		sourceValue := source.String()
		if _, found := sourceValues[sourceValue]; found {
			continue
		}
		sourceValues[sourceValue] = struct{}{}
		mergedSources = append(mergedSources, source)
	}

	directive.SetSources(mergedSources)
}

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

func PatchCspImageSrc(contentSecurityPolicy *csp.ContentSecurityPolicy, urls ...*url.URL) {
	if contentSecurityPolicy == nil {
		return
	}

	var newSources []csp.SourceI
	for _, u := range urls {
		if u == nil {
			continue
		}
		if u.Scheme == "data" {
			newSources = append(newSources, &csp.SchemeSource{Scheme: "data"})
		} else if hostSource := csp.HostSourceFromUrl(u); hostSource != nil {
			newSources = append(newSources, hostSource)
		}
	}

	if len(newSources) == 0 {
		return
	}

	imgSrcDirective := &csp.ImgSrcDirective{
		SourceDirective: csp.SourceDirective{
			Sources: slices.Concat(
				[]csp.SourceI{
					&csp.KeywordSource{Keyword: "self"},
				},
				newSources,
			),
		},
	}

	if existingImgSrcDirective := contentSecurityPolicy.GetImgSrc(); existingImgSrcDirective != nil {
		sourceMap := make(map[string]struct{})
		for _, source := range existingImgSrcDirective.Sources {
			sourceMap[source.String()] = struct{}{}
		}

		for _, newSource := range newSources {
			if _, found := sourceMap[newSource.String()]; !found {
				existingImgSrcDirective.Sources = append(existingImgSrcDirective.Sources, newSource)
			}
		}
	} else {
		contentSecurityPolicy.Directives = append(contentSecurityPolicy.Directives, imgSrcDirective)
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
