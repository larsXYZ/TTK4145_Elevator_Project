package network_order_handler

//-----------------------------------------------------------------------------------------
//---------------Sends/Receives orders to/from other elevators-----------------------------
//-----------------------------------------------------------------------------------------

import(
  d "./../datatypes"
  "./../network_go/bcast"
  "fmt"
)

//States
var id = ""


//====== FUNCTIONS =======

func Run(netstate_order_channel chan d.State_order_message, id_in string ){ //Starts order handler system

  //Store id
  id = id_in

  //Set up Channels
  order_tx_chn := make(chan d.Network_order_message,1)
  order_rx_chn := make(chan d.Network_order_message,1)

  //Activate bcast library functions
  go bcast.Transmitter(14002, order_tx_chn)
  go bcast.Receiver(14002, order_rx_chn)

  //Start operation
  for{
    select{

        case net_message := <- order_rx_chn: //Receive order from net
          fmt.Println("RECEIVED ORDER FROM NETWORK")
          fmt.Println(net_message)

        case netstate_message := <- netstate_order_channel: //Receive order from netstatemachine
          fmt.Println("RECEIVED ORDER FROM NETSTATE")
          order_tx_chn <- d.Network_order_message{netstate_message.Order, netstate_message.Id_slave}


    }
  }
}
