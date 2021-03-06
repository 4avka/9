package spv
import (
	"git.parallelcoin.io/dev/9/cmd/ll"
	cl "git.parallelcoin.io/dev/9/pkg/util/cl"
)
// logClosure is used to provide a closure over expensive logging operations so
// don't have to be performed when the logging level doesn't warrant it.
type logClosure func() string
// Log is the logger for node
var Log = cl.NewSubSystem("cmd/spv", ll.DEFAULT)
var log = Log.Ch
// String invokes the underlying function and returns the result.
func (c logClosure) String() string {
	return c()
}
// UseLogger uses a specified Logger to output package logging info. This should be used in preference to SetLogWriter if the caller is also using log.
func UseLogger(
	logger *cl.SubSystem) {
	Log = logger
	log = Log.Ch
}
// newLogClosure returns a new closure over a function that returns a string
// which itself provides a Stringer interface so that it can be used with the
// logging system.
func newLogClosure(
	c func() string) logClosure {
	return logClosure(c)
}
