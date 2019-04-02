package addrmgr

import (
	"git.parallelcoin.io/dev/9/cmd/ll"
	cl "git.parallelcoin.io/dev/9/pkg/util/cl"
)

// Log is the logger for the addrmgr package
var Log = cl.NewSubSystem("peer/addrmgr", ll.DEFAULT)
var log = Log.Ch
