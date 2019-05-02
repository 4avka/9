package bola

import (
	"git.parallelcoin.io/dev/9/pkg/util/cl"
)

// var _debuglevel = ll.DEFAULT
var _debuglevel = "trace"
var Log = cl.NewSubSystem("pkg/util/bola", _debuglevel)
var log = Log.Ch
