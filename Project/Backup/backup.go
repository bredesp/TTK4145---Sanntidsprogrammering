package backup

import (
	"encoding/json"
	"os"
	"sync"
)

var FLOORS int
var ELEVATORS int
var Mtx sync.Mutex = sync.Mutex{}

type UpdateMessage struct {
	MessageType int
	// 0 = hall request
	// 1 = new behaviour
	// 2 = arrived at floor
	// 3 = new direction
	// 4 = cab request
	// 5 = delete elevator
	// 9 = motorBroken
	Elevator    	string
	Floor       	int
	Button      	int
	Behaviour   	string
	Direction   	string
	OrderCompleted 	bool
}

type StateValues struct {
	Behaviour   string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type StatusStruct struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]*StateValues `json:"states"` //key kan be changed to int if more practical but remember to cast to string before JSON encoding!
}


func Backup(ch_status_broadcast chan<- StatusStruct, ch_status_update <-chan UpdateMessage, ch_elevator_status chan<- StatusStruct, ch_status_refresh <-chan StatusStruct, init bool, id string) {

	file, err := os.OpenFile("backup.txt", os.O_RDWR|os.O_CREATE, 0671)
	ifError(err)

	systemInfo := new(StatusStruct) //main StatusStruct, continually updated
	Mtx.Lock()
	if init { //clean initialization
		file, err = os.Create("backup.txt")
		ifError(err)

		systemInfo.HallRequests = make([][2]bool, FLOORS)
		systemInfo.States = make(map[string]*StateValues)
		initElevator(id, systemInfo, "idle", 0, "stop", make([]bool, FLOORS))
		file.Seek(0, 0)
		res := json.NewEncoder(file).Encode(systemInfo)
		ifError(res)

	} else { // recover systemInfo from file
		res := json.NewDecoder(file).Decode(systemInfo)
		ifError(res)
	}
	Mtx.Unlock()
	for {
		select {
		case message := <-ch_status_update:
			Mtx.Lock()
			if message.Elevator != "" {
				_, elevStatus := systemInfo.States[message.Elevator]
				deleteElev := 5
				if !elevStatus && message.MessageType != deleteElev { //Elevator is not in systemInfo struct, initialized with best guess
					initElevator(message.Elevator, systemInfo, message.Behaviour, message.Floor, message.Direction, make([]bool, FLOORS))
				}
			}
      updateStates(systemInfo, message, id)
			Mtx.Unlock()
			file, err = os.Create("backup.txt")
			ifError(err)
			e := json.NewEncoder(file).Encode(systemInfo)
			ifError(e)

		case inputState := <-ch_status_refresh: //only add orders and update states
			//refresh hall requests
			Mtx.Lock()
			for floor := 0; floor < FLOORS; floor++ {
				for button := 0; button < 2; button++ {
					if inputState.HallRequests[floor][button] {
						systemInfo.HallRequests[floor][button] = true
					}
				}
			}
			for elev, estate := range systemInfo.States {
				_, fail := systemInfo.States[elev]
				if !fail {
					initElevator(elev, systemInfo, estate.Behaviour, estate.Floor, estate.Direction, estate.CabRequests)
				} else {
					systemInfo.States[elev].Behaviour = estate.Behaviour
					systemInfo.States[elev].Floor = estate.Floor
					systemInfo.States[elev].Direction = estate.Direction
					for floor := 0; floor < FLOORS; floor++ {
						if estate.CabRequests[floor] {
							systemInfo.States[elev].CabRequests[floor] = true
						}
					}
				}
			}
			Mtx.Unlock()

		case ch_elevator_status <- *systemInfo:
		case ch_status_broadcast <- *systemInfo:
		}
	}
}


// Used to update the states and MessageType
func updateStates(systemInfo *StatusStruct, message UpdateMessage, id string){
	switch message.MessageType{
		case 0: //cab request
			if message.OrderCompleted {
				systemInfo.States[message.Elevator].CabRequests[message.Floor] = false
			} else {
				systemInfo.States[message.Elevator].CabRequests[message.Floor] = true
			}
		case 1: //hall request
			if message.OrderCompleted {
			systemInfo.HallRequests[message.Floor][message.Button] = false
			}else{
			systemInfo.HallRequests[message.Floor][message.Button] = true
			}

		case 2: //new Behaviour
			systemInfo.States[message.Elevator].Behaviour = message.Behaviour

		case 3: //new direction
			systemInfo.States[message.Elevator].Direction = message.Direction

		case 4: //arrived at floor
			systemInfo.States[message.Elevator].Floor = message.Floor

		case 5: //lost Elevator
			if message.Elevator != id { //dont delete ourselves
				delete(systemInfo.States, message.Elevator)
			}
		}
}

//Used to initialize a new elevator in @param systemInfo with the gven paramters
func initElevator(elevName string, systemInfo *StatusStruct, Behaviour string, Floor int, Direction string, cabRequests []bool) {

	// Set elevator systemInfo
	systemInfo.States[elevName] = new(StateValues)
	if systemInfo.States[elevName].Behaviour == "" {
		systemInfo.States[elevName].Behaviour = "idle"
	} else {
		systemInfo.States[elevName].Behaviour = Behaviour
	}

	// Set elevator floor
	systemInfo.States[elevName].Floor = Floor
	if systemInfo.States[elevName].Direction == "" {
		systemInfo.States[elevName].Direction = "up"
	} else {
		systemInfo.States[elevName].Direction = Behaviour
	}

	// Set number of floors
	if len(cabRequests) != FLOORS {
		systemInfo.States[elevName].CabRequests = make([]bool, FLOORS)
	} else {
		systemInfo.States[elevName].CabRequests = cabRequests
	}
	return
}

func ifError(err error) {
	if err != nil {
		panic(err)
	}
}
