package main

//------------------------------------------------------------------------------------
//--------------- Starts all the other processes in the elevator system---------------
//------------------------------------------------------------------------------------

import (
	d "./datatypes"
	"./network"
	"./statemachine"
	//"time"
	"flag"
	//	"fmt"
	//	"os"
)

func main() {

	//Determines port number
	var port int
	flag.IntVar(&port, "port", 15000, "portnumber")

	//Initializes channels
	state_elev_channel := make(chan d.State_elev_message)
	state_network_channel := make(chan d.State_network_message)

	//Runs statemachine
	go statemachine.Init(state_elev_channel)

	//Runs network module
	go network.Init(state_network_channel, port)

	select {}

}
