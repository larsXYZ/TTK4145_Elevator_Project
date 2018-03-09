package main

import(
  "./statemachine"
  "./network/bcast"
  //"./network/localip"
  //"./network/peers"
  "time"
  "fmt"
  "os"
)

func main(){

  //Initializes channels
  floor_sensors_channel := make(chan int)
  heartbeat_tx_channel  := make(chan string)
  heartbeat_rx_channel  := make(chan string)

  //Runs statemachine
  go statemachine.Init(floor_sensors_channel)

  //Spam broadcast
  go bcast.Transmitter(16569, heartbeat_tx_channel)
  go bcast.Receiver(16569, heartbeat_rx_channel)
  go func() {
		heartbeat := fmt.Sprintf("Heartbeat signal from: %d", os.Getpid())
		for {
			heartbeat_tx_channel <- heartbeat
			time.Sleep(1 * time.Second)
		}
	}()

  //Scan for heartbeat
  for{
    select{
    case heartbeat := <-heartbeat_rx_channel:
      fmt.Println(heartbeat)
    }
  }

}
