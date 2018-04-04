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

func Run(netstate_order_channel chan d.State_order_message, order_elev_channel chan d.Order_elev_message, id_in string ){ //Starts order handler system

  //Store id
  id = id_in

  //Set up Channels
  delegate_order_tx_chn := make(chan d.Network_order_message,1)
  delegate_order_rx_chn := make(chan d.Network_order_message,1)
  new_order_tx_chn := make(chan d.Order_struct,1)
  new_order_rx_chn := make(chan d.Order_struct,1)

  //Activate bcast library functions
  go bcast.Transmitter(14002, delegate_order_tx_chn)
  go bcast.Receiver(14002, delegate_order_rx_chn)
  go bcast.Transmitter(14003, new_order_tx_chn)
  go bcast.Receiver(14003, new_order_rx_chn)

  //Start operation
  for{
    select{

        case net_message := <- delegate_order_rx_chn: //Receive order from net
          if (!net_message.ACK && !net_message.NACK){ //Filters out ACK and NACK messages
            fmt.Println("RECEIVED ORDER FROM NETWORK")

            if (net_message.Id_slave == id){ //Check if the message is for us

              // ASK ELEVATOR STATEMACHINE IF IT CAN EXECUTE ORDER NOW
              busystate := false
              order_elev_channel <- d.Order_elev_message{d.Order_struct{},false}
              select{
              case response := <-order_elev_channel:
                fmt.Println("Response received")
                busystate = response.BusyState
              }

              net_message.ACK = !busystate
              net_message.NACK = busystate
              delegate_order_tx_chn <- net_message
              fmt.Println("CORRECT ID")
              if busystate{
                fmt.Println("BUSY")
              } else {
                fmt.Println("EXECUTED")
              }
            } else {
              fmt.Println("WRONG ID")
            }
            fmt.Println("")
          }

        case netstate_message := <- netstate_order_channel: //Receive order from netstatemachine
          execution_state := send_order(delegate_order_tx_chn, delegate_order_rx_chn, netstate_message) //Tells slave to execute order
          netstate_message.ACK = execution_state
          netstate_order_channel <-netstate_message                                   //Sends result to network statemachine

        case elevstate_message := <- order_elev_channel: //Receive order from elev state (button press)
          //Broadcast on network
          fmt.Println("BROADCAST ORDER TO MASTER")
          fmt.Println(elevstate_message)
          new_order_tx_chn <- elevstate_message.Order

        case new_order := <-new_order_rx_chn: //Receive new order update
          //Send to network_statemachine
          fmt.Println("RECEIVED NEW ORDER FROM NET ")
          netstate_order_channel <- d.State_order_message{new_order, "", false}



    }
  }
}

func send_order(order_tx_chn chan d.Network_order_message, order_rx_chn chan d.Network_order_message, netstate_message d.State_order_message) bool{ //Sends order to slave and ensures ACK
                                                                                                                                                    //Returns true if slave executes
  //Broadcasts order and wait for ACK
  for{

    //Setting up timout signal
    timeOUT := time.NewTimer(time.Millisecond * 300)

    //Send order
    order_tx_chn <- d.Network_order_message{netstate_message.Order, netstate_message.Id_slave, false, false}


    //Wait for response
    for{
      select{
      case ack_mes := <- order_rx_chn: //Receive ACK
        if ack_mes.ACK && ack_mes.Id_slave == netstate_message.Id_slave{
          fmt.Println("ACK received, order handled.")
          return true
        } else if (ack_mes.NACK && ack_mes.Id_slave == netstate_message.Id_slave){
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
