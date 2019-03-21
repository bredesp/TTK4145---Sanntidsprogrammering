package FSM

import (
	"fmt"
	"time"

	"github.com/bredesp/TTK4145---Sanntidsprogrammering/blob/master/Project/driver-go-master/elevio/elevator_io.go"
)

type STATE int

const (
	IDLE
	DOOROPEN
	MOVING
)

// elevator directon
var elevDir int

// function returns the floor the elevator is at
func getElevFloor() int{
    return getFloor()
}

// move the elevator up
func moveUp() {
    setMotorDircetion(MD_up)
    elevDir = 1
}

// move the elevator down
func moveDown() {
    setMotorDirection(MD_down)
    elevDir =  -1
}

// stop the elevator
func stopElev() {
    setMotorDirection
}

//

func FSM() {

	state = DOOROPEN

	for {
		switch state {

		case IDLE: {
			elevDir = 0
			elevFloor = getElevFloor()
			for floor := 0; floor < N_FLOORS; floor ++ {
				if internalOrders(floor) || externalOrders(floor){
					if floor < elevFloor {
						moveDown()
						state = MOVING
					}
					else if floor > elevFloor {
						moveUp()
						sate = MOVING
					}
					else if floor == elevFloor {
						orderCompleteChan <- elevFloor (???)
						state = DOOROPEN
					}
				}
			}

		}
		case MOVING: {
			elevFloor = getElevFloor()
			ElevSetFloorIndicator(elevFloor)
			if floor != -1 && floor < N_FLOOR {

			}
			if internalOrders(elevFloor) || externalOrders(elevFloor){
				stop()
				orderCompleteChan <- elevFloor
				state = DOOROPEN
			}
			else if elevDir = -1 {
				ordersBelow := 0
				for floor := elevFloor; floor > -1; floor -- {
					if internalOrders(floor) || externalOrders(floor){
						ordersBelow += 1
					}
					if ordersBelow = 0 {
						stopElev()
						state = IDLE
					}
				}
			}
			else if elevDir = 1 {
				ordersAbove := 0
				for floor := elevFloor; floor < N_FLOORS; floor ++ {
					if internalOrders(floor) || externalOrders(floor){
						ordersAbove += 1
					}
					if ordersBelow = 0 {
						stopElev()
						state = IDLE
					}
				}

			}
		}
		case DOOROPEN: {
			elevFloor = getElevFloor()
			ElevSetDoorOpenLamp(true)
			// Timer
			doorTimer := time.NewTimer(3 * time.Second)
			The <-timer1.C
			fmt.Println("doorTimer expired")
			if elevDir = -1	{
				state = MOVING
			}
			if elevDir = -1 {
				for floor := 0; floor < N_FLOORS; floor -- {
					if internalOrders(floor) || externalOrders(floor){
						if floor < elevFloor {
							moveDown()
							state = MOVING
						}
						if floor > elevFlor {
							moveUp()
							state = MOVING
						}
					}
				}
				state = IDLE
			}
			if elevDir = 1 {
				for floor := N_FLOORS - 1; floor > -1; floor -- {
					if internalOrders(floor) || externalOrders(floor){
						if floor < elevFloor {
							moveDown()
							state = MOVING
						}
						if floor > elevFlor {
							moveUp()
							state = MOVING
						}
					}
				}
				state = IDLE
			}
			else if elevDir == 0{
				state = IDLE
			}
		}
	}
}
