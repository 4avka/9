package connmgr

import (
	"git.parallelcoin.io/dev/9/cmd/ll"
	cl "git.parallelcoin.io/dev/9/pkg/util/cl"
)

// Log is the logger for the connmgr package
var Log = cl.NewSubSystem("peer/connmgr", ll.DEFAULT)
var log = Log.Ch

// UseLogger uses a specified Logger to output package logging info.
func UseLogger(
	logger *cl.SubSystem) {

	Log = logger
	log = Log.Ch
}
