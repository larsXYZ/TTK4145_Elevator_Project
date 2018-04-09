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
//var order_queue = make([]int,numFloors)


//Determines current floor at startup
func init_floor_finder(floor_sensors_channel chan int) int {

  var current_direction elevio.MotorDirection = elevio.MD_Up
	elevio.SetMotorDirection(current_direction)

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

//Runs elevator interface logic
func Run(
  state_elev_channel chan d.State_elev_message,
  order_elev_ch_busypoll chan bool,
	order_elev_ch_neworder chan d.Order_struct,
	order_elev_ch_finished chan d.Order_struct,
  simIp string,
  ){

  //Initializes driver
  elevio.Init(simIp, numFloors)

  //Channel to driver
  buttons := make(chan elevio.ButtonEvent)
  floor_sensors_channel := make(chan int,100)

  //Starts polling buttons and sensors
  go elevio.PollButtons(buttons)
  go elevio.PollFloorSensor(floor_sensors_channel)

  //Determine starting floor
  current_floor := init_floor_finder(floor_sensors_channel)

  fmt.Printf("Elevator initialized: Floor determined: %d\n",current_floor)

  //Start listening for commands and buttonpresses
  for{
    select{

    case button_event := <- buttons: //Reads button inputs
      if button_event.Button == elevio.BT_Cab && busystate == false{
        go execute_order(d.Order_struct{Floor: button_event.Floor},floor_sensors_channel,order_elev_ch_finished) //Executes cab orders
      }else{
        //Create and send order update
        new_order := d.Order_struct{button_event.Floor,button_event.Button == 0,button_event.Button == 1,false}
        order_elev_ch_neworder <- new_order

      }


    case message := <- state_elev_channel:
      if (message.UpdateLights){  //Updates lights
        update_lights(message.Button_matrix)
      }

    case <- order_elev_ch_busypoll: //Sends busystate to ordehandler
      order_elev_ch_busypoll <- busystate

    case order := <-order_elev_ch_neworder: //Executes received order
      go execute_order(order,floor_sensors_channel,order_elev_ch_finished)
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

  finished := false //Wait until we reach floor
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

func execute_order(order d.Order_struct, floor_sensors_channel chan int, order_elev_ch_finished chan d.Order_struct) { //Executes order and stays busy while doing it

  busystate = true

  //Moves to floor
  elevio.SetDoorOpenLamp(false) //Just in case
  go_to_floor(order.Floor, floor_sensors_channel)

  //Notify master that elevator has arrived
  order.Fin = true
  order_elev_ch_finished <- order

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
