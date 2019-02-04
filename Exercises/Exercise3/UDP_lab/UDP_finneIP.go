package main

import (
	"fmt"
	"net"
)

func main() {

	ServerConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: []byte{0, 0, 0, 0}, Port: 20008, Zone: ""})
	if err != nil {
		fmt.Println("Error: ", err)
	}
	defer ServerConn.Close()

	buf := make([]byte, 1024)

	for {
		fmt.Println("Venter på meldinger:")
		n, addr, err := ServerConn.ReadFromUDP(buf)
		fmt.Println("Received ", string(buf[0:n]), " from ", addr)

		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}
