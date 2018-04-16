package net_fsm

//-----------------------------------------------------------------------------------------
//--------------- Receives, synchronizes, and delegates orders. ---------------------------
//-----------------------------------------------------------------------------------------

//Import packages
import (
	"fmt"
	d "../datatypes"
	"../network_go/peers"
	"../network_go/bcast"
	"../timer"
	s "../settings"
	u "../utilities"
	"time"
	"math/rand"
	"sync"
)

//=======States==========
var Master_state = false
var master_change sync.Mutex
var current_master_id = ""
var id = ""
var localIp = ""
var current_peers = []string{}
var peers_port = 0
var connected_elevator_count = 1
var State = d.State{}


//=======Functions=======

//Runs statemachine logic
func Run(
	netfsm_sync_ch_send_state chan d.State_sync_message,
	netfsm_sync_ch_arriving_state chan d.State,
	netfsm_sync_ch_error chan bool,
	netfsm_order_channel chan d.State_order_message,
	id_in string,
	){

	id = id_in

	//Initializes mutexes
	master_change = sync.Mutex{}

	//Starts timer
	timer_chan := make(chan bool)
	go timer.Run(timer_chan, s.DELEGATE_ORDER_DELAY)

	//Prefetch channels
	fetch_rx_ch := make(chan d.Network_fetch_message, 100)
	fetch_tx_ch := make(chan d.Network_fetch_message, 100)
	go bcast.Transmitter(s.FETCH_PORT, fetch_tx_ch)
	go bcast.Receiver(s.FETCH_PORT, fetch_rx_ch)

	//Fetch state before we enable peers system
	fetch_state(fetch_tx_ch,fetch_rx_ch)

	//Channels for determining peers
	peers_tx_channel := make(chan bool)
	peers_rx_channel := make(chan peers.PeerUpdate)

	//Start peer system
	go peers.Receiver(s.PEERS_PORT, peers_rx_channel)
	go peers.Transmitter(s.PEERS_PORT, id, peers_tx_channel)

	//Start regular operation
	for {

		select {

		//-----------------Answers to hello message
	case msg := <-fetch_rx_ch:
			if Check_master_state() && msg.Hello{
				fetch_tx_ch <- d.Network_fetch_message{State,false}
			}

		//-----------------Receive update on connected elevators
		case pu := <-peers_rx_channel:
			current_peers = pu.Peers
			connected_elevator_count = len(pu.Peers)
			reevaluate_Master_state(pu, timer_chan,fetch_tx_ch,fetch_rx_ch)
			if Check_master_state() {
				sync_state(netfsm_sync_ch_send_state)
			}
			fmt.Printf("\nNetwork FSM: Network change detected: %q, %d, %v\n", pu.Peers, connected_elevator_count, Master_state) //Print current info


		//-----------------Delegates orders on regular timing intervals
		case <-timer_chan:
			if Check_master_state(){

				distributed_order := find_order() //Finds new order to delegate

				if distributed_order.Floor != s.NO_FLOOR_FOUND { //Delegate the found order
					if delegate_order(netfsm_order_channel, distributed_order) {//This is true if the order is executed

						fmt.Printf("Network FSM: Order Delegated: Floor %d: at time: %d \n", distributed_order.Floor,int(time.Now().Unix()))
						update_timetable_delegation(distributed_order)
						sync_state(netfsm_sync_ch_send_state)
					}
				}
			}


		//-----------------Receives update from sync module
	case new_state := <-netfsm_sync_ch_arriving_state:
			fmt.Println("Network FSM: State variable updated")
			State = new_state


		//-----------------Receives update from order handler
		case message := <-netfsm_order_channel:

			if Check_master_state() && !message.Order.Fin { //It is a new order

				if new_order_check(message.Order) { //Checks if this order means we must update State

					if len(current_peers) < 2{ //We should not accept orders if there are fewer than two elevators connected
						fmt.Printf("Net FSM: Order received, not enough elevators to ensure execution\n")
						continue
					}

					add_order(message.Order)
					update_timetable_received(message.Order)
					sync_state(netfsm_sync_ch_send_state)
				}

			} else if Check_master_state() && message.Order.Fin { //An order has been finished
				fmt.Printf("Network FSM: Order completed, floor %d, up: %v, down: %v\n", message.Order.Floor, message.Order.Up, message.Order.Down)
				clear_order(message.Order)
				sync_state(netfsm_sync_ch_send_state)
			}


		//-----------------If sync fails, we resync
		case <- netfsm_sync_ch_error:
			fmt.Println("Network FSM: Resyncing")
			sync_state(netfsm_sync_ch_send_state)
		}
	}
}

