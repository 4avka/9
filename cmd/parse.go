package cmd

import (
	"time"

	"git.parallelcoin.io/dev/9/pkg/util/cl"
)

func Parse(args []string) int {
	// parse commandline
	// collect set of items in commandline
	if len(args) < 2 {
		log <- cl.Info{"no args given, printing help"}
	}
	for i, x := range args[1:] {
		for j, y := range commands {
			if y.RE.Match([]byte(x)) {
				log <- cl.Info{"found", i, j, x}
			}
		}
	}
	// read configuration
	time.Sleep(time.Second)
	return 0
}
