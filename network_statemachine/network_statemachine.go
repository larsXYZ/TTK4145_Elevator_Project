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
	/*
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
	*/
	//Start regular operation
	for {
		select {

		case pu := <-peers_rx_channel: //Receive update on connected elevators
			fmt.Printf("  Elevators in system: %q\n", pu.Peers) //Print all current elevators
			determine_new_master(pu)

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

//Determines current master from peerupdate, aka elevator with lowest id
func determine_new_master(pu peers.PeerUpdate){

	if len(pu.Peers) == 1{ //If we are only elevator on network we become master
		fmt.Println("No other elevators")
		enableMasterState()
		return
	}

	id_lowest := 999999999999999999 //Determines lowest id -> master
	for i := 0; i < len(pu.Peers); i++{
		if (id_lowest >= u.StrToInt(pu.Peers[i])){
			id_lowest = u.StrToInt(pu.Peers[i])
		}
	}

	if (id_lowest == u.StrToInt(id)){ //If this elevator has lowest id, we become master
		fmt.Printf("This is elevator with lowest id, (our id: %s)\n",id)
		if (!is_master){
			enableMasterState()
		} else{
			fmt.Println("Already master, no change necessarry, still MASTER")
		}

	} else if is_master{ //If we are master we remove this status, since an other master has arrived
		fmt.Printf("This elevator does not have lowest id anymore, (our id: %s, other id: %d)\n",id, id_lowest)
		removeMasterState()
	} else{ //Do nothing
		fmt.Println("Elevator network change detected, no change necessary, still SLAVE")
	}


}
