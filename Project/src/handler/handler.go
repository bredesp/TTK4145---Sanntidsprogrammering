package handler

import (
	. "elevator_io"
	. "network"
	. "time"
	. "fmt"
)

var internalOrders [4]bool //Bolsk vektor str 4
var externalOrders [4]bool // bolsk vektor str 4
var otherOrders [4]int // om noen skal til floor = otherOrders[2]
var externalButtons [4][2]Time // time - floor x up/down
var internalButtons [4]Time // setter "start-tid" på registrert trykk
var prevInternalButton [4]bool //den forrige interne utførte
var prevExternalButton [4][2]bool //den forrige utførte ex.ordren

var MaxWaitTime float64 = 20 //maks ventetid 20sek

var internalButtonCh = make(chan int) //kanal for int.butts - sender etasje nr
var externalButtonCh = make(chan int)

var DoneCh = make(chan int) // im done chanel - with floor

// --- //



func int2Button(num int) ButtonType {
	if num == 0 {
		return BT_HallUp // 0 = opp
	} else if num == 1 {
		return BT_HallDown  // 1 = ned
	} else if num == 2{
		return BT_Cab  //cabcall
	}else{
		fmt.Println("ButtonType not valid")
	}
}


// funksjon som skjekker om det er en intern ordre på etasjen floor
func InternalOrders(floor int) bool {
	return internalOrders[floor]
}

// funksjon som skjekker om det er en ekstern ordre på etasjen floor
func ExternalOrders(floor int) bool {
	return externalOrders[floor]
}

func ReadButtons() {
	for {
		for floor := 0; floor < N_FLOORS; floor++ {
			internalButton := getButton(BT_Cab, floor)
			if internalButton && !prevInternalButton[floor] {
				internalButtonCh <- floor
				internalButtons[floor] = Now()
				SetButtonLamp(BT_Cab, floor, true)
			}
			prevInternalButton[floor] = internalButton
			for upDown := 0; upDown < 2; upDown++ {
				if floor == N_FLOORS-1 && upDown == 0 {
					continue
				}
				if floor == 0 && upDown == 1 { /
					continue
				}
				externalButton := getButton(int2Button(upDown), floor)
				if externalButton && !prevExternalButton[floor][upDown] {
					externalButtonCh <- floor
					SendButtonCh <- [2]int{floor, upDown}
					SetButtonLamp(int2Button(upDown), floor, true)
				}
				prevExternalButton[floor][upDown] = externalButton
			}
		}
	}
}


func watchDog() {
	for {
		timer := NewTimer(5 * Second)
		<-timer.C
		for floor := 0; floor < N_FLOORS; floor++ {
			if !internalButtons[floor].IsZero() && Now().Sub(internalButtons[floor]).Seconds() > MaxWaitTime {
				internalButtonCh <- floor
			}
			for upDown := 0; upDown < 2; upDown++ {
				if otherOrders[floor] > 0 && Now().Sub(externalButtons[floor][upDown]).Seconds() > MaxWaitTime {
					externalButtonCh <- floor
					SendButtonCh <- [2]int{floor, upDown}
				} else if !externalButtons[floor][upDown].IsZero() && Now().Sub(externalButtons[floor][upDown]).Seconds() > MaxWaitTime {
					externalButtonCh <- floor
					SendButtonCh <- [2]int{floor, upDown}
				}
			}
		}
	}
}


func AddOrders() {
	go watchDog()

	for {
		select {
		case floor := <-internalButtonCh:
			if !internalOrders[floor] {
				internalOrders[floor] = true
				SendGoingCh <- floor
			}
		case floor := <-externalButtonCh:
			if internalOrders[floor] {
				externalOrders[floor] = true
				SendGoingCh <- floor
			} else if externalOrders[floor] {
				SendGoingCh <- floor
			} else if otherOrders[floor] > 0 {
				continue
			} else {
				externalOrders[floor] = true
				SendGoingCh <- floor
			}
		case button := <-RecieveButtonCh:
			floor := button[0]
			buttonType := button[1]
			if !(floor == N_FLOORS-1 && buttonType == 0) && !(floor == 0 && buttonType == 1) {
				SetButtonLamp(int2Button(buttonType), floor, true)
			}
			if internalOrders[floor] {
				externalOrders[floor] = true
				SendGoingCh <- floor
			} else if externalOrders[floor] {
				SendGoingCh <- floor
			} else if otherOrders[floor] > 0 {
				continue
			} else {
				continue
			}
		case floor := <-RecieveGoingCh:
			otherOrders[floor] += 1
		}
	}
}


func RemoveOrders() {

	for {
		select {
		case floor := <- DoneCh:
			SendStoppedCh <- floor
			internalOrders[floor] = false
			internalButtons[floor] = Time{}
			externalOrders[floor] = false
			for upDown:= 0; upDown < 2; upDown++ {
				externalButtons[floor][upDown] = Time{}
			}
			for upDown := 0; upDown < 3; upDown++ {
				if floor == N_FLOORS-1 && upDown == 0 {
					continue
				}
				if floor == 0 && upDown == 1 {
					continue
				}
				SetButtonLamp(int2Button(upDown), floor, false)
			}

		case floor := <- RecieveStoppedCh:
			externalOrders[floor] = false
			for upDown := 0; upDown < 2; upDown++ {
				externalButtons[floor][upDown] = Time{}
				if floor == N_FLOORS-1 && upDown == 0 {
					continue
				}
				if floor == 0 && upDown == 1 {
					continue
				}
				SetButtonLamp(int2Button(upDown), floor, false)
			}
			otherOrders[floor] -= 1
		}
	}
}
