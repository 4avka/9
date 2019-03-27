package cmd

import (
	"fmt"
	"time"

	"git.parallelcoin.io/dev/9/pkg/util/cl"
)

func Start(args []string) int {
	log <- cl.Dbg("starting 9")
	for i, x := range testargs[0] {
		if err := Parse(x); err != 0 {
			fmt.Println("error item", i, x)
		}
	}
	// pause to let logger finish
	time.Sleep(time.Second)

	return 0
}

var testargs = [][][]string{
	{
		// positive
		{"9", "h"},
		{"9", "help"},
		{"9", "h", "node"},
		{"9", "help", "conf"},
		{"9", "test/", "c"},
		{"9", "node", "h"},
		{"9", "C", "basename", "9"},
		{"9", "n", "test/"},
		{"9", "t", "basename"},
		{"9", "create", "testnet", "9"},
		{"9", "create", "testnet", "9", "h"},
		{"9", "w", "test/"},
		{"9", "n", "wallet"},
		{"9", "n", "wallet", "c", "h"},
	},
	{
		// negative
	},
}
