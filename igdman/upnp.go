package igdman

import (
	"fmt"
	"github.com/oxtoacart/byteexec"
	"strings"
	"time"
)

const (
	IGD_URL_LABEL             = "Found valid IGD : "
	LOCAL_IP_ADDRESS_LABEL    = "Local LAN ip address : "
	EXTERNAL_IP_ADDRESS_LABEL = "ExternalIPAddress = "
)

type upnpIGD struct {
	upnpc      *byteexec.ByteExec
	igdUrl     string
	internalIP string
	externalIP string
}

func newUpnpIGD() (igd *upnpIGD, err error) {
	upnpcBytes, err := Asset("upnpc")
	if err != nil {
		return nil, err
	}
	be, err := byteexec.NewByteExec(upnpcBytes)
	if err != nil {
		return nil, err
	}
	return &upnpIGD{upnpc: be}, nil
}

func (igd *upnpIGD) GetExternalIP() (ip string, err error) {
	err = igd.updateStatus()
	if err != nil {
		return "", err
	}
	return igd.externalIP, nil
}

func (igd *upnpIGD) AddPortMapping(proto protocol, internalIP string, internalPort int, externalPort int, expiration time.Duration) error {
	if err := igd.updateStatus(); err != nil {
		return fmt.Errorf("Unable to add port mapping: %s", err)
	}
	params := []string{
		"-url", igd.igdUrl,
		"-a", internalIP, fmt.Sprintf("%d", internalPort), fmt.Sprintf("%d", externalPort), string(proto),
	}
	if expiration > 0 {
		params = append(params, fmt.Sprintf("%d", expiration/time.Second))
	}
	out, err := igd.upnpc.Command(params...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("Unable to add port mapping: %s\n%s", err, out)
	} else {
		return nil
	}
}

func (igd *upnpIGD) RemovePortMapping(proto protocol, externalPort int) error {
	if err := igd.updateStatus(); err != nil {
		return fmt.Errorf("Unable to add port mapping: %s", err)
	}
	params := []string{
		"-url", igd.igdUrl,
		"-d", fmt.Sprintf("%d", externalPort), string(proto),
	}
	out, err := igd.upnpc.Command(params...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("Unable to remove port mapping: %s\n%s", err, out)
	} else {
		return nil
	}
}

func (igd *upnpIGD) Close() error {
	return igd.upnpc.Close()
}

// updateStatus updates the IGD's status fields
func (igd *upnpIGD) updateStatus() error {
	skipDiscovery := igd.igdUrl != ""
	params := []string{"-s"}
	if skipDiscovery {
		params = []string{"-url", igd.igdUrl, "-s"} // -s has to be at the end for some reason
	}
	out, err := igd.upnpc.Command(params...).CombinedOutput()
	if err != nil {
		if skipDiscovery {
			// Clear remembered url and try again
			igd.igdUrl = ""
			return igd.updateStatus()
		} else {
			return fmt.Errorf("Unable to call upnpc to get status: %s\n%s", err, out)
		}
	}
	resp := string(out)
	if igd.igdUrl, err = igd.extractFromStatusResponse(resp, IGD_URL_LABEL); err != nil {
		return err
	}
	if igd.internalIP, err = igd.extractFromStatusResponse(resp, LOCAL_IP_ADDRESS_LABEL); err != nil {
		return err
	}
	if igd.externalIP, err = igd.extractFromStatusResponse(resp, EXTERNAL_IP_ADDRESS_LABEL); err != nil {
		return err
	}
	return nil
}

func (igd *upnpIGD) extractFromStatusResponse(resp string, label string) (string, error) {
	i := strings.Index(resp, label)
	if i < 0 {
		return "", fmt.Errorf("%s not available from upnpc", label)
	}
	resp = resp[i+len(label):]
	s := strings.Index(resp, "\n")
	if s < 0 {
		return "", fmt.Errorf("Unable to find newline after %s", label)
	}
	return resp[:s], nil
}
