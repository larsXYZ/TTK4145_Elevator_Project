package elev_fsm

//-----------------------------------------------------------------------------------------
//-------------------------------------------Controls elevator-----------------------------
//-----------------------------------------------------------------------------------------


import (
  "../elevio_go"
  d "../datatypes"
  "fmt"
  "time"
  "sync"
  s "../settings"
)

var numFloors = 4
var busystate = false   //True when elevator is doing an order
var current_floor = -1
var cab_array = [4]bool{false,false,false,false}
var current_direction elevio.MotorDirection = elevio.MD_Stop
var motor = false
var _mtx sync.Mutex

//Determines current floor at startup
func init_floor_finder(floor_sensors_channel chan int) int {

  _mtx = sync.Mutex{}

  current_direction = elevio.MD_Up
	elevio.SetMotorDirection(current_direction)

	//Wait until we hit a floor
	select {

  //Now we set current_floor and stop
	case floor := <-floor_sensors_channel:
		current_floor = floor
	}

  elevio.SetFloorIndicator(current_floor)

	if current_floor == numFloors-1{
    current_direction = elevio.MD_Down
  }

	elevio.SetMotorDirection(elevio.MD_Stop)
  return current_floor
}

//Runs elevator interface logic
func Run(
  netfsm_elev_light_update chan d.Button_matrix_struct,
  order_elev_ch_busypoll chan bool,
	order_elev_ch_neworder chan d.Order_struct,
	order_elev_ch_finished chan d.Order_struct,
  elevIp string,
  ){

  //Initializes driver
  elevio.Init(elevIp, numFloors)

  //Channels
  buttons := make(chan elevio.ButtonEvent,100)  //Polls order buttons
  floor_sensors_channel := make(chan int,100)  //Polls floor sensors
  interrupt := make(chan bool,100)            //Interrupts an ongoing order

  //Set all order lights off
  for floor:=0; floor < numFloors; floor++{
    elevio.SetButtonLamp(elevio.BT_Cab,floor,false)
    elevio.SetButtonLamp(elevio.BT_HallUp,floor,false)
    elevio.SetButtonLamp(elevio.BT_HallDown,floor,false)
  }

  //Starts tick for checking cab orders
  check_cab := time.Tick(s.ELEV_CAB_TIMEOUT*time.Millisecond)

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

      if button_event.Button == elevio.BT_Cab{ //Updates cab_array
        cab_array[button_event.Floor] = true
        elevio.SetButtonLamp(button_event.Button, button_event.Floor, true)
        interrupt <- false

      }else{  //Creates and send order update
        new_order := d.Order_struct{button_event.Floor,button_event.Button == 0,button_event.Button == 1,false}
        order_elev_ch_neworder <- new_order
      }


    case Button_matrix := <- netfsm_elev_light_update:
      update_hall_lights(Button_matrix)


    case <- order_elev_ch_busypoll: //Sends busystate to
      order_elev_ch_busypoll <- busystate


    case received_order := <-order_elev_ch_neworder: //Executes received order

      if next_cab_target() == -1{  //Executes order only if there exist no cab orders
        _mtx.Lock()
        busystate = true
        go execute_order(received_order,floor_sensors_channel,order_elev_ch_finished,true,interrupt)
      }

    case <- check_cab: //Checks cab orders

      if busystate == false{
        var next_target = next_cab_target()

        if next_target != -1{
          _mtx.Lock()
          busystate = true
          fmt.Println("Elev_fsm: Executing cab order: Floor",next_target)
          go execute_order(d.Order_struct{Floor: next_target},floor_sensors_channel,order_elev_ch_finished,false,interrupt) //Executes cab orders
        }
      }
    }
  }
}

func go_to_floor(target_floor int,floor_sensors_channel chan int, interrupt chan bool, master_order bool){

  if (target_floor == elevio.GetFloor()){ //If we already are at the floor we exit
    return
  }

  motor = true

  //Activate motor
  if target_floor > current_floor{
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

    case <- interrupt:   //Checks if a pressed cab order should change target floor
        var next = next_cab_target()
        if next != -1 && next != current_floor && !master_order{
          fmt.Println("Elev_fsm: Changing target floor",next)
          target_floor = next
        }
      }

    if target_floor == current_floor {
      elevio.SetMotorDirection(elevio.MD_Stop)

      //Orientate the direction of the elevator at endpoints
      if current_floor == numFloors-1{
        current_direction = elevio.MD_Down

      }else if current_floor == 0{
        current_direction = elevio.MD_Up
      }

      motor = false
      finished = true
    }
  }
}

func execute_order(order d.Order_struct, floor_sensors_channel chan int, order_elev_ch_finished chan d.Order_struct,master_order bool,interrupt chan bool) { //Executes order and stays busy while doing it


  //Moves to floor
  elevio.SetDoorOpenLamp(false) //Just in case
  go_to_floor(order.Floor, floor_sensors_channel,interrupt,master_order)

  //Updates cab
  if !master_order{
    order_elev_ch_finished <- d.Order_struct{current_floor, current_direction == 1, current_direction == -1, true}
    cab_array[current_floor] = false
    elevio.SetButtonLamp(elevio.BT_Cab,current_floor,false)
    fmt.Println("Elev_fsm: Cab order finished: Floor",current_floor)
  }

  //Notify master that elevator has arrived
  if master_order{
    cab_array[current_floor] = false
    elevio.SetButtonLamp(elevio.BT_Cab,current_floor,false)
    order.Fin = true
    order_elev_ch_finished <- order
  }

  door()

  busystate = false

  _mtx.Unlock()
}

//Open door and wait
func door(){
  if motor == false{
    elevio.SetDoorOpenLamp(true)
    time.Sleep(4*time.Second)
    elevio.SetDoorOpenLamp(false)
  }

}

//Updates lights
func update_hall_lights(button_matrix d.Button_matrix_struct){

  //Up lights
  for floor := 0; floor < numFloors; floor++{
    elevio.SetButtonLamp(0,floor, button_matrix.Up[floor])
  }
  //Down lights
  for floor := 0; floor < numFloors; floor++{
    elevio.SetButtonLamp(1,floor, button_matrix.Down[floor])
  }
}

//Finds the next target floor for cab orders
func next_cab_target() int{

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
