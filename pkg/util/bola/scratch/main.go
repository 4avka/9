package main

import (
	"fmt"
	"os"
	"time"
)

var packetchan = make(chan int, 32)
var incomingchan = make(chan int, 32)
var returnchan = make(chan int, 32)

func main() {
	go pusher()
	go incomer()
	go returner()
	select {}
}

func pusher() {
	counter := 0
	done := false
	for !done {
		packetchan <- counter
		fmt.Println(counter, "-> packetchan")
		time.Sleep(time.Second)
		counter++
		if counter > 3 {
			incomingchan = nil
			returnchan = nil
			os.Exit(0)
		}
	}
}

func incomer() {
	counter := 0
	bundled := []int{}
	done := false
	for !done {
		select {
		case packet := <-packetchan:
			if counter == 3 {
				fmt.Println("-> packetchan {", bundled, "}")
				counter = 0
				bundled = []int{}
			} else {
				fmt.Println("packetchan <-", packet)
				bundled = append(bundled, packet)
				returnchan <- packet
				fmt.Println(packet, "-> returnchan")
				counter++
			}
		default:
		}
	}
}

func returner() {
	counter := 0
	done := false
	for !done {
		select {
		case ret := <-returnchan:
			fmt.Println("returnchan <-", ret)
			incomingchan <- ret
			fmt.Println(ret, "-> incomingchan")
			counter++
		}
	}
}
