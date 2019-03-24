package acknowledge

/*This module aims to maximize the chances for any message sent over the network to be received of all peers.
The mid layer for the communication of updatemessages and statusmessages. Every message is stored in a struct,
and then resent up to 10 times if there are peers no ack has been received from. When all acks have been received
or the message is sent 10 times, it is deleted from the struct. This module communicates with the other peers on the network
in addition to the Network module.
*/
import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"../../../Backup"
	"../bcast"
	"../peers"
)

//TODO: Fix maps

/*---------------------Ack message struct---------------------
Id: ID of the Elevator
SeqNo: Sequence number of the UDP packet
MessageType: Type of UDP message that was sent/received out of:
			- 0: UpdateMessages
			- 1: StatusMessages
-------------------------------------------------------------*/
type AckMsg struct {
	Id      	string
	SeqNo   	int
	MessageType 	int
}

type SentMessages struct {
	UpdateMessages    map[int]backup.UpdateMessage
	StatusMessages    map[int]backup.StatusStruct
	NumberOfTimesSent map[int]int
	NotRecFromPeer    map[int][]string //Acks not received from active peers per seq. no.
}

type AckStruct struct {
	AckMessage AckMsg
	AckTimer   *time.Timer
}

type UpdateMessageStruct struct {
	Message backup.UpdateMessage
	SeqNo   int
}

type StatusMessageStruct struct {
	Message backup.StatusStruct
	SeqNo   int
}

var ID string
var PORT string
var seqNo = 0
var _mtx sync.Mutex = sync.Mutex{}
var updateMessageToSend UpdateMessageStruct
var statusMessageToSend StatusMessageStruct
var sentMessages = new(SentMessages)
var peerlist peers.PeerUpdate

var ch_TX_update = make(chan UpdateMessageStruct)
var ch_TX_state = make(chan StatusMessageStruct)
var ch_RX_update = make(chan UpdateMessageStruct)
var ch_RX_state = make(chan StatusMessageStruct)
var ch_TX_ack = make(chan AckMsg)
var ch_RX_ack = make(chan AckMsg)
var TimeoutAckChan = make(chan AckMsg)

//Main ack-goroutine. Communicates with enclosing Network module and other peers on the network through the bcast submodule.
func Ack(ch_new_update chan<- backup.UpdateMessage, ch_new_status chan<- backup.StatusStruct, ch_ack_peer_update <-chan peers.PeerUpdate) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(r, " ACK fatal panic, unable to recover. Rebooting...", "./main -init=false -port="+PORT, " -id="+ID)
			err := exec.Command("gnome-terminal", "-x", "sh", "-c", "./main -init=false -port="+PORT+" -id="+ID).Run()
			if err != nil {
				fmt.Println("Unable to reboot process, crashing...")
			}
		}
		os.Exit(0)
	}()
	sentMessages.UpdateMessages = make(map[int]backup.UpdateMessage)
	sentMessages.StatusMessages = make(map[int]backup.StatusStruct)
	sentMessages.NumberOfTimesSent = make(map[int]int)
	sentMessages.NotRecFromPeer = make(map[int][]string)

	// Start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(16569, ch_TX_update, ch_TX_state, ch_TX_ack)
	go bcast.Receiver(16569, ch_RX_update, ch_RX_state, ch_RX_ack)
	//Main loop in Ack-module
	for {
		select {
		case update := <-ch_RX_update:
			ackMessage := AckMsg{
				Id:      		ID,
				SeqNo:   		update.SeqNo,
				MessageType: 	0,
			}
			if update.Message.Elevator != ID {
				ch_new_update <- update.Message
			}
			ch_TX_ack <- ackMessage

		case status := <-ch_RX_state:
			ackMessage := AckMsg{
				Id:      		ID,
				SeqNo:   		status.SeqNo,
				MessageType: 	1,
			}
			ch_TX_ack <- ackMessage
			ch_new_status <- status.Message
		case notReceivedAck := <-TimeoutAckChan:
			switch notReceivedAck.MessageType {
			case 0: //UpdateMessages
				_mtx.Lock()
				_, ok := sentMessages.UpdateMessages[notReceivedAck.SeqNo]
				if ok && (sentMessages.NumberOfTimesSent[notReceivedAck.SeqNo] < 10) { //Resend if sent <10 times
					updateMessageToSend.Message = sentMessages.UpdateMessages[notReceivedAck.SeqNo]
					updateMessageToSend.SeqNo = notReceivedAck.SeqNo
					ch_TX_update <- updateMessageToSend
					sentMessages.NumberOfTimesSent[notReceivedAck.SeqNo] += 1
					newAckStruct := AckStruct{
						AckMessage: AckMsg{
							Id:      		notReceivedAck.Id,
							SeqNo:   		notReceivedAck.SeqNo,
							MessageType: 	0,
						},
						AckTimer: time.NewTimer(15 * time.Millisecond),
					}
					go ackTimer(TimeoutAckChan, newAckStruct)
				}
				_mtx.Unlock()
			case 1: //StatusMessages
				_mtx.Lock()
				_, ok := sentMessages.StatusMessages[notReceivedAck.SeqNo]
				if ok && (sentMessages.NumberOfTimesSent[notReceivedAck.SeqNo] < 10) {
					statusMessageToSend.Message = sentMessages.StatusMessages[notReceivedAck.SeqNo]
					statusMessageToSend.SeqNo = notReceivedAck.SeqNo
					ch_TX_state <- statusMessageToSend
					sentMessages.NumberOfTimesSent[notReceivedAck.SeqNo] += 1
					newAckStruct := AckStruct{
						AckMessage: AckMsg{
							Id:      		notReceivedAck.Id,
							SeqNo:   		notReceivedAck.SeqNo,
							MessageType: 	1,
						},
						AckTimer: time.NewTimer(15 * time.Millisecond),
					}
					go ackTimer(TimeoutAckChan, newAckStruct)
				}
				_mtx.Unlock()
			}
		case recAck := <-ch_RX_ack:
			_mtx.Lock()
			_, ok := sentMessages.NotRecFromPeer[recAck.SeqNo] //In case the SeqNo has been deleted unexpectedly
			if ok {
				ind := stringInSlice(recAck.Id, sentMessages.NotRecFromPeer[recAck.SeqNo])
				if ind != -1 {
					sentMessages.NotRecFromPeer[recAck.SeqNo] = removeFromSlice(sentMessages.NotRecFromPeer[recAck.SeqNo], ind)
					if len(sentMessages.NotRecFromPeer[recAck.SeqNo]) == 0 {
						delete(sentMessages.NotRecFromPeer, recAck.SeqNo)
						delete(sentMessages.NumberOfTimesSent, recAck.SeqNo) //Delete from NumberOfTimesSent
						switch recAck.MessageType {
						case 0: //UpdateMessages
							delete(sentMessages.UpdateMessages, recAck.SeqNo) //Delete from UpdateMessages
						case 1: //StatusMessages
							delete(sentMessages.StatusMessages, recAck.SeqNo) //Delete from StatusMessages
						}
					}
				}
			}
			_mtx.Unlock()
			//Delete lost peer from NotRecFromPeer
		case newpeerlist := <-ch_ack_peer_update:
			_mtx.Lock()
			peerlist = newpeerlist
			if peerlist.Lost != "" {
				for seqNo, peers := range sentMessages.NotRecFromPeer {
					ind := stringInSlice(peerlist.Lost, peers)
					if ind != -1 {
						sentMessages.NotRecFromPeer[seqNo] = removeFromSlice(sentMessages.NotRecFromPeer[seqNo], ind)
					}
				}
			}
			_mtx.Unlock()
		}
	}
}

