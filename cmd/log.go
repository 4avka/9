package cmd

import (
	"git.parallelcoin.io/dev/9/pkg/util/cl"
)

// Log is the logger for the peer package
var Log = cl.NewSubSystem("cmd", "info")

var log = Log.Ch
