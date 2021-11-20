package network

import (
	"bytes"
	"io"
	"log"
	"net"
	"testing"
)

func TestRequestDiscoveryPacket(t *testing.T) {
	var address = net.IPv4(127, 0, 0, 1)
	var port uint16 = 8400
	packet := NewRequestDiscoveryPacket(address, port)
	packetString, err := packet.String()
	if err != nil {
		log.Fatalf("%v", err)
	}
	expectedString := "DYLLABLE-DISCOVERY\r\n" +
		"TYPE: DISCOVERY\r\n" +
		"HOST: 127.0.0.1:8400\r\n" +
		"\r\n"
	if packetString != expectedString {
		t.Fatalf("RequestDiscoveryPacket String() does not "+
			"match to the expected.\ncurrent:\n%#v.\nexpected:\n%#v\n", packetString, expectedString)
	}

	packetBytes, err := packet.Bytes()
	if err != nil {
		log.Fatalf("%v", err)
	}
	expectedBytes := []byte(expectedString)
	if !bytes.Equal(packetBytes, expectedBytes) {
		t.Fatalf("RequestDiscoveryPacket Bytes() does not "+
			"match to the expected.\ncurrent:\n%#v.\nexpected:\n%#v\n", packetBytes, expectedBytes)
	}
}

func TestResponseDiscoveryPacket(t *testing.T) {
	var address = net.IPv4(127, 0, 0, 1)
	var port uint16 = 8400
	packet := NewResponseDiscoveryPacket(address, port)
	packetString, err := packet.String()
	if err != nil {
		log.Fatalf("%v", err)
	}
	expectedString := "DYLLABLE-DISCOVERY\r\n" +
		"TYPE: RUNNING-APP\r\n" +
		"HOST: 127.0.0.1:8400\r\n" +
		"\r\n"
	if packetString != expectedString {
		t.Fatalf("ResponseDiscoveryPacket String() does not "+
			"match to the expected.\ncurrent:\n%#v.\nexpected:\n%#v\n", packetString, expectedString)
	}

	packetBytes, err := packet.Bytes()
	if err != nil {
		log.Fatalf("%v", err)
	}
	expectedBytes := []byte(expectedString)
	if !bytes.Equal(packetBytes, expectedBytes) {
		t.Fatalf("ResponseDiscoveryPacket Bytes() does not "+
			"match to the expected.\ncurrent:\n%#v.\nexpected:\n%#v\n", packetBytes, expectedBytes)
	}
}

func TestParseRequestDiscoveryPacket(t *testing.T) {
	var address = net.IPv4(127, 0, 0, 1)
	var port uint16 = 8400
	packetString := "DYLLABLE-DISCOVERY\r\n" +
		"TYPE: DISCOVERY\r\n" +
		"HOST: 127.0.0.1:8400\r\n" +
		"\r\n"
	packetBuffer := bytes.NewBuffer([]byte(packetString))
	packet, err := ParsePacket(packetBuffer)
	if err != nil {
		log.Fatalf("%v", err)
	}
	discoveryPacket, ok := packet.(*DiscoveryPacket)
	if !ok {
		log.Fatalf("expected DiscoveryPacket object instead of %T", packet)
	}
	if !discoveryPacket.Address.Equal(address) {
		log.Fatalf("expected parsed packet IP equals to \"127.0.0.1\" instead of \"%v\"", discoveryPacket.Address)
	}
	if discoveryPacket.Port != port {
		log.Fatalf("expected parsed packet port equals to \"8400\" instead of \"%v\"", discoveryPacket.Port)
	}
	if discoveryPacket.Type != requestDiscoveryType {
		log.Fatalf("expected parsed packet type equals to \"%s\" instead of \"%s\"", requestDiscoveryType, discoveryPacket.Type)
	}
}

func TestParseResponseDiscoveryPacket(t *testing.T) {
	var address = net.IPv4(127, 0, 0, 1)
	var port uint16 = 8400
	packetString := "DYLLABLE-DISCOVERY\r\n" +
		"TYPE: RUNNING-APP\r\n" +
		"HOST: 127.0.0.1:8400\r\n" +
		"\r\n"
	packetBuffer := bytes.NewBuffer([]byte(packetString))
	packet, err := ParsePacket(packetBuffer)
	if err != nil {
		log.Fatalf("%v", err)
	}
	discoveryPacket, ok := packet.(*DiscoveryPacket)
	if !ok {
		log.Fatalf("expected DiscoveryPacket object instead of %T", packet)
	}
	if !discoveryPacket.Address.Equal(address) {
		log.Fatalf("expected parsed packet IP equals to \"127.0.0.1\" instead of \"%v\"", discoveryPacket.Address)
	}
	if discoveryPacket.Port != port {
		log.Fatalf("expected parsed packet port equals to \"8400\" instead of \"%v\"", discoveryPacket.Port)
	}
	if discoveryPacket.Type != responseDiscoveryType {
		log.Fatalf("expected parsed packet type equals to \"%s\" instead of \"%s\"", responseDiscoveryType, discoveryPacket.Type)
	}
}

