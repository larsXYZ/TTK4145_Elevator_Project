package utilities

//-----------------------------------------------------------------------------------------
//--------------- Utility functions for our own use----------------------------------------
//-----------------------------------------------------------------------------------------

import (
	"strconv"
	r "math/rand"
)

//=======Functions=======

func StrToInt(s string) int { //Turns strings to ints, but only returns one value

	i, _ := strconv.Atoi(s)
	return i
}

func IpToString(s string) string { //Removes . from ip, allowing it to be used as an id

	new_string := ""

	for i := 0; i < len(s); i++{ //Parse through filtering out non-numbers
		if string(s[i]) != "."{
			new_string = new_string + string(s[i])
		}
	}

	return new_string
}

//Checks if array contains id
func IdInArray(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

//Turns list into string, elements seperated by ,
func ListToString(list []string) string{

	result := ""

	for i := 0; i < len(list); i++{
		result += list[i]
		if i < len(list)-1{
			result += ","
		}
	}

	return result
}

//Simulating packetloss
func PacketLossSim(chance int) bool {

	return r.Intn(100) < chance
}
