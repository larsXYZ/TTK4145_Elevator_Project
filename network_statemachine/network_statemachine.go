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
var current_peers = []string{}
var peers_port = 0
var connected_elevator_count = 1
var State = d.State{}

//=======Functions=======

//Runs statemachine logic
func Run(state_elev_channel chan d.State_elev_message, state_sync_channel chan d.State_sync_message, state_order_channel chan d.State_order_message , port int, id_in string) {

	id = id_in

	//Channels for determining peers
	peers_tx_channel := make(chan bool)
	peers_rx_channel := make(chan peers.PeerUpdate)

	//Start peer system
	go peers.Receiver(port, peers_rx_channel)
	go peers.Transmitter(port, id, peers_tx_channel)

	//Starts timer
	timer_chan := make(chan bool)
	go timer.Run(timer_chan)


	//Start regular operation
	for {
		fmt.Println("") //Starts new line in debug

		select {

		case pu := <-peers_rx_channel: //Receive update on connected elevators
			reevaluate_master_state(pu,timer_chan)
			current_peers = pu.Peers
			if is_master{
				sync_state(state_sync_channel)
			}
			fmt.Printf("\nCHANGE IN NETWORK DETECTED: %q\n", pu.Peers) //Print all current elevators
			fmt.Printf("Connected elevator counter: %d connections\n", connected_elevator_count)
			fmt.Printf("MASTER STATE: %t\n\n", is_master)

		//case <- timer_chan: //Tests sending order and other things -----------------------------------
		//	if is_master && len(current_peers) > 1{
		//		delegate_order(state_order_channel, d.Order_struct{3,true,false,false})
		//	}

		case message := <- state_sync_channel: //Receives update from sync module
			fmt.Println("State variable updated")
			fmt.Println(State)
			fmt.Println(message.SyncState)
			fmt.Println("")
			State = message.SyncState
			//update_lights(state_elev_channel)

		case message := <-state_order_channel: //Receives update from order handler
			if (is_master) {
				fmt.Printf("Master: New order received: ")
				update_state(message.Order)
				sync_state(state_sync_channel)
				fmt.Printf("Syncing state with slaves\n")
				fmt.Println(State)
				//update_lights(state_elev_channel)
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

//Removes master state
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


func delegate_order(state_order_channel chan d.State_order_message, order d.Order_struct){ //Delegates order to available slaves

	fmt.Println("-----Delegating order-----")
	finished_state := false

	for (!finished_state){

		for i := 0; i < len(current_peers); i++{

			//Dont send order to self
			if current_peers[i] == id{
				continue
			}

			//For debugging..
			fmt.Printf("Sending order to id: %s (%d of %d) - ", current_peers[i],i+1,len(current_peers))

			//Notify order order_handler
			message := d.State_order_message{order, current_peers[i], false}
			state_order_channel <- message

			select{

			case response := <- state_order_channel: //Receives order update from order handler
				if (response.ACK){ //If we ACK the order has been executed
					fmt.Println("")
					return
				}
			}
		}
	}
}

func update_state(order d.Order_struct){ //Updates state from new order
	if (order.Up){
		State.Button_matrix.Up[order.Floor] = true;
	} else if (order.Down){
		State.Button_matrix.Down[order.Floor] = true;
	}
}

func sync_state(state_sync_channel chan d.State_sync_message){ //Syncs state with slave-elevators
	state_sync_channel <- d.State_sync_message{State,connected_elevator_count}//Inform sync module
}

func update_lights(state_elev_channel chan d.State_elev_message){ //Tells elevator to update lights
	fmt.Println("Netstate: update_lights()")
	state_elev_channel <- d.State_elev_message{State.Button_matrix, true}
}
