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
	netstate_sync_channel						:= make(chan d.State_sync_message,100)
	netstate_elev_channel 					:= make(chan d.State_elev_message,100)

	netstate_order_channel					:= make(chan d.State_order_message,100)
	order_elev_ch_busypoll					:= make(chan bool ,100)
	order_elev_ch_neworder					:= make(chan d.Order_struct,100)
	order_elev_ch_finished					:= make(chan d.Order_struct,100)

	fmt.Println("-----Activating Modules-----")

	//Runs interface module
	go elevator_statemachine.Run(
		netstate_elev_channel,
		order_elev_ch_busypoll,
		order_elev_ch_neworder,
		order_elev_ch_finished,
		elevSimIp)

	//Waiting for elevator to find floor
	time.Sleep(5*time.Second)

	//Run order handler module
	go network_order_handler.Run(
		netstate_order_channel,
		order_elev_ch_busypoll,
		order_elev_ch_neworder,
		order_elev_ch_finished,
		id)

	//Runs sync module
	go sync.Run(netstate_sync_channel, id)

	//Runs network statemachine
	go network_statemachine.Run(
		netstate_elev_channel,
		netstate_sync_channel,
		netstate_order_channel,
		*portPtr,
		id)


	//Waits
	select {}
}
