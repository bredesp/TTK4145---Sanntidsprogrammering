package main

import (
	"fmt"
	"net"
	//"strconv"
	"time"
)


func receiver(conn net.Conn){
	for {
		//fmt.Println("Er i toppen av for")
		//Simple read from connection
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		fmt.Println("Received ", string(buffer[0:n]))

		if err != nil {
			fmt.Println("Feil ved lesing!")
		}
	}
}

func sender(conn net.Conn){
	msgStart := "Connect to: 10.100.23.233:34933\x00"

	buf := []byte(msgStart)
	_, err := conn.Write(buf)
	if err != nil {
		fmt.Println(msgStart, err)
	}
	time.Sleep(time.Second * 1)

	i := 0
	for {
		//msg := strconv.Itoa(i)
		msg := "test\x00"
		i++
		buf := []byte(msg)
		_, err := conn.Write(buf)
		if err != nil {
			fmt.Println(msg, err)
		}
		time.Sleep(time.Second * 1)
	}
}


func main() {
	//Connect TCP
	conn, err := net.Dial("tcp", "10.100.23.242:34933")
	if err != nil {
		fmt.Println("Error ved dial")
	}
	fmt.Println("Før goroutines")
	//defer conn.Close()

	go sender(conn)
	go receiver(conn)

	for  {
		time.Sleep(time.Second*10)
	}

}
