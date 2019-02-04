package main

import (
	"fmt"
	"net"
)

func main() {

	/*
		ServerAddr, err := net.ResolveUDPAddr("udp", "129.241.187.159:20008")
		if err != nil {
			fmt.Println("Error: ", err)
		}

		if err != nil {
			fmt.Println("Error: ", err)
		}

		Conn, err := net.DialUDP("udp", nil, ServerAddr)
		if err != nil {
			fmt.Println("Error: ", err)
		}

		defer Conn.Close()
		i := 0
		for {
			fmt.Println("Sender en melding")
			msg := strconv.Itoa(i)
			i++
			buf := []byte(msg)
			_, err := Conn.Write(buf)
			if err != nil {
				fmt.Println(msg, err)
			}
			time.Sleep(time.Second * 1)
		}
	*/

	// 129.241.187.159 ???
	/*
		ServerAddr, err := net.ResolveUDPAddr("udp", ":30000")
		if err != nil {
			fmt.Println("Error: ", err)
		}

		// Now listen at selected port
		ServerConn, err := net.ListenUDP("udp", ServerAddr)
	*/

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
