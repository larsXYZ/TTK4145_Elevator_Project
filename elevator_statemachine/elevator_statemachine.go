//package main
package elevator_statemachine

import (
  "./../elevio_go"
  d "./../datatypes"
  "fmt"
  "math/rand"
)

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
  numFloors := 4
  elevio.Init(simIp, numFloors)


  //Channel to driver
  buttons := make(chan elevio.ButtonEvent)
  floor_sensors_channel := make(chan int)

  //Starts polling buttons and sensors
  go elevio.PollButtons(buttons)
  go elevio.PollFloorSensor(floor_sensors_channel)

  current_floor := init_floor_finder(floor_sensors_channel)

  update := d.State_elev_message{}
  update.Button_matrix = d.Button_matrix_init()
  fmt.Println(update.Button_matrix)

  for{
    select{

    case button_event := <- buttons: //Reads button inputs
      if button_event.Button == elevio.BT_HallUp {
        update.Button_matrix.Up[button_event.Floor] = true
      }else if button_event.Button == elevio.BT_HallDown{
        update.Button_matrix.Down[button_event.Floor] = true
      }else if button_event.Button == elevio.BT_Cab{
        update.Button_matrix.Cab[button_event.Floor] = true
      }

    case floor := <- floor_sensors_channel: //Checks floor sensors
      current_floor = floor
      elevio.SetFloorIndicator(floor)

    case new_order := <- state_elev_channel: //Receives order from state machine, (NOT NEEDED)
      fmt.Println(new_order.Floor)
      floor := new_order.Floor
      if floor == current_floor {
        return
      }else if floor < current_floor{
        elevio.SetMotorDirection(elevio.MD_Down)
      }else{
        elevio.SetMotorDirection(elevio.MD_Up)
      }

    case <-order_elev_channel: //Respond to busyrequest, if busy send busy signal, else execute order

      busystate := false
      if rand.Intn(100) < 70{
        busystate = true
      }
      fmt.Println("A")
      order_elev_channel <- d.Order_elev_message{d.Order_struct{},busystate}

    }
  }
}
