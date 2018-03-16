package network

//-----------------------------------------------------------------------------------------
//--------------- Communicates with other elevators and statemachine-----------------------
//-----------------------------------------------------------------------------------------

import (
	d "./../datatypes"
	"./../network_go/bcast"
	"./../network_go/localip"
	"./../network_go/peers"
	"fmt"
	"os"
	"time"
)

func Init(state_network_channel chan d.State_network_message, port int) {

	//Channels to communicate with network_go package
	heartbeat_tx_channel := make(chan d.Heartbeat)
	heartbeat_rx_channel := make(chan d.Heartbeat)
	peers_tx_channel := make(chan bool)
	peers_rx_channel := make(chan peers.PeerUpdate)

	//Finds ip & id
	localIp, _ := localip.LocalIP()
	var id string = fmt.Sprintf("%s - %d", localIp, os.Getpid())

	//Start peer system
	go peers.Transmitter(port, id, peers_tx_channel)
	go peers.Receiver(port, peers_rx_channel)

	//Spam broadcast
	go bcast.Transmitter(15647, heartbeat_tx_channel)
	go bcast.Receiver(15647, heartbeat_rx_channel)
	go func() {
		message_tx := d.Heartbeat{Id: id, Ip: localIp}
		for {
			heartbeat_tx_channel <- message_tx
			time.Sleep(2 * time.Second)
		}
	}()

	//Notify user that network module is operational
	fmt.Printf("Network module online, id: %s, port: %d\n", id, port)

	//Look at pipes
	for {
		select {
		case message_rx := <-heartbeat_rx_channel:
			if message_rx.Id != id {
				fmt.Println(message_rx)
			}

		case p := <-peers_rx_channel:
			fmt.Printf("  Peers:    %q\n", p.Peers)

		}
	}
}
