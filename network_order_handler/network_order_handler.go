package network_order_handler

//-----------------------------------------------------------------------------------------
//---------------Sends/Receives orders to/from other elevators-----------------------------
//-----------------------------------------------------------------------------------------

import (
	"fmt"
	"time"
	d "../datatypes"
	"../network_go/bcast"
	"../network_statemachine"
	u "../utilities"
)

//States
var id = ""

//====== FUNCTIONS =======

//Starts order handler system
func Run(
	netfsm_order_channel chan d.State_order_message,
	order_elev_ch_busypoll chan bool,
	order_elev_ch_neworder chan d.Order_struct,
	order_elev_ch_finished chan d.Order_struct,
	elev_id string,
	){

	//Store id
	id = elev_id

	//Set up networking channels
	delegate_order_tx_chn := make(chan d.Network_delegate_order_message, 100)
	delegate_order_rx_chn := make(chan d.Network_delegate_order_message, 100)
	new_order_tx_chn := make(chan d.Network_new_order_message, 100)
	new_order_rx_chn := make(chan d.Network_new_order_message, 100)

	//Activate bcast library functions
	go bcast.Transmitter(14002, delegate_order_tx_chn)
	go bcast.Receiver(14002, delegate_order_rx_chn)
	go bcast.Transmitter(14003, new_order_tx_chn)
	go bcast.Receiver(14003, new_order_rx_chn)

	//Start operation
	for {
		select {

		//-----------------Receive order from net
		case msg := <-delegate_order_rx_chn:

		//Simulates packetloss
		if (u.PacketLossSim(70)) { continue }

 		//Filters out ACK and NACK messages and checks if order is for this elevator
		if msg.ACK || msg.NACK || msg.Id_slave != id { continue }

		//Otherwise we ask elevator if it can execute
		execution_state := give_order_to_local_elevator(order_elev_ch_neworder, order_elev_ch_busypoll, msg.Order)

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
			if network_statemachine.Master_state{
				netfsm_order_channel <- d.State_order_message{new_order, "", false}
			} else {
				transmit_order(new_order, new_order_tx_chn, new_order_rx_chn)
			}


		//-----------------Receive new order update
		case message := <-new_order_rx_chn:

			//Simulates packetloss
			if (u.PacketLossSim(30)) { continue }

			//Filters out ACK messages
			if (message.ACK) { continue }

			//Send to network_statemachine
			if (network_statemachine.Master_state){
				netfsm_order_channel <- d.State_order_message{message.Order, "", false}
				new_order_tx_chn <- d.Network_new_order_message{message.Order,true}
				fmt.Printf("Order handler: Confirming order\n")
			}


		//-----------------Receive update from elevator. The elevator has finished an order, notify master
		case finished_order := <-order_elev_ch_finished:
			if network_statemachine.Master_state{
				netfsm_order_channel <- d.State_order_message{finished_order, "", false}
			}	else {
				transmit_order(finished_order, new_order_tx_chn, new_order_rx_chn)
			}

		}
	}
}

//Sends order to slave and ensures ACK
//Returns true if slave executes
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
		timeOUT := time.NewTimer(time.Millisecond * 50)

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
			fmt.Printf("Order handler: send_order() timed out.. %d\n", timeout_count)
			if (timeout_count > 5) { return false }
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

func transmit_order(order d.Order_struct, //Transmits new order to master
										tx_chn chan d.Network_new_order_message,
										rx_chn chan d.Network_new_order_message){

	fmt.Printf("Order handler: Sending order to master: ")
	finished := false

	//Tries to send, several times if we do not receive confirmation
	for i := 0; i < 5; i++{

		//Sending sync-message to target
		tx_chn <- d.Network_new_order_message{order,false}

		//Setting up timout signal
		timeOUT := time.NewTimer(time.Millisecond * 50)

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
