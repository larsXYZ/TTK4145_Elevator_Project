package order_delegator

//-----------------------------------------------------------------------------------------
//---------------Sends/Receives orders to/from other elevators-----------------------------
//-----------------------------------------------------------------------------------------

//=======PACKAGES==========
import (
	"fmt"
	"time"
	d "../datatypes"
	"../network_go/bcast"
	"../net_fsm"
	u "../utilities"
	s "../settings"
)

//=======States==========
var id = ""

//======FUNCTIONS=======

//Starts order delegator system
func Run(
	netfsm_order_channel chan d.State_order_message,
	order_elev_ch_busypoll chan bool,
	order_elev_ch_neworder chan d.Order_struct,
	order_elev_ch_finished chan d.Order_struct,
	id_in string,
	){

	//Store id
	id =  id_in

	//Set up networking channels
	delegate_order_tx_chn := make(chan d.Network_delegate_order_message)
	delegate_order_rx_chn := make(chan d.Network_delegate_order_message)
	new_order_tx_chn := make(chan d.Network_new_order_message)
	new_order_rx_chn := make(chan d.Network_new_order_message)

	//Activate bcast library functions
	go bcast.Transmitter(s.DELEGATE_ORDER_PORT, delegate_order_tx_chn)
	go bcast.Receiver(s.DELEGATE_ORDER_PORT, delegate_order_rx_chn)
	go bcast.Transmitter(s.NEW_ORDER_PORT, new_order_tx_chn)
	go bcast.Receiver(s.NEW_ORDER_PORT, new_order_rx_chn)

	//Start operation
	for {
		select {

		//-----------------Receive order from net
		case msg := <-delegate_order_rx_chn:

			//Simulates packetloss
			if (u.PacketLossSim(s.DELEGATE_ORDER_PACKET_LOSS_SIM_CHANCE)) { continue }

	 		//Filters out ACK and NACK messages and checks if order is for this elevator
			if msg.ACK || msg.NACK || msg.Id_slave != id { continue }

			//Otherwise we ask elevator if it can execute
			execution_state := false
			execution_state = give_order_to_local_elevator(order_elev_ch_neworder, order_elev_ch_busypoll, msg.Order)

			//Sends ACK
			msg.ACK = execution_state
			msg.NACK = !execution_state
			delegate_order_tx_chn <- msg


		//-----------------Receive order from netFSM
		case msg := <-netfsm_order_channel:

			msg.ACK = send_order(delegate_order_tx_chn, //Tells slave to execute order
													delegate_order_rx_chn,	//Save true if the order is executed
													msg,
													order_elev_ch_neworder,
													order_elev_ch_busypoll)

			netfsm_order_channel <- msg //Sends result to network statemachine


		//-----------------Receive update from elevator. It has a new order to transmit to master
	case new_order := <-order_elev_ch_neworder:
			if  net_fsm.Check_master_state(){
				netfsm_order_channel <- d.State_order_message{new_order, "", false}
			} else if net_fsm.Get_number_of_peers() > 1{
				transmit_order_to_master(new_order, new_order_tx_chn, new_order_rx_chn)
			} else {
				fmt.Println("Order Delegator: Alone on network, no Network-action allowed")
			}


		//-----------------Receive new order update
		case message := <-new_order_rx_chn:

			//Simulates packetloss
			if (u.PacketLossSim(s.NEW_ORDER_PACKET_LOSS_SIM_CHANCE)) { continue }

			//Filters out ACK messages
			if (message.ACK) { continue }

			//Send to  net_fsm
			if ( net_fsm.Check_master_state()){
				netfsm_order_channel <- d.State_order_message{message.Order, "", false}
				new_order_tx_chn <- d.Network_new_order_message{message.Order,true}
				fmt.Printf("Order delegator: Confirming order\n")
			}


		//-----------------Receive update from elevator. The elevator has finished an order, notify master
		case finished_order := <-order_elev_ch_finished:
			if net_fsm.Check_master_state(){
				netfsm_order_channel <- d.State_order_message{finished_order, "", false}
			}	else {
				transmit_order_to_master(finished_order, new_order_tx_chn, new_order_rx_chn)
			}

		}
	}
}

//Sends order to slave and ensures ACK, Returns true if slave executes
func send_order(delegate_order_tx_chn chan d.Network_delegate_order_message,
	delegate_order_rx_chn chan d.Network_delegate_order_message,
	netfsm_msg d.State_order_message,
	order_elev_ch_neworder chan d.Order_struct,
	order_elev_ch_busypoll chan bool,
	) bool {

	if netfsm_msg.Id_slave == id { //If we ask ourself, we give order to our elevator
		return give_order_to_local_elevator(order_elev_ch_neworder, order_elev_ch_busypoll, netfsm_msg.Order)
	}

	//If not we broadcast order on network
	timeout_count := 0
	for {

		//Setting up timout signal
		timeOUT := time.NewTimer(time.Millisecond * s.SEND_ORDER_TIMEOUT_DELAY)

		//Send order
		delegate_order_tx_chn <- d.Network_delegate_order_message{netfsm_msg.Order, netfsm_msg.Id_slave, false, false}

		select {
		case message := <-delegate_order_rx_chn: //Receive ACK
			if message.ACK && message.Id_slave == netfsm_msg.Id_slave {
				return true
			} else if message.NACK && message.Id_slave == netfsm_msg.Id_slave {
				return false
			}

		case <-timeOUT.C: //If we time out we return false

			timeout_count += 1
			fmt.Printf("Order delegator: send_order() timed out.. %d\n", timeout_count)
			if (timeout_count > s.SEND_ORDER_MAX_TIMEOUT_COUNT) { return false }
		}
	}
}

//Asks elevator if it can execute order, and returns true if order is executed
func give_order_to_local_elevator(
	order_elev_ch_neworder chan d.Order_struct,
	order_elev_ch_busypoll chan bool,
	order d.Order_struct) bool{

	//Poll elevator if it can execute order
	busystate := false
	order_elev_ch_busypoll <- true
	select {
	case busystate = <-order_elev_ch_busypoll:
	}

	//If the elevator is not busy we give it the new order
	if !busystate{ order_elev_ch_neworder <- order }

	return !busystate
}

//Transmits order via network to master elevator, uses ACK messages to ensure arrival of order
func transmit_order_to_master(order d.Order_struct,
										tx_chn chan d.Network_new_order_message,
										rx_chn chan d.Network_new_order_message){

	fmt.Printf("Order delegator: Sending order to master: ")
	finished := false

	//Tries to send, several times if we do not receive confirmation
	for i := 0; i < s.TRANSMIT_ORDER_TO_MASTER_MAX_TIMOUT_COUNT; i++{

		//Sending sync-message to target
		tx_chn <- d.Network_new_order_message{order,false}

		//Setting up timout signal
		timeOUT := time.NewTimer(time.Millisecond * s.TRANSMIT_ORDER_TO_MASTER_TIMEOUT_DELAY)

		//Waiting for response
		for{

			resend := false

			select{

			case msg := <- rx_chn: //Check ack message
				if (msg.ACK){
					fmt.Printf(" [TRANSMITTED]\n")
					finished = true
				}

			case <-timeOUT.C: //Message timed out
				fmt.Printf("X, ")
				resend = true

			}
			if resend || finished { break }
		}
		if finished { break }
	}
	if (!finished) { fmt.Printf(" [ORDER LOST]\n") }
}
