package proxy

import (
	"go.uber.org/zap"
	"net"
	"sync"
	"time"
)

const localAddr = "localhost:8888"

const remoteAddr = "localhost:7777"

// We are going to select a random port on execution time
const clientAddr = "localhost:2000"

// This buffer size should be defined by a configuration
const bufferSize = 1500

const connectionTimeout = time.Minute * 1

const idleConnectionsSyncPeriod = time.Second

type Proxy struct {
	connectionsCache          sync.Map
	datagramsBufferSize       int
	mainPacketListenerConn    *net.UDPConn
	clientPacketChannel       chan clientPacket
	gameServerPacketChannel   chan gameServerPacket
	connectionTimeout         time.Duration
	idleConnectionsSyncPeriod time.Duration
}

func NewProxy() *Proxy {
	return &Proxy{
		connectionsCache:          sync.Map{},
		datagramsBufferSize:       bufferSize,
		clientPacketChannel:       make(chan clientPacket),
		gameServerPacketChannel:   make(chan gameServerPacket),
		connectionTimeout:         connectionTimeout,
		idleConnectionsSyncPeriod: idleConnectionsSyncPeriod,
	}
}

func (p *Proxy) readLoop() {
	datagramBuffer := make([]byte, p.datagramsBufferSize)
	for {
		size, srcAddress, err := p.mainPacketListenerConn.ReadFromUDP(datagramBuffer)
		if err != nil {
			zap.L().Error("error", zap.Error(err))
			continue
		}
		if size > 0 {
			//TODO: We should parse the packet here to extract the game room address prefix
			gameRoomAddress, err := p.parseClientPacket(datagramBuffer)
			if err != nil {
				// TODO: think what should we do here
				zap.L().Error("Could not parse client packet")
			}

			p.clientPacketChannel <- clientPacket{
				sourceAddress:   srcAddress,
				gameRoomAddress: gameRoomAddress,
				data:            datagramBuffer[:size],
			}
		}
	}
}

func (p *Proxy) parseClientPacket(buffer []byte) (string, error) {
	// TODO: we must implement the parse logic here
	return remoteAddr, nil
}

func (p *Proxy) handleClientPackets() {
	for packt := range p.clientPacketChannel {
		packetSourceString := packt.sourceAddress.String()
		zap.L().Debug("packet received",
			zap.String("src address", packetSourceString),
			zap.Int("src port", packt.sourceAddress.Port),
			zap.String("packet", string(packt.data)),
			zap.Int("size", len(packt.data)),
		)

		conn, found := p.connectionsCache.Load(BuildConnectionCacheKey(packt.sourceAddress.String(), packt.gameRoomAddress))
		if !found {
			gameServerConnection, err := net.ListenUDP("udp", nil)
			zap.L().Debug("new client connection",
				zap.String("local port", gameServerConnection.LocalAddr().String()),
			)

			if err != nil {
				zap.L().Error("upd proxy failed to dial", zap.Error(err))
				return
			}
			gameServerAddrResolved, err := net.ResolveUDPAddr("udp", packt.gameRoomAddress)

			p.connectionsCache.Store(BuildConnectionCacheKey(packetSourceString, packt.gameRoomAddress), &GameServerConnection{
				udp:                    gameServerConnection,
				gameServerAddrResolved: gameServerAddrResolved,
				lastActivity:           time.Now(),
			})

			gameServerConnection.WriteToUDP(packt.data, gameServerAddrResolved)

			// Start process to keep listening from incoming packets from game server
			go p.listenToGameServerPackets(packt.sourceAddress, gameServerAddrResolved, gameServerConnection)
		} else {
			gameServerConnection := conn.(*GameServerConnection)
			gameServerConnection.udp.WriteTo(packt.data, gameServerConnection.gameServerAddrResolved)

			// TODO: Here we should update the connection last activity
		}
		p.updateClientLastActivity(BuildConnectionCacheKey(packt.sourceAddress.String(), packt.gameRoomAddress))
	}
}

func (p *Proxy) updateClientLastActivity(connectionCacheKey string) {
	zap.L().Debug("updating client last activity", zap.String("client", connectionCacheKey))
	if connWrapper, found := p.connectionsCache.Load(connectionCacheKey); found {
		connWrapper.(*GameServerConnection).lastActivity = time.Now()
	}
}

func (p *Proxy) listenToGameServerPackets(clientAddr *net.UDPAddr, gameServerAddr *net.UDPAddr, gameServerConn *net.UDPConn) {
	clientAddrString := clientAddr.String()
	gameServerAddrString := gameServerAddr.String()
	datagramBuffer := make([]byte, p.datagramsBufferSize)
	for {
		size, _, err := gameServerConn.ReadFromUDP(datagramBuffer)
		if err != nil {
			gameServerConn.Close()
			p.connectionsCache.Delete(clientAddrString)
			return
		}
		// TODO: update client last activity
		p.updateClientLastActivity(BuildConnectionCacheKey(clientAddrString, gameServerAddrString))
		p.gameServerPacketChannel <- gameServerPacket{
			clientAddress: clientAddr,
			data:          datagramBuffer[:size],
		}
	}
}

func (p *Proxy) handleGameServerPackets() {
	for packt := range p.gameServerPacketChannel {
		zap.L().Debug("forwarded data from game server back to client", zap.Int("size", len(packt.data)), zap.String("data", string(packt.data)))
		p.mainPacketListenerConn.WriteTo(packt.data, packt.clientAddress)
	}
}

func (p *Proxy) setupMainListener() {
	localAddrUdp, err := net.ResolveUDPAddr("udp", localAddr)
	if err != nil {
		panic(err)
	}

	udpListener, err := net.ListenUDP("udp", localAddrUdp)
	if err != nil {
		zap.L().Error("Error creating main listener")
		panic(err)
	}

	p.mainPacketListenerConn = udpListener
}

func (p *Proxy) freeIdleSocketsLoop() {
	for {
		time.Sleep(p.idleConnectionsSyncPeriod)
		var clientsToTimeout []string

		p.connectionsCache.Range(func(k, conn interface{}) bool {
			if conn.(*GameServerConnection).lastActivity.Before(time.Now().Add(-p.connectionTimeout)) {
				clientsToTimeout = append(clientsToTimeout, k.(string))
			}
			return true
		})

		for _, client := range clientsToTimeout {
			zap.L().Debug("client timeout", zap.String("client", client))
			conn, ok := p.connectionsCache.Load(client)
			if ok {
				conn.(*GameServerConnection).udp.Close()
				p.connectionsCache.Delete(client)
			}
		}
	}
}

func (p *Proxy) RunProxy() {
	p.setupMainListener()

	go p.freeIdleSocketsLoop()
	go p.readLoop()
	go p.handleClientPackets()
	go p.handleGameServerPackets()

}
