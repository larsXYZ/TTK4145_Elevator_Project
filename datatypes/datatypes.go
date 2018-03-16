package datatypes

//--------------------------------------------------------------------
//--------------- VARIOUS DATATYPES USED IN THE PROJECT---------------
//--------------------------------------------------------------------

type Heartbeat struct { //Used for notifying the other elevators that master is alive
	Id string
	Ip string
}

type State_elev_message struct { //Used to communicate between statemachine and elevator_interface
	NewFloor       uint
	ArrivedAtFloor bool
	NewOrder       uint
}

type State_network_message struct { //Used to communicate between statemachine and elevator_interface
	NewFloor       uint
	ArrivedAtFloor bool
	NewOrder       uint
}
