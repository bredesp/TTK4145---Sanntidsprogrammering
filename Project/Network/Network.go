package network

/* The Network module is the top layer for communication between internal modules of one elevator and communication with other peers on the network.
The Network module receives information from the Fsm module and sends its received information to the status module. It can also enable or disable itself
if it receives a message from the Fsm that the motor is broken.
*/
import (
	"fmt"

	"../Backup"
	"./network/acknowledge"
	"./network/peers"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//  will be received as zero-values.
func Network(ch_status_update chan<- backup.UpdateMessage, ch_status_refresh chan<- backup.StatusStruct, ch_status_broadcast <-chan backup.StatusStruct, ch_network_update <-chan backup.UpdateMessage, id string) {
	var peerlist peers.PeerUpdate

	ch_new_update := make(chan backup.UpdateMessage)
	ch_new_status := make(chan backup.StatusStruct)
	ch_ack_peer_update := make(chan peers.PeerUpdate, 2)

	go acknowledge.Ack(ch_new_update, ch_new_status, ch_ack_peer_update)
	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	ch_peer_update := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	ch_peer_TX_enable := make(chan bool)
	peerTxEnableVar := true
	go peers.Transmitter(16016, id, ch_peer_TX_enable)
	go peers.Receiver(16016, ch_peer_update)

	for {
		select {
		case peerlist = <-ch_peer_update:

			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", peerlist.Peers)
			fmt.Printf("  New:      %q\n", peerlist.New)
			fmt.Printf("  Lost:     %q\n", peerlist.Lost)
			ch_ack_peer_update <- peerlist
			if peerlist.Lost != "" {
				update := backup.UpdateMessage{
					MessageType:	5,
					Elevator:		peerlist.Lost,
				}
				ch_status_update <- update
			}
			if peerlist.New != "" {
				acknowledge.SendStatus(<-ch_status_broadcast)
			}
		case update := <-ch_network_update:
			if update.MessageType == 8 { //motor is broken, disconnect from network
				peerTxEnableVar = false
				ch_peer_TX_enable <- peerTxEnableVar
				continue
			} else if peerTxEnableVar == false && update.MessageType == 2 {
				ch_peer_TX_enable <- true
			}
			acknowledge.SendUpdate(update)

			ch_status_update <- update
		case update := <-ch_new_update:
			if update.Elevator != id {
				ch_status_update <- update
			}
		case status := <-ch_new_status:
			ch_status_refresh <- status
		}
	}
}
