package mempool

import (
	"git.parallelcoin.io/dev/9/cmd/ll"
	cl "git.parallelcoin.io/dev/9/pkg/util/cl"
)

// Log is the logger for the peer package
var Log = cl.NewSubSystem("node/mempool", ll.DEFAULT)
var log = Log.Ch

// UseLogger uses a specified Logger to output package logging info. This should be used in preference to SetLogWriter if the caller is also using log.
func UseLogger(
	logger *cl.SubSystem) {

	Log = logger
	log = Log.Ch
}

// pickNoun returns the singular or plural form of a noun depending on the count n.
func pickNoun(
	n int, singular, plural string) string {

	if n == 1 {
		return singular
	}
	return plural
}
