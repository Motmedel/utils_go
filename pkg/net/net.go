package net

import (
	"bytes"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelNetErrors "github.com/Motmedel/utils_go/pkg/net/errors"
	"net"
	"strconv"
)

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

type DomainBreakdown struct {
	RegisteredDomain string `json:"registered_domain,omitempty"`
	Subdomain        string `json:"subdomain,omitempty"`
	TopLevelDomain   string `json:"top_level_domain,omitempty"`
}
