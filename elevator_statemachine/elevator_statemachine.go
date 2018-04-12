package elevator_statemachine

//-----------------------------------------------------------------------------------------
//-------------------------------------------Controls elevator-----------------------------
//-----------------------------------------------------------------------------------------


import (
  "../elevio_go"
  d "../datatypes"
  "fmt"
  "time"
  //"../timer"
  //"sync"
  //"math/rand"
)

//--STATES
var numFloors = 4
var busystate = false
var current_floor = -1
var cab_array = [4]bool{false,false,false,false}
var current_direction elevio.MotorDirection = elevio.MD_Stop

//Determines current floor at startup
func init_floor_finder(floor_sensors_channel chan int) int {

  current_direction = elevio.MD_Up
	elevio.SetMotorDirection(current_direction)

	//Wait until we hit a floor
	select {

  //Now we set current_floor and stop
	case floor := <-floor_sensors_channel:
		current_floor = floor
	}
  elevio.SetFloorIndicator(current_floor)


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
  cab := make(chan elevio.ButtonEvent)

  //Starts tick for checking cab orders
  check_cab := time.Tick(time.Second)

  //Starts polling buttons and sensors
  go elevio.PollButtons(buttons)
  go elevio.PollFloorSensor(floor_sensors_channel)
  go elevio.PollButtons(cab)


  //Determine starting floor
  current_floor := init_floor_finder(floor_sensors_channel)

  fmt.Printf("Elevator initialized: Floor determined: %d\n",current_floor)

  //Start listening for commands and buttonpresses
  for{
    select{

    case button_event := <- buttons: //Reads button inputs
      if button_event.Button == elevio.BT_Cab{ //updates cab_array
        cab_array[button_event.Floor] = true
        elevio.SetButtonLamp(button_event.Button, button_event.Floor, true)
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
      if next_cab_target() == -1{  //Executes order only if there exist no cab orders
        go execute_order(order,floor_sensors_channel,order_elev_ch_finished,true,cab)
      }

    case <- check_cab: //Polls cab orders
      if busystate == false{
        var next_target = next_cab_target()
        if next_target != -1{
          go execute_order(d.Order_struct{Floor: next_target},floor_sensors_channel,order_elev_ch_finished,false,cab) //Executes cab orders
        }
      }
    }
  }
}

func go_to_floor(target_floor int,floor_sensors_channel chan int, cab chan elevio.ButtonEvent){

  if (target_floor == elevio.GetFloorTest()){ //If we already are at the floor we exit
    return
  }

  if target_floor > current_floor{ //Activate motor
    current_direction = elevio.MD_Up
    elevio.SetMotorDirection(current_direction)
  }else if target_floor < current_floor{
    current_direction = elevio.MD_Down
    elevio.SetMotorDirection(current_direction)
  }

  finished := false //Wait until we reach floor
  for !finished{

    select{
    case arrived_floor := <-floor_sensors_channel:
      current_floor = arrived_floor
      elevio.SetFloorIndicator(current_floor)

    case button_event := <- cab:   //Checks if a pressed cab order should change target floor
      var next = next_cab_target()
      if  next != -1 && button_event.Button == elevio.BT_Cab{
        //fmt.Println("Changing target floor",next)
        target_floor = next
      }
    }

    if target_floor == current_floor {
      //current_direction = elevio.MD_Stop
      elevio.SetMotorDirection(elevio.MD_Stop)
      finished = true
    }
  }

}

func execute_order(order d.Order_struct, floor_sensors_channel chan int, order_elev_ch_finished chan d.Order_struct,master_order bool,chab chan elevio.ButtonEvent) { //Executes order and stays busy while doing it

  busystate = true

  //Moves to floor
  elevio.SetDoorOpenLamp(false) //Just in case
  go_to_floor(order.Floor, floor_sensors_channel,chab)

  //Internal cab orders
  if !master_order{
    cab_array[current_floor] = false
    elevio.SetButtonLamp(elevio.BT_Cab,current_floor,false)
  }

  //Notify master that elevator has arrived
  if master_order{
    order.Fin = true
    order_elev_ch_finished <- order
  }

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

func next_cab_target() int{ //Finds the next target floor for cab orders
  //fmt.Println("current_direction",current_direction)
  if current_floor == numFloors-1{
    current_direction = elevio.MD_Down
  }

  if current_direction == elevio.MD_Up{
    for floor := current_floor+1; floor < numFloors; floor++{
      if cab_array[floor]{
        return floor
      }
    }
  }else if current_direction == elevio.MD_Down{
    for floor := current_floor-1; floor >= 0; floor--{
      if cab_array[floor]{
        return floor
      }
    }
  }

  for floor := 0; floor < numFloors; floor++{
      if cab_array[floor]{
        return floor
      }
    }

  return -1
}
