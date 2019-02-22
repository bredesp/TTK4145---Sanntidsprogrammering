package main

import (
  "encoding/binary"
  "fmt"
  "net"
  "time"
)


func main() {
  var number uint64
  var buf = make([]byte, 16)

  ServerAddr, err := net.ResolveUDPAddr("udp", "10.100.23.233:10001")
  if err != nil {
    fmt.Println("Error: ", err)
  }
  numberConn, _ := net.DialUDP("udp", nil, ServerAddr)

  // Master loop
  for {

    n, _, err := numberConn.ReadFromUDP(buf)
    if err != nil {
      // KJØR NY BACKUP
      fmt.Println("Kunne ikke lese")
    } else {
      number = binary.BigEndian.Uint64(buf[:n])
      fmt.Println("\t Mottok tallet: ", number, "\t")
      // Fikk heartbeat fra backup. Kan trygt fortsette å telle
    }

    // Update number
    // Sleep
    time.Sleep(1 * time.Second)

  }
}
