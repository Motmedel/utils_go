package net

import (
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	"golang.org/x/net/publicsuffix"
	"math"
	"net"
	"strconv"
	"strings"
)

var cidr2mask = []uint32{
	0x00000000, 0x80000000, 0xC0000000,
	0xE0000000, 0xF0000000, 0xF8000000,
	0xFC000000, 0xFE000000, 0xFF000000,
	0xFF800000, 0xFFC00000, 0xFFE00000,
	0xFFF00000, 0xFFF80000, 0xFFFC0000,
	0xFFFE0000, 0xFFFF0000, 0xFFFF8000,
	0xFFFFC000, 0xFFFFE000, 0xFFFFF000,
	0xFFFFF800, 0xFFFFFC00, 0xFFFFFE00,
	0xFFFFFF00, 0xFFFFFF80, 0xFFFFFFC0,
	0xFFFFFFE0, 0xFFFFFFF0, 0xFFFFFFF8,
	0xFFFFFFFC, 0xFFFFFFFE, 0xFFFFFFFF,
}

func SplitAddress(address string) (string, int, error) {
	ip, portString, err := net.SplitHostPort(address)
	if err != nil {
		return "", 0, &motmedelErrors.InputError{
			Message: "An error occurred when splitting an address into host and port.",
			Cause:   err,
			Input:   address,
		}
	}

	port, err := strconv.Atoi(portString)
	if err != nil {
		return ip, 0, &motmedelErrors.InputError{
			Message: "An error occurred when parsing an address port string as an integer.",
			Cause:   err,
			Input:   portString,
		}
	}

	return ip, port, nil
}

func ipv4ToUint32(iPv4 string) uint32 {
	ipOctets := [4]uint64{}
	for i, v := range strings.SplitN(iPv4, ".", 4) {
		ipOctets[i], _ = strconv.ParseUint(v, 10, 32)
	}
	result := (ipOctets[0] << 24) | (ipOctets[1] << 16) | (ipOctets[2] << 8) | ipOctets[3]

	return uint32(result)
}

func uint32ToIpv4(iPuInt32 uint32) (iP string) {
	return fmt.Sprintf(
		"%d.%d.%d.%d",
		iPuInt32>>24,
		(iPuInt32&0x00FFFFFF)>>16,
		(iPuInt32&0x0000FFFF)>>8,
		iPuInt32&0x000000FF,
	)
}

func Ipv4RangeToCidrRange(ipStart string, ipEnd string) ([]string, error) {
	ipStartUint32 := ipv4ToUint32(ipStart)
	ipEndUint32 := ipv4ToUint32(ipEnd)

	if ipStartUint32 > ipEndUint32 {
		return nil, fmt.Errorf("start IP:%s must be less than end IP:%s", ipStart, ipEnd)
	}

	var cidrs []string

	for ipEndUint32 >= ipStartUint32 {
		maxSize := 32
		for maxSize > 0 {

			maskedBase := ipStartUint32 & cidr2mask[maxSize-1]

			if maskedBase != ipStartUint32 {
				break
			}
			maxSize--
		}

		x := math.Log(float64(ipEndUint32-ipStartUint32+1)) / math.Log(2)
		maxDiff := 32 - int(math.Floor(x))
		if maxSize < maxDiff {
			maxSize = maxDiff
		}

		cidrs = append(cidrs, uint32ToIpv4(ipStartUint32)+"/"+strconv.Itoa(maxSize))

		ipStartUint32 += uint32(math.Exp2(float64(32 - maxSize)))
	}

	return cidrs, nil
}

func GetIpVersion(ip *net.IP) int {
	if ip.To4() != nil {
		return 4
	} else if ip.To16() != nil {
		return 6
	} else {
		return 0
	}
}

type DomainBreakdown struct {
	RegisteredDomain string `json:"registered_domain,omitempty"`
	Subdomain        string `json:"subdomain,omitempty"`
	TopLevelDomain   string `json:"top_level_domain,omitempty"`
}

func GetDomainBreakdown(domainString string) *DomainBreakdown {
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

	domainBreakdown := DomainBreakdown{
		TopLevelDomain:   etld,
		RegisteredDomain: registeredDomain,
	}

	if subdomain := strings.TrimSuffix(domainString, "."+registeredDomain); subdomain != domainString {
		domainBreakdown.Subdomain = subdomain
	}

	return &domainBreakdown
}
