package acknowledge

/*
This module will send every message up to 10 times if it doesen't
receive acknowledgement from all the peers on the network.
*/

import (
	"../../../Backup"
	"../bcast"
	"../peers"

	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

type SentMessages struct {
	UpdateMessages    map[int]backup.UpdateMessage
	StatusMessages    map[int]backup.StatusStruct
	NumbTimesSent 		map[int]int
	NotAckFromPeer    map[int][]string
}

type StatusMessageStruct struct {
	Message 			backup.StatusStruct
	SequenceNum   int
}

type UpdateMessageStruct struct {
	Message 			backup.UpdateMessage
	SequenceNum   int
}

type AcknowledgeMessage struct {
	ID      			string
	SequenceNum   int
	MessageType 	int				// 0 = UpdateMessages, 1 = StatusMessages
}

type AcknowledgeStruct struct {
	AcknowledgeMsg 				AcknowledgeMessage
	AcknowledgeTimer   		*time.Timer
}

var ID string
var PORT string

var seqNumb = 0
var _mtx sync.Mutex = sync.Mutex{}

var ch_timeout_ack = make(chan AcknowledgeMessage)
var ch_TX_update = make(chan UpdateMessageStruct)
var ch_RX_update = make(chan UpdateMessageStruct)
var ch_TX_state = make(chan StatusMessageStruct)
var ch_RX_state = make(chan StatusMessageStruct)
var ch_TX_ack = make(chan AcknowledgeMessage)
var ch_RX_ack = make(chan AcknowledgeMessage)

var updateMessageToSend UpdateMessageStruct
var statusMessageToSend StatusMessageStruct
var sentMessages = new(SentMessages)
var peerlist peers.PeerUpdate

func Ack(ch_new_update chan<- backup.UpdateMessage, ch_new_status chan<- backup.StatusStruct, ch_ack_peer_update <-chan peers.PeerUpdate) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r, " Panic in Ack, did not manage to recover. Trying to reboot: ", "port="+PORT, "id="+ID)
			err := exec.Command("gnome-terminal", "-x", "sh", "-c", "./main -init=false -port="+PORT+" -id="+ID).Run()
			if err != nil {
				fmt.Println("ERROR: Did not manage to reboot!")
			}
		}
		os.Exit(0)
	}()
	sentMessages.UpdateMessages = make(map[int]backup.UpdateMessage)
	sentMessages.StatusMessages = make(map[int]backup.StatusStruct)
	sentMessages.NumbTimesSent = make(map[int]int)
	sentMessages.NotAckFromPeer = make(map[int][]string)

	go bcast.Transmitter(16569, ch_TX_update, ch_TX_state, ch_TX_ack)
	go bcast.Receiver(16569, ch_RX_update, ch_RX_state, ch_RX_ack)

	for {
		select {
			// When received new state, send acknowledge to peers and send state to
			// Network-module and then to Backup-module
		case newStatus := <-ch_RX_state:
			ackMessage := AcknowledgeMessage{
				ID:      			ID,
				SequenceNum:  newStatus.SequenceNum,
				MessageType: 	2,
			}
			ch_TX_ack <- ackMessage
			ch_new_status <- newStatus.Message

			// When received new update, send acknowledge to peers and send
			// update to Network-module and then to Backup-module
		case newUpdate := <-ch_RX_update:
			ackMessage := AcknowledgeMessage{
				ID:      			ID,
				SequenceNum:  newUpdate.SequenceNum,
				MessageType: 	1,
			}
			if newUpdate.Message.Elevator != ID {
				ch_new_update <- newUpdate.Message
			}
			ch_TX_ack <- ackMessage

			// If 15 ms has gone by, we have not received all the neccesary acks
			// and we tried to send the message less than 10 times,
			// we will try to resend the same message, and update the number of times
			// the message is sent. Restart the acknowledge timer for another 15 ms.
			// If sent more than 10 times, or if we have received all neccesary acks
			// ("exist" will be false), don't try to resend. We will find out
			// by other means if one or more peers is lost.
			// This is only to "assure" against packet loss.
		case timeoutAck := <-ch_timeout_ack:
			switch timeoutAck.MessageType {
			case 1:
				_mtx.Lock()
				_, exists := sentMessages.UpdateMessages[timeoutAck.SequenceNum]
				if exists && (sentMessages.NumbTimesSent[timeoutAck.SequenceNum] < 10) {
					updateMessageToSend.Message = sentMessages.UpdateMessages[timeoutAck.SequenceNum]
					updateMessageToSend.SequenceNum = timeoutAck.SequenceNum
					ch_TX_update <- updateMessageToSend
					sentMessages.NumbTimesSent[timeoutAck.SequenceNum] += 1
					ackStruct := AcknowledgeStruct{
						AcknowledgeMsg: AcknowledgeMessage{
							ID:      			timeoutAck.ID,
							SequenceNum:  timeoutAck.SequenceNum,
							MessageType: 	1,
						},
						AcknowledgeTimer: time.NewTimer(15*time.Millisecond),
					}
					go AcknowledgeTimeout(ch_timeout_ack, ackStruct)
				}
				_mtx.Unlock()
			case 2:
				_mtx.Lock()
				_, exists := sentMessages.StatusMessages[timeoutAck.SequenceNum]
				if exists && (sentMessages.NumbTimesSent[timeoutAck.SequenceNum] < 10) {
					statusMessageToSend.Message = sentMessages.StatusMessages[timeoutAck.SequenceNum]
					statusMessageToSend.SequenceNum = timeoutAck.SequenceNum
					ch_TX_state <- statusMessageToSend
					sentMessages.NumbTimesSent[timeoutAck.SequenceNum] += 1
					ackStruct := AcknowledgeStruct{
						AcknowledgeMsg: AcknowledgeMessage{
							ID:      			timeoutAck.ID,
							SequenceNum:  timeoutAck.SequenceNum,
							MessageType: 	2,
						},
						AcknowledgeTimer: time.NewTimer(15*time.Millisecond),
					}
					go AcknowledgeTimeout(ch_timeout_ack, ackStruct)
				}
				_mtx.Unlock()
			}

			// When we receive ack from peers, update info on who we have received ack from.
			// If we have received acks from all the peers (len() == 0), empty the associated sentMesseges
			// category (either UpdateMessages or StatusMesseges) for the given message number
			// (SequenceNum)
		case receivedAck := <-ch_RX_ack:
			_mtx.Lock()
			_, exists := sentMessages.NotAckFromPeer[receivedAck.SequenceNum]
			if exists {
				index := findString(receivedAck.ID, sentMessages.NotAckFromPeer[receivedAck.SequenceNum])
				if index != -1 {
					sentMessages.NotAckFromPeer[receivedAck.SequenceNum] = removeString(sentMessages.NotAckFromPeer[receivedAck.SequenceNum], index)
					if len(sentMessages.NotAckFromPeer[receivedAck.SequenceNum]) == 0 {
						delete(sentMessages.NotAckFromPeer, receivedAck.SequenceNum)
						delete(sentMessages.NumbTimesSent, receivedAck.SequenceNum)
						switch receivedAck.MessageType {
						case 1:
							delete(sentMessages.UpdateMessages, receivedAck.SequenceNum)
						case 2:
							delete(sentMessages.StatusMessages, receivedAck.SequenceNum)
						}
					}
				}
			}
			_mtx.Unlock()

			// A new peer-overview has been received in Network-module, and is forwarded here.
			// If we have lost a peer since last time, delete it from our list over peers we want to
			// receive acknowledge from.
		case newPeerList := <-ch_ack_peer_update:
			_mtx.Lock()
			peerlist = newPeerList
			if peerlist.Lost != "" {
				for seqNumb, peers := range sentMessages.NotAckFromPeer {
					index := findString(peerlist.Lost, peers)
					if index != -1 {
						sentMessages.NotAckFromPeer[seqNumb] = removeString(sentMessages.NotAckFromPeer[seqNumb], index)
					}
				}
			}
			_mtx.Unlock()
		}
	}
}

