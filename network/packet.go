package network

import (
	"fmt"
	"net"
	"strings"
)

const preffix = "DYLLABLE-"
const discoveryPreffix = preffix + "DISCOVERY"

const headerSeparator = "\r\n"
const endPacket = "\r\n"

const requestDiscoveryType = "DISCOVERY"
const responseDiscoveryType = "RUNNING-APP"

type DiscoveryPacket struct {
	Address net.IP
	Port    uint16
	Type    string
}

func (packet *DiscoveryPacket) String() string {
	typeHeader := fmt.Sprintf("TYPE: %s", packet.Type)
	hostHeader := fmt.Sprintf("HOST: %s:%d", packet.Address, packet.Port)
	headers := []string{discoveryPreffix, typeHeader, hostHeader}
	return strings.Join(headers, headerSeparator) + headerSeparator + endPacket
}

func (packet *DiscoveryPacket) Bytes() []byte {
	return []byte(packet.String())
}

func NewRequestDiscoveryPacket(address net.IP, port uint16) DiscoveryPacket {
	return DiscoveryPacket{address, port, requestDiscoveryType}
}

func NewResponseDiscoveryPacket(address net.IP, port uint16) DiscoveryPacket {
	return DiscoveryPacket{address, port, responseDiscoveryType}
}
