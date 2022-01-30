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

func DiscoveryService(ctx context.Context, discoveredNodes chan *net.TCPAddr, discoverySocket *net.UDPAddr, appSocket *net.TCPAddr) (err error) {
	conn, err := net.ListenUDP("udp4", discoverySocket)
	if err != nil {
		return
	}
	defer conn.Close()

	closeChannel := make(chan error, 1)

	go func() {
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
				//Resolve Addresses should not block to receive incoming DiscoveryPackets
				go func() {
					address := fmt.Sprintf("%s:%d", discoveryPacket.Address, discoveryPacket.Port)
					theirAppSocket, err := net.ResolveTCPAddr("tcp4", address)
					if err == nil {
						discoveredNodes <- theirAppSocket
					}
				}()
			}
		}
	}()

	select {
	case <-ctx.Done():
		err = ctx.Err()
	}
	return
}
