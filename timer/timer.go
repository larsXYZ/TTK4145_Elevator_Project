package timer

//-----------------------------------------------------------------------------------------
//---------------Generates timer interrupts on channel, can be turned on and of-------------------------------------
//-----------------------------------------------------------------------------------------

import(
  "time"
)

//States
var enabled = false

func Run(timing_channel chan bool){ //Runs interrupts and respons to toggle message

  //Starts internal timer
  ticker := time.NewTicker(time.Millisecond * 1000)

  for{
    select{

    case <-ticker.C: //Sends timer signal if enabled
      if enabled {
        timing_channel <- true
      }

    case state :=  <- timing_channel: //Turns on/off
      enabled = state
    }
  }




}
