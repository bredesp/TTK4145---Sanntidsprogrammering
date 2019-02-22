package main

import (
  "encoding/binary"
  "fmt"
  "net"
  "time"
)

func main(){
  var number uint64
  var buf = make([]byte, 16)

  // ----------------------- SETT OPP UDP-KOBLING -----------------------
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


  for {
    binary.BigEndian.PutUint64(buf, number)
    _, errW := Conn.Write(buf)

    if errW != nil {
      fmt.Println("Kunne ikke skrive")
      //fmt.Println(errW)
    } else {
      fmt.Println("\t Number: ", number, "\t")
      number++
    }

    time.Sleep(1 * time.Second)
    }
}
