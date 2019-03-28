package stateMachine

import (
	"../Cost"
	"../Elevio"
	"../Backup"
	"fmt"
	"os"
	"os/exec"
	"time"
)

var FLOORS int
var ID string
var PORT string

func StateMachine(ch_network_update chan<- backup.UpdateMessage, ch_fsm_info <-chan cost.AssignedOrderInformation, init bool) {

	// Handling unexpected panic errors. Spawns a new process and initializing from a saved state.
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r, " -FSM fatal panic, unable to recover. Rebooting...", "./main -init=false -port="+PORT+" -id="+ID)
			err := exec.Command("gnome-terminal", "-x", "sh", "-c", "./main -init=false -port="+PORT+" -id="+ID).Run()
			if err != nil {
				fmt.Println("Unable to reboot process, crashing...")
			}
		}
		os.Exit(0)
	}()

	//----- Initializing subroutines and variables ----//
	var updateMessage backup.UpdateMessage
	var elevState cost.AssignedOrderInformation

	updateMessage.Elevator = ID
	elevState = <-ch_fsm_info // Contains all the StateMachine needs to, being constatlly refreshed
	ch_button_press := make(chan elevio.ButtonEvent)
	ch_new_floor := make(chan int)
	doorTimedOut := time.NewTimer(3 * time.Second)
	doorTimedOut.Stop()
	motorTimedOut := time.NewTimer(4 * time.Second)
	motorTimedOut.Stop()
	go elevio.PollButtons(ch_button_press)
	go elevio.PollFloorSensor(ch_new_floor)

	//---- Initializing the elevator----//
	if init { // Initializing a clean elevator
		elevio.SetMotorDirection(elevio.MD_Down)
		elevio.SetFloorIndicator(0)
		elevio.SetDoorOpenLamp(false)
		turnOffAllLights()
		updateMessage.Floor = 0
		updateMessage.Behaviour = "idle"
		updateMessage.Direction = "up"
	L:
		for {
			select {
			case floor := <-ch_new_floor:
				if floor == 0 {
					elevio.SetMotorDirection(elevio.MD_Stop)
					break L
				}
			}
		}
	} else { // Recovers the elevator from backup
		updateMessage.Floor = elevState.States[ID].Floor
		updateMessage.Behaviour = elevState.States[ID].Behaviour
		updateMessage.Direction = elevState.States[ID].Direction
		if elevState.States[ID].Behaviour == "doorOpen" {
			doorTimedOut.Reset(3 * time.Second)
		}
		if elevState.States[ID].Behaviour == "moving" {
			if elevState.States[ID].Direction == "up" {
				elevio.SetMotorDirection(elevio.MD_Up)
			} else {
				elevio.SetMotorDirection(elevio.MD_Down)
			}
		}
	}

	// Main loop in StateMachine
	for {
		select {
		case elevState = <-ch_fsm_info: // From cost module
			setAllLights(elevState, ID)
			switch elevState.States[ID].Behaviour {
			case "doorOpen":
				if shouldElevStop(elevState, ID, elevState.States[ID].Floor) {
					if clearAtFloor(elevState, ID, elevState.States[ID].Floor, ch_network_update) {
						doorTimedOut.Reset(3 * time.Second)
					}
					setAllLights(elevState, ID)
				}
			case "idle":
				newDirection := chooseDirection(elevState, ID, elevState.States[ID].Floor)
				switch newDirection {
				case elevio.MD_Stop:
					if elevState.AssignedOrders[ID][elevState.States[ID].Floor][0] ||
						elevState.AssignedOrders[ID][elevState.States[ID].Floor][1] ||
						elevState.States[ID].CabRequests[elevState.States[ID].Floor] { // Check if buttons at current floor is pressed
						elevio.SetDoorOpenLamp(true)
						doorTimedOut.Reset(3 * time.Second)

						// Behaviour message
						updateMessage.MessageType = 2
						updateMessage.Behaviour = "doorOpen"
						updateMessage.Elevator = ID
						ch_network_update <- updateMessage

						clearAtFloor(elevState, ID, elevState.States[ID].Floor, ch_network_update)
						setAllLights(elevState, ID)
					}
				case elevio.MD_Up:
					elevio.SetMotorDirection(newDirection)
					motorTimedOut.Reset(4 * time.Second)

					// Direction Message
					updateMessage.MessageType = 3
					updateMessage.Direction = "up"
					updateMessage.Elevator = ID
					ch_network_update <- updateMessage

					// Behaviour message
					updateMessage.MessageType = 2
					updateMessage.Behaviour = "moving"
					updateMessage.Elevator = ID
					ch_network_update <- updateMessage

				case elevio.MD_Down:
					elevio.SetMotorDirection(newDirection)
					motorTimedOut.Reset(4 * time.Second)

					// Direction Message
					updateMessage.MessageType = 3
					updateMessage.Direction = "down"
					updateMessage.Elevator = ID
					ch_network_update <- updateMessage

					// Behaviour message
					updateMessage.MessageType = 2
					updateMessage.Behaviour = "moving"
					updateMessage.Elevator = ID
					ch_network_update <- updateMessage

				}
			case "moving":
			}
		case buttonEvent := <-ch_button_press: // When a button press occurs it forwarded to the network module
			/*------------ Make update message ------------*/
			if buttonEvent.Button < 2 { // If hall request
				updateMessage.MessageType = 1
				updateMessage.Button = int(buttonEvent.Button)
				updateMessage.OrderCompleted = false // New buttton press
				updateMessage.Elevator = ID
				updateMessage.Behaviour = elevState.States[ID].Behaviour
				updateMessage.Floor = buttonEvent.Floor

			} else {
				updateMessage.MessageType = 0 // Cab request
				updateMessage.Button = int(buttonEvent.Button)
				updateMessage.OrderCompleted = false // New buttton press
				updateMessage.Elevator = ID
				updateMessage.Behaviour = elevState.States[ID].Behaviour
				updateMessage.Floor = buttonEvent.Floor

			}
			ch_network_update <- updateMessage
		case floor := <-ch_new_floor: // Checks if the elevator should stop and clear orders the new floor
			/*-------------- Make update message --------------------*/
			updateMessage.MessageType = 4
			updateMessage.Floor = floor
			updateMessage.Behaviour = elevState.States[ID].Behaviour
			updateMessage.Elevator = ID
			ch_network_update <- updateMessage

			motorTimedOut.Stop()
			elevio.SetFloorIndicator(floor)
			if shouldElevStop(elevState, ID, floor) {
				elevio.SetMotorDirection(elevio.MD_Stop)
				clearAtFloor(elevState, ID, floor, ch_network_update)
				elevio.SetDoorOpenLamp(true)
				doorTimedOut.Reset(3 * time.Second)

				// Behaviour message
				updateMessage.MessageType = 2
				updateMessage.Behaviour = "doorOpen"
				updateMessage.Elevator = ID
				ch_network_update <- updateMessage

				setAllLights(elevState, ID)
			} else {
				motorTimedOut.Reset(4 * time.Second)
			}
		case <-doorTimedOut.C: // The door closes and a new direction set
			elevio.SetDoorOpenLamp(false)
			newDirection := chooseDirection(elevState, ID, elevState.States[ID].Floor)
			switch newDirection {
			case elevio.MD_Stop:

				// Behaviour message
				updateMessage.MessageType = 2
				updateMessage.Behaviour = "idle"
				updateMessage.Elevator = ID
				ch_network_update <- updateMessage

			case elevio.MD_Up:
				elevio.SetMotorDirection(newDirection)
				motorTimedOut.Reset(4 * time.Second)

				// Direction Message
				updateMessage.MessageType = 3
				updateMessage.Direction = "up"
				updateMessage.Elevator = ID
				ch_network_update <- updateMessage

				// Behaviour message
				updateMessage.MessageType = 2
				updateMessage.Behaviour = "moving"
				updateMessage.Elevator = ID
				ch_network_update <- updateMessage

			case elevio.MD_Down:
				elevio.SetMotorDirection(newDirection)
				motorTimedOut.Reset(4 * time.Second)

				// Direction Message
				updateMessage.MessageType = 3
				updateMessage.Direction = "down"
				updateMessage.Elevator = ID
				ch_network_update <- updateMessage

				// Behaviour message
				updateMessage.MessageType = 2
				updateMessage.Behaviour = "moving"
				updateMessage.Elevator = ID
				ch_network_update <- updateMessage

			}
		case <-motorTimedOut.C: // No elevator is deteced at floor sensor within 4 seconds
			updateMessage.MessageType = 9
			updateMessage.Elevator = ID
			fmt.Println("motor broke")
			ch_network_update <- updateMessage
			motorTimedOut.Reset(4 * time.Second)
			motorTimedOut.Stop()
			turnOffAllLights()
		F:
			for {
				select {
				case floor := <-ch_new_floor:
					if floor != -1 {
						break F
					}
				}
			}
		}
	}
}

