package network_order_handler

//-----------------------------------------------------------------------------------------
//---------------Sends/Receives orders to/from other elevators-----------------------------
//-----------------------------------------------------------------------------------------

import (
	"fmt"
	"time"
	d "../datatypes"
	"../network_go/bcast"
)

//States
var id = ""

//====== FUNCTIONS =======

//Starts order handler system
func Run(
	netstate_order_channel chan d.State_order_message,
	order_elev_ch_busypoll chan bool,
	order_elev_ch_neworder chan d.Order_struct,
	order_elev_ch_finished chan d.Order_struct,
	elev_id string,
	){

	//Store id
	id = elev_id

	//Set up networking channels
	delegate_order_tx_chn := make(chan d.Network_order_message, 100)
	delegate_order_rx_chn := make(chan d.Network_order_message, 100)
	new_order_tx_chn := make(chan d.Order_struct, 100)
	new_order_rx_chn := make(chan d.Order_struct, 100)

	//Activate bcast library functions
	go bcast.Transmitter(14002, delegate_order_tx_chn)
	go bcast.Receiver(14002, delegate_order_rx_chn)
	go bcast.Transmitter(14003, new_order_tx_chn)
	go bcast.Receiver(14003, new_order_rx_chn)

	//Start operation
	for {
		select {

		case net_message := <-delegate_order_rx_chn: //Receive order from net
			give_order_to_elevator(order_elev_ch_neworder, order_elev_ch_busypoll, delegate_order_tx_chn, net_message)

		case netstate_message := <-netstate_order_channel: //Receive order from netstatemachine
			netstate_message.ACK = send_order(delegate_order_tx_chn, //Tells slave to execute order
																				delegate_order_rx_chn,
																				netstate_message,
																				order_elev_ch_neworder,
																				order_elev_ch_busypoll)
			netstate_order_channel <- netstate_message //Sends result to network statemachine

		case order := <-order_elev_ch_neworder: //Receive update from elevator. Finished or new order
			new_order_tx_chn <- order

		case order := <-order_elev_ch_finished: //Receive update from elevator. Finished or new order
			new_order_tx_chn <- order

		case new_order := <-new_order_rx_chn: //Receive new order update
			netstate_order_channel <- d.State_order_message{new_order, "", false} //Send to network_statemachine
		}
	}
}

//Sends order to slave and ensures ACK
//Returns true if slave executes
func send_order(order_tx_chn chan d.Network_order_message,
	order_rx_chn chan d.Network_order_message,
	netstate_message d.State_order_message,
	order_elev_ch_neworder chan d.Order_struct,
	order_elev_ch_busypoll chan bool,
	) bool {

	if netstate_message.Id_slave == id { //If we ask ourself, we poll elevator
		busystate := false
		order_elev_ch_busypoll <- true
		select {
		case busystate =  <-order_elev_ch_busypoll:
		}

		if !busystate{
			order_elev_ch_neworder <- netstate_message.Order
		}

		return !busystate
	}

	//Broadcasts order and wait for ACK
	for {

		//Setting up timout signal
		timeOUT := time.NewTimer(time.Millisecond * 300)

		//Send order
		order_tx_chn <- d.Network_order_message{netstate_message.Order, netstate_message.Id_slave, false, false}

		//Wait for response
		for {
			select {
			case message := <-order_rx_chn: //Receive ACK
				if message.ACK && message.Id_slave == netstate_message.Id_slave {
					return true
				} else if message.NACK && message.Id_slave == netstate_message.Id_slave {
					return false
				}

			case <-timeOUT.C: //If we time out we
				fmt.Printf("|SYNC TIMEOUT| ")
				return false
			}
		}
	}
}

//Asks elevator if it can execute order, and gives result back
func give_order_to_elevator(
	order_elev_ch_neworder chan d.Order_struct,
	order_elev_ch_busypoll chan bool,
	delegate_order_tx_chn chan d.Network_order_message,
	net_message d.Network_order_message){

	if !net_message.ACK && !net_message.NACK { //Filters out ACK and NACK messages

		if net_message.Id_slave == id { //Check if the message is for us

			fmt.Printf("Order handler: Order for local elevator received: ")

			// ASK ELEVATOR STATEMACHINE IF IT CAN EXECUTE ORDER NOW
			busystate := false
			order_elev_ch_busypoll <- true
			select {
			case busystate = <-order_elev_ch_busypoll:
			}

			if !busystate{
				order_elev_ch_neworder <- net_message.Order
			}

			if busystate {
				fmt.Printf("[BUSY]\n\n")
			} else {
				fmt.Printf("[EXECUTES]\n\n")
			}

			//Sends ACK
			net_message.ACK = !busystate
			net_message.NACK = busystate
			delegate_order_tx_chn <- net_message
		}
	}
}
