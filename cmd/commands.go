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
type Commands map[string]Command

type Command struct {
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
	Handler func(*Command) int
}

type opts []string
type precedent []string

const (
	HELP    = "help"
	CONF    = "conf"
	NEW     = "new"
	COPY    = "copy"
	LIST    = "list"
	CTL     = "ctl"
	NODE    = "node"
	WALLET  = "wallet"
	SHELL   = "shell"
	TEST    = "test"
	CREATE  = "create"
	LOG     = "log"
	DATADIR = "datadir"
	INTEGER = "integer"
	FLOAT   = "float"
	WORD    = "word"
)

var commandsList = []string{
	HELP, CONF, NEW, COPY, LIST, CTL, NODE, WALLET, SHELL,
	TEST, CREATE, LOG, DATADIR, INTEGER, FLOAT, WORD,
}

var commands = Commands{
	HELP: {
		regexp.MustCompile("^(h|help)$"),
		"show help text and quit",
		`	any other command also mentioned with help/h 
	will have its detailed help information printed`,
		nil,
		precedent{"help"},
		Help,
	},
	CONF: {
		regexp.MustCompile("^(C|conf)$"),
		"run interactive configuration CLI",
		"	<datadir> sets the data directory to read and write to",
		opts{"datadir"},
		precedent{"help"},
		Conf,
	},
	NEW: {
		regexp.MustCompile("^(N|new)$"),
		"create new configuration with optional basename and count for testnets",
		`	<word> is the basename for the data directories
	<integer> is the number of numbered data directories to create`,
		opts{"word", "integer"},
		precedent{"help"},
		New,
	},
	COPY: {
		regexp.MustCompile("^(cp|copy)$"),
		"create a set of testnet configurations based on a datadir",
		`	<datadir> is the base to work from
	<word> is a basename 
	<integer> is a number for how many to create`,
		opts{"datadir", "word", "integer"},
		precedent{"help"},
		Copy,
	},
	LIST: {
		regexp.MustCompile("^(l|list|listcommands)$"),
		"lists commands available at the RPC endpoint",
		`	<datadir> is the enabled data directory
	<ctl> is optional and implied by list
	<wallet> indicates to connect to the wallet RPC
	<node>, or wallet not specified to connect to full node RPC`,
		opts{"datadir", "ctl", "wallet", "node"},
		precedent{"help"},
		List,
	},
	CTL: {
		regexp.MustCompile("^(c|ctl)$"),
		"sends rpc requests and prints the results",
		`	<datadir> sets the data directory to read configurations from
	<node> indicates we are connecting to a full node RPC (overrides wallet and is default)
	<wallet> indicates we are connecting to a wallet RPC
	<word> and <integer> just cover the items that follow in RPC commands
	the RPC command is expected to be everything after the ctl keyword`,
		opts{"datadir", "node", "wallet", "word", "integer"},
		precedent{"help", "list"},
		Ctl,
	},
	NODE: {
		regexp.MustCompile("^(n|node)$"),
		"runs a full node",
		`	<datadir> sets the data directory to read configuration and store data`,
		opts{"datadir"},
		precedent{"help", "ctl"},
		Node,
	},
	WALLET: {
		regexp.MustCompile("^(w|wallet)$"),
		"runs a wallet server",
		`	<datadir> sets the data directory to read configuration and store data
	<create> runs the wallet create prompt`,
		opts{"datadir", "create"},
		precedent{"help", "ctl"},
		Wallet,
	},
	SHELL: {
		regexp.MustCompile("^(S|shell)$"),
		"runs a combined node/wallet server",
		`	<datadir> sets the data directory to read configuration and store data
	<create> runs the wallet create prompt`,
		opts{"datadir", "create"},
		precedent{"help", "ctl"},
		Shell,
	},
	TEST: {
		regexp.MustCompile("^(t|test)$"),
		"run multiple full nodes from given <word> logging optionally to <datadir>",
		`	<word> indicates the basename to search for as the path to the test configurations
	<log> indicates to write logs to the individual data directories instead of print to stdout`,
		opts{"word", "log"},
		precedent{"help"},
		Test,
	},
	CREATE: {
		regexp.MustCompile("^(create)$"),
		"runs the create new wallet prompt",
		"	<datadir> sets the data directory where the wallet will be stored",
		opts{"datadir"},
		precedent{"wallet", "shell", "help"},
		Create,
	},
	LOG: {
		regexp.MustCompile("^(log)$"),
		"write to log in <datadir> file instead of printing to stderr",
		"",
		nil,
		nil,
		nil,
	},
	DATADIR: {
		regexp.MustCompile("^(.*/)$"),
		"directory to look for configuration or other, must end in a '/'",
		"",
		nil,
		nil,
		nil,
	},
	INTEGER: {
		regexp.MustCompile("^([0-9]+)$"),
		"number of items to create",
		"",
		nil,
		nil,
		nil,
	},
	FLOAT: {
		regexp.MustCompile("^([0-9]*[.][0-9]+)$"),
		"a floating point value",
		"",
		nil,
		nil,
		nil,
	},
	WORD: {
		regexp.MustCompile("^([a-zA-Z][a-zA-Z0-9._-]*)$"),
		"mostly used for testnet datadir basenames",
		"",
		nil,
		nil,
		nil,
	},
}
