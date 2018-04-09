package elevator_statemachine

//-----------------------------------------------------------------------------------------
//-------------------------------------------Controls elevator-----------------------------
//-----------------------------------------------------------------------------------------


import (
  "../elevio_go"
  d "../datatypes"
  "fmt"
  "time"
  //"sync"
  //"math/rand"
)

//--STATES
var numFloors = 4
var busystate = false
var current_floor = -1


//Determines current floor at startup
func init_floor_finder(floor_sensors_channel chan int) int {

  var current_direction elevio.MotorDirection = elevio.MD_Up
	elevio.SetMotorDirection(current_direction)

  //current_floor := -1
	//Wait until we hit a floor
	select {

	case floor := <-floor_sensors_channel:
		current_floor = floor
	}
  elevio.SetFloorIndicator(current_floor)

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
  floor_sensors_channel := make(chan int,100)

  //Starts polling buttons and sensors
  go elevio.PollButtons(buttons)
  go elevio.PollFloorSensor(floor_sensors_channel)

  current_floor := init_floor_finder(floor_sensors_channel)

  fmt.Printf("Elevator initialized: Floor determined: %d\n",current_floor)


  for{
    select{

    case button_event := <- buttons: //Reads button inputs
      //Create and send order update
      new_order := d.Order_struct{button_event.Floor,button_event.Button == 0,button_event.Button == 1,button_event.Button == 2,false}
      order_elev_channel <- d.Order_elev_message{new_order,false}


    case message := <- state_elev_channel:
      if (message.UpdateLights){  //Updates lights
        update_lights(message.Button_matrix)
      }

    case order := <-order_elev_channel: //Respond to busyrequest, if busy send busy signal, else execute order
      order_elev_channel <- d.Order_elev_message{d.Order_struct{},busystate}
      if busystate == false{
          go execute_order(order.Order,floor_sensors_channel)
      }
    }
  }
}

func go_to_floor(target_floor int,floor_sensors_channel chan int){

  if (target_floor == elevio.GetFloorTest()){ //If we already are at the floor we exit
    return
  }

  if target_floor > current_floor{ //Activate motor
    elevio.SetMotorDirection(elevio.MD_Up)
  }else if target_floor < current_floor{
    elevio.SetMotorDirection(elevio.MD_Down)
  }

  finished := false //Wait untill we reach floor
  for !finished{

    select{
    case arrived_floor := <-floor_sensors_channel:
      current_floor = arrived_floor
      elevio.SetFloorIndicator(current_floor)
    }

    if target_floor == current_floor {
      elevio.SetMotorDirection(elevio.MD_Stop)
      finished = true
    }
  }

}

func execute_order(order d.Order_struct, floor_sensors_channel chan int) { //Executes order and stays busy while doing it

  busystate = true

  //Moves to floor
  elevio.SetDoorOpenLamp(false) //Just in case
  go_to_floor(order.Floor, floor_sensors_channel)

  //Open door and wait
  elevio.SetDoorOpenLamp(true)
  time.Sleep(5*time.Second)
  elevio.SetDoorOpenLamp(false)

  busystate = false

}

func update_lights(button_matrix d.Button_matrix_struct){ //Updates lights

  //Up lights
  for floor := 0; floor < numFloors; floor++{
    elevio.SetButtonLamp(0,floor, button_matrix.Up[floor])
  }

  //Down lights
  for floor := 0; floor < numFloors; floor++{
    elevio.SetButtonLamp(1,floor, button_matrix.Down[floor])
  }
}
