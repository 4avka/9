package main

import (
	"fmt"

	"git.parallelcoin.io/dev/9/pkg/util/interrupt"
)

func main() {

	interrupt.AddHandler(func() {

		fmt.Println("IT'S THE END OF THE WORLD!")
	})
	<-interrupt.HandlersDone
}
