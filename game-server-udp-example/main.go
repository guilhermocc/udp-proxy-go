package main

import (
	"flag"
	"fmt"
	"log"
	"net"
)

var serverAddr *string = flag.String("r", "localhost:7777", "server address")
var proxyAddr *string = flag.String("bla", "localhost:8888", "server address")

func main() {
	flag.Parse()
	fmt.Printf("Listening: %v\n", *serverAddr)

	proxyAddrUdp, err := net.ResolveUDPAddr("udp", *proxyAddr)
	if err != nil {
		panic(err)
	}

	udpListener, err := net.ListenPacket("udp", *serverAddr)
	if err != nil {
		panic(err)
	}
	for {
		packet := make([]byte, 1024)
		// Reads a packet from the connection
		numberOfBytes, sourceAddress, err := udpListener.ReadFrom(packet)
		if numberOfBytes > 0 {
			log.Println("New packet from", sourceAddress)
			log.Println("Packet:", string(packet[:numberOfBytes]))
		}
		if err != nil {
			log.Println("error reading packet", err)
			continue
		}

		proxyConn, err := net.DialUDP("udp", nil, proxyAddrUdp)
		if err != nil {
			log.Println("error dialing back to proxy", err)
			return
		}
		defer proxyConn.Close()

		_, err = proxyConn.Write([]byte("Hello from game server"))
		if err != nil {
			log.Println("error writing packet to proxy", err)
			continue
		}
	}

}
