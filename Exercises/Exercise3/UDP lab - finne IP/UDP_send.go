package main

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

func main() {

	ServerAddr, err := net.ResolveUDPAddr("udp", "10.100.23.242:20008")
	if err != nil {
		fmt.Println("Error: ", err)
	}

	// LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
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
		msg := strconv.Itoa(i)
		i++
		buf := []byte(msg)
		_, err := Conn.Write(buf)
		if err != nil {
			fmt.Println(msg, err)
		}
		time.Sleep(time.Second * 1)
	}
}
