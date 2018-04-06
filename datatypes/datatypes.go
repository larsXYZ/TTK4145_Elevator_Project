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
	Button_matrix Button_matrix_struct
	UpdateLights bool		//True if we should update lights
}

type State_order_message struct { //Used to communicate between net-statemachine and order_handler
	Order Order_struct 	//The order sent
	Id_slave string			//The id of the slave to execute order
	ACK bool						//If order is executed this is true
}

type Order_elev_message struct { //Sent between order handler and elevator statemachine
	Order Order_struct	//Order
	BusyState bool			//Returns busystate
}

//---------Network types--------

type Network_sync_message struct { //Used to Sync states between elevators
	SyncState State //State variable to be synced
	SyncAck bool	//Signal acknowlinging receiving sync-message
	Sender string //ID of sender
}

type Network_order_message struct { //Used to send and receive orders
	Order Order_struct 	//The order to execute
	Id_slave string 		//The id of the slave to execute order
	ACK bool	//If slave CAN execute order this is returned
	NACK bool //If slave CANT execute order this is returned
}


//---------Other types---------

type Button_matrix_struct struct { //Used to keep track of which buttons are pressed
	Up   [4]bool
	Down [4]bool
	Cab  [4]bool
}

type State struct { //The state struct, this includes order list etc..
	Button_matrix Button_matrix_struct
}

type Order_struct struct { //The order object
	Floor int //Floor which the elevator should move to
	Up bool		//True if passanger wants to go up
	Down bool	//True if passanger wants to go down
	Cab bool	//True if cab button is pressed
	Fin bool
}

//===DATATYPE CONSTRUCTORS===

func Button_matrix_init() Button_matrix_struct { //Initializes a button matrix object
	m := Button_matrix_struct{}
	m.Up = [4]bool{false, false, false, false}
	m.Down = [4]bool{false, false, false, false}
	m.Cab = [4]bool{false, false, false, false}

	return m
}

func State_init() State{ //Creates empty State variable
	return State{}
}
