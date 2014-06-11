package igdman

import (
	"fmt"
	"github.com/oxtoacart/framed"
	"net"
	"os"
	"testing"
	"time"
)

var (
	PUBLIC_IP = "PUBLIC_IP"
)

// TestExternalIP only works when there is a valid IGD device with an external
// IP.  The environment variable EXTERNAL_IP needs to be set for this test to
// work.
func TestExternalIP(t *testing.T) {
	expectedExternalIP := os.Getenv("EXTERNAL_IP")
	if expectedExternalIP == "" {
		t.Fatalf("Please set the environment variable EXTERNAL_IP to provide your expected public IP address")
	}

	igd, err := NewIGD()
	if err != nil {
		t.Fatalf("Unable to create IGD: %s", err)
	}
	defer igd.Close()

	start := time.Now()
	publicIp, err := igd.GetExternalIP()
	if err != nil {
		t.Fatalf("Unable to get Public IP: %s", err)
	}
	delta1 := time.Now().Sub(start)
	if publicIp != expectedExternalIP {
		t.Errorf("External ip '%s' did not match expected '%s'", publicIp, expectedExternalIP)
	}
	start = time.Now()
	publicIp, err = igd.GetExternalIP()
	if err != nil {
		t.Fatalf("Unable to get External IP 2nd time: %s", err)
	}
	delta2 := time.Now().Sub(start)
	if delta2 > delta1/10 {
		t.Fatalf("2nd external ip lookup should have been much faster than first because of cached IGD url.  1st lookup: %d ms, 2nd lookup: %d ms", delta1/time.Millisecond, delta2/time.Millisecond)
	}
}

func TestMapping(t *testing.T) {
	port := 15067

	igd, err := NewIGD()
	if err != nil {
		t.Fatalf("Unable to create IGD: %s", err)
	}
	defer igd.Close()

	externalIP, err := igd.GetExternalIP()
	if err != nil {
		t.Fatalf("Unable to get external IP: %s", err)
	}

	// Run echo server
	go func() {
		l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
		defer l.Close()
		if err != nil {
			t.Fatalf("Unable to start server: %s", err)
		}
		for {
			conn, err := l.Accept()
			if err != nil {
				t.Fatalf("Unable to accept connection: %s", err)
			}
			f := framed.Framed{conn}
			defer f.Close()
			frame := make([]byte, 1000)
			for {
				n, err := f.Read(frame)
				if err != nil {
					return
				}
				_, err = f.Write(frame[:n])
				if err != nil {
					return
				}
			}
		}
	}()

	// Add port mapping
	internalIP, err := getFirstNonLoopbackAdapterAddr()
	if err != nil {
		t.Fatalf("Unable to get internal ip: %s", err)
	}

	err = igd.AddPortMapping(TCP, internalIP, port, port, 0)
	if err != nil {
		t.Fatalf("Unable to add port mapping: %s", err)
	}

	// Run echo client
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", externalIP, port))
	if err != nil {
		t.Fatalf("Unable to connect to echo server")
	}
	f := framed.Framed{conn}
	defer f.Close()
	testString := "Hello strange port mapped world"
	_, err = f.Write([]byte(testString))
	if err != nil {
		t.Fatalf("Unable to write to echo server")
	}
	frame := make([]byte, 1000)
	n, err := f.Read(frame)
	if err != nil {
		t.Fatalf("Unable to read from echo server")
	}
	response := string(frame[:n])
	if response != testString {
		t.Errorf("Response '%s' did not match expected value '%s'", response, testString)
	}

	// Remove port mapping
	err = igd.RemovePortMapping(TCP, port)
	if err != nil {
		t.Fatalf("Unable to remove port mapping: %s", err)
	}
	conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", externalIP, port))
	if err == nil {
		t.Errorf("Connecting to address with closed port mapping should have resulted in an error")
	}
}

func TestFailedMapping(t *testing.T) {
	port := 15068

	igd, err := NewIGD()
	if err != nil {
		t.Fatalf("Unable to create IGD: %s", err)
	}
	defer igd.Close()

	// Add port mapping
	internalIP, err := getFirstNonLoopbackAdapterAddr()
	if err != nil {
		t.Fatalf("Unable to get internal ip: %s", err)
	}

	err = igd.AddPortMapping(TCP, internalIP, port, 0, 0)
	if err == nil {
		t.Error("Adding mapping for bad port should have resulted in error")
	}
}

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
