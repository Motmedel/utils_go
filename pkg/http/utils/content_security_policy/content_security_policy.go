package content_security_policy

import (
	"net/url"
	"slices"

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
