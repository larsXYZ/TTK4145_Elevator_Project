package statemachine

//-----------------------------------------------------------------------------------------
//--------------- Receives, and delegates orders, the brain of each elevator---------------
//-----------------------------------------------------------------------------------------

//Import packages
import d "./../datatypes"
import "fmt"

//=======Functions=======

//Initializes the statemachine
func Init(state_elev_channel chan d.State_elev_message) {

	select {
	case new_message := <-state_elev_channel:
		fmt.Println(new_message)
	}
}
