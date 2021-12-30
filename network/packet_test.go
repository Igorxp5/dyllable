package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"regexp"
	"testing"

	"github.com/google/uuid"
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

	log.Println("Test if the function returns EOF when it does not find the delimeter")
	bufferValue = []byte("DYLLABLE-DISCOVERY")
	buffer = bytes.NewBuffer(bufferValue)
	read, err = readUntil(buffer, delim)
	if err != io.EOF {
		log.Fatalf("1: when the delimiter is not in the buffer, the function should return io.EOF")
	}
	if !bytes.Equal(read, bufferValue) {
		log.Fatalf("2: when the delimiter is not in the buffer, the function should return the same value in the buffer.  %v != %v", read, bufferValue)
	}

	log.Println("Test if the function allows call multiple times in the same buffer (already readed buffer)")
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

func TestRequestActionPacket(t *testing.T) {
	log.Println("Create RequestActionPacket including content")
	actionId := 1
	parameters := make(map[string]interface{})
	parameters["key"] = "value"
	parameters["key2"] = 10
	parameters["key3"] = true
	parameters["key4"] = 15.7

	packet := NewRequestActionPacket(uint8(actionId), parameters)
	packetString, err := packet.String()
	if err != nil {
		log.Fatalf("%v", err)
	}

	requestUUID := packet.RequestUUID.String()
	match, _ := regexp.MatchString("[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", requestUUID)

	if !match {
		log.Fatalf("RequestUUID is following the pattern \"xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx\". Current: %s", requestUUID)
	}

	expectedString := "DYLLABLE-ACTION-REQUEST\r\n" +
		fmt.Sprintf("REQUEST-UUID: %s\r\n", requestUUID) +
		fmt.Sprintf("ACTION-ID: %d\r\n", actionId) +
		"\r\n" +
		"{\"key\":\"value\",\"key2\":10,\"key3\":true,\"key4\":15.7}\r\n" +
		"\r\n"

	if packetString != expectedString {
		t.Fatalf("RequestActionPacket String() does not "+
			"match to the expected.\ncurrent:\n%#v.\nexpected:\n%#v\n", packetString, expectedString)
	}

	packetBytes, err := packet.Bytes()
	if err != nil {
		log.Fatalf("%v", err)
	}
	expectedBytes := []byte(expectedString)
	if !bytes.Equal(packetBytes, expectedBytes) {
		t.Fatalf("RequestActionPacket Bytes() does not "+
			"match to the expected.\ncurrent:\n%#v.\nexpected:\n%#v\n", packetBytes, expectedBytes)
	}

	log.Println("Create RequestActionPacket without content")
	packet = NewRequestActionPacket(uint8(actionId), nil)
	packetString, err = packet.String()
	if err != nil {
		log.Fatalf("%v", err)
	}

	requestUUID = packet.RequestUUID.String()
	match, _ = regexp.MatchString("[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", requestUUID)

	if !match {
		log.Fatalf("RequestUUID is following the pattern \"xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx\". Current: %s", requestUUID)
	}

	expectedString = "DYLLABLE-ACTION-REQUEST\r\n" +
		fmt.Sprintf("REQUEST-UUID: %s\r\n", requestUUID) +
		fmt.Sprintf("ACTION-ID: %d\r\n", actionId) +
		"\r\n"

	if packetString != expectedString {
		t.Fatalf("RequestActionPacket String() does not "+
			"match to the expected.\ncurrent:\n%#v.\nexpected:\n%#v\n", packetString, expectedString)
	}

	packetBytes, err = packet.Bytes()
	if err != nil {
		log.Fatalf("%v", err)
	}
	expectedBytes = []byte(expectedString)
	if !bytes.Equal(packetBytes, expectedBytes) {
		t.Fatalf("RequestActionPacket Bytes() does not "+
			"match to the expected.\ncurrent:\n%#v.\nexpected:\n%#v\n", packetBytes, expectedBytes)
	}
}

