package statemachine

import "./../elevio"

//States
var current_floor int = -1
var current_target int = -1
var current_direction elevio.MotorDirection = elevio.MD_Up

//Functions
func Init(){

  //Channels
  floor_sensors_channel  := make(chan int)

  //Connect to elevator
  elevio.Init("localhost:15657", 4)

  //Listen to floor floor_sensors
  go elevio.PollFloorSensor(floor_sensors_channel)

  //Determines current floor
  elevio.SetMotorDirection(current_direction)

  //Wait until we hit a floor
  select{
  case floor := <-floor_sensors_channel:
    current_floor = floor
  }

  //Now we stop
  current_direction = elevio.MD_Stop
  elevio.SetMotorDirection(current_direction)







}