func TestParseInvalidDiscoveryPacket(t *testing.T) {
	invalidPackets := []string{
		"TYPE: DISCOVERY\r\n" +
			"HOST: 127.0.0.1:8400\r\n" +
			"\r\n",
		"DYLLABLE-DYSCOVERY\r\n" +
			"TYPE: DISCOVERY\r\n" +
			"HOST: 127.0.0.1:8400\r\n" +
			"\r\n",
		"DYLLABLE-DISCOVERY\r\n" +
			"TYPE: INVALID\r\n" +
			"HOST: 127.0.0.1:8400\r\n" +
			"\r\n",
		"DYLLABLE-DISCOVERY\r\n" +
			"TYPE: RUNNING-APP\r\n" +
			"\r\n",
		"DYLLABLE-DISCOVERY\r\n" +
			"HOST: 127.0.0.1:8400\r\n" +
			"\r\n",
		"DYLLABLE-DISCOVERY\n" +
			"TYPE: INVALID\r\n" +
			"HOST: 127.0.0.1:8400\r\n" +
			"\r\n",
		"DYLLABLE-DISCOVERY\r\n" +
			"TYPE: INVALID\n" +
			"HOST: 127.0.0.1:8400\r\n" +
			"\r\n",
		"DYLLABLE-DISCOVERY\r\n" +
			"TYPE: INVALID\n" +
			"HOST: 127.0.0.1:8400\r\n" +
			"\n",
	}

	var packetBuffer *bytes.Buffer
	for _, packetString := range invalidPackets {
		packetBuffer = bytes.NewBuffer([]byte(packetString))
		_, err := ParsePacket(packetBuffer)
		if err == nil {
			log.Fatalf("following packet should be invalid: \n%s", packetString)
		}
	}
}

func TestReadUntil(t *testing.T) {
	var bufferValue []byte
	var buffer *bytes.Buffer
	var delim = []byte("\r\n")
	var read []byte
	var err error

	//Test if the function returns EOF when it does not find the delimeter
	bufferValue = []byte("DYLLABLE-DISCOVERY")
	buffer = bytes.NewBuffer(bufferValue)
	read, err = readUntil(buffer, delim)
	if err != io.EOF {
		log.Fatalf("1: when the delimiter is not in the buffer, the function should return io.EOF")
	}
	if !bytes.Equal(read, bufferValue) {
		log.Fatalf("2: when the delimiter is not in the buffer, the function should return the same value in the buffer.  %v != %v", read, bufferValue)
	}

	//Test if the function allows call multiple times in the same buffer (already readed buffer)
	bufferValue = []byte("DYLLABLE-DISCOVERY\r\nTYPE: DISCOVERY\r\n\r\n")

	buffer = bytes.NewBuffer(bufferValue)
	read, err = readUntil(buffer, delim)
	if err != nil {
		log.Fatalf("3: %v", err)
	}
	if !bytes.Equal(read, []byte("DYLLABLE-DISCOVERY\r\n")) {
		log.Fatalf("4: %#v != %#v", string(read), "DYLLABLE-DISCOVERY\r\n")
	}
	read, err = readUntil(buffer, delim)
	if err != nil {
		log.Fatalf("4: %v", err)
	}
	if !bytes.Equal(read, []byte("TYPE: DISCOVERY\r\n")) {
		log.Fatalf("5: %#v != %#v", string(read), "TYPE: DISCOVERY\r\n")
	}
	read, err = readUntil(buffer, delim)
	if err != nil {
		log.Fatalf("6: %v", err)
	}
	if !bytes.Equal(read, delim) {
		log.Fatalf("7: %#v != %#v", string(read), string(delim))
	}
	read, err = readUntil(buffer, delim)
	if err != io.EOF {
		log.Fatalf("8: the function should return io.EOF when buffer has reached to the end")
	}
}
