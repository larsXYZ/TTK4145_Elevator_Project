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
	"./elevator_interface"
)

func main() {

	//Determines port number
	portPtr := flag.Int("port", 15000, "an int")
	flag.Parse()

	//Determine ip & id
	localIp, _ := localip.LocalIP()
	id := fmt.Sprintf("%s%d", u.IpToString(localIp), os.Getpid())
	fmt.Printf("ELEVATOR ID: %s\n",id)

	//Initializes channels
	netstate_elevstate_channel 	:= make(chan d.State_elev_message)
	netstate_sync_channel				:= make(chan d.State_sync_message)
	state_elev_channel 					:= make(chan d.State_elev_message)


	//Runs interface module
	go elevator_interface.Run(state_elev_channel)

	//Runs sync module
	go sync.Run(netstate_sync_channel, id)

	//Runs network statemachine
	go network_statemachine.Run(netstate_elevstate_channel, netstate_sync_channel, *portPtr, id)


	//Waits
	select {}
}
