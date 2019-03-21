package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	"git.parallelcoin.io/dev/9/cmd"
	"git.parallelcoin.io/dev/9/pkg/util/limits"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())
	debug.SetGCPercent(10)

	if err := limits.SetLimits(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to set limits: %v\n", err)
		os.Exit(1)
	}

	os.Exit(cmd.Start(os.Args))
}
