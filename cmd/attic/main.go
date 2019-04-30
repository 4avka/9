package cmd
import (
	"time"
	"git.parallelcoin.io/dev/9/cmd/ll"
	"git.parallelcoin.io/dev/9/pkg/util/cl"
)
// Log is the logger for the peer package
var log = cl.NewSubSystem("cmd", ll.DEFAULT).Ch
func Start(args []string) int {
	log <- cl.Dbg("starting 9")
	if err := Parse(args); err != 0 {
		time.Sleep(time.Second)
		panic(err)
	}
	// pause to let logger finish
	time.Sleep(time.Second / 4)
	return 0
}
