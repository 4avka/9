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
	pattern string
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
	Handler func(args []string, cmds, tokens, all Commands) int
}

type opts []string
type precedent []string

const (
	APPNAME = "9"
	APPDESC = "all in one everything for parallelcoin"

	HELP, RE_HELP       = "help", "(h|help)"
	CONF, RE_CONF       = "conf", "(C|conf)"
	NEW, RE_NEW         = "new", "(N|new)"
	COPY, RE_COPY       = "copy", "(cp|copy)"
	LIST, RE_LIST       = "list", "(l|list|listcommands)"
	CTL, RE_CTL         = "ctl", "(c|ctl)"
	NODE, RE_NODE       = "node", "(n|node)"
	WALLET, RE_WALLET   = "wallet", "(w|wallet)"
	SHELL, RE_SHELL     = "shell", "(s|shell)"
	TEST, RE_TEST       = "test", "(t|test)"
	CREATE, RE_CREATE   = "create", "(cr|create)"
	LOG, RE_LOG         = "log", "(L|log)"
	DATADIR, RE_DATADIR = "datadir", "([~/.]+.*/)"
	INTEGER, RE_INTEGER = "integer", "[0-9]+"
	FLOAT, RE_FLOAT     = "float", "([0-9]*[.][0-9]+)"
	WORD, RE_WORD       = "word", "([a-zA-Z0-9._/:-]*)"
)

var commandsList = []string{
	HELP, CONF, NEW, COPY, LIST, CTL, NODE, WALLET, SHELL,
	TEST, CREATE, LOG, DATADIR, INTEGER, FLOAT, WORD,
}

func match(s string) *regexp.Regexp {
	return regexp.MustCompile("^" + s + "$")
}

var commands = Commands{
	HELP: {
		RE_HELP,
		match(RE_HELP),
		"show help text and quit",
		`	any other command also mentioned with help/h 
	will have its detailed help information printed`,
		nil,
		precedent{"help"},
		Help,
	},
	CONF: {
		RE_CONF,
		match(RE_CONF),
		"run interactive configuration CLI",
		"	<datadir> sets the data directory to read and write to",
		opts{"datadir"},
		precedent{"help"},
		Conf,
	},
	NEW: {
		RE_NEW,
		match(RE_NEW),
		"create new configuration with optional basename and count for testnets",
		`	<word> is the basename for the data directories
	<integer> is the number of numbered data directories to create`,
		opts{"word", "integer"},
		precedent{"help"},
		New,
	},
	COPY: {
		RE_COPY,
		match(RE_COPY),
		"create a set of testnet configurations based on a datadir",
		`	<datadir> is the base to work from
	<word> is a basename 
	<integer> is a number for how many to create`,
		opts{"datadir", "word", "integer"},
		precedent{"help"},
		Copy,
	},
	LIST: {
		RE_LIST,
		match(RE_LIST),
		"lists commands available at the RPC endpoint",
		`	<datadir> is the enabled data directory
	<ctl> is optional and implied by list
	<wallet> indicates to connect to the wallet RPC
	<node> (or wallet not specified) connect to full node RPC`,
		opts{"datadir", "ctl", "wallet", "node"},
		precedent{"help"},
		List,
	},
	CTL: {
		RE_CTL,
		match(RE_CTL),
		"sends rpc requests and prints the results",
		`	<datadir> sets the data directory to read configurations from
	<node> indicates we are connecting to a full node RPC (overrides wallet and is default)
	<wallet> indicates we are connecting to a wallet RPC
	<word>, <float> and <integer> just cover the items that follow in RPC commands
	the RPC command is expected to be everything after the ctl keyword`,
		opts{"datadir", "node", "wallet", "word", "integer", "float"},
		precedent{"help", "list"},
		Ctl,
	},
	NODE: {
		RE_NODE,
		match(RE_NODE),
		"runs a full node",
		`	<datadir> sets the data directory to read configuration and store data`,
		opts{"datadir"},
		precedent{"help", "ctl"},
		Node,
	},
	WALLET: {
		RE_WALLET,
		match(RE_WALLET),
		"runs a wallet server",
		`	<datadir> sets the data directory to read configuration and store data
	<create> runs the wallet create prompt`,
		opts{"datadir", "create"},
		precedent{"help", "ctl"},
		Wallet,
	},
	SHELL: {
		RE_SHELL,
		match(RE_SHELL),
		"runs a combined node/wallet server",
		`	<datadir> sets the data directory to read configuration and store data
	<create> runs the wallet create prompt`,
		opts{"datadir", "create"},
		precedent{"help", "ctl"},
		Shell,
	},
	TEST: {
		RE_TEST,
		match(RE_TEST),
		"run multiple full nodes from given <word> logging optionally to <datadir>",
		`	<word> indicates the basename to search for as the path to the test configurations
	<log> indicates to write logs to the individual data directories instead of print to stdout`,
		opts{"word", "log"},
		precedent{"help"},
		Test,
	},
	CREATE: {
		RE_CREATE,
		match(RE_CREATE),
		"runs the create new wallet prompt",
		"	<datadir> sets the data directory where the wallet will be stored",
		opts{"datadir"},
		precedent{"wallet", "shell", "help"},
		Create,
	},
	LOG: {
		RE_LOG,
		match(RE_LOG),
		"write to log in <datadir> file instead of printing to stderr",
		"",
		nil,
		nil,
		nil,
	},
	DATADIR: {
		RE_DATADIR,
		match(RE_DATADIR),
		"directory to look for configuration or other, must end in a '/'",
		"",
		nil,
		nil,
		nil,
	},
	INTEGER: {
		RE_INTEGER,
		match(RE_INTEGER),
		"number of items to create",
		"",
		nil,
		nil,
		nil,
	},
	FLOAT: {
		RE_FLOAT,
		match(RE_FLOAT),
		"a floating point value",
		"",
		nil,
		nil,
		nil,
	},
	WORD: {
		RE_WORD,
		match(RE_WORD),
		"mostly used for testnet datadir basenames",
		"",
		nil,
		nil,
		nil,
	},
}
