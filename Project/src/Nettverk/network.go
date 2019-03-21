package Nettverk

import (
	. "fmt"
	. "strconv"
	. "strings"
	"net"
	"os"
)

var ch_send_stopped = make(chan int)
var ch_send_going = make(chan int)
var ch_send_button = make(chan [2]int)

var ch_receive_stopped = make(chan int)
var ch_receive_going = make(chan int)
var ch_receive_button = make(chan [2]int)

var ch_send = make(chan Udp_message, 40)
var ch_receive = make(chan Udp_message, 40)

var udp_port int = 20008
var my_IP []byte = Find_My_IP()


func Network() {
	err := UDP_init(udp_port, udp_port, 1024, ch_send, ch_receive)
	if err != nil {
		Print("UDP initialized: %s\n", err)
	}

	go Order_Receiver()
	go Order_Sender()

	for {
		time.Sleep(1000 * time.Millisecond)
	}
}

func Order_Receiver() {
	for {
		received_msg := <-ch_receive
		msg_type, button, floor, looped_msg := UDP_message_parser(received_msg)

		switch msg_type {
			case 0:
				Println("Received button messsage")
				if !looped_msg {
					ch_receive_button <- button
				}
			case 1:
				Println("Received stoppet message")
				if !looped_msg {
					ch_receive_stopped <- floor
				}
			case 2:
				Println("Received going message")
				if !looped_msg {
					ch_receive_going <- floor
				}
		}
	}
}

func Order_Sender() {
	for {
		select {
		case button := <-ch_send_button:
			ch_send <- UDP_message_generator(0, button[0], button[1])
		case floor := <-ch_send_stopped:
			ch_send <- UDP_message_generator(1, floor, 0)
		case floor := <-ch_send_going:
			ch_send <- UDP_message_generator(2, floor, 0)
		}
	}
}

func UDP_message_generator(msg_type, floor, button int) Udp_message {
	s := " _ _ "
	if msg_type > -1 && msg_type < 3 && floor > -1 && floor < 4 && button > -1 && button < 3 {
		s = Itoa(msg_type) + "_" + Itoa(floor) + "_" + Itoa(button) + "_"
	} else {
		Println("Bad arguments in call to network.UDP_message_generator(msg_type, floor, button), tried to send: ", s)
	}
	return Udp_message{Raddr: "broadcast", Data: s, Length: len(s)}
}

func UDP_message_parser(received_msg Udp_message) (int, [2]int, int, bool) {
	addr := received_msg.Raddr
	data := received_msg.Data
	looped_msg := false
	myID, _ := Atoi(Split((Split(addr, ".")[3]), ":")[0])
	if myID == int(my_IP[3]) {
		looped_msg = true
	}
	dataArr := Data_to_array(data)
	var button [2]int
	msg_type := dataArr[0]
	for i := 0; i < 2; i++ {
		button[i] = dataArr[i+1]
	}
	floor := button[0]
	return msg_type, button, floor, looped_msg
}

func Data_to_array(received_data string) [3]int {
	splitted_data := Split(received_data, "_")
	var message_array [3]int
	for i := 0; i < 3; i++ {
		message_array[i], _ = Atoi(splitted_data[i])
	}
	return message_array
}

func UDP_message_printer(msg Udp_message) {
	Printf("Message:  \n \t Raddr = %s \n \t Data = %s \n \t Length = %v \n", msg.Raddr, msg.Data, msg.Length)
}

func Find_My_IP() []byte {
	result := make([]byte, 4)
	interface_adresses, err := net.InterfaceAddrs()
	if err != nil {
		Println("Error trying to find my own IP: ", err)
		emptyResult := make([]byte, 4)
		return emptyResult
	}

	for _, address := range interface_adresses {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.To4()
			}
		}
	}
	return result
}
