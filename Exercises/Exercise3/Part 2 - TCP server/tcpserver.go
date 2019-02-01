package main

import (
	"fmt"
	"net"
)

func main() {
	// listen to incoming tcp connections
	l, err := net.Listen("tcp", "127.0.0.1:8081")
	if err != nil {
		fmt.Println("Feil ved listen")
	}
	defer l.Close()

	// A common pattern is to start a loop to continously accept connections
	for {
		fmt.Println("Er i for-løkken")
		//accept connections using Listener.Accept()
		c, err := l.Accept()
		fmt.Println("Kom forbi accept")
		if err != nil {
			fmt.Println("Feil ved accept")
		}

		//It's common to handle accepted connection on different goroutines
		go handleConnection(c)
	}
}

func handleConnection(c net.Conn) {
	//some code...
	fmt.Println("Tar meg av conn")

	for {
		//Simple read from connection
		buffer := make([]byte, 1024)
		n, err := c.Read(buffer)
		fmt.Println("Received ", string(buffer[0:n]))

		if err != nil {
			fmt.Println("Feil ved lesing!")
		}
	}
}
