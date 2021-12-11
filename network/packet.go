package network

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

const preffix = "DYLLABLE-"
const actionPreffix = preffix + "ACTION-"
const discoveryIdentifier = preffix + "DISCOVERY"

const headerSeparator = "\r\n"
const endPacket = "\r\n"

const requestDiscoveryType = "DISCOVERY"
const responseDiscoveryType = "RUNNING-APP"

const requestActionIdentifier = actionPreffix + "REQUEST"
const responseActionIdentifier = actionPreffix + "RESPONSE"

type Packet interface {
	String() (string, error)
	Bytes() ([]byte, error)
}

type DiscoveryPacket struct {
	Address net.IP
	Port    uint16
	Type    string
}

func (packet *DiscoveryPacket) String() (string, error) {
	typeHeader := fmt.Sprintf("TYPE: %s", packet.Type)
	hostHeader := fmt.Sprintf("HOST: %s:%d", packet.Address, packet.Port)
	headers := []string{discoveryIdentifier, typeHeader, hostHeader}
	return strings.Join(headers, headerSeparator) + headerSeparator + endPacket, nil
}

func (packet *DiscoveryPacket) Bytes() (out []byte, err error) {
	packetString, err := packet.String()
	out = []byte(packetString)
	return
}

func NewRequestDiscoveryPacket(address net.IP, port uint16) DiscoveryPacket {
	return DiscoveryPacket{address, port, requestDiscoveryType}
}

func NewResponseDiscoveryPacket(address net.IP, port uint16) DiscoveryPacket {
	return DiscoveryPacket{address, port, responseDiscoveryType}
}

type RequestActionPacket struct {
	RequestUUID uuid.UUID
	ActionId    uint8
	Parameters  map[string]interface{}
}

type ResponseActionPacket struct {
	RequestUUID uuid.UUID
	Approved    bool
	Content     map[string]interface{}
}

func (packet *RequestActionPacket) String() (out string, err error) {
	requestUUIDHeader := fmt.Sprintf("REQUEST-UUID: %s", packet.RequestUUID.String())
	actionHeader := fmt.Sprintf("ACTION-ID: %d", packet.ActionId)
	parametersJSON, err := json.Marshal(packet.Parameters)
	if err != nil {
		return
	}
	headers := []string{requestActionIdentifier, requestUUIDHeader, actionHeader}
	out = strings.Join(headers, headerSeparator) + headerSeparator
	if packet.Parameters != nil {
		out += headerSeparator + string(parametersJSON) + headerSeparator
	}
	out += endPacket
	return out, err
}

func (packet *RequestActionPacket) Bytes() (out []byte, err error) {
	packetString, err := packet.String()
	out = []byte(packetString)
	return out, err
}

func (packet *ResponseActionPacket) String() (out string, err error) {
	requestUUIDHeader := fmt.Sprintf("REQUEST-UUID: %s", packet.RequestUUID.String())
	var approved string
	if packet.Approved {
		approved = "True"
	} else {
		approved = "False"
	}
	approvedHeader := fmt.Sprintf("APPROVED: %s", approved)
	contentJSON, err := json.Marshal(packet.Content)
	if err != nil {
		return
	}
	headers := []string{responseActionIdentifier, requestUUIDHeader, approvedHeader}
	out = strings.Join(headers, headerSeparator) + headerSeparator
	if packet.Content != nil {
		out += headerSeparator + string(contentJSON) + headerSeparator
	}
	out += endPacket
	return out, err
}

func (packet *ResponseActionPacket) Bytes() (out []byte, err error) {
	packetString, err := packet.String()
	out = []byte(packetString)
	return
}

func NewRequestActionPacket(actionId uint8, parameters map[string]interface{}) RequestActionPacket {
	return RequestActionPacket{uuid.New(), actionId, parameters}
}

func NewResponseActionPacket(requestUUID uuid.UUID, approved bool, content map[string]interface{}) ResponseActionPacket {
	return ResponseActionPacket{requestUUID, approved, content}
}

