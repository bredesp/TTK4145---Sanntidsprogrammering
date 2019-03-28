package main

import (
	"./Cost"
	"./Elevio"
	"./Backup"
	"./StateMachine"
	"./Network"
	"./Network/network/acknowledge"
	"./Network/network/localip"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"time"
)

// Global variable for number of elevators and floors
const FLOORS = 4
const ELEVATORS = 1

// Id and port number for elevator
var id string
var port string

func main() {
	time.Sleep(400 * time.Millisecond)

	//  The id/port/intit can set by `go run main.go -id="id" -port="port" -init="true/false"` in the command line
	var init bool
	flag.BoolVar(&init, "init", true, "false if elev is recovering")
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.StringVar(&port, "port", "15657", "set port to connect to elevator")
	flag.Parse()

	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	// Handling unexpected panic errors. Spawns a new process and initializing from a saved state, makeeing sure the elevator server is running
	defer func() {
		if r := recover(); r != nil {
			if r == "dial tcp 127.0.0.1:"+port+": getsockopt: connection refused" {
				err := exec.Command("gnome-terminal", "-x", "sh", "-c", "ElevatorServer").Run()

				if err != nil {
					fmt.Println("Unable to reboot process, crashing...")
				}
			}
			fmt.Println(r, " MAIN fatal panic, unable to recover. Rebooting...", "./main -init=false -port="+port+" -id="+id)
			err := exec.Command("gnome-terminal", "-x", "sh", "-c", "./main -init=false -port="+port+" -id="+id).Run()
			if err != nil {
				fmt.Println("Unable to reboot process, crashing...")
			}
		}
		os.Exit(0)
	}()

	AssignGlobals(id, port)

	ch_status_update := make(chan backup.UpdateMessage)
	ch_network_update := make(chan backup.UpdateMessage)
	ch_elevator_update := make(chan backup.StatusStruct)
	ch_fsm_info := make(chan cost.AssignedOrderInformation)
	ch_status_broadcast := make(chan backup.StatusStruct)
	ch_status_refresh:= make(chan backup.StatusStruct)

	elevio.Init("localhost:"+port, FLOORS)

	go atExit()
	go network.Network(ch_status_broadcast, ch_network_update, ch_status_update, ch_status_refresh, id)
	go backup.Backup(ch_status_broadcast, ch_status_update, ch_elevator_update, ch_status_refresh, init, id)
	go stateMachine.StateMachine(ch_network_update, ch_fsm_info, init)
	go cost.Cost(ch_elevator_update, ch_fsm_info)

	select {}
}


// Restarts program when it is killed from terminal
func atExit() {
	sigchan := make(chan os.Signal, 10)
	signal.Notify(sigchan, os.Interrupt)
	<-sigchan
	elevio.SetMotorDirection(elevio.MD_Stop)
	log.Println("Rebooting", "sh", "-c", "./main -init=false -port="+port+" -id="+id)
	err := exec.Command("gnome-terminal", "-x", "sh", "-c", "./main -init=false -port="+port+" -id="+id).Run()
	if err != nil {
		fmt.Println("Unable to reboot process, crashing...")
	}
	log.Println("Program killed !")
	os.Exit(0)
}

// Assigning the global variables
func AssignGlobals(id string, port string) {
	backup.FLOORS = FLOORS
	backup.ELEVATORS = ELEVATORS
	stateMachine.FLOORS = FLOORS
	stateMachine.ID = id
	stateMachine.PORT = port
	acknowledge.ID = id
	acknowledge.PORT = port
}
