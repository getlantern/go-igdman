package igdman

import (
	"time"
)

type protocol string

const (
	TCP = protocol("TCP")
	UDP = protocol("UDP")
)

type IGD interface {
	GetExternalIP() (ip string, err error)

	AddPortMapping(proto protocol, internalIP string, internalPort int, externalPort int, duration time.Duration) error

	RemovePortMapping(proto protocol, externalPort int) error

	Close() error
}

// NewIGD obtains a new IGD (either UPnP or NAT-PMP, depending on what's available)
func NewIGD() (igd IGD, err error) {
	return newUpnpIGD()
}
