package proxy

import (
	"flag"
	"fmt"
	"log"
	"net"
	"syscall"
)

const localAddr = "localhost:8888"

const remoteAddr = "localhost:7777"

// We are going to select a random port on execution time
const clientAddr = "localhost:2000"

// This buffer size should be defined by a configuration
const bufferSize = 1500

func RunProxy() {
	flag.Parse()
	fmt.Printf("Listening: %v\nProxying: %v\n\n", localAddr, remoteAddr)

	localAddrUdp, err := net.ResolveUDPAddr("udp", localAddr)
	if err != nil {
		panic(err)
	}

	//clientAddrUdp, err := net.ResolveUDPAddr("udp", *clientAddr)
	//if err != nil {
	//	panic(err)
	//}
	//
	//remoteAddrUdp, err := net.ResolveUDPAddr("udp", *remoteAddr)
	//if err != nil {
	//	panic(err)
	//}

	udpListener, err := net.ListenUDP("udp", localAddrUdp)
	if err != nil {
		panic(err)
	}
	for {
		packet := make([]byte, 1500)

		// Reads a packet from the connection
		numberOfBytes, _, flags, clientAddress, err := udpListener.ReadMsgUDP(packet, make([]byte, 0))

		// If this flag is set, it means that we were unable to read the entire packet payload
		if flags&syscall.MSG_TRUNC != 0 {
			panic("unable")
			fmt.Println("truncated read")
		}
		if numberOfBytes > 0 {
			log.Println("New packet from", clientAddress)
			log.Println("Packet:", string(packet[:numberOfBytes]))
		}
		if err != nil {
			log.Println("error reading packet", err)
			continue
		}
		//
		//gameServerConn, err := net.ListenUDP("udp", clientAddrUdp)
		//if err != nil {
		//	log.Println("error dialing game server addr", err)
		//	return
		//}
		//defer gameServerConn.Close()
		//
		//// send data to gamer server conn
		//go func() {
		//	_, err = gameServerConn.WriteTo(packet[:numberOfBytes], remoteAddrUdp)
		//	if err != nil {
		//		log.Println("error writing packet to game server", err)
		//		return
		//	}
		//}()
		//
		//gameServerResponsePacket := make([]byte, 1024)
		//
		//numberOfBytes, gruAddress, err := gameServerConn.ReadFromUDP(gameServerResponsePacket)
		//if numberOfBytes > 0 {
		//	log.Println("New packet from game room", gruAddress)
		//	log.Println("Packet:", string(packet[:numberOfBytes]))
		//}
		//if err != nil {
		//	log.Println("error reading packet", err)
		//	return
		//}
		//
		//gameServerConn.WriteTo(gameServerResponsePacket, clientAddress)

	}
}
