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
  (exec.Command("gnome-terminal", "-x", "go", "run", "/home/student/TTK4145---Sanntidsprogrammering/Exercises/Exercise6/phoenix.go")).Run()
}



func main() {
  var number uint64
  var buf = make([]byte, 16)

  // Master variable
  isMaster := false

  // Creat Server Address,
  ServerAddr, err := net.ResolveUDPAddr("udp", "10.100.23.233:10001")
	if err != nil {
		fmt.Println("Error: ", err)
	}

	LocalAddr, err := net.ResolveUDPAddr("udp", "10.100.23.233:0")
	if err != nil {
		// fmt.Println("Error: ", err)
	}

  Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
  if err != nil {
    // fmt.Println("Error: ", err)
  }

  fmt.Println("Backup")

  //Conn.SetDeadline(time.Second * 1)
  // BackUP loop (Creats BackUP/Master)
  for !(isMaster) {
    // Look for Master: listning for numbers
    binary.BigEndian.PutUint64(buf, number)
    _, errW := Conn.Write(buf)
    if errW != nil {
      //
    }

    Conn.SetReadDeadline(time.Now().Add(2 * time.Second))
    fmt.Println("Waiting for number")
    n, _, err := Conn.ReadFromUDP(buf)

    if (err != nil) {
      isMaster = true;
    } else {
      number = binary.BigEndian.Uint64(buf[:n])
      fmt.Println("\t Number: ", number, "\t")
    }
  }

  Conn.Close()

  spawnBackup()
  fmt.Println("Master")


  numberConn, _ := net.DialUDP("udp", nil, ServerAddr)

  // Master loop
  for {
    // Print Number

    fmt.Println("\t Number: ", number, "\t")
    // Send pulse signal with number
    binary.BigEndian.PutUint64(buf, number)
    _, err := numberConn.Write(buf)
		if err != nil {
			//fmt.Println(buf, err)
		}

    numberConn.SetReadDeadline(time.Now().Add(2*time.Second))
    n, _, err := numberConn.ReadFromUDP(buf)
    if err != nil {
      // KJØR NY BACKUP
      if backupLimit == 0 {
        spawnBackup()
        backupLimit++
      }
    } else {
      number = binary.BigEndian.Uint64(buf[:n])
      fmt.Println("\t Mottok tallet: ", number, "\t")
      // Fikk heartbeat fra backup. Kan trygt fortsette å telle
      number++
    }

    // Update number
    // Sleep
    time.Sleep(1 * time.Second)

  }

  // Check if backup exists: listning for heartbeat

}
