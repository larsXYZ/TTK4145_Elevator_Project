package network_order_handler

//-----------------------------------------------------------------------------------------
//---------------Sends/Receives orders to/from other elevators-----------------------------
//-----------------------------------------------------------------------------------------

import(
  d "./../datatypes"
  "./../network_go/bcast"
  "fmt"
  "time"
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
          if (!net_message.ACK && !net_message.NACK){ //Filters out ACK and NACK messages
            fmt.Println("RECEIVED ORDER FROM NETWORK")
            if (net_message.Id_slave == id){

              // ASK ELEVATOR STATEMACHINE IF IT CAN EXECUTE ORDER NOW
              // RESPOND ACK OR NACK ACCORDING TO RESPONSE FROM ELEVATOR STATEMACHINE

              net_message.ACK = true
              net_message.NACK = false
              order_tx_chn <- net_message
              fmt.Println("CORRECT ID, EXECUTING + ACK")
            } else {
              fmt.Println("WRONG ID")
            }
            fmt.Println("")
          }

        case netstate_message := <- netstate_order_channel: //Receive order from netstatemachine
          fmt.Println("RECEIVED ORDER FROM NETSTATE")
          send_order(order_tx_chn, order_rx_chn, netstate_message)
          fmt.Println("")
    }
  }
}

func send_order(order_tx_chn chan d.Network_order_message, order_rx_chn chan d.Network_order_message, netstate_message d.State_order_message) { //Sends order to slave and ensures ACK

  //Broadcasts order and wait for ACK
  for{
    //Setting up timout signal
    timeOUT := time.NewTimer(time.Millisecond * 200)

    //Send order
    fmt.Printf("Sending order to slave, id = %s\n", netstate_message.Id_slave)
    order_tx_chn <- d.Network_order_message{netstate_message.Order, netstate_message.Id_slave, false, false}

    //Wait for response
    for{
      select{
      case ack_mes := <- order_rx_chn: //Receive ACK
        if ack_mes.ACK{
          fmt.Println("ACK received, order handled.")
          return
        }

      case <-timeOUT.C: //If we do not get response withing a timelimit we resend
        fmt.Println("SYNC TIMEOUT, resending...")
        break
      }
    }
  }
}
