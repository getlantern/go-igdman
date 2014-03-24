package igdman

type protocol string

const (
	TCP = protocol("TCP")
	UDP = protocol("UDP")
)

type IGD interface {
	GetPublicIP() (ip string, err error)

	AddPortMapping(proto protocol, internalIp string, internalPort int, externalPort int) error

	RemovePortMapping(proto protocol, externalPort int) error
}
