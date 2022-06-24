package proxy

import (
	"net"
	"time"
)

type GameServerConnection struct {
	udp                    *net.UDPConn
	gameServerAddrResolved *net.UDPAddr
	lastActivity           time.Time
}

type clientPacket struct {
	sourceAddress   *net.UDPAddr
	gameRoomAddress string
	data            []byte
}

type gameServerPacket struct {
	clientAddress *net.UDPAddr
	data          []byte
}

func BuildConnectionCacheKey(clientAddress, gameServerAddress string) string {
	return clientAddress + ":" + gameServerAddress
}