func TestResponseActionPacket(t *testing.T) {
	log.Println("Create ResponseActionPacket including content")

	content := make(map[string]interface{})
	content["key"] = "value"
	content["key2"] = 10
	content["key3"] = true
	content["key4"] = 15.7

	requestUUID := uuid.New()
	packet := NewResponseActionPacket(requestUUID, false, content)
	packetString, err := packet.String()
	if err != nil {
		log.Fatalf("%v", err)
	}

	match, _ := regexp.MatchString("[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", requestUUID.String())

	if !match {
		log.Fatalf("RequestUUID is following the pattern \"xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx\". Current: %s", requestUUID)
	}

	expectedString := "DYLLABLE-ACTION-RESPONSE\r\n" +
		fmt.Sprintf("REQUEST-UUID: %s\r\n", requestUUID) +
		"APPROVED: False\r\n" +
		"\r\n" +
		"{\"key\":\"value\",\"key2\":10,\"key3\":true,\"key4\":15.7}\r\n" +
		"\r\n"

	if packetString != expectedString {
		t.Fatalf("RequestActionPacket String() does not "+
			"match to the expected.\ncurrent:\n%#v.\nexpected:\n%#v\n", packetString, expectedString)
	}

	packetBytes, err := packet.Bytes()
	if err != nil {
		log.Fatalf("%v", err)
	}
	expectedBytes := []byte(expectedString)
	if !bytes.Equal(packetBytes, expectedBytes) {
		t.Fatalf("RequestActionPacket Bytes() does not "+
			"match to the expected.\ncurrent:\n%#v.\nexpected:\n%#v\n", packetBytes, expectedBytes)
	}

	log.Println("Create ResponseActionPacket without content")
	requestUUID = uuid.New()
	packet = NewResponseActionPacket(requestUUID, true, nil)
	packetString, err = packet.String()
	if err != nil {
		log.Fatalf("%v", err)
	}

	match, _ = regexp.MatchString("[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}", requestUUID.String())

	if !match {
		log.Fatalf("RequestUUID is following the pattern \"xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx\". Current: %s", requestUUID)
	}

	expectedString = "DYLLABLE-ACTION-RESPONSE\r\n" +
		fmt.Sprintf("REQUEST-UUID: %s\r\n", requestUUID) +
		"APPROVED: True\r\n" +
		"\r\n"

	if packetString != expectedString {
		t.Fatalf("RequestActionPacket String() does not "+
			"match to the expected.\ncurrent:\n%#v.\nexpected:\n%#v\n", packetString, expectedString)
	}

	packetBytes, err = packet.Bytes()
	if err != nil {
		log.Fatalf("%v", err)
	}
	expectedBytes = []byte(expectedString)
	if !bytes.Equal(packetBytes, expectedBytes) {
		t.Fatalf("RequestActionPacket Bytes() does not "+
			"match to the expected.\ncurrent:\n%#v.\nexpected:\n%#v\n", packetBytes, expectedBytes)
	}
}

func TestParseRequestActionPacket(t *testing.T) {
	var requestUUID = uuid.New()
	var actionId uint8 = 1

	parameters := make(map[string]interface{})
	parameters["key"] = "value"
	parameters["key2"] = 10.0 // by default json.Unmarshal parse JSON numbers to float64
	parameters["key3"] = true
	parameters["key4"] = 15.7

	parametersJSON, err := json.Marshal(parameters)
	if err != nil {
		log.Fatalf("%v", err)
	}

	packetString := "DYLLABLE-ACTION-REQUEST\r\n" +
		fmt.Sprintf("REQUEST-UUID: %s\r\n", requestUUID) +
		fmt.Sprintf("ACTION-ID: %d\r\n", actionId) +
		"\r\n" +
		string(parametersJSON) + "\r\n" +
		"\r\n"

	packetBuffer := bytes.NewBuffer([]byte(packetString))
	packet, err := ParsePacket(packetBuffer)
	if err != nil {
		log.Fatalf("%v", err)
	}
	actionPacket, ok := packet.(*RequestActionPacket)
	if !ok {
		log.Fatalf("expected RequestActionPacket object instead of %T", packet)
	}
	if actionPacket.RequestUUID != requestUUID {
		log.Fatalf("expected parsed packet REQUEST-UUID equals to \"%s\" instead of \"%s\"", requestUUID, actionPacket.RequestUUID)
	}
	if actionPacket.ActionId != actionId {
		log.Fatalf("expected parsed packet ACTION-ID equals to \"%d\" instead of \"%v\"", actionId, actionPacket.ActionId)
	}
	if !reflect.DeepEqual(parameters, actionPacket.Parameters) {
		log.Fatalf("expected parsed packet paramaters equals to \"%s\" instead of \"%v\"", string(parametersJSON), actionPacket.Parameters)
	}
}

