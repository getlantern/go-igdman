// Package igdman provides a basic management interface for Internet Gateway
// Devices (IGDs), primarily intended to help with creating port mappings from
// external ports to ports on internal ips.
//
// igdman uses either UPnP or NAT-PMP, depending on what's discovered on the
// network.
package igdman

import (
	"time"
)

// protocol is TCP or UDP
type protocol string

const (
	TCP = protocol("TCP")
	UDP = protocol("UDP")
)

// Interface IGD represents an Internet Gateway Device.
type IGD interface {
	// GetExternalIP returns the IGD's external (public) IP address
	GetExternalIP() (ip string, err error)

	// AddPortMapping maps the given external port on the IGD to the internal
	// port, with an optional expiration.
	AddPortMapping(proto protocol, internalIP string, internalPort int, externalPort int, expiration time.Duration) error

	// RemovePortMapping removes the mapping from the given external port.
	RemovePortMapping(proto protocol, externalPort int) error

	// Close closes the IGD instance to clean up any open resources
	Close() error
}

// NewIGD obtains a new IGD (either UPnP or NAT-PMP, depending on what's available)
func NewIGD() (igd IGD, err error) {
	return newUpnpIGD()
}
