package main

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

func main() {
	//Connect TCP
	conn, err := net.Dial("tcp", "10.22.228.184:30000")
	if err != nil {
		fmt.Println("Error ved dial")
	}
	defer conn.Close()

	i := 0
	for {
		msg := strconv.Itoa(i)
		i++
		buf := []byte(msg)
		_, err := conn.Write(buf)
		if err != nil {
			fmt.Println(msg, err)
		}
		time.Sleep(time.Second * 1)
	}

}
