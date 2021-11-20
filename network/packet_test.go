package network

import (
	"bytes"
	"net"
	"testing"
)

func TestRequestDiscoveryPacket(t *testing.T) {
	var address = net.IPv4(127, 0, 0, 1)
	var port uint16 = 8400
	packet := NewRequestDiscoveryPacket(address, port)
	packet_str := packet.String()
	expected_str := "DYLLABLE-DISCOVERY\r\n" +
		"TYPE: DISCOVERY\r\n" +
		"HOST: 127.0.0.1:8400\r\n" +
		"\r\n"
	if packet_str != expected_str {
		t.Fatalf("Request DiscoveryPacket String() does not "+
			"match to the expected.\nCurrent:\n%#v.\nExpected:\n%#v\n", packet_str, expected_str)
	}

	packet_bytes := packet.Bytes()
	expected_bytes := []byte(expected_str)
	if bytes.Compare(packet_bytes, expected_bytes) != 0 {
		t.Fatalf("Request DiscoveryPacket Bytes() does not "+
			"match to the expected.\nCurrent:\n%#v.\nExpected:\n%#v\n", packet_bytes, expected_bytes)
	}
}

func TestResponseDiscoveryPacket(t *testing.T) {
	var address = net.IPv4(127, 0, 0, 1)
	var port uint16 = 8400
	packet := NewResponseDiscoveryPacket(address, port)
	packet_str := packet.String()
	expected_str := "DYLLABLE-DISCOVERY\r\n" +
		"TYPE: RUNNING-APP\r\n" +
		"HOST: 127.0.0.1:8400\r\n" +
		"\r\n"
	if packet_str != expected_str {
		t.Fatalf("Request DiscoveryPacket String() does not "+
			"match to the expected.\nCurrent:\n%#v.\nExpected:\n%#v\n", packet_str, expected_str)
	}

	packet_bytes := packet.Bytes()
	expected_bytes := []byte(expected_str)
	if bytes.Compare(packet_bytes, expected_bytes) != 0 {
		t.Fatalf("Request DiscoveryPacket Bytes() does not "+
			"match to the expected.\nCurrent:\n%#v.\nExpected:\n%#v\n", packet_bytes, expected_bytes)
	}
}
