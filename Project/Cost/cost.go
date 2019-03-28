package cost

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"../Backup"
)

var FLOORS int
var ELEVATORS int

type AssignedOrderInformation struct {
	AssignedOrders map[string][][]bool
	HallRequests   [][2]bool
	States         map[string]*backup.StateValues
}

func Cost(ch_elevator_status <-chan backup.StatusStruct, ch_fsm_info chan<- AssignedOrderInformation) {
	for {
		select {
		case state := <-ch_elevator_status: //when systemInfo is updated
			backup.Mtx.Lock()
			arg, err := json.Marshal(state)
			backup.Mtx.Unlock()
			if err != nil {
				fmt.Println("error:", err)
			}
			result, err := exec.Command("sh", "-c", "./hall_request_assigner --input '"+string(arg)+"'").Output()
			if err != nil {
				fmt.Println("error:", err, "cost function")
				fmt.Println("recived:", string(arg))
				continue
			}
			orders := new(map[string][][]bool)
			json.Unmarshal(result, orders)

			backup.Mtx.Lock()
			output := AssignedOrderInformation{
				AssignedOrders: *orders,
				HallRequests:   state.HallRequests,
				States:         state.States,
			}
			backup.Mtx.Unlock()
			ch_fsm_info <- output
		}
	}

}
