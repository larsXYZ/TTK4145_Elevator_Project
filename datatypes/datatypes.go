package datatypes

//--------------------------------------------------------------------
//--------------- VARIOUS DATATYPES USED IN THE PROJECT---------------
//--------------------------------------------------------------------

type State_sync_message struct { //Used to communicate between statemachine and sync module
	SyncState bool //If this is true sync module should sync following statevariable
	State string //State variable to be synced
}

type Button_matrix struct { //Used to keep track of which buttons are pressed
	Up   [4]bool
	Down [4]bool
	Cab  [4]bool
}

type State_elev_message struct { //Used to communicate between statemachine and elevator_interface
	button_matrix_update Button_matrix
}

type Network_message struct { //Used to communicate between elevators on network
	Greeting bool
	Greeting_response bool
}


//===DATATYPE CONSTRUCTORS===

func Button_matrix_init() Button_matrix { //Initializes a button matrix object
	m := Button_matrix{}
	m.Up = [4]bool{false, false, false, false}
	m.Down = [4]bool{false, false, false, false}
	m.Cab = [4]bool{false, false, false, false}

	return m
}
