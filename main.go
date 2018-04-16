package main

//------------------------------------------------------------------------------------
//--------------- Starts all the other processes in the elevator system---------------
//------------------------------------------------------------------------------------

import (
	d "./datatypes"
	"./net_fsm"
	"./network_go/localip"
	"./sync_module"
	"flag"
	"os"
	"fmt"
	u "./utilities"
	"./elev_fsm"
	"./network_order_handler"
	"time"
)

func main() {

	fmt.Printf("WELCOME TO TTK4145 - ELEVATOR SOFTWARE - GROUP 78\n")

	//Determines port number
	elevPortPtr := flag.Int("port", 15657, "the port of the elevator")
	flag.Parse()

	//Create elevator ip
	elevIp := fmt.Sprintf("localhost:%d",*elevPortPtr)

	//Determine ip & id
	localIp, _ := localip.LocalIP()
	id := fmt.Sprintf("%s%d", u.IpToString(localIp), os.Getpid())
	fmt.Printf("ELEVATOR ID: %s\n",id)

	//Initializes channels
	netfsm_sync_ch_tx_state					:= make(chan d.State_sync_message,100)
	netfsm_sync_ch_rx_state					:= make(chan d.State,100)
	netfsm_sync_ch_error						:= make(chan bool)

	netfsm_order_channel						:= make(chan d.State_order_message,100)

	order_elev_ch_busypoll					:= make(chan bool ,100)
	order_elev_ch_neworder					:= make(chan d.Order_struct,100)
	order_elev_ch_finished					:= make(chan d.Order_struct,100)

	fmt.Println("-----Activating Modules-----")

	//Runs elev fsm
	go elev_fsm.Run(
		order_elev_ch_busypoll,
		order_elev_ch_neworder,
		order_elev_ch_finished,
		elevIp)

	//Waiting for elevator to find floor
	time.Sleep(5*time.Second)

	//Runs sync module
	go sync.Run(netfsm_sync_ch_tx_state,
							netfsm_sync_ch_rx_state,
							netfsm_sync_ch_error,
							id)

	//Run order handler module
	go network_order_handler.Run(
		netfsm_order_channel,
		order_elev_ch_busypoll,
		order_elev_ch_neworder,
		order_elev_ch_finished,
		id)

	//Runs net fsm
	go net_fsm.Run(
		netfsm_sync_ch_tx_state,
		netfsm_sync_ch_rx_state,
		netfsm_sync_ch_error,
		netfsm_order_channel,
		id)

	fmt.Println("-----Activation Completed-----")

	//Waits
	select {}
}
