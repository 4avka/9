package bola

import (
	"git.parallelcoin.io/dev/9/cmd/ll"

	"git.parallelcoin.io/dev/9/pkg/util/cl"
)

var Log = cl.NewSubSystem("pkg/util/bola", ll.DEFAULT)
var log = Log.Ch
