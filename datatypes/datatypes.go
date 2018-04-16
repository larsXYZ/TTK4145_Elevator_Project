package datatypes

//--------------------------------------------------------------------
//--------------- VARIOUS DATATYPES USED IN THE PROJECT---------------
//--------------------------------------------------------------------

//---------Channel types---------

type State_sync_message struct { //Used to communicate between statemachine and sync module
	State       		State //State variable to be synced
	Peers						string //Special string containing the ids of the slaves to be synchronized
	Sync						bool		//True if we should sync state
}

type State_order_message struct { //Used to communicate between net-statemachine and order_handler
	Order    Order_struct //The order sent
	Id_slave string       //The id of the slave to execute order
	ACK      bool         //If order is executed this is true
}

//---------Network types--------

type Network_sync_message struct { //Used to Sync states between elevators
	State State  			//State variable to be synced
	SyncAck   bool   	//Signal acknowlinging receiving sync-message
	Sender    string 	//ID of sender
	Target		string 	//ID of target
}

type Network_delegate_order_message struct { //Used to send and receive orders between orderhandlers
	Order    Order_struct //The order to execute
	Id_slave string       //The id of the slave to execute order
	ACK      bool         //If slave CAN execute order this is returned
	NACK     bool         //If slave CANT execute order this is returned
}

type Network_new_order_message struct { //Used to communicate new orders between orderhandlers, implements ACK's
	Order Order_struct		//The order struct sent over internet
	ACK		bool						//Confirming that master has received it
}

type Network_fetch_message struct {
	State State	//State to be synced
	Hello bool	//True if this is an hello message
}

//---------Other types---------

type Button_matrix_struct struct { //Used to keep track of which buttons are pressed
	Up   [4]bool
	Down [4]bool
	Cab  [4]bool
}

type State struct { //The state struct, this includes order list etc..
	Button_matrix Button_matrix_struct

	Time_table_received_up [4]int		//Holds the time of when an order was received
	Time_table_received_down [4]int

	Time_table_delegated_up [4]int	//Holds the time of when an order was delegated
	Time_table_delegated_down [4]int
}

type Order_struct struct { //The order object
	Floor int  //Floor which the elevator should move to
	Up    bool //True if passanger wants to go up
	Down  bool //True if passanger wants to go down
	Fin   bool //True if order is completed
}

//===DATATYPE CONSTRUCTORS===

func Button_matrix_init() Button_matrix_struct { //Initializes a button matrix object
	m := Button_matrix_struct{}
	m.Up = [4]bool{false, false, false, false}
	m.Down = [4]bool{false, false, false, false}

	return m
}

func State_init() State { //Creates empty State variable
	return State{}
}
