package net

import (
	"bytes"
	"encoding/binary"
	"fmt"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelNetErrors "github.com/Motmedel/utils_go/pkg/net/errors"
	"net"
	"strconv"
)

const (
	ProtocolIcmp  = 1
	ProtocolTcp   = 6
	ProtocolUdp   = 17
	ProtocolIcmp6 = 132
)

func SplitAddress(address string) (string, int, error) {
	ip, portString, err := net.SplitHostPort(address)
	if err != nil {
		return "", 0, &motmedelErrors.Error{
			Message: "An error occurred when splitting an address into host and port.",
			Cause:   err,
			Input:   address,
		}
	}

	port, err := strconv.Atoi(portString)
	if err != nil {
		return ip, 0, &motmedelErrors.Error{
			Message: "An error occurred when parsing an address port string as an integer.",
			Cause:   err,
			Input:   portString,
		}
	}

	return ip, port, nil
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

// Calculate the last address in the network
func lastAddress(network net.IPNet) net.IP {
	var last net.IP
	ip := network.IP.To16()
	mask := network.Mask

	last = make(net.IP, len(ip))
	for i := range last {
		last[i] = ip[i] | ^mask[i]
	}
	return last
}

func ParseAddressNet(addressNet string) (*net.IPNet, error) {
	if addressNet == "" {
		return nil, nil
	}

	networkString := addressNet

	if ip := net.ParseIP(addressNet); ip != nil {
		var mask int
		switch ipVersion := GetIpVersion(&ip); ipVersion {
		case 4:
			mask = 32
		case 6:
			mask = 128
		default:
			return nil, motmedelErrors.NewWithTrace(
				fmt.Errorf("%w: %d", motmedelNetErrors.ErrUnexpectedIpVersion, ipVersion),
				ipVersion,
			)
		}

		networkString += fmt.Sprintf("/%d", mask)
	}

	_, network, err := net.ParseCIDR(networkString)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(
			fmt.Errorf("net parse cidr: %w", err),
			networkString,
		)
	}

	return network, nil
}

func GetStartEndCidr(startIpAddress *net.IP, endIpAddress *net.IP, checkBoundary bool) (string, error) {
	if startIpAddress == nil || endIpAddress == nil {
		return "", nil
	}

	if (startIpAddress.To4() == nil) != (endIpAddress.To4() == nil) {
		return "", motmedelNetErrors.ErrIpVersionMismatch
	}

	startBytes := startIpAddress.To16()
	endBytes := endIpAddress.To16()

	byteComparison := bytes.Compare(startBytes, endBytes)

	if byteComparison > 0 {
		return "", motmedelNetErrors.ErrStartAfterEnd
	}

	// Find the first byte where the two IP addresses differ
	maskLength := 0
	found := false
	for i := 0; i < len(startBytes); i++ {
		if startBytes[i] != endBytes[i] {
			// Calculate the mask length up to this point
			maskLength = i * 8
			diff := startBytes[i] ^ endBytes[i]
			// Count the number of leading zeros in the differing byte
			for j := 7; j >= 0; j-- {
				if diff&(1<<j) != 0 {
					maskLength += 8 - j - 1
					found = true
					break
				}
			}
			if found {
				break
			}
		}
	}

	if !found {
		if startIpAddress.To4() != nil {
			maskLength = 32
		} else {
			maskLength = 128
		}
	}

	mask := net.CIDRMask(maskLength, len(startBytes)*8)
	network := net.IPNet{IP: startIpAddress.Mask(mask), Mask: mask}

	if checkBoundary && byteComparison != 0 {
		// Ensure start IP is network's base address and end IP is the last address in the network
		networkBase := network.IP
		networkLast := lastAddress(network)

		if !networkBase.Equal(*startIpAddress) || !networkLast.Equal(*endIpAddress) {
			return "", motmedelNetErrors.ErrNotOnSubnetBoundaries
		}
	}

	return network.String(), nil
}

// IntToIpv4 converts IPv4 number to net.IP
func IntToIpv4(ipNum uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, ipNum)
	return ip
}

func NetworkFromTarget(target string) (*net.IPNet, error) {
	if target == "" {
		return nil, nil
	}

	if _, network, _ := net.ParseCIDR(target); network != nil {
		return network, nil
	}

	if ip := net.ParseIP(target); ip != nil {
		useIpv4 := ip.To4() != nil
		useIpv6 := ip.To16() != nil && ip.To4() == nil

		var targetCidrString string

		if useIpv6 {
			targetCidrString = target + "/128"
		} else if useIpv4 {
			targetCidrString = target + "/32"
		} else {
			return nil, motmedelErrors.NewWithTrace(motmedelNetErrors.ErrUndeterminableIpVersion, ip)
		}

		_, network, _ := net.ParseCIDR(targetCidrString)
		if network == nil {
			return nil, motmedelErrors.NewWithTrace(
				fmt.Errorf("%w (single target)", motmedelNetErrors.ErrNilIpNet), targetCidrString,
			)
		}

		return network, nil
	}

	return nil, motmedelErrors.NewWithTrace(motmedelNetErrors.ErrUndeterminableTargetFormat)
}

type DomainBreakdown struct {
	RegisteredDomain string `json:"registered_domain,omitempty"`
	Subdomain        string `json:"subdomain,omitempty"`
	TopLevelDomain   string `json:"top_level_domain,omitempty"`
}
