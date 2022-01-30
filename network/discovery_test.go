package network

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestDiscoveryServiceAnySocket(t *testing.T) {
	discoveredNodes := make(chan *net.TCPAddr, 5)
	discoverySocket, err := net.ResolveUDPAddr("udp4", "0.0.0.0:8400")
	if err != nil {
		t.Fatalf("%v", err)
	}

	ifaces, err := listIPV4LocalInterfaces()

	appIPPort := fmt.Sprintf("%s:8401", ifaces[0])
	appSocket, err := net.ResolveTCPAddr("tcp4", appIPPort)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	go DiscoveryService(ctx, discoveredNodes, discoverySocket, appSocket)

	conn, err := net.DialUDP("udp4", nil, discoverySocket)
	if err != nil {
		t.Fatalf("%v", err)
	}

	defer conn.Close()

	otherAppSocket, err := net.ResolveTCPAddr("tcp4", "10.0.10.0:8401")
	if err != nil {
		t.Fatalf("%v", err)
	}

	discoveryPacket := NewRequestDiscoveryPacket(otherAppSocket.IP, uint16(otherAppSocket.Port))
	discoveryPacketBytes, _ := discoveryPacket.Bytes()
	conn.Write(discoveryPacketBytes)

	readBuffer := make([]byte, 1024)
	read, _, err := conn.ReadFrom(readBuffer)
	if err != nil {
		t.Fatalf("%v", err)
	}
	buffer := bytes.NewBuffer(readBuffer[:read])
	packet, err := ParsePacket(buffer)
	if err != nil {
		t.Fatalf("%v", err)
	}
	discoveryResponsePacket, ok := packet.(*DiscoveryPacket)
	if !ok {
		t.Fatalf("%v", err)
	}
	if discoveryResponsePacket.Address.String() == appIPPort {
		t.Fatalf("Expecting the app running in a valid IP not %s", discoveryResponsePacket.Address)
	}
	if discoveryResponsePacket.Port != 8401 {
		t.Fatalf("Expecting the app running in port 8401 instead of %d", discoveryResponsePacket.Port)
	}

	var discoveredNode *net.TCPAddr

	select {
	case discoveredNode = <-discoveredNodes:
	case <-time.After(1 * time.Second):
		t.Fatal("Discovery Service did return the other app socket")
	}

	if discoveredNode.IP.String() != otherAppSocket.IP.String() {
		t.Fatalf("Expecting other app socket IP equals to %d instead of %d", discoveredNode.IP, otherAppSocket.IP)
	}
	if discoveredNode.Port != otherAppSocket.Port {
		t.Fatalf("Expecting other app socket port equals to %d instead of %d", discoveredNode.Port, otherAppSocket.Port)
	}

	deadline := time.Now().Add(1 * time.Second)
	err = conn.SetReadDeadline(deadline)
	if err != nil {
		t.Fatalf("%v", err)
	}

	//Sending after the Discovery Service is down by context timeout
	time.Sleep(2 * time.Second)

	conn.Write(discoveryPacketBytes)

	_, open := <-discoveredNodes

	if open {
		t.Fatal("Discovered Nodes channel shuold be closed after timeout has reached")
	}

	read, _, err = conn.ReadFrom(readBuffer)
	if err == nil {
		t.Fatal("DiscoveryService should not send a packet when it is down")
	}
}

func TestDiscoveryServiceMulticastSocket(t *testing.T) {
	t.Error("Not implemented yet")
}

func listIPV4LocalInterfaces() ([]string, error) {
	var addresses []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			switch v := a.(type) {
			case *net.IPNet:
				if strings.Count(v.IP.String(), ":") < 2 {
					addresses = append(addresses, v.IP.String())
				}
			}

		}
	}
	return addresses, nil
}
