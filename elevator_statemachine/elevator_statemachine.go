package elevator_statemachine

//-----------------------------------------------------------------------------------------
//-------------------------------------------Controls elevator-----------------------------
//-----------------------------------------------------------------------------------------


import (
  "../elevio_go"
  d "../datatypes"
  "fmt"
  "math/rand"
)

var numFloors = 4

//Determines current floor at startup
func init_floor_finder(floor_sensors_channel chan int) int {

  var current_direction elevio.MotorDirection = elevio.MD_Up
	elevio.SetMotorDirection(current_direction)

  current_floor := -1
	//Wait until we hit a floor
	select {

	case floor := <-floor_sensors_channel:
		current_floor = floor
	}

	//Now we set current_floor and stop
	current_direction = elevio.MD_Stop
	elevio.SetMotorDirection(current_direction)
  return current_floor
}

func Run(state_elev_channel chan d.State_elev_message, order_elev_channel chan d.Order_elev_message, simIp string){

  //Initializes driver
  elevio.Init(simIp, numFloors)

  //Channel to driver
  buttons := make(chan elevio.ButtonEvent)
  floor_sensors_channel := make(chan int)

  //Starts polling buttons and sensors
  go elevio.PollButtons(buttons)
  go elevio.PollFloorSensor(floor_sensors_channel)

  current_floor := init_floor_finder(floor_sensors_channel)
  fmt.Printf("Elevator initialized: Floor determined: %d",current_floor)

  for{
    select{

    case button_event := <- buttons: //Reads button inputs
      //Create and send order update
      fmt.Println("CASE BUTTON_EVENT")
      new_order := d.Order_struct{button_event.Floor,button_event.Button == 0,button_event.Button == 1,button_event.Button == 2,false}
      order_elev_channel <- d.Order_elev_message{new_order,false}

    case floor := <- floor_sensors_channel: //Checks floor sensors
      current_floor = floor
      elevio.SetFloorIndicator(floor)

    case message := <- state_elev_channel:
      fmt.Println("CASE MESSAGE")
      if (message.UpdateLights){  //Updates lights
        update_lights(message.Button_matrix)
      }

    case <-order_elev_channel: //Respond to busyrequest, if busy send busy signal, else execute order

      busystate := false
      if rand.Intn(100) < 30{
        busystate = true
      }
      order_elev_channel <- d.Order_elev_message{d.Order_struct{},busystate}


    }


  }
}

func update_lights(button_matrix d.Button_matrix_struct){ //Updates lights
  fmt.Println("Elev: update_lights() START")

  //Up lights
  for floor := 0; floor < numFloors; floor++{
    elevio.SetButtonLamp(0,floor, button_matrix.Up[floor])
  }

  //Down lights
  for floor := 0; floor < numFloors; floor++{
    elevio.SetButtonLamp(1,floor, button_matrix.Down[floor])
  }

  fmt.Println("Elev: update_lights() END")
}
