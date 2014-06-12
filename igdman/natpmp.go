package igdman

import (
	"fmt"
	"math"
	"net"
	"strings"
	"time"

	natpmp "code.google.com/p/go-nat-pmp"
)

type natpmpIGD struct {
	client *natpmp.Client
}

func newNATPMPIGD() (igd *natpmpIGD, err error) {
	ip, err := defaultGatewayIp()
	if err != nil {
		return nil, fmt.Errorf("Unable to find default gateway: %s", err)
	}
	return &natpmpIGD{natpmp.NewClient(net.ParseIP(ip))}, nil
}

func (igd *natpmpIGD) GetExternalIP() (ip string, err error) {
	response, err := igd.client.GetExternalAddress()
	if err != nil {
		return "", fmt.Errorf("Unable to get external address: %s", err)
	}
	ip = net.IPv4(response.ExternalIPAddress[0],
		response.ExternalIPAddress[1],
		response.ExternalIPAddress[2],
		response.ExternalIPAddress[3]).String()
	return
}

func (igd *natpmpIGD) AddPortMapping(proto protocol, internalIP string, internalPort int, externalPort int, expiration time.Duration) error {
	expirationInSeconds := int(expiration.Seconds())
	if expirationInSeconds == 0 {
		expirationInSeconds = int(math.MaxInt32)
	}
	result, err := igd.client.AddPortMapping(natpmpProtoFor(proto), internalPort, externalPort, expirationInSeconds)
	if err != nil {
		return fmt.Errorf("Unable to add port mapping: %s", err)
	}
	if int(result.MappedExternalPort) != externalPort {
		igd.RemovePortMapping(proto, externalPort)
		return fmt.Errorf("Mapped port didn't match requested")
	}
	return nil
}

func (igd *natpmpIGD) RemovePortMapping(proto protocol, externalPort int) error {
	someInternalPort := 15670 // actual value doesn't matter
	_, err := igd.client.AddPortMapping(natpmpProtoFor(proto), someInternalPort, externalPort, 0)
	if err != nil {
		return fmt.Errorf("Unable to remove port mapping: %s", err)
	}
	return nil
}

func (igd *natpmpIGD) Close() error {
	// nothing to close
	return nil
}

func natpmpProtoFor(proto protocol) string {
	return strings.ToLower(string(proto))
}
