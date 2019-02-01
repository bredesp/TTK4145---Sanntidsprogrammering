package main

import (
	"fmt"
	"net"
)

func main() {
	// listen to incoming tcp connections
	// serverAdress, err := ResolveTCPAddr("tcp", "10.100.23.242:34933")
	l, err := net.Dial("tcp", "10.100.23.242:39970")
	if err != nil {
		fmt.Println("Feil ved listen")
	}
	fmt.Println("Kom foran defer")
	//defer l.Close()


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
