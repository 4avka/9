package fork

import (
	cl "git.parallelcoin.io/dev/9/pkg/util/cl"
)

// Log is the logger for the peer package
var Log = cl.NewSubSystem("chain/fork", "info")

var log = Log.Ch
