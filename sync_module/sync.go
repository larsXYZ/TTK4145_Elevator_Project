package sync

//-----------------------------------------------------------------------------------------
//---------------Sync information with other elevators-------------------------------------
//-----------------------------------------------------------------------------------------

import(
  "./../network_go/bcast"
  d "./../datatypes"
  "fmt"
  //"time"
)

//States
var id = ""

//Runs the sync module
func Run(state_sync_channel chan d.State_sync_message, id_in string){

  id = id_in

  //Set up Channels
  sync_tx_chn := make(chan d.Network_sync_message,1)
  sync_rx_chn := make(chan d.Network_sync_message,1)

  //Activate bcast library functions
  go bcast.Transmitter(16569, sync_tx_chn)
  go bcast.Receiver(16569, sync_rx_chn)

  fmt.Println("Sync module: Listening started")
  for{ //Handles messages over network and commands from network statemachine
    select{

    //Responds to Network_message
  case message := <-sync_rx_chn:
      if message.Sender != id{
        network_sync_handler(sync_tx_chn, sync_rx_chn, state_sync_channel, message)
      }

    case command := <-state_sync_channel:
      fmt.Println("Sync module: Command Received from NetworkStatemachine")
      command_handler(sync_tx_chn, sync_rx_chn, state_sync_channel, command)
    }
  }

}

//Synchronizes state with other elevators on network
func sync_state(sync_tx_chn chan d.Network_sync_message, sync_rx_chn chan d.Network_sync_message, state d.State){

  //Creates Network_sync_message
  sync_message := d.Network_sync_message{state, false, id}

  //Broadcasts state 10 times for now
  for i := 0; i < 10; i++{
    sync_tx_chn <- sync_message
  }
}

//Handles received network messages
func network_sync_handler(tx_chn chan d.Network_sync_message, rx_chn chan d.Network_sync_message, state_sync_channel chan d.State_sync_message, m d.Network_sync_message){
  fmt.Println("Sync module: Sync message received")
}

//Handles commands from network statemachine
func command_handler(sync_tx_chn chan d.Network_sync_message, sync_rx_chn chan d.Network_sync_message, state_sync_channel chan d.State_sync_message, ms d.State_sync_message){
  sync_state(sync_tx_chn, sync_rx_chn, ms.Test_state)
}
