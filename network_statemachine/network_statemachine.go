package network_statemachine

//-----------------------------------------------------------------------------------------
//--------------- Receives, and delegates orders, the brain of each elevator---------------
//-----------------------------------------------------------------------------------------

//Import packages
import (
	d "./../datatypes"
	"./../timer"
	"./../network_go/peers"
	u "./../utilities"
	"fmt"

)

//=======States==========
var is_master = false
var current_master_id = ""
var id = ""
var localIp = ""
var peers_port = 0
var connected_elevator_count = 1

var sync_state = d.State{""}


//var bcast_port = peers_port + 1

//=======Functions=======

//Runs statemachine logic
func Run(state_elev_channel chan d.State_elev_message, state_sync_channel chan d.State_sync_message, port int, id_in string) {

	id = id_in

	//Channels for determining peers
	peers_tx_channel := make(chan bool)
	peers_rx_channel := make(chan peers.PeerUpdate)

	//Start peer system
	go peers.Receiver(port, peers_rx_channel)
	go peers.Transmitter(port, id, peers_tx_channel)

	//Customizes test state variable
	sync_state = d.State_init()
	sync_state.Word = "VAR: " + id

	//Starts timer
	timer_chan := make(chan bool)
	go timer.Run(timer_chan)


	//Start regular operation
	for {
		select {

		case pu := <-peers_rx_channel: //Receive update on connected elevators
			reevaluate_master_state(pu,timer_chan) //Redetermine MASTER
			if is_master{
				state_sync_channel <- d.State_sync_message{sync_state,connected_elevator_count}//Inform sync module
			}
			fmt.Printf("\nCHANGE IN NETWORK DETECTED: %q\n", pu.Peers) //Print all current elevators
			fmt.Printf("Connected elevator counter: %d connections\n", connected_elevator_count)
			fmt.Printf("MASTER STATE: %t\n\n", is_master)

		case <- timer_chan: //Synchronizes state if master
			fmt.Printf("TEST SYNC VARIABLE: " + sync_state.Word)
			if is_master{
				state_sync_channel <- d.State_sync_message{sync_state, connected_elevator_count}
				fmt.Println("")
			}

		case message := <- state_sync_channel: //Receives update from sync module
			if (sync_state != message.SyncState){ //Updates state variable
				fmt.Printf("State variable updated, was %s is now %s", sync_state, message.SyncState)
				sync_state = message.SyncState
			}



		}




	}
}

//Enables master state
func enableMasterState(timer_chan chan bool) {
	is_master = true
	current_master_id = id

	timer_chan<-true //Activate timer

	fmt.Println("Enables master state")
}

//Enables master state
func removeMasterState(timer_chan chan bool) {
	is_master = false

	timer_chan<-false //Deactivate timer

	fmt.Println("Removes master state")
}

//Determines current master from peerupdate, aka elevator with lowest id. Fills in current_master variable
func reevaluate_master_state(pu peers.PeerUpdate, timer_chan chan bool){

	update_connected_count(pu) //Updates connected elevator count



	id_lowest := 999999999999999999 //Determines lowest id -> master
	for i := 0; i < len(pu.Peers); i++{
		if (id_lowest >= u.StrToInt(pu.Peers[i])){
			id_lowest = u.StrToInt(pu.Peers[i])
		}
	}

	if (id_lowest == u.StrToInt(id)){ //If this elevator has lowest id, we become master
		fmt.Printf("This is elevator with lowest id, (our id: %s)\n",id)
		if (!is_master){
			enableMasterState(timer_chan)
		} else{
			fmt.Println("Already master, no change necessarry, still MASTER")
		}

	} else if is_master{ //If we are master we remove this status, since an other master has arrived
		fmt.Printf("This elevator does not have lowest id anymore, (our id: %s, other id: %s)\n",id, current_master_id)
		current_master_id = fmt.Sprintf("%d",id_lowest) //Changes master ID
		removeMasterState(timer_chan)
	} else{ //Do nothing
		fmt.Println("Elevator network change detected, no change necessary, still SLAVE")
	}



}

func update_connected_count	(pu peers.PeerUpdate){ //Updates connected elevator counter
	connected_elevator_count = len(pu.Peers)
}
