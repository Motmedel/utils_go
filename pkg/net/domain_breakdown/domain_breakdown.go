package domain_breakdown

import (
	motmedelNet "github.com/Motmedel/utils_go/pkg/net"
	"golang.org/x/net/publicsuffix"
	"strings"
)

func GetDomainBreakdown(domainString string) *motmedelNet.DomainBreakdown {
	if domainString == "" {
		return nil
	}

	etld, icann := publicsuffix.PublicSuffix(domainString)
	if !icann && strings.IndexByte(etld, '.') == -1 {
		return nil
	}

	registeredDomain, err := publicsuffix.EffectiveTLDPlusOne(domainString)
	if err != nil {
		return nil
	}

	domainBreakdown := motmedelNet.DomainBreakdown{
		TopLevelDomain:   etld,
		RegisteredDomain: registeredDomain,
	}

	if subdomain := strings.TrimSuffix(domainString, "."+registeredDomain); subdomain != domainString {
		domainBreakdown.Subdomain = subdomain
	}

	return &domainBreakdown
}
