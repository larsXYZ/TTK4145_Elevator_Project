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

//Runs the sync module
func Run(state_sync_channel chan d.State_sync_message){

  //Set up Channels
  rx_chn := make(chan d.Network_message,1)
  tx_chn := make(chan d.Network_message,1)

  //Activate bcast library functions
  go bcast.Transmitter(16569, tx_chn)
  go bcast.Receiver(16569, rx_chn)

  fmt.Println("Sync module: Listening started")
  for{ //Handles messages over network and commands from network statemachine
    select{

    //Responds to Network_message
    case message := <-rx_chn:
      fmt.Println("Sync module: Message Received from Network")
      network_message_handler(tx_chn, rx_chn, state_sync_channel, message)

    case command := <-state_sync_channel:
      fmt.Println("Sync module: Command Received from NetworkStatemachine")
      command_handler(tx_chn, rx_chn, state_sync_channel, command)
    }
  }

}
/*
//Detects if helicopter is alone on network by sending numerous UDP messages
func presence_check(tx_chn chan d.Network_message, rx_chn chan d.Network_message, state_sync_channel chan d.State_sync_message){

  fmt.Println("Presence check started")

  //Send UDP greeting message and listen for response
  for i := 0; i < 5; i++{

    //Sends greeting message
    fmt.Printf("    ...%d",i)
    tx_chn <- d.Network_message{true,false}

    //Creating timeout channel and function
    timer := time.NewTimer(time.Millisecond * 100)

    //Listens for response
    for{
      norespone := false

      select{
      case <- timer.C:{ //Break if we timout
        fmt.Printf(".. No response\n")
        norespone = true
      }

      case message := <-rx_chn:{ //Handle received greeting message
        if message.Greeting_response{
          fmt.Printf(".. Response received\n")
          fmt.Println("Presence check ended")
          state_sync_channel<-d.State_sync_message{false,true} //We have gotten response
          return
        } else{ //If the message received was not greeting response, we continue waiting. This filters out our own message sent earlier.
          continue
        }
      }
      }
      if norespone{break} //Breaks if we receive no response
    }

    timer.Stop()
  }

  fmt.Println("\nPresence check ended")
  state_sync_channel<-d.State_sync_message{false,false} //We have gotten response

}
*/
//Handles received network messages
func network_message_handler(tx_chn chan d.Network_message, rx_chn chan d.Network_message, state_sync_channel chan d.State_sync_message, m d.Network_message){

  if m.Greeting{ //If it is greeting message, respond
    fmt.Println("Greeting received, responding..")
    tx_chn <- d.Network_message{false,true}
  }
}

//Handles commands from network statemachine
func command_handler(tx_chn chan d.Network_message, rx_chn chan d.Network_message, state_sync_channel chan d.State_sync_message, c d.State_sync_message){


}
