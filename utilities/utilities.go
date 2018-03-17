package utilities

//-----------------------------------------------------------------------------------------
//--------------- Utility functions for our own use----------------------------------------
//-----------------------------------------------------------------------------------------

import (
	"strconv"
	"fmt"
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
