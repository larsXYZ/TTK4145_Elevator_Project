package datatypes

//--------------------------------------------------------------------
//--------------- VARIOUS DATATYPES USED IN THE PROJECT---------------
//--------------------------------------------------------------------

type State_sync_message struct { //Used to communicate between statemachine and sync module
	SyncState State //State variable to be synced
	Connected_count int //Number of connected elevators, important to make sure all elevators has been synchronized
}

type Network_sync_message struct { //Used to Sync states between elevators
	SyncState State //State variable to be synced
	SyncAck bool
	Sender string //ID of sender
}

type State_elev_message struct { //Used to communicate between statemachine and elevator_interface
	button_matrix_update Button_matrix
}

type Button_matrix struct { //Used to keep track of which buttons are pressed
	Up   [4]bool
	Down [4]bool
	Cab  [4]bool
}

type State struct { //The state struct to be synchronized
	Word string
}

//===DATATYPE CONSTRUCTORS===

func Button_matrix_init() Button_matrix { //Initializes a button matrix object
	m := Button_matrix{}
	m.Up = [4]bool{false, false, false, false}
	m.Down = [4]bool{false, false, false, false}
	m.Cab = [4]bool{false, false, false, false}

	return m
}
