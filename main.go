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
	netfsm_sync_ch_command					:= make(chan d.State_sync_message,100)
	netfsm_sync_ch_error						:= make(chan bool,100)

	netfsm_elev_light_update				:= make(chan d.Button_matrix_struct,100)

	netfsm_order_channel						:= make(chan d.State_order_message,100)

	order_elev_ch_busypoll					:= make(chan bool ,100)
	order_elev_ch_neworder					:= make(chan d.Order_struct,100)
	order_elev_ch_finished					:= make(chan d.Order_struct,100)

	fmt.Println("-----Activating Modules-----")

	//Runs interface module
	go elev_fsm.Run(
		netfsm_elev_light_update,
		order_elev_ch_busypoll,
		order_elev_ch_neworder,
		order_elev_ch_finished,
		elevIp)

	//Waiting for elevator to find floor
	time.Sleep(5*time.Second)

	//Runs sync module
	go sync.Run(netfsm_sync_ch_command, netfsm_sync_ch_error, id)

	//Run order handler module
	go network_order_handler.Run(
		netfsm_order_channel,
		order_elev_ch_busypoll,
		order_elev_ch_neworder,
		order_elev_ch_finished,
		id)

	//Runs network statemachine
	go net_fsm.Run(
		netfsm_elev_light_update,
		netfsm_sync_ch_command,
		netfsm_sync_ch_error,
		netfsm_order_channel,
		id)

	fmt.Println("-----Activation Completed-----")

	//Waits
	select {}
}