func TestParseRequestActionPacketWithoutParameters(t *testing.T) {
	var requestUUID = uuid.New()
	var actionId uint8 = 10

	packetString := "DYLLABLE-ACTION-REQUEST\r\n" +
		fmt.Sprintf("REQUEST-UUID: %s\r\n", requestUUID) +
		fmt.Sprintf("ACTION-ID: %d\r\n", actionId) +
		"\r\n"

	packetBuffer := bytes.NewBuffer([]byte(packetString))
	packet, err := ParsePacket(packetBuffer)
	if err != nil {
		log.Fatalf("%v", err)
	}
	actionPacket, ok := packet.(*RequestActionPacket)
	if !ok {
		log.Fatalf("expected RequestActionPacket object instead of %T", packet)
	}
	if actionPacket.RequestUUID != requestUUID {
		log.Fatalf("expected parsed packet REQUEST-UUID equals to \"%s\" instead of \"%s\"", requestUUID, actionPacket.RequestUUID)
	}
	if actionPacket.ActionId != actionId {
		log.Fatalf("expected parsed packet ACTION-ID equals to \"%d\" instead of \"%v\"", actionId, actionPacket.ActionId)
	}
}

func TestParseInvalidRequestActionPacket(t *testing.T) {
	var requestUUID = uuid.New()

	invalidPackets := []string{
		fmt.Sprintf("REQUEST-UUID: %s\r\n", requestUUID) +
			"\r\n",
		"ACTION-ID: 100\r\n" +
			"\r\n",
		"DYLLABLE-ACTION-REQUEST\r\n" +
			fmt.Sprintf("REQUEST-UUID: %s\r\n", requestUUID) +
			"\r\n",
		"DYLLABLE-ACTION-REQUEST\r\n" +
			"ACTION-ID: 100\r\n" +
			"\r\n",
		"DYLLABLE-ACTION-RQUEST\r\n" +
			fmt.Sprintf("REQUEST-UUID: %s\r\n", requestUUID) +
			"ACTION-ID: 100\r\n" +
			"\r\n",
		"DYLLABLE-ACTION-REQUEST\r\n" +
			"ACTION-ID: 100\r\n" +
			"{\"key\":\"value\",\"key2\":10,\"key3\":true,\"key4\":15.7}\r\n" +
			"\r\n",
		"DYLLABLE-ACTION-REQUEST\r\n" +
			fmt.Sprintf("REQUEST-UUID: %s\r\n", requestUUID) +
			"ACTION-ID: 100\r\n" +
			"{\"key\":\"value\",key2\":10,\"key3\":true,\"key4\":15.7}\r\n" +
			"\r\n",
		"DYLLABLE-ACTION-REQUEST\r\n" +
			"REQUEST-UUID: 1\r\n" +
			"ACTION-ID: 100\r\n" +
			"{\"key\":\"value\",\"key2\":10,\"key3\":true,\"key4\":15.7}\r\n" +
			"\r\n",
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

func TestParseResponseActionPacket(t *testing.T) {
	var requestUUID = uuid.New()

	approvedValues := map[string]bool{
		"True":  true,
		"False": false,
	}

	content := make(map[string]interface{})
	content["key"] = "value"
	content["key2"] = 10.0 // by default json.Unmarshal parse JSON numbers to float64
	content["key3"] = true
	content["key4"] = 15.7

	contentJSON, err := json.Marshal(content)
	if err != nil {
		log.Fatalf("%v", err)
	}

	for key, approved := range approvedValues {
		packetString := "DYLLABLE-ACTION-RESPONSE\r\n" +
			fmt.Sprintf("REQUEST-UUID: %s\r\n", requestUUID) +
			fmt.Sprintf("APPROVED: %s\r\n", key) +
			"\r\n" +
			string(contentJSON) + "\r\n" +
			"\r\n"

		packetBuffer := bytes.NewBuffer([]byte(packetString))
		packet, err := ParsePacket(packetBuffer)
		if err != nil {
			log.Fatalf("%v", err)
		}
		responsePacket, ok := packet.(*ResponseActionPacket)
		if !ok {
			log.Fatalf("expected ResponseActionPacket object instead of %T", packet)
		}
		if responsePacket.RequestUUID != requestUUID {
			log.Fatalf("expected parsed packet REQUEST-UUID equals to \"%s\" instead of \"%s\"", requestUUID, responsePacket.RequestUUID)
		}
		if responsePacket.Approved != approved {
			log.Fatalf("expected parsed packet APPROVED equals to \"%v\" instead of \"%v\"", approved, responsePacket.Approved)
		}
		if !reflect.DeepEqual(content, responsePacket.Content) {
			log.Fatalf("expected parsed packet content equals to \"%s\" instead of \"%v\"", string(contentJSON), responsePacket.Content)
		}
	}
}

func TestParseResponseActionPacketWithoutContent(t *testing.T) {
	var requestUUID = uuid.New()

	approvedValues := map[string]bool{
		"True":  true,
		"False": false,
	}

	for key, approved := range approvedValues {
		packetString := "DYLLABLE-ACTION-RESPONSE\r\n" +
			fmt.Sprintf("REQUEST-UUID: %s\r\n", requestUUID) +
			fmt.Sprintf("APPROVED: %s\r\n", key) +
			"\r\n"

		packetBuffer := bytes.NewBuffer([]byte(packetString))
		packet, err := ParsePacket(packetBuffer)
		if err != nil {
			log.Fatalf("%v", err)
		}
		responsePacket, ok := packet.(*ResponseActionPacket)
		if !ok {
			log.Fatalf("expected RequestActionPacket object instead of %T", packet)
		}
		if responsePacket.RequestUUID != requestUUID {
			log.Fatalf("expected parsed packet REQUEST-UUID equals to \"%s\" instead of \"%s\"", requestUUID, responsePacket.RequestUUID)
		}
		if responsePacket.Approved != approved {
			log.Fatalf("expected parsed packet APPROVED equals to \"%v\" instead of \"%v\"", approved, responsePacket.Approved)
		}
	}
}

func TestParseInvalidResponseActionPacket(t *testing.T) {
	var requestUUID = uuid.New()

	invalidPackets := []string{
		fmt.Sprintf("REQUEST-UUID: %s\r\n", requestUUID) +
			"\r\n",
		"APPROVED: True\r\n" +
			"\r\n",
		"DYLLABLE-ACTION-RESPONSE\r\n" +
			fmt.Sprintf("REQUEST-UUID: %s\r\n", requestUUID) +
			"\r\n",
		"DYLLABLE-ACTION-RESPONSE\r\n" +
			"APPROVED: False\r\n" +
			"\r\n",
		"DYLLABLE-ACTION-RESPON\r\n" +
			fmt.Sprintf("REQUEST-UUID: %s\r\n", requestUUID) +
			"APPROVED: True\r\n" +
			"\r\n",
		"DYLLABLE-ACTION-RESPONSE\r\n" +
			"APPROVED: False\r\n" +
			"{\"key\":\"value\",\"key2\":10,\"key3\":true,\"key4\":15.7}\r\n" +
			"\r\n",
		"DYLLABLE-ACTION-RESPONSE\r\n" +
			fmt.Sprintf("REQUEST-UUID: %s\r\n", requestUUID) +
			"APPROVED: True\r\n" +
			"{\"key\":\"value\",key2\":10,\"key3\":true,\"key4\":15.7}\r\n" +
			"\r\n",
		"DYLLABLE-ACTION-RESPONSE\r\n" +
			"REQUEST-UUID: 1\r\n" +
			"APPROVED: FAlse\r\n" +
			"{\"key\":\"value\",\"key2\":10,\"key3\":true,\"key4\":15.7}\r\n" +
			"\r\n",
		"DYLLABLE-ACTION-RESPONSE\r\n" +
			fmt.Sprintf("REQUEST-UUID: %s\r\n", requestUUID) +
			"Approved: True\r\n" +
			"\r\n",
		"DYLLABLE-ACTION-RESPONSE\r\n" +
			fmt.Sprintf("REQUEST-UUID: %s\r\n", requestUUID) +
			"APPROVED: true\r\n" +
			"\r\n",
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
