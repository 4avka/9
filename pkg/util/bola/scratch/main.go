package main

import (
	"fmt"
	"runtime"
	"time"
)

const queuesize = 1000000

var queue = make(chan int, 1)
var finished = make(chan struct{})
var readyreceive = make(chan struct{})
var factors = make([][]int, queuesize)

func main() {
	fmt.Println(runtime.NumCPU())
	timenow := time.Now()
	go reception()
	go emitter()
	<-finished
	comptime := time.Now().Sub(timenow) / queuesize
	// println("highest common factor for numbers between 2 and", queuesize)
	// println()
	for i, x := range factors {
		if len(x) == 0 {
			continue
		}
		// if i == 1 {
		// 	println("prime numbers:")
		print(i, ": ")
		for _, y := range x {
			print(y, " ")
		}
		println()
		// }
		// time.Sleep(time.Millisecond * 200)
	}
	fmt.Println(int(comptime), "ns/op")
}

func ranger(max int) (o []int) {
	for i := 2; i < max; i++ {
		o = append(o, i)
	}
	return
}

func emitter() {
	<-readyreceive
	// print("receiver is ready, emitting:")
	for _, x := range ranger(queuesize) {
		// print(x, ",")
		queue <- x
	}
	// println("finished emitting")
	finished <- struct{}{}
	// println()
}

func reception() {
	// println("reception opening")
	readyreceive <- struct{}{}
	for x := range queue {
		for i := queuesize; i > 0; i-- {
			if x%i == 0 && i != x {
				factors[i] = append(factors[i], x)
				break
			}
		}
	}
}
