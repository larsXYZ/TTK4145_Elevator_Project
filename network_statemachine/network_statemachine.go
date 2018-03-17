package network_statemachine

//-----------------------------------------------------------------------------------------
//--------------- Receives, and delegates orders, the brain of each elevator---------------
//-----------------------------------------------------------------------------------------

//Import packages
import (
	d "./../datatypes"
	"./../network_go/localip"
	"./../network_go/peers"
	u "./../utilities"
	"fmt"
	"os"

)

//=======States==========
var is_master = false
var id = ""
var localIp = ""
var peers_port = 0

//var bcast_port = peers_port + 1

//=======Functions=======

//Runs statemachine logic
func Run(state_elev_channel chan d.State_elev_message, state_sync_channel chan d.State_sync_message, port int) {

	//Set up portnumbers
	peers_port := port

	//Determine ip & id
	localIp, _ = localip.LocalIP()
	id = fmt.Sprintf("%s%d", u.IpToString(localIp), os.Getpid())

	//Channels for determining peers
	peers_tx_channel := make(chan bool)
	peers_rx_channel := make(chan peers.PeerUpdate)

	//Start peer system
	go peers.Receiver(peers_port, peers_rx_channel)
	go peers.Transmitter(peers_port, id, peers_tx_channel)

	//Check if we are alone
	state_sync_channel <- d.State_sync_message{true,false}
	presence_check_result := (<-state_sync_channel).GreetingResponse
	if (presence_check_result){ //Check result
		fmt.Println("|||Active elevators found\n")
		removeMasterState()
	} else{
		fmt.Println("|||No elevators found\n")
		enableMasterState()
	}

	for {
		select {

		case p := <-peers_rx_channel: //Check the other elevators on the system
			fmt.Printf("  Peers:    %q\n", p.Peers)

		}
	}
}

//Enables master state
func enableMasterState() {
	is_master = true
	fmt.Println("Enables master state")
}

//Enables master state
func removeMasterState() {
	is_master = false
	fmt.Println("Removes master state")
}
