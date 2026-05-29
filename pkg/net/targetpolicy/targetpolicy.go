package targetpolicy

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"strings"
)

var (
	ErrDNSLookupFailed = errors.New("DNS lookup failed")
	ErrEmptyTarget     = errors.New("empty target")
	ErrNilLookupIP     = errors.New("nil lookup IP function")
	ErrNoIPAddresses   = errors.New("DNS lookup returned no IP addresses")
	ErrNonRoutableIP   = errors.New("target resolves to non-routable IP")
)

var deniedPrefixStrings = []string{
	"0.0.0.0/8",
	"10.0.0.0/8",
	"100.64.0.0/10",
	"127.0.0.0/8",
	"169.254.0.0/16",
	"172.16.0.0/12",
	"192.0.0.0/24",
	"192.0.2.0/24",
	"192.88.99.0/24",
	"192.168.0.0/16",
	"198.18.0.0/15",
	"198.51.100.0/24",
	"203.0.113.0/24",
	"224.0.0.0/4",
	"240.0.0.0/4",
	"::/128",
	"::1/128",
	"64:ff9b::/96",
	"64:ff9b:1::/48",
	"100::/64",
	"100:0:0:1::/64",
	"2001::/32",
	"2001:10::/28",
	"2001:20::/28",
	"2001:2::/48",
	"2001:db8::/32",
	"2002::/16",
	"3fff::/20",
	"5f00::/16",
	"fc00::/7",
	"fe80::/10",
	"ff00::/8",
}

var deniedPrefixes = mustParsePrefixes(deniedPrefixStrings)

// LookupIPFunc resolves host into IP addresses.
type LookupIPFunc func(ctx context.Context, network string, host string) ([]net.IP, error)

// TargetInfo describes the normalized target data used by policy checks.
type TargetInfo struct {
	Target   string
	Host     string
	IPs      []net.IP
	DirectIP bool
}

// DeniedPrefixStrings returns the CIDR deny list as strings.
func DeniedPrefixStrings() []string {
	return append([]string(nil), deniedPrefixStrings...)
}

// DeniedPrefixes returns the parsed CIDR deny list.
func DeniedPrefixes() []netip.Prefix {
	return append([]netip.Prefix(nil), deniedPrefixes...)
}

// AppendDeniedPrefixStrings appends the shared CIDR deny list to values,
// preserving existing values and avoiding duplicates.
func AppendDeniedPrefixStrings(values []string) []string {
	return appendUniqueStrings(values, deniedPrefixStrings...)
}

// ValidateTarget parses and validates a target using the default resolver.
func ValidateTarget(ctx context.Context, target string) (*TargetInfo, error) {
	return ValidateTargetWithLookup(ctx, target, net.DefaultResolver.LookupIP)
}

// ValidateTargetWithLookup parses and validates a target with lookupIP.
//
// target may be a raw IP, hostname, host:port, or URL. Hostnames are resolved
// and every resulting IP must be outside the shared deny list. DNS errors and
// empty DNS answers fail closed.
func ValidateTargetWithLookup(ctx context.Context, target string, lookupIP LookupIPFunc) (*TargetInfo, error) {
	targetInfo, err := ResolveTarget(ctx, target, lookupIP)
	if err != nil {
		return targetInfo, err
	}

	for _, checkIP := range targetInfo.IPs {
		if !IsRoutableIP(checkIP) {
			return targetInfo, fmt.Errorf("%w: target=%q host=%q ip=%s", ErrNonRoutableIP, target, targetInfo.Host, checkIP.String())
		}
	}

	return targetInfo, nil
}

// IsRoutableTarget reports whether target is valid under the shared policy.
func IsRoutableTarget(ctx context.Context, target string) bool {
	_, err := ValidateTarget(ctx, target)
	return err == nil
}

// IsRoutableTargetWithLookup reports whether target is valid under the shared
// policy using lookupIP.
func IsRoutableTargetWithLookup(ctx context.Context, target string, lookupIP LookupIPFunc) bool {
	_, err := ValidateTargetWithLookup(ctx, target, lookupIP)
	return err == nil
}

// IsRoutableIP reports whether ip is outside the shared deny list.
func IsRoutableIP(ip net.IP) bool {
	addr, ok := addrFromIP(ip)
	if !ok {
		return false
	}
	return IsRoutableAddr(addr)
}

// IsRoutableAddr reports whether addr is outside the shared deny list.
func IsRoutableAddr(addr netip.Addr) bool {
	addr = normalizeAddr(addr)
	if !addr.IsValid() {
		return false
	}

	for _, denied := range deniedPrefixes {
		if sameFamily(addr, denied.Addr()) && denied.Contains(addr) {
			return false
		}
	}

	return true
}

