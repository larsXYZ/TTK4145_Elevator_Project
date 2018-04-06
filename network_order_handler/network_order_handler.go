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

func Run(netstate_order_channel chan d.State_order_message, order_elev_channel chan d.Order_elev_message, id_in string) { //Starts order handler system

	//Store id
	id = id_in

	//Set up Channels
	delegate_order_tx_chn := make(chan d.Network_order_message, 1)
	delegate_order_rx_chn := make(chan d.Network_order_message, 1)
	new_order_tx_chn := make(chan d.Order_struct, 1)
	new_order_rx_chn := make(chan d.Order_struct, 1)

	//Activate bcast library functions
	go bcast.Transmitter(14002, delegate_order_tx_chn)
	go bcast.Receiver(14002, delegate_order_rx_chn)
	go bcast.Transmitter(14003, new_order_tx_chn)
	go bcast.Receiver(14003, new_order_rx_chn)

	//Start operation
	for {
		select {

		case net_message := <-delegate_order_rx_chn: //Receive order from net
			give_order_to_elevator(order_elev_channel, delegate_order_tx_chn, net_message)

		case netstate_message := <-netstate_order_channel: //Receive order from netstatemachine
			execution_state := send_order(delegate_order_tx_chn, delegate_order_rx_chn, netstate_message, order_elev_channel) //Tells slave to execute order
			netstate_message.ACK = execution_state
			netstate_order_channel <- netstate_message //Sends result to network statemachine

		case elevstate_message := <-order_elev_channel: //Receive order from elev state (button press)
			//Broadcast on network
			fmt.Println("BROADCAST NEW ORDER TO MASTER")
			fmt.Println(elevstate_message)
			new_order_tx_chn <- elevstate_message.Order

		case new_order := <-new_order_rx_chn: //Receive new order update
			//Send to network_statemachine
			netstate_order_channel <- d.State_order_message{new_order, "", false}

		}
	}
}

//Sends order to slave and ensures ACK
//Returns true if slave executes
func send_order(order_tx_chn chan d.Network_order_message, order_rx_chn chan d.Network_order_message, netstate_message d.State_order_message, order_elev_channel chan d.Order_elev_message) bool {


	if netstate_message.Id_slave == id { //If we ask ourself, we poll elevator
		busystate := false
		order_elev_channel <- d.Order_elev_message{netstate_message.Order, false}
		select {
		case response := <-order_elev_channel:
			busystate = response.BusyState
		}
		if busystate {
			fmt.Println("NACK received, slave busy")
		} else {
			fmt.Println("ACK received, order handled.")
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
					fmt.Println("ACK received, order handled.")
					return true
				} else if !message.ACK && message.Id_slave == netstate_message.Id_slave {
					fmt.Println("NACK received, slave busy")
					return false
				}

			case <-timeOUT.C: //If we do not get response withing a timelimit we resend
				fmt.Println("SYNC TIMEOUT")
				return false
			}
		}
	}
}

func give_order_to_elevator(order_elev_channel chan d.Order_elev_message, delegate_order_tx_chn chan d.Network_order_message, net_message d.Network_order_message) { //Asks elevator if it can execute order, and gives result back
	if !net_message.ACK && !net_message.NACK { //Filters out ACK and NACK messages
		fmt.Println("RECEIVED ORDER FROM NETWORK")

		if net_message.Id_slave == id { //Check if the message is for us

			// ASK ELEVATOR STATEMACHINE IF IT CAN EXECUTE ORDER NOW
			busystate := false
			order_elev_channel <- d.Order_elev_message{net_message.Order, false}
			select {
			case response := <-order_elev_channel:
				fmt.Println("Response received")
				busystate = response.BusyState
			}

			net_message.ACK = !busystate
			net_message.NACK = busystate
			delegate_order_tx_chn <- net_message

		}
	}
}
