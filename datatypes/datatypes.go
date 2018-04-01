package datatypes

//--------------------------------------------------------------------
//--------------- VARIOUS DATATYPES USED IN THE PROJECT---------------
//--------------------------------------------------------------------

//---------Channel types---------

type State_sync_message struct { //Used to communicate between statemachine and sync module
	SyncState State //State variable to be synced
	Connected_count int //Number of connected elevators, important to make sure all elevators has been synchronized
}

type State_elev_message struct { //Used to communicate between network statemachine and elevator statemachine
	Button_matrix Button_matrix
	Floor int
}

type State_order_message struct { //Used to communicate between net-statemachine and order_handler
	Order Order_struct 	//The order sent
	Id_slave string			//The id of the slave to execute order
	ACK bool						//If order is executed this is true
}

//---------Network types--------

type Network_sync_message struct { //Used to Sync states between elevators
	SyncState State //State variable to be synced
	SyncAck bool
	Sender string //ID of sender
}

type Network_order_message struct { //Used to send and receive orders
	Order Order_struct 	//The order to execute
	Id_slave string 		//The id of the slave to execute order
	ACK bool	//If slave CAN execute order this is returned
	NACK bool //If slave CANT execute order this is returned
}


//---------Other types---------

type Button_matrix struct { //Used to keep track of which buttons are pressed
	Up   [4]bool
	Down [4]bool
	Cab  [4]bool
}

type State struct { //The state struct, this includes order list etc..
	Word string
}

type Order_struct struct { //The order object
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

func State_init() State{ //Creates empty State variable
	return State{""}
}