// PrefixOverlapsDenied reports whether ipNet overlaps any denied prefix.
func PrefixOverlapsDenied(ipNet *net.IPNet) (bool, error) {
	prefix, err := PrefixFromIPNet(ipNet)
	if err != nil {
		return false, err
	}

	for _, denied := range deniedPrefixes {
		if sameFamily(prefix.Addr(), denied.Addr()) && prefix.Overlaps(denied) {
			return true, nil
		}
	}

	return false, nil
}

// PrefixFromIPNet converts ipNet into a normalized netip.Prefix.
func PrefixFromIPNet(ipNet *net.IPNet) (netip.Prefix, error) {
	if ipNet == nil {
		return netip.Prefix{}, fmt.Errorf("nil IP network")
	}

	ones, bits := ipNet.Mask.Size()
	if ones < 0 {
		return netip.Prefix{}, fmt.Errorf("invalid IP network mask")
	}

	var (
		addr netip.Addr
		ok   bool
	)

	switch bits {
	case 32:
		ip4 := ipNet.IP.To4()
		if ip4 == nil {
			return netip.Prefix{}, fmt.Errorf("invalid IPv4 network address: %s", ipNet.IP.String())
		}
		addr, ok = netip.AddrFromSlice(ip4)
	case 128:
		ip16 := ipNet.IP.To16()
		if ip16 == nil || ipNet.IP.To4() != nil {
			return netip.Prefix{}, fmt.Errorf("invalid IPv6 network address: %s", ipNet.IP.String())
		}
		addr, ok = netip.AddrFromSlice(ip16)
	default:
		return netip.Prefix{}, fmt.Errorf("unsupported IP network mask size: %d", bits)
	}

	if !ok {
		return netip.Prefix{}, fmt.Errorf("invalid IP network address: %s", ipNet.IP.String())
	}

	return netip.PrefixFrom(normalizeAddr(addr), ones).Masked(), nil
}

// ResolveTarget parses target and returns the IPs used for policy checks.
func ResolveTarget(ctx context.Context, target string, lookupIP LookupIPFunc) (*TargetInfo, error) {
	if strings.TrimSpace(target) == "" {
		return &TargetInfo{Target: target}, ErrEmptyTarget
	}
	if lookupIP == nil {
		return &TargetInfo{Target: target}, ErrNilLookupIP
	}

	host := HostForTarget(target)
	targetInfo := &TargetInfo{
		Target: target,
		Host:   host,
	}
	if strings.TrimSpace(host) == "" {
		return targetInfo, ErrEmptyTarget
	}

	if ip := net.ParseIP(host); ip != nil {
		targetInfo.IPs = []net.IP{ip}
		targetInfo.DirectIP = true
		return targetInfo, nil
	}

	resolvedIPs, err := lookupIP(ctx, "ip", host)
	if err != nil {
		return targetInfo, fmt.Errorf("%w: target=%q host=%q: %w", ErrDNSLookupFailed, target, host, err)
	}
	if len(resolvedIPs) == 0 {
		return targetInfo, fmt.Errorf("%w: target=%q host=%q", ErrNoIPAddresses, target, host)
	}

	targetInfo.IPs = append([]net.IP(nil), resolvedIPs...)
	return targetInfo, nil
}

// HostForTarget extracts the hostname or IP portion from a raw target.
func HostForTarget(target string) string {
	target = strings.TrimSpace(target)

	if parsedURL, err := url.Parse(target); err == nil && parsedURL.Scheme != "" && parsedURL.Host != "" {
		return parsedURL.Hostname()
	}

	if host, _, err := net.SplitHostPort(target); err == nil {
		return strings.Trim(host, "[]")
	}

	return strings.Trim(target, "[]")
}

func appendUniqueStrings(values []string, additional ...string) []string {
	seen := make(map[string]struct{}, len(values)+len(additional))
	for _, value := range values {
		seen[value] = struct{}{}
	}

	for _, value := range additional {
		if _, ok := seen[value]; ok {
			continue
		}
		values = append(values, value)
		seen[value] = struct{}{}
	}

	return values
}

func addrFromIP(ip net.IP) (netip.Addr, bool) {
	if ip == nil {
		return netip.Addr{}, false
	}

	if ip4 := ip.To4(); ip4 != nil {
		addr, ok := netip.AddrFromSlice(ip4)
		return addr, ok
	}

	ip16 := ip.To16()
	if ip16 == nil {
		return netip.Addr{}, false
	}

	addr, ok := netip.AddrFromSlice(ip16)
	if !ok {
		return netip.Addr{}, false
	}

	return normalizeAddr(addr), true
}

func normalizeAddr(addr netip.Addr) netip.Addr {
	if addr.Is4In6() {
		return addr.Unmap()
	}
	return addr
}

func sameFamily(a netip.Addr, b netip.Addr) bool {
	return a.Is4() == b.Is4()
}

func mustParsePrefixes(values []string) []netip.Prefix {
	prefixes := make([]netip.Prefix, 0, len(values))
	for _, value := range values {
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			panic(err)
		}
		prefixes = append(prefixes, prefix.Masked())
	}
	return prefixes
}