// Chooses the direction of the elevator
func chooseDirection(elevState cost.AssignedOrderInformation, ID string, floor int) elevio.MotorDirection {
	switch elevState.States[ID].Direction {
	case "stop":
		fallthrough
	case "down":
		if ordersBelow(elevState, ID, floor) {
			return elevio.MD_Down
		} else if ordersAbove(elevState, ID, floor) {
			return elevio.MD_Up
		} else {
			return elevio.MD_Stop
		}
	case "up":
		if ordersAbove(elevState, ID, floor) {
			return elevio.MD_Up
		} else if ordersBelow(elevState, ID, floor) {
			return elevio.MD_Down
		} else {
			return elevio.MD_Stop
		}
	default:
		return elevio.MD_Stop
	}
	return elevio.MD_Stop
}

// Checks for any orders below the current floor
func ordersBelow(elevState cost.AssignedOrderInformation, ID string, reachedFloor int) bool {
	for floor := 0; floor < reachedFloor; floor++ {
		if elevState.States[ID].CabRequests[floor] {
			return true
		}
		for button := 0; button < 2; button++ {
			if elevState.AssignedOrders[ID][floor][button] {
				return true
			}
		}
	}
	return false
}

// Checks for any orders above the current floor
func ordersAbove(elevState cost.AssignedOrderInformation, ID string, reachedFloor int) bool {
	for floor := reachedFloor + 1; floor < FLOORS; floor++ {
		if elevState.States[ID].CabRequests[floor] {
			return true
		}
		for button := 0; button < 2; button++ {
			if elevState.AssignedOrders[ID][floor][button] {
				return true
			}
		}
	}
	return false
}

