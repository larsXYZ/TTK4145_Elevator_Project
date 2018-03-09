package statemachine

import "./../elevio"
import "fmt"

//States
var current_floor int = -1
var current_target int = -1
var current_direction elevio.MotorDirection = elevio.MD_Up
var is_primary bool = false;

//=======Functions=======

//Initializes the statemachine
func Init(floor_sensors_channel chan int){

  //Print info
  fmt.Printf("elevator initialization started\n")

  //Connect to elevator
  elevio.Init("localhost:15657", 4)

  //Listen to floor floor_sensors
  go elevio.PollFloorSensor(floor_sensors_channel)

  //Determine current floor
  init_floor_finder(floor_sensors_channel)

  //Print info
  fmt.Printf("current_floor = %d\n", current_floor)

}

//Determines current floor at startup
func init_floor_finder(floor_sensors_channel chan int){

  elevio.SetMotorDirection(current_direction)

  //Wait until we hit a floor
  select{
  case floor := <-floor_sensors_channel:
    current_floor = floor
  }

  //Now we set current_floor and stop
  current_direction = elevio.MD_Stop
  elevio.SetMotorDirection(current_direction)
}