//Enables master state
func enableMasterState(timer_chan chan bool) {

	//Locking mutex
	master_change.Lock()
	defer master_change.Unlock()

	Master_state = true
	timer_chan <- true //Activate timer
	fmt.Println("Network FSM: Enables master state")
}

//Removes master state
func removeMasterState(timer_chan chan bool) {

	//Locking mutex
	master_change.Lock()
	defer master_change.Unlock()

	Master_state = false
	timer_chan <- false //Deactivate timer
	fmt.Println("Network FSM: Removes master state")
}

func Check_master_state() bool{

	//Locking mutex
	master_change.Lock()
	defer master_change.Unlock()

	return Master_state
}

func Get_number_of_peers() int{
	return len(current_peers)
}

//Determines current master from peerupdate, aka elevator with lowest id. Fills in current_master variable
func reevaluate_Master_state(pu peers.PeerUpdate,
														timer_chan chan bool,
														fetch_tx_ch chan d.Network_fetch_message,
														fetch_rx_ch chan d.Network_fetch_message) {

	fmt.Printf("Network FSM: Redetermining master state: ")

	if len(pu.Peers) == 2{
		fetch_state(fetch_tx_ch,fetch_rx_ch)
	}

	update_connected_count(pu) //Updates connected elevator count

	id_lowest := 999999999999999999 //Determines lowest id -> master
	for i := 0; i < len(pu.Peers); i++ {
		if id_lowest >= u.StrToInt(pu.Peers[i]) {
			id_lowest = u.StrToInt(pu.Peers[i])
		}
	}

	if id_lowest == u.StrToInt(id) { //If this elevator has lowest id, we become master
		if !Master_state {
			fmt.Printf("Become MASTER\n")
			enableMasterState(timer_chan)
		} else {
			fmt.Printf("Already MASTER, still MASTER\n")
		}

	} else if Check_master_state() { //If we are master we remove this status, since an other master has arrived

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

func update_timetable_received(order d.Order_struct){ //Updates timetable with order, keeps track of the time it was received
	time :=  int(time.Now().Unix())
	if order.Up {
		State.Time_table_received_up[order.Floor] = time
	} else if order.Down {
		State.Time_table_received_down[order.Floor] = time
	}
}

func update_timetable_delegation(order d.Order_struct){ //Updates timetable with order, keeps track of the time it was delegated
	time :=  int(time.Now().Unix())
	if order.Up {
		State.Time_table_delegated_up[order.Floor] = time
	} else if order.Down {
		State.Time_table_delegated_down[order.Floor] = time
	}
}

func time_check(order_time int) bool {	//Checks if order time has expired. Then we must send another elevator

	result := int(time.Now().Unix())-order_time > s.ORDER_TIMOUT_DELAY || order_time == s.ORDER_INACTIVE //Order timeouts after 10 seconds

	if (result && order_time != s.ORDER_INACTIVE) {
		fmt.Printf("Network FSM: Order timed out\n")
	}

	return result
}

func delegate_order(netfsm_order_channel chan d.State_order_message, order d.Order_struct) bool { //Delegates order to available slaves

	//The direction we iterate through is randomized
	dir := 1	//1 -> up
	if 50 < rand.Intn(100){
		dir = -1 //-1 -> down
	}

	if (dir == 1){ //We iterate up
		for i := 0; i < len(current_peers); i += dir {
			//Notify order order_handler
			message := d.State_order_message{order, current_peers[i], false}
			netfsm_order_channel <- message

			select {
			case response := <-netfsm_order_channel: //Receives order update from order handler
				if response.ACK { //If we ACK the order has been executed
					return true
				}
			}
		}
	} else if (dir == -1){ //We iterate down
		for i := len(current_peers)-1; i >= 0; i += dir {
			//Notify order order_handler
			message := d.State_order_message{order, current_peers[i], false}
			netfsm_order_channel <- message

			select {
			case response := <-netfsm_order_channel: //Receives order update from order handler
				if response.ACK { //If we ACK the order has been executed
					return true
				}
			}
		}
	}


	return false
}

func add_order(order d.Order_struct) { //Updates state from new order
	if order.Up {
		State.Button_matrix.Up[order.Floor] = true
	} else if order.Down {
		State.Button_matrix.Down[order.Floor] = true
	}
}

func sync_state(netfsm_sync_ch_send_state chan d.State_sync_message) { //Syncs state with slave-elevators

	//Converting connected elevator list to string for sync
	current_peers_string :=  u.ListToString(current_peers)
	netfsm_sync_ch_send_state <- d.State_sync_message{State, current_peers_string, true} //Inform sync module
}

func clear_order(order d.Order_struct) { //Updates state when an order has been executed

	if order.Up{
		State.Button_matrix.Up[order.Floor] = false
		State.Time_table_delegated_up[order.Floor] = s.ORDER_INACTIVE
		State.Time_table_received_up[order.Floor] = s.ORDER_INACTIVE
	}

	if order.Down {
		State.Button_matrix.Down[order.Floor] = false
		State.Time_table_delegated_down[order.Floor] = s.ORDER_INACTIVE
		State.Time_table_received_up[order.Floor] = s.ORDER_INACTIVE
	}
}

func new_order_check(order d.Order_struct) bool { //Figures out new order means that we must sync state / update state
	if order.Up{
		return !State.Button_matrix.Up[order.Floor]
	} else if order.Down{
		return !State.Button_matrix.Down[order.Floor]
	}
	return true
}

func find_order() d.Order_struct{ //Finds next order to delegate, depending on how old the order is

	up := false
	down := false
	floor := s.NO_FLOOR_FOUND

	current_time := int(time.Now().Unix())
	max_wait_time := -1

	for i := 0; i < 4; i++ { //Look through state
		if State.Button_matrix.Up[i] && time_check(State.Time_table_delegated_up[i]){ //If order is present and time has run out we delegate order

			if(current_time - State.Time_table_received_up[i] >= max_wait_time){ //Checks if this order is the olders one
				up = true
				down = false
				floor = i
				max_wait_time = current_time - State.Time_table_received_up[i]
			}
		} else if State.Button_matrix.Down[i] && time_check(State.Time_table_delegated_down[i]){ //If order is present and time has run out we delegate order

			if(current_time - State.Time_table_received_down[i] >= max_wait_time){ //Checks if this order is the olders one
				down = true
				up = false
				floor = i
				max_wait_time = current_time - State.Time_table_received_down[i]
			}
		}
	}
	return d.Order_struct{floor, up, down, false} //Order to be executed
}

//Prefetches State from existing network (if there is one)
func fetch_state(fetch_tx_ch chan d.Network_fetch_message, fetch_rx_ch chan d.Network_fetch_message){

	fmt.Printf("Net FSM: Prefetching state ")

	for i := 0; i < s.FETCH_MAX_TIMEOUT_COUNT; i++{

		//Send hello message
		fetch_tx_ch <- d.Network_fetch_message{d.State{}, true}

		//Setting up timout signal
    timeOUT := time.NewTimer(time.Millisecond * s.FETCH_TIMEOUT_DELAY)

		timeout := false
		for !timeout{
			select{
			case message := <-fetch_rx_ch: //Receives new state from existing network
			if !message.Hello{
				State = message.State
				fmt.Printf("0 \n")
				return
			}

			case <-timeOUT.C:	//If we time out we try again / quit
				fmt.Printf("X ")
				timeout = true
			}
		}
	}
	fmt.Printf("\n")
}
