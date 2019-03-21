package Nettverk

import (
	"fmt"
	"net"
	"strconv"
)

var local_adress *net.UDPAddr
var broadcast_adress *net.UDPAddr

type UDP_message struct {
	Raddr  string
	Data   string
	Length int
}

func UDP_init(local_port, broadcast_port, msg_size int, ch_send, ch_receive chan UDP_message) (err error) {
	// Resolves broadcasting adress
	broadcast_adress, err = net.ResolveUDPAddr("udp4", "255.255.255.255:"+strconv.Itoa(broadcast_port))
	if err != nil {
		return err
	}

	// Finding and resolving the local adress
	temp_conn, err := net.DialUDP("udp4", nil, broadcast_adress)
	defer temp_conn.Close()
	temp_address := temp_conn.LocalAddr()
	local_adress, err = net.ResolveUDPAddr("udp4", temp_address.String())
	local_adress.Port = local_port

	// Creates connection for the local adress
	local_conn, err := net.ListenUDP("udp4", local_adress)
	if err != nil {
		return err
	}

	// Creates  connection for the broadcasting adress
	broadcast_conn, err := net.ListenUDP("udp", broadcast_adress)
	if err != nil {
		local_conn.Close()
		return err
	}

	go UDP_sender(local_conn, broadcast_conn, ch_send)
	go UDP_receiver(local_conn, broadcast_conn, msg_size, ch_receive)

	return err
}

func UDP_sender(local_conn, broadcast_conn *net.UDPConn, ch_send chan UDP_message) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error in UDP_sender, closing connections: %s\n", r)
			local_conn.Close()
			broadcast_conn.Close()
		}
	}()

	var err error
	var n int

	for {
		msg := <-ch_send

		if msg.Raddr == "broadcast" {
			n, err = local_conn.WriteToUDP([]byte(msg.Data), broadcast_adress)
		} else {
			raddr, err := net.ResolveUDPAddr("udp", msg.Raddr)

			if err != nil {
				fmt.Printf("Error in UDP_sender, could not resolve UDP adress.\n")
				panic(err)
			}

			n, err = local_conn.WriteToUDP([]byte(msg.Data), raddr)
		}

		if err != nil || n < 0 {
			fmt.Printf("Error in UDP_sender while writing.\n")
			panic(err)
		}
	}
}

func UDP_receiver(local_conn, broadcast_conn *net.UDPConn, msg_size int, ch_receive chan UDP_message) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error in UDP_receiver, closing connections: %s\n", r)
			local_conn.Close()
			broadcast_conn.Close()
		}
	}()

	ch_receive_broadcast_conn := make(chan UDP_message)
	ch_receive_local_conn := make(chan UDP_message)

	go UDP_conn_reader(local_conn, msg_size, ch_receive_local_conn)
	go UDP_conn_reader(broadcast_conn, msg_size, ch_receive_broadcast_conn)

	for {
		select {
			case buf := <-ch_receive_broadcast_conn:
				ch_receive <- buf
			case buf := <-ch_receive_local_conn:
				ch_receive <- buf
		}
	}
}

func UDP_conn_reader(conn *net.UDPConn, msg_size int, ch_receive chan UDP_message) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error in UDP_conn_reader, closing connection: %s \n", r)
			conn.Close()
		}
	}()

	for {
		buf := make([]byte, msg_size)
		n, raddr, err := conn.ReadFromUDP(buf)

		if err != nil || n < 0 {
			fmt.Printf("Error in UDP_conn_reader while reading\n")
			panic(err)
		}

		ch_receive <- UDP_message{Raddr: raddr.String(), Data: string(buf), Length: n}
	}
}
