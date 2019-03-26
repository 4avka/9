package cmd

import (
	"regexp"
)

// Command is a set-based syntax for command line invocation. A set has a
// more limited range of possibilities when an item type cannot appear more
// than once, but for any given task, there usually is only one thing for
// each type required.
//
// If more detailed stuff is needed, we provide access through a configuration
// interactive CLI.
//
// nil values in opt/prec indicate wildcards, empty means no other acceptable.
type Command map[string]struct {
	// How to identify a specific item
	RE *regexp.Regexp
	// Short help information to show
	Usage string
	// Detailed information to show on specific help page
	Detail string
	// Optional keywords that may appear with the command
	Optional opts
	// Precedent indicates other commands that will preferentially match
	Precedent precedent
	// Handler
	Handler func(Command) int
}

type opts []string
type precedent []string

var commands = Command{
	"help": {
		regexp.MustCompile("^(h|help)$"),
		"show help text and quit",
		`	any other command also mentioned with a help line 
	will have its detailed help information printed`,
		nil,
		nil,
		nil,
	},
	"conf": {
		regexp.MustCompile("^(C|conf)$"),
		"run interactive configuration CLI",
		"",
		opts{"datadir"},
		precedent{"help"},
		nil,
	},
	"new": {
		regexp.MustCompile("^(N|new)$"),
		"create new configuration with optional basename and count for testnets",
		"",
		opts{"word", "integer"},
		precedent{"help"},
		nil,
	},
	"copy": {
		regexp.MustCompile("^(cp|copy)$"),
		"create a set of testnet configurations based on a datadir",
		"<datadir> is the base to work from <word> is a basename and <integer> is a number for how many to create",
		opts{"datadir", "word", "integer"},
		precedent{"help"},
		nil,
	},
	"list": {
		regexp.MustCompile("^(l|list|listcommands)$"),
		"lists commands available at the RPC endpoint",
		"",
		opts{"datadir", "ctl", "cli"},
		precedent{"help"},
		nil,
	},
	"ctl": {
		regexp.MustCompile("^(c|ctl)$"),
		"sends rpc requests and prints the results",
		"",
		opts{"datadir", "node", "wallet", "word", "integer"},
		precedent{"help", "list"},
		nil,
	},
	"node": {
		regexp.MustCompile("^(n|node)$"),
		"runs a full node",
		"",
		opts{"datadir"},
		precedent{"help", "ctl"},
		nil,
	},
	"wallet": {
		regexp.MustCompile("^(w|wallet)$"),
		"runs a wallet server",
		"",
		opts{"datadir", "create"},
		precedent{"help", "cli"},
		nil,
	},
	"test": {
		regexp.MustCompile("^(t|test)$"),
		"run multiple full nodes from given <word> logging optionally to <datadir>",
		"",
		opts{"datadir", "word"},
		precedent{"help"},
		nil,
	},
	"create": {
		regexp.MustCompile("^(create)$"),
		"runs the create new wallet prompt",
		"",
		opts{"datadir"},
		precedent{"wallet", "shell", "help"},
		nil,
	},
	"datadir": {
		regexp.MustCompile("^(.*/)$"),
		"directory to look for configuration, must end in a '/'",
		"",
		nil,
		nil,
		nil,
	},
	"integer": {
		regexp.MustCompile("^([0-9]+)$"),
		"number of items to create",
		"",
		nil,
		nil,
		nil,
	},
	"word": {
		regexp.MustCompile("^([a-zA-Z0-9._-]*)$"),
		"mostly used for testnet datadir basenames",
		"",
		nil,
		nil,
		nil,
	},
}

var allcommands = func() (s []string) {
	for i := range commands {
		s = append(s, i)
	}
	return
}()

var positivetests = [][]string{
	{"9", "h"},
	{"9", "help"},
	{"9", "h", "node"},
	{"9", "help", "conf"},
	{"9", "test/", "c"},
	{"9", "node", "h"},
	{"9", "C", "basename", "9"},
	{"9", "n", "test/"},
	{"9", "t", "basename"},
	{"9", "create", "testnet", "9"},
}
