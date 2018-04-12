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
  "strings"
)

//States
var id = ""

//Runs the sync module
func Run(netfsm_sync_ch_command chan d.State_sync_message,
        netfsm_sync_ch_error chan bool,
        id_in string){

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
      network_sync_handler(sync_tx_chn, sync_rx_chn, netfsm_sync_ch_command, sync_message)

    //Synchronizes state
  case command := <-netfsm_sync_ch_command:

      //Synchronize state, if we fail we request resync
      if (!sync_state(sync_tx_chn, sync_rx_chn, command, netfsm_sync_ch_command)){
        netfsm_sync_ch_error <- true
      }
    }
  }

}

//Synchronizes state with other elevators on network
func sync_state(sync_tx_chn chan d.Network_sync_message,
                sync_rx_chn chan d.Network_sync_message,
                command d.State_sync_message,
                netfsm_sync_ch_command chan d.State_sync_message) bool{

  fmt.Printf("Sync module: Syncing state: ")

  //Converting connected elevator string to list
  PeersList := strings.Split(command.Peers, ",")

  //Starting synchronization
  for i := 0; i < len(PeersList); i++{

    targetId := PeersList[i]

    if (targetId == id){ //No need to sync with ourself
      continue
    }

    timeout_count := 0

    for{

      finished := false

      //Sending sync-message to target
      sync_tx_chn <- d.Network_sync_message{command.State, false, id, targetId}

      //Setting up timout signal
      timeOUT := time.NewTimer(time.Millisecond * 50)

      //Waiting for response
      for{

        resend := false

        select{

        case msg := <- sync_rx_chn: //Check ack message
          if (msg.Sender == targetId && msg.SyncAck && msg.Target == id){
            fmt.Printf("%d,",i)
            finished = true
          }

        case <-timeOUT.C: //Message timed out
          fmt.Printf("X, ")
          resend = true
          timeout_count += 1

        }

        if resend || finished { break }
      }

      if finished { break }

      if timeout_count > 5{
        fmt.Printf(": [FAILED]\n")
        return false
      }
    }
  }

  fmt.Printf(": [COMPLETED]\n")
  return true

}

//Handles received network messages
func network_sync_handler(tx_chn chan d.Network_sync_message,
                          rx_chn chan d.Network_sync_message,
                          netfsm_sync_ch_command chan d.State_sync_message,
                          m d.Network_sync_message){

if u.PacketLossSim(25){ return }

  if m.Sender != id && !m.SyncAck && m.Target == id{ //Ignores messages sent by ourself and ACK messages

    fmt.Println("Sync module: State update received, sending ACK\n")
    netfsm_sync_ch_command <- d.State_sync_message{m.State,""}
    tx_chn <- d.Network_sync_message{d.State_init(),true, id, m.Sender}
  }
}
