package cmd

import (
	"time"

	"git.parallelcoin.io/dev/9/pkg/util/cl"
)

// Log is the logger for the peer package
var Log = cl.NewSubSystem("cmd", "info")

var log = Log.Ch

func Start(args []string) int {
	log <- cl.Dbg("starting 9")

	if err := Parse(args); err != 0 {
		panic(err)
	}
	// pause to let logger finish
	time.Sleep(time.Second / 4)

	return 0
}
