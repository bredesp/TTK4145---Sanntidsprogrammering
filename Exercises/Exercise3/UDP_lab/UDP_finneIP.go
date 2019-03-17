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
	//c := exec.Command("cmd", "/C", "go run", "D:\\Desktop\\NTNU\\Semester 2\\TTK4145 - Sanntidsprogrammering\\GitHub\\TTK4145---Sanntidsprogrammering\\Exercises\\Exercise3\\UDP_lab\\UDP_send.go")
	for {
		fmt.Println("Venter på meldinger:")
		n, addr, err := ServerConn.ReadFromUDP(buf)
		fmt.Println("Received ", string(buf[0:n]), " from ", addr)

		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}
