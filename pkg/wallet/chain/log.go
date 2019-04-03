package chain

import (
	"git.parallelcoin.io/dev/9/cmd/ll"
	cl "git.parallelcoin.io/dev/9/pkg/util/cl"
)

// Log is the logger for the peer package
var Log = cl.NewSubSystem("wallet/chain", ll.DEFAULT)
var log = Log.Ch
