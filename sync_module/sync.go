package sync

//-----------------------------------------------------------------------------------------
//---------------Sync information with other elevators-------------------------------------
//-----------------------------------------------------------------------------------------

import(
  "./../network_go/bcast"
  d "./../datatypes"
  "fmt"
  "time"
  u "./../utilities"
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
    case sync_message := <-sync_rx_chn:
      network_sync_handler(sync_tx_chn, sync_rx_chn, state_sync_channel, sync_message)

    case command := <-state_sync_channel:
      fmt.Println("Sync module: Command Received from NetworkStatemachine")
      command_handler(sync_tx_chn, sync_rx_chn, state_sync_channel, command)
    }
  }

}

//Synchronizes state with other elevators on network
func sync_state(sync_tx_chn chan d.Network_sync_message, sync_rx_chn chan d.Network_sync_message, command d.State_sync_message, state_sync_channel chan d.State_sync_message){


  //If this is only elevator, it doesnt need to sync state
  if command.Connected_count == 1{
    fmt.Printf("This is the only elevator, not need to sync state\n\n")
    return
  }

  //Creates Network_sync_message
  sync_message := d.Network_sync_message{command.SyncState, false, id}

  //Broadcasts state and wait for ACK
  for{

    //Array to keep track of elevators which have ACK-ed
    ack_elevators := make([]string,0)

    //Setting up timout signal
    timeOUT := time.NewTimer(time.Millisecond * 200)

    sync_tx_chn <- sync_message //Broadcast state
    for{
      select{
      case ack_mes := <- sync_rx_chn: //Receive ACK
        if !u.IdInArray(ack_mes.Sender,ack_elevators) && ack_mes.SyncAck{
          ack_elevators = append(ack_elevators,ack_mes.Sender) //Adds it to the list
          fmt.Printf("ACK received, (%d of %d)\n", len(ack_elevators), command.Connected_count-1)
          fmt.Printf(" -> %q\n", ack_elevators)
        }

        if len(ack_elevators) >= command.Connected_count-1{ //If all elevators ack we are finished
          fmt.Printf("Sync completed, all %d elevators ACK\n\n", command.Connected_count-1)
          return
        }

      case <-timeOUT.C: //If we do not get response withing a timelimit we resend
        fmt.Println("SYNC TIMEOUT, resending...")
        break;

      case command :=<-state_sync_channel: //If there is a network change we try another elevator
          sync_state(sync_tx_chn, sync_rx_chn, command, state_sync_channel)
          return
      }
    }
  }
}

//Handles received network messages
func network_sync_handler(tx_chn chan d.Network_sync_message, rx_chn chan d.Network_sync_message, state_sync_channel chan d.State_sync_message, m d.Network_sync_message){

  if m.Sender != id && !m.SyncAck{ //Ignores messages sent by ourself and ACK messages
    state_sync_channel <- d.State_sync_message{m.SyncState,0}

    //Send ACK
    fmt.Printf("State update received, sending ACK\n")
    tx_chn <- d.Network_sync_message{d.State_init(),true,id}
  }
}

//Handles commands from network statemachine
func command_handler(sync_tx_chn chan d.Network_sync_message, sync_rx_chn chan d.Network_sync_message, state_sync_channel chan d.State_sync_message, ms d.State_sync_message){
  sync_state(sync_tx_chn, sync_rx_chn, ms, state_sync_channel)
}
