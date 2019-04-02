package txscript

import (
	"git.parallelcoin.io/dev/9/cmd/ll"

	cl "git.parallelcoin.io/dev/9/pkg/util/cl"
)

// Log is the logger for the peer package
var Log = cl.NewSubSystem("chain/tx/script", ll.DEFAULT)
var log = Log.Ch

// UseLogger uses a specified Logger to output package logging info. This should be used in preference to SetLogWriter if the caller is also using log.
func UseLogger(
	logger *cl.SubSystem) {

	Log = logger
	log = Log.Ch
}

// LogClosure is a closure that can be printed with %v to be used to generate expensive-to-create data for a detailed log level and avoid doing the work if the data isn't printed.

type logClosure func() string

func (c logClosure) String() string {

	return c()
}
func newLogClosure(
	c func() string) logClosure {

	return logClosure(c)
}
