package main

import (
  "elevio"
  "handler"
)


func main(){
  elevio.Init("localhost:15657", N_FLOORS)
  go handler.readButtons()
}