// Check if the elevator should stop at the floor and return true
func shouldElevStop(elevState cost.AssignedOrderInformation, ID string, floor int) bool {
	switch elevState.States[ID].Direction {
	case "down":
		return (elevState.AssignedOrders[ID][floor][elevio.BT_HallDown] ||
			elevState.States[ID].CabRequests[floor] ||
			!ordersBelow(elevState, ID, floor))
	case "up":
		return (elevState.AssignedOrders[ID][floor][elevio.BT_HallUp] ||
			elevState.States[ID].CabRequests[floor] ||
			!ordersAbove(elevState, ID, floor))
	case "stop":
		fallthrough
	default:
		return true
	}
}

// Check if the elevator is travelling in the right direction to clear the floor and then return true
func clearAtFloor(elevState cost.AssignedOrderInformation, ID string, floor int, ch_network_update chan<- backup.UpdateMessage) bool {
	// For cab requests
	cleared := false
	update := backup.UpdateMessage{
		MessageType:    	0,
		Floor:       		floor,
		Button:      		2,
		Behaviour:   		elevState.States[ID].Behaviour,
		Direction:   		elevState.States[ID].Direction,
		OrderCompleted: 	true,
		Elevator:    		ID,
	}
	if elevState.States[ID].CabRequests[elevState.States[ID].Floor] {
		ch_network_update <- update
		cleared = true
	}

	// For hall requests
	update.MessageType = 1
	switch elevState.States[ID].Direction {
	case "up":
		if elevState.HallRequests[elevState.States[ID].Floor][int(elevio.BT_HallUp)] {
			update.Button = int(elevio.BT_HallUp)
			ch_network_update <- update
			cleared = true
		}
		if !ordersAbove(elevState, ID, floor) &&
			elevState.HallRequests[elevState.States[ID].Floor][int(elevio.BT_HallDown)] {
			update.Button = int(elevio.BT_HallDown)
			ch_network_update <- update
			cleared = true
		}
	case "down":
		if elevState.HallRequests[elevState.States[ID].Floor][int(elevio.BT_HallDown)] {
			update.Button = int(elevio.BT_HallDown)
			ch_network_update <- update
			cleared = true
		}
		if !ordersBelow(elevState, ID, floor) &&
			elevState.HallRequests[elevState.States[ID].Floor][int(elevio.BT_HallUp)] {
			update.Button = int(elevio.BT_HallUp)
			ch_network_update <- update
			cleared = true
		}
	case "stop":
		update.Button = int(elevio.BT_HallDown)
		ch_network_update <- update
		update.Button = int(elevio.BT_HallUp)
		ch_network_update <- update
		cleared = true
	}
	return cleared

}

// Lights are updated based on orders from Cost module
func setAllLights(elevState cost.AssignedOrderInformation, ID string) {
	for floor := 0; floor < FLOORS; floor++ {
		elevio.SetButtonLamp(elevio.BT_Cab, floor, elevState.States[ID].CabRequests[floor])
		for button := elevio.BT_HallUp; button < elevio.BT_Cab; button++ {
			elevio.SetButtonLamp(button, floor, elevState.HallRequests[floor][button])
		}
	}
}

func turnOffAllLights() {
	for floor := 0; floor < FLOORS; floor++ {
		elevio.SetButtonLamp(elevio.BT_Cab, floor, false)
		for button := elevio.BT_HallUp; button < elevio.BT_Cab; button++ {
			elevio.SetButtonLamp(button, floor, false)
		}
	}
}
