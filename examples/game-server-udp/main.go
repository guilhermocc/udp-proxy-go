package main

import (
	"flag"
	"fmt"
	"log"
	"net"
)

var serverAddr *string = flag.String("r", "localhost:7777", "server address")

func main() {
	flag.Parse()
	fmt.Printf("Listening: %v\n", *serverAddr)

	serverAddrUdp, err := net.ResolveUDPAddr("udp", *serverAddr)
	if err != nil {
		panic(err)
	}

	udpListener, err := net.ListenUDP("udp", serverAddrUdp)
	if err != nil {
		panic(err)
	}
	for {
		packet := make([]byte, 1024)
		// Reads a packet from the connection
		numberOfBytes, proxyAddress, err := udpListener.ReadFromUDP(packet)
		if numberOfBytes > 0 {
			log.Println("New packet from", proxyAddress)
			log.Println("Packet:", string(packet[:numberOfBytes]))
		}
		if err != nil {
			log.Println("error reading packet", err)
			continue
		}

		proxyConn, err := net.DialUDP("udp", nil, proxyAddress)
		if err != nil {
			log.Println("error dialing back to proxy", err)
			return
		}
		defer proxyConn.Close()

		_, err = proxyConn.Write([]byte(fmt.Sprintf("Hello from game server, i received %s\n", packet)))
		if err != nil {
			log.Println("error writing packet to proxy", err)
			continue
		}
	}

}
