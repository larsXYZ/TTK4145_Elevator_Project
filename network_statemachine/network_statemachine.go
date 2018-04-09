package network_statemachine

//-----------------------------------------------------------------------------------------
//--------------- Receives, synchronizes, and delegates orders. ---------------------------
//-----------------------------------------------------------------------------------------

//Import packages
import (
	"fmt"
	d "../datatypes"
	"../network_go/peers"
	"../timer"
	u "../utilities"
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
func Run(state_elev_channel chan d.State_elev_message, state_sync_channel chan d.State_sync_message, state_order_channel chan d.State_order_message, port int, id_in string) {

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

	//Clear lights
	update_lights(state_elev_channel)

	//Start regular operation
	for {
		fmt.Println("") //Starts new line in debug

		select {

		case pu := <-peers_rx_channel: //Receive update on connected elevators
			reevaluate_master_state(pu, timer_chan)
			current_peers = pu.Peers
			if is_master {
				sync_state(state_sync_channel)
			}
			fmt.Printf("\nNetwork FSM: Network change detected: %q, %d, %v\n", pu.Peers, connected_elevator_count, is_master) //Print current info

		case <-timer_chan: //Tests sending order and other things -----------------------------------
			if is_master{

				fmt.Printf("Network FSM: Delegating order: ")

				//Check if we have an order to distribute
				up := false
				down := false
				floor := -1 //-1 indicates that no new order has been found

				for i := 0; i < 4; i++ { //Look through state
					if State.Button_matrix.Up[i] {
						up = true
						floor = i
						break
					} else if State.Button_matrix.Down[i] {
						down = true
						floor = i
						break
					}
				}

				order := d.Order_struct{floor, up, down, false, false} //Order to be executed

				if floor != -1 { //Delegate the found order
					fmt.Printf("Order found: floor %d: \n", floor)
					if delegate_order(state_order_channel, order) {//This is true if the order is executed
						remove_order(order)
						sync_state(state_sync_channel)
						update_lights(state_elev_channel)
						fmt.Printf("Order executed\n")
					} else {
						fmt.Printf("All elevators busy\n")
					}
				} else {
					fmt.Printf("No order found..\n")
				}
			}

		case message := <-state_sync_channel: //Receives update from sync module
			fmt.Println("Network FSM: State variable updated")
			State = message.SyncState
			update_lights(state_elev_channel)

		case message := <-state_order_channel: //Receives update from order handler
			if is_master {
				update_state(message.Order)
				sync_state(state_sync_channel)
				fmt.Printf("Network FSM: Syncing new state with slaves\n")
				update_lights(state_elev_channel)
			}
		}
	}
}

//Enables master state
func enableMasterState(timer_chan chan bool) {
	is_master = true
	current_master_id = id

	timer_chan <- true //Activate timer

	fmt.Println("Network FSM: Enables master state")
}

//Removes master state
func removeMasterState(timer_chan chan bool) {
	is_master = false

	timer_chan <- false //Deactivate timer

	fmt.Println("Network FSM: Removes master state")
}

//Determines current master from peerupdate, aka elevator with lowest id. Fills in current_master variable
func reevaluate_master_state(pu peers.PeerUpdate, timer_chan chan bool) {

	fmt.Printf("Network FSM: Redetermining master state: ")

	update_connected_count(pu) //Updates connected elevator count

	id_lowest := 999999999999999999 //Determines lowest id -> master
	for i := 0; i < len(pu.Peers); i++ {
		if id_lowest >= u.StrToInt(pu.Peers[i]) {
			id_lowest = u.StrToInt(pu.Peers[i])
		}
	}

	if id_lowest == u.StrToInt(id) { //If this elevator has lowest id, we become master
		if !is_master {
			fmt.Printf("Become MASTER\n")
			enableMasterState(timer_chan)
		} else {
			fmt.Printf("Already MASTER, still MASTER\n")
		}

	} else if is_master { //If we are master we remove this status, since an other master has arrived

		current_master_id = fmt.Sprintf("%d", id_lowest) //Changes master ID
		fmt.Printf("Removes master state, becomes SLAVE\n")
		removeMasterState(timer_chan)
	} else { //Do nothing
		fmt.Printf("Already SLAVE, still SLAVE\n")
	}

}

func update_connected_count(pu peers.PeerUpdate) { //Updates connected elevator counter
	connected_elevator_count = len(pu.Peers)
}

func delegate_order(state_order_channel chan d.State_order_message, order d.Order_struct) bool { //Delegates order to available slaves

	for i := 0; i < len(current_peers); i++ {

		//Notify order order_handler
		message := d.State_order_message{order, current_peers[i], false}
		state_order_channel <- message

		select {

		case response := <-state_order_channel: //Receives order update from order handler
			if response.ACK { //If we ACK the order has been executed
				return true
			}
		}
	}
	return false
}

func update_state(order d.Order_struct) { //Updates state from new order
	if order.Up {
		State.Button_matrix.Up[order.Floor] = true
	} else if order.Down {
		State.Button_matrix.Down[order.Floor] = true
	}
}

func sync_state(state_sync_channel chan d.State_sync_message) { //Syncs state with slave-elevators
	state_sync_channel <- d.State_sync_message{State, connected_elevator_count} //Inform sync module
}

func update_lights(state_elev_channel chan d.State_elev_message) { //Tells elevator to update lights
	state_elev_channel <- d.State_elev_message{State.Button_matrix, true}
}

func remove_order(order d.Order_struct) { //Updates state when an order has been executed
	if order.Up{
		State.Button_matrix.Up[order.Floor] = false
	} else if order.Down {
		State.Button_matrix.Down[order.Floor] = false
	}
}
