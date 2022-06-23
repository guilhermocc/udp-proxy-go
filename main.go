package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"
)

var localAddr = flag.String("l", "localhost:8888", "local address")
var remoteAddr = flag.String("r", "localhost:7777", "remote address")

func main() {
	flag.Parse()
	fmt.Printf("Listening: %v\nProxying: %v\n\n", *localAddr, *remoteAddr)

	localAddrUdp, err := net.ResolveUDPAddr("udp", *localAddr)
	if err != nil {
		panic(err)
	}

	remoteAddrUdp, err := net.ResolveUDPAddr("udp", *remoteAddr)
	if err != nil {
		panic(err)
	}

	udpListener, err := net.ListenUDP("udp", localAddrUdp)
	if err != nil {
		panic(err)
	}
	for {
		packet := make([]byte, 1024)
		// Reads a packet from the connection
		numberOfBytes, clientAddress, err := udpListener.ReadFromUDP(packet)

		if numberOfBytes > 0 {
			log.Println("New packet from", clientAddress)
			log.Println("Packet:", string(packet[:numberOfBytes]))
		}
		if err != nil {
			log.Println("error reading packet", err)
			continue
		}

		gameServerConn, err := net.DialUDP("udp", nil, remoteAddrUdp)
		if err != nil {
			log.Println("error dialing game server addr", err)
			return
		}
		defer gameServerConn.Close()

		// read data from game server conn
		go func(conn *net.UDPConn) {
			gameServerResponsePacket := make([]byte, 1024)

			numberOfBytes, clientAddress, err = conn.ReadFromUDP(gameServerResponsePacket)
			if numberOfBytes > 0 {
				log.Println("New packet from game room", clientAddress)
				log.Println("Packet:", string(packet[:numberOfBytes]))
			}
			if err != nil {
				log.Println("error reading packet", err)
				return
			}
		}(gameServerConn)

		// send data to gamer server conn
		go func() {
			_, err = gameServerConn.Write(packet[:numberOfBytes])
			if err != nil {
				log.Println("error writing packet to game server", err)
				return
			}
		}()

		time.Sleep(time.Minute * 5)

	}
}
