package network

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"time"
)

const bufferSize = 1024
const writingSocketTimeout = 5 * time.Second
const lookForNodesInterval = 30 * time.Second

func DiscoveryService(ctx context.Context, discoveredNodes chan *net.TCPAddr, discoverySocket *net.UDPAddr, appSocket *net.TCPAddr) (err error) {
	conn, err := net.ListenUDP("udp4", discoverySocket)
	if err != nil {
		return
	}
	defer conn.Close()
	defer close(discoveredNodes)

	closeChannel := make(chan error, 1)

	go func() {
		defer close(closeChannel)

		readBuffer := make([]byte, bufferSize)
		for {
			read, addr, err := conn.ReadFrom(readBuffer)
			if err != nil {
				closeChannel <- err
				return
			}
			buffer := bytes.NewBuffer(readBuffer[:read])
			packet, err := ParsePacket(buffer)
			if err != nil {
				continue
			}
			discoveryPacket, ok := packet.(*DiscoveryPacket)
			if !ok {
				continue
			}
			deadline := time.Now().Add(writingSocketTimeout)
			err = conn.SetWriteDeadline(deadline)
			if err != nil {
				continue
			}
			responsePacket := NewResponseDiscoveryPacket(appSocket.IP, uint16(appSocket.Port))
			responsePacketBytes, _ := responsePacket.Bytes()
			_, err = conn.WriteTo(responsePacketBytes, addr)
			if discoveryPacket.Type == requestDiscoveryType {
				go resolveDiscoveryPacketTCPAddress(discoveryPacket, discoveredNodes)
			}
		}
	}()

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-closeChannel:
	}
	return
}

func LookForNodes(ctx context.Context, discoveredNodes chan *net.TCPAddr, dstAddress *net.UDPAddr, appSocket *net.TCPAddr) (err error) {
	conn, err := net.DialUDP("udp4", nil, dstAddress)
	if err != nil {
		return
	}

	defer conn.Close()
	defer close(discoveredNodes)

	closeSenderChannel := make(chan error, 1)

	//Send Discovery Packet to the network
	go func() {
		defer close(closeSenderChannel)
		for {
			packet := NewRequestDiscoveryPacket(appSocket.IP, uint16(appSocket.Port))
			packetBytes, _ := packet.Bytes()
			_, err := conn.WriteToUDP(packetBytes, dstAddress)
			if err != nil {
				closeSenderChannel <- err
				return
			}
			time.Sleep(lookForNodesInterval)
		}
	}()

	closeReceiverChannel := make(chan error, 1)

	//Receive Discovery Packet from the network
	go func() {
		for {
			readBuffer := make([]byte, bufferSize)
			read, _, err := conn.ReadFrom(readBuffer)
			if err != nil {
				closeReceiverChannel <- err
				return
			}
			buffer := bytes.NewBuffer(readBuffer[:read])
			packet, err := ParsePacket(buffer)
			if err != nil {
				continue
			}
			discoveryPacket, ok := packet.(*DiscoveryPacket)
			if !ok {
				continue
			}
			go resolveDiscoveryPacketTCPAddress(discoveryPacket, discoveredNodes)
		}
	}()

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-closeSenderChannel:
	case err = <-closeReceiverChannel:
	}
	return
}

func resolveDiscoveryPacketTCPAddress(discoveryPacket *DiscoveryPacket, resultChannel chan *net.TCPAddr) {
	address := fmt.Sprintf("%s:%d", discoveryPacket.Address, discoveryPacket.Port)
	theirAppSocket, err := net.ResolveTCPAddr("tcp4", address)
	if err == nil {
		resultChannel <- theirAppSocket
	}
}
