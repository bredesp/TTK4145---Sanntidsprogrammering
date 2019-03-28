# Code we've loaned
- elevator_io.go
- bcast.go
- bcast_conn.go
- localip.go
- peers.go
- hall_request_assigner

# Module overview
## Backup
Module for making a “backup.txt” file for the elevators using JSON format. Logging the struct “StatusStruck”, shown below, in a txt-file. This module also sends a the state to the cost function.

```go
type StateValues struct {
	Behaviour     string  `json:"behaviour"`
	Floor         int     `json:"floor"`
	Direction     string  `json:"direction"`
	CabRequests   []bool  `json:"cabRequests"`
}
```

```go
type StatusStruct struct {
	HallRequests  [][2]bool               `json:"hallRequests"`
	States        map[string]*StateValues `json:"states"`
}
```

## StateMachine
The StateMachine handles the main logic for the elevator software behavior and connection between the different modules. StateMachine uses the reassigned information from the cost function to place the elevators in the different states, it receives button presses and sensor signals, and sends the state to the network module every time the state changes.

## Network
### network.go
Receives new states from the state machine when the state has changed. For every state change, it sends the new state to the net through the acknowledge-module, so we can be sure every peer has received the new state. We then forward the new state to the backup-module who then feeds this into the cost-function which then informs the state machine (...).

Receives a peer update everytime there is a change in peers. Forward peerlist to acknowledge so it knows who it has to receive acknowledges from. If we have lost peers, send this info to the backup-module who then feeds this into the cost-function which then informs the state machine. If we have new peers, request the newest complete state information, and then send this to the net.

### acknowledge.go
If 15 ms has gone by, we have not received all the neccesary acks and we tried to send the message less than 10 times, we will try to resend the same message, and update the number of times the message is sent. Restart the acknowledge timer for another 15 ms. If sent more than 10 times, or if we have received all neccesary acks, don't try to resend. We will find out by other means if one or more peers is lost. This is only to "assure" against packet loss.

When we receive ack from peers, update info on who we have received ack from.

When received new state or update, send acknowledge to peers and forward said state or update to the network module and then to the backup module.

Send update- and status-messages to the net. If there are other peers on the net, start timer for acknowledgements.

## Cost
The cost module receives updated elevator states and then transforms it to the proper format for it to be used by the given program "hall_request_assigner". "hall_request_assigner" calculates the “cost” for each elevator and reassigns the orderbook for each elevator. We then catch the resulting program output, converts it back to the format we use in our program and forwards it to the state machine.

The code and explanation for the hall_request_assigner is given here: https://github.com/TTK4145/Project-resources/tree/master/cost_fns/hall_request_assigner.

## Main
Starting point for the program. Starts the necessary threads and sets up the channels. 


# Channel overview
![alt text](https://github.com/simonkvammen/Test/blob/master/0001.jpg "Module communication")