// Send update-message to the net. If there are other peers on the net,
// start timer for acknowledgements.
func SendUpdate(newUpdate backup.UpdateMessage) {
	_mtx.Lock()
	defer _mtx.Unlock()
	seqNumb += 1
	updateMessageToSend.Message = newUpdate
	updateMessageToSend.SequenceNum = seqNumb
	ch_TX_update <- updateMessageToSend

	if len(peerlist.Peers) != 0 {
		sentMessages.UpdateMessages[seqNumb] = newUpdate
		sentMessages.NumbTimesSent[seqNumb] = 1
		sentMessages.NotAckFromPeer[seqNumb] = peerlist.Peers

		ackStruct := AcknowledgeStruct{
			AcknowledgeMsg: AcknowledgeMessage{
				ID:      			ID,
				SequenceNum:  seqNumb,
				MessageType: 	1,
			},
			AcknowledgeTimer: time.NewTimer(15*time.Millisecond),
		}
		go AcknowledgeTimeout(ch_timeout_ack, ackStruct)
	}
}

// Send status-message to the net. If there are other peers on the net,
// start timer for acknowledgements.
func SendStatus(newStatus backup.StatusStruct) {
	_mtx.Lock()
	defer _mtx.Unlock()
	seqNumb += 1
	statusMessageToSend.Message = newStatus
	statusMessageToSend.SequenceNum = seqNumb
	ch_TX_state <- statusMessageToSend

	if len(peerlist.Peers) != 0 {
		sentMessages.StatusMessages[seqNumb] = statusMessageToSend.Message
		sentMessages.NumbTimesSent[seqNumb] = 1
		sentMessages.NotAckFromPeer[seqNumb] = peerlist.Peers

		ackStruct := AcknowledgeStruct{
			AcknowledgeMsg: AcknowledgeMessage{
				ID:      			ID,
				SequenceNum:  seqNumb,
				MessageType: 	2,
			},
			AcknowledgeTimer: time.NewTimer(15*time.Millisecond),
		}
		go AcknowledgeTimeout(ch_timeout_ack, ackStruct)
	}
}

// If 15 ms has gone by, send the associated acknowledge-message and let
// for-select above check if all the acknowledgements has been received or not.
func AcknowledgeTimeout(ch_timeout_ack chan<- AcknowledgeMessage, ackStruct AcknowledgeStruct) {
	for {
		select {
		case <-ackStruct.AcknowledgeTimer.C:
			ch_timeout_ack <- ackStruct.AcknowledgeMsg
			return
		}
	}
}

//Looks for a string in an array and returns the index.
func findString(a string, list []string) int {
	for ind, b := range list {
		if b == a {
			return ind
		}
	}
	return -1
}

//Removes a string from an array, does not care about sorting.
func removeString(s []string, i int) []string {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}
