package main

//------------------------------------------------------------------------------------
//--------------- Starts all the other processes in the elevator system---------------
//------------------------------------------------------------------------------------

import (
	d "./datatypes"
	"./network_statemachine"
	"./sync_module"
	"flag"
)

func main() {

	//Determines port number
	portPtr := flag.Int("port", 15000, "an int")
	flag.Parse()

	//Initializes channels
	netstate_elevstate_channel 	:= make(chan d.State_elev_message)
	netstate_sync_channel				:= make(chan d.State_sync_message)

	//Runs sync module
	go sync.Run(netstate_sync_channel)

	//Runs network statemachine
	go network_statemachine.Run(netstate_elevstate_channel, netstate_sync_channel, *portPtr)


	//Waits
	select {}
}
