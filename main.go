package main

//------------------------------------------------------------------------------------
//--------------- Starts all the other processes in the elevator system---------------
//------------------------------------------------------------------------------------

import (
	d "./datatypes"
	"./network_statemachine"
	"./network_go/localip"
	"./sync_module"
	"flag"
	"os"
	"fmt"
	u "./utilities"
	"./elevator_statemachine"
	"./network_order_handler"
	"time"
)

func main() {

	//Determines port number
	portPtr := flag.Int("port", 15000, "an int")
	elevPortPtr := flag.Int("elevport", 15657, "the port of the elevator sim")
	flag.Parse()

	//Create sim-elevator ip
	elevSimIp := fmt.Sprintf("localhost:%d",*elevPortPtr)

	//Determine ip & id
	localIp, _ := localip.LocalIP()
	id := fmt.Sprintf("%s%d", u.IpToString(localIp), os.Getpid())
	fmt.Printf("ELEVATOR ID: %s\n",id)

	//Initializes channels
	netstate_elevstate_channel 	:= make(chan d.State_elev_message)
	netstate_sync_channel				:= make(chan d.State_sync_message)
	netstate_elev_channel 			:= make(chan d.State_elev_message)
	netstate_order_channel			:= make(chan d.State_order_message)
	netorder_elev_channel				:= make(chan d.Order_elev_message)

	fmt.Println("-----Activating Modules-----")

	//Runs interface module
	go elevator_statemachine.Run(netstate_elev_channel,netorder_elev_channel, elevSimIp)
	//Waiting for elevator to find floor
	time.Sleep(2*time.Second)

	//Run order handler module
	go network_order_handler.Run(netstate_order_channel,netorder_elev_channel, id)

	//Runs sync module
	go sync.Run(netstate_sync_channel, id)

	//Runs network statemachine
	go network_statemachine.Run(netstate_elevstate_channel, netstate_sync_channel, netstate_order_channel, *portPtr, id)


	//Waits
	select {}
}
