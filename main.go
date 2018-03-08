package main

import "fmt"

func main(){
    test := [5]int{1,2,3,4,5}
    for  i := 0;  i < 5;  i++ {
        fmt.Printf("%d ",test[i])
    }
}
