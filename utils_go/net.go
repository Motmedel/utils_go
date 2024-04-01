package utils_go

import (
	"net"
	"strconv"
)

func SplitAddress(address string) (string, int, error) {
	ip, port, err := net.SplitHostPort(address)
	if err != nil {
		return "", 0, err
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil {
		return ip, 0, err
	}

	return ip, portNumber, nil
}
