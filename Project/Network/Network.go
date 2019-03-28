package network

import (
	"fmt"
	"../Backup"
	"./network/acknowledge"
	"./network/peers"
)

func Network( ch_status_broadcast <-chan backup.StatusStruct, ch_network_update <-chan backup.UpdateMessage, ch_status_update chan<- backup.UpdateMessage, ch_status_refresh chan<- backup.StatusStruct, id string) {
	var peerlist peers.PeerUpdate
	elevatorAvaliable := true

	ch_ack_peer_update := make(chan peers.PeerUpdate, 2)
	ch_new_update := make(chan backup.UpdateMessage)
	ch_new_status := make(chan backup.StatusStruct)
	ch_peer_update := make(chan peers.PeerUpdate)
	ch_peer_TX_enable := make(chan bool)

	go acknowledge.Ack(ch_new_update, ch_new_status, ch_ack_peer_update)
	go peers.Transmitter(16016, id, ch_peer_TX_enable)
	go peers.Receiver(16016, ch_peer_update)

	for {
		select {
			// Acknowledge has received a new update from the net, and forwards this
			// to the network module. If the message was from another elevator,
			// send the message to the backup-module, who then transmits this
			// to the cost-function who informs the state-machine
		case update := <-ch_new_update:
			if update.Elevator != id {
				ch_status_update <- update
			}

			// Same as above, but for new states
		case status := <-ch_new_status:
			ch_status_refresh <- status

			// Receives new state from the state machine when the state has changed.
			// If the new state is that the motor is broken, then disconnect from
			// the network. If the new state is that the motor is not broken anymore,
			// re-connect to the network. For every other state change,
			// send the new state to the net through the acknowledge-module,
			// so we can be sure every peer has received the new state,
			// and then forward the new state to the backup-module who then feeds
			// this into the cost-function which then informs the state machine (...)
		case update := <-ch_network_update:
			if update.MessageType == 9 {
				elevatorAvaliable = false
				ch_peer_TX_enable <- elevatorAvaliable
				continue
			} else if elevatorAvaliable == false && update.MessageType == 4 {
				ch_peer_TX_enable <- true
			}
			acknowledge.SendUpdate(update)
			ch_status_update <- update

			// Receives a peer update everytime there is a change in peers.
			// Forward peerlist to acknowledge so it knows who it has to receive
			// acknowledges from. If we have lost peers, send this info to the backup-module
			// who then feeds this into the cost-function which then informs the state machine.
			// If we have new peers, request the newest complete state information,
			// and then send this to the net through the ack-module, so it knows
			// the new peer (and everyone else) has received the state information.
		case peerlist = <-ch_peer_update:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", peerlist.Peers)
			fmt.Printf("  New:      %q\n", peerlist.New)
			fmt.Printf("  Lost:     %q\n", peerlist.Lost)
			ch_ack_peer_update <- peerlist
			if peerlist.Lost != "" {
				update := backup.UpdateMessage{
					MessageType:	5,
					Elevator:			peerlist.Lost,
				}
				ch_status_update <- update
			}
			if peerlist.New != "" {
				acknowledge.SendStatus(<-ch_status_broadcast)
			}
		}
	}
}