func ParsePacket(buffer *bytes.Buffer) (packet Packet, err error) {
	identifier, err := readUntil(buffer, []byte(headerSeparator))
	if err != nil {
		return
	}
	var headers map[string]string
	switch strings.TrimRight(string(identifier), "\r\n") {
	case requestActionIdentifier:
		var packetRequestUUID uuid.UUID
		var packetActionId uint64
		var parametersBytes []byte
		var parametersJSON map[string]interface{}
		headers, err = readHeaders(buffer)
		if err != nil {
			return
		}
		packetRequestUUIDString, ok := headers["REQUEST-UUID"]
		if !ok {
			return packet, errors.New("malformed packet: \"REQUEST-UUID\" header not found")
		}
		packetRequestUUID, err = uuid.Parse(packetRequestUUIDString)
		if err != nil {
			return
		}
		packetActionIdString, ok := headers["ACTION-ID"]
		if !ok {
			return packet, errors.New("malformed packet: \"ACTION-ID\" header not found")
		}
		packetActionId, err = strconv.ParseUint(packetActionIdString, 10, 8)
		if err != nil {
			return
		}
		parametersBytes, err = readUntil(buffer, []byte(headerSeparator))
		if err == nil {
			err = json.Unmarshal(parametersBytes, &parametersJSON)
			if err != nil {
				return
			}
		}
		err = nil
		packetObj := RequestActionPacket{packetRequestUUID, uint8(packetActionId), parametersJSON}
		packet = &packetObj
	case responseActionIdentifier:
		//TODO: Response action packet
	case discoveryIdentifier:
		headers, err = readHeaders(buffer)
		if err != nil {
			return
		}
		packetType, ok := headers["TYPE"]
		if !ok {
			return packet, errors.New("malformed packet: \"TYPE\" header not found")
		}
		packetHost, ok := headers["HOST"]
		if !ok {
			return packet, errors.New("malformed packet: \"HOST\" header not found")
		}
		var host, portString string
		host, portString, err = net.SplitHostPort(packetHost)
		if err != nil {
			return packet, errors.New("malformed packet: invalid ip/port pair")
		}
		ip := net.ParseIP(host)
		if ip == nil {
			return packet, errors.New("malformed packet: hostname is not supported")
		}
		var port uint64
		port, err = strconv.ParseUint(portString, 10, 16)
		if err != nil {
			return
		}
		switch packetType {
		case requestDiscoveryType:
			packetObj := NewRequestDiscoveryPacket(ip, uint16(port))
			packet = &packetObj
		case responseDiscoveryType:
			packetObj := NewResponseDiscoveryPacket(ip, uint16(port))
			packet = &packetObj
		default:
			return packet, errors.New(fmt.Sprintf("malformed packet: type \"%s\" not supported", packetType))
		}
	default:
		err = errors.New("malformed packet: identifier not knwon")
	}
	return packet, err
}

func readUntil(buffer *bytes.Buffer, delim []byte) ([]byte, error) {
	var err error
	var readByte byte
	out := make([]byte, 0, buffer.Cap())
	read := make([]byte, len(delim))
	for !bytes.Equal(read, delim) && err != io.EOF {
		readByte, err = buffer.ReadByte()
		if err != io.EOF {
			out = append(out, readByte)
		}
		read[0] = read[1]
		read[1] = readByte
	}
	readOut := make([]byte, len(out))
	copy(readOut, out)
	return readOut, err
}

func readHeaders(buffer *bytes.Buffer) (headers map[string]string, err error) {
	headers = make(map[string]string)
	headerRaw, err := readUntil(buffer, []byte(headerSeparator))
	var headerPair []string
	var header string
	for err == nil && string(headerRaw) != "\r\n" {
		header = strings.TrimRight(string(headerRaw), "\r\n")
		headerPair, err = parseHeader(header)
		if err != nil {
			return
		}
		headers[headerPair[0]] = headerPair[1]
		headerRaw, err = readUntil(buffer, []byte(headerSeparator))
	}
	return headers, err
}

func parseHeader(raw string) ([]string, error) {
	headerValue := strings.SplitN(raw, ": ", 2)
	if len(headerValue) == 1 {
		return []string{}, errors.New("invalid header line")
	}
	return headerValue, nil
}
