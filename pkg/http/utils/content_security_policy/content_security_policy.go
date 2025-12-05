package content_security_policy

import (
	"net/url"
	"slices"

	"github.com/Motmedel/utils_go/pkg/http/types/content_security_policy"
)

func PatchCspConnectSrcWithHostSrc(contentSecurityPolicy *content_security_policy.ContentSecurityPolicy, hostUrls ...*url.URL) *content_security_policy.ContentSecurityPolicy {
	var hostSources []content_security_policy.SourceI
	for _, hostUrl := range hostUrls {
		if hostSource := content_security_policy.HostSourceFromUrl(hostUrl); hostSource != nil {
			hostSources = append(hostSources, hostSource)
		}
	}

	if len(hostSources) == 0 {
		return contentSecurityPolicy
	}

	connectSrcDirective := &content_security_policy.ConnectSrcDirective{
		SourceDirective: content_security_policy.SourceDirective{
			Sources: slices.Concat(
				[]content_security_policy.SourceI{
					&content_security_policy.KeywordSource{Keyword: "self"},
				},
				hostSources,
			),
		},
	}

	if contentSecurityPolicy != nil {
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
	} else {
		contentSecurityPolicy = &content_security_policy.ContentSecurityPolicy{
			Directives: []content_security_policy.DirectiveI{
				connectSrcDirective,
			},
		}
	}

	return contentSecurityPolicy
}
