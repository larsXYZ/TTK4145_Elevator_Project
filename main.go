package main

import(
  "./statemachine"
  //"./network/bcast"
  "./network/localip"
  "./network/peers"
  //"time"
  "fmt"
  "os"
  "flag"
)

type heartbeat struct{
  Id string
  Ip string
}

func main(){

  //Determines port number
  var port int
  flag.IntVar(&port,"port",15000,"portnumber")
  flag.Parse()

  //Initializes channels
  floor_sensors_channel := make(chan int)
  //heartbeat_tx_channel  := make(chan heartbeat)
  //heartbeat_rx_channel  := make(chan heartbeat)
  peers_tx_channel      := make(chan bool)
  peers_rx_channel      := make(chan peers.PeerUpdate)

  //Finds ip & id
  localIp, _ := localip.LocalIP()
  var id string = fmt.Sprintf("%s - %d",localIp, os.Getpid())

  //Runs statemachine
  go statemachine.Init(floor_sensors_channel)

  //Start peer system
  go peers.Transmitter(port, id, peers_tx_channel)
  go peers.Receiver(port, peers_rx_channel)

  //Spam broadcast
  /*go bcast.Transmitter(15647, heartbeat_tx_channel)
  go bcast.Receiver(15647, heartbeat_rx_channel)
  go func() {
		message_tx := heartbeat{Id: id, Ip: localIp}
		for {
			heartbeat_tx_channel <- message_tx
			time.Sleep(2 * time.Second)
		}
	}()*/



  //Look at pipes
  for{
    select{
    /*case message_rx := <-heartbeat_rx_channel:
      if (message_rx.Id != id){
        fmt.Println(message_rx)
      }*/

    case p := <-peers_rx_channel:
      fmt.Printf("  Peers:    %q\n", p.Peers)

    }
  }

}
