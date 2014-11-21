package igdman

import (
	"fmt"
	"net"
	"os"
)

func getFirstNonLoopbackAdapterAddr() (string, error) {
	name, err := os.Hostname()
	if err != nil {
		return "", err
	}

	addrs, err := net.LookupHost(name)
	if err != nil {
		return "", err
	}

	for _, a := range addrs {
		ip := net.ParseIP(a)
		if !ip.IsLoopback() {
			return a, nil
		}
	}

	return "", fmt.Errorf("No non-loopback adapter found")
}