//Function for sending update messages over the network. Adds the messages to the sentMessages struct. Spawns an ack-timer to wait for acks on the message.
func SendUpdate(update backup.UpdateMessage) {
	_mtx.Lock()
	defer _mtx.Unlock()
	seqNo += 1
	updateMessageToSend.Message = update
	updateMessageToSend.SeqNo = seqNo
	ch_TX_update <- updateMessageToSend
	if len(peerlist.Peers) != 0 {
		sentMessages.UpdateMessages[seqNo] = update
		sentMessages.NumberOfTimesSent[seqNo] = 1
		sentMessages.NotRecFromPeer[seqNo] = peerlist.Peers

		newAckStruct := AckStruct{
			AckMessage: AckMsg{
				Id:      		ID,
				SeqNo:   		seqNo,
				MessageType: 	0,
			},
			AckTimer: time.NewTimer(15 * time.Millisecond),
		}
		go ackTimer(TimeoutAckChan, newAckStruct)
	}
}

//Function for sending status messages over the network. Adds the messages to the sentMessages struct. Spawns an ack-timer to wait for acks on the message.
func SendStatus(statusUpdate backup.StatusStruct) {
	_mtx.Lock()
	defer _mtx.Unlock()
	seqNo += 1
	statusMessageToSend.Message = statusUpdate
	statusMessageToSend.SeqNo = seqNo
	ch_TX_state <- statusMessageToSend

	if len(peerlist.Peers) != 0 {
		sentMessages.StatusMessages[seqNo] = statusMessageToSend.Message
		sentMessages.NumberOfTimesSent[seqNo] = 1
		sentMessages.NotRecFromPeer[seqNo] = peerlist.Peers

		newAckStruct := AckStruct{
			AckMessage: AckMsg{
				Id:      		ID,
				SeqNo:   		seqNo,
				MessageType: 	1,
			},
			AckTimer: time.NewTimer(15 * time.Millisecond),
		}
		go ackTimer(TimeoutAckChan, newAckStruct)
	}
}

//Goroutine for waiting for ack messages. When the timer finishes, the ack-message is sent to the TimeoutAckChan.
func ackTimer(TimeoutAckChan chan<- AckMsg, ackStruct AckStruct) {
	for {
		select {
		case <-ackStruct.AckTimer.C:
			TimeoutAckChan <- ackStruct.AckMessage
			return
		}
	}
}
