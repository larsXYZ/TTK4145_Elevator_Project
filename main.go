package main

//------------------------------------------------------------------------------------
//--------------- Starts all the other processes in the elevator system---------------
//------------------------------------------------------------------------------------

import (
	d "./datatypes"
	ns "./network_statemachine"
	"flag"
)

func main() {

	//Determines port number
	portPtr := flag.Int("port", 15000, "an int")
	flag.Parse()

	//Initializes channels
	netstate_elevstate_channel := make(chan d.State_elev_message)

	//Runs network statemachine
	go ns.Run(netstate_elevstate_channel, *portPtr)

	//Waits
	select {}

}
