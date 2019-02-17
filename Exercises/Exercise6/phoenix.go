package main

import (
  "encoding/binary"
  "fmt"
  "net"
  "os/exec"
  "time"
)

func spawnBackup(){
  // gnome-terminal -x ["commands"]
  exec.Command("gnome-terminal", "-x", "go", "run", "/home/student/TTK4145---Sanntidsprogrammering/Exercises/Exercise6/phoenix.go")
}



func main() {
  var number uint64
  // Master variable
  isMaster := false
  // Creat Server Address,
  ServerAddr, err := net.ResolveUDPAddr("udp", "10.100.23.233:10001")
	if err != nil {
		fmt.Println("Error: ", err)
	}

	LocalAddr, err := net.ResolveUDPAddr("udp", "10.100.23.233:0")
	if err != nil {
		fmt.Println("Error: ", err)
	}

	Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	if err != nil {
		fmt.Println("Error: ", err)
	}

  defer Conn.Close()


  buf := make([]byte, 4)
  //Conn.SetDeadline(time.Second * 1)
  // BackUP loop (Creats BackUP/Master)
  for !(isMaster) {
    // Look for Master: listning for numbers
    fmt.Println("Waiting for number")
    n, _, err := Conn.ReadFromUDP(buf)

    if (err != nil) {
      isMaster = true;
    } else {
      number = binary.BigEndian.Uint64(buf[:n])
    }
  }

  // spawnBackup
  spawnBackup()
  number = 1


  // Master loop
  for {
    // Print Number
    fmt.Println("\t Number: ", number, "\t")
    // Send pulse signal with number
    binary.BigEndian.PutUint64(buf, number)
    _, err := Conn.Write(buf)
		if err != nil {
			fmt.Println(buf, err)
		}
    // Update number
    number++
    // Sleep
    time.Sleep(time.Second * 1)

  }

  // Check if backup exists: listning for heartbeat

}
