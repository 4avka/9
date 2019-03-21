package cmd

// Line is a configuration line, made into map becomes a
// configuration thingy that has set-like properties.
type Line struct {
	// Default is the default for this value
	Default interface{}
	// Type is basically an empty version of the possible thing. Slices with contents are assumed to be toggles, empty slices are arrays, type must match the value and default type
	Validator func(string) bool
	// The help string that will be shown by the interactive config system
	Comment string
	// Value is where this value is actually stored
	Value interface{}
}

var LogLevels = []string{"off", "critical", "warning", "error", "info", "debug", "trace"}
var Networks = []string{"mainnet", "testnet", "simnet", "regtestnet"}

func logLevelValidate(s string) bool {
	for _, x := range LogLevels {
		if x == s {
			return true
		}
	}
	return false
}

func LogLevel(def, usage string) *Line {
	var p *string
	if !logLevelValidate(def) {
		panic("log level was not in available set")
	}
	return &Line{def, logLevelValidate, usage, p}
}

func Path(def, usage string) *Line {
	p := new(string)
	*p = CleanAndExpandPath(def)
	return &Line{def, func(s string) bool {
		*p = CleanAndExpandPath(s)
		return true
	}, usage, p}
}

// SubSystem is just a list of alphanumeric names followed by a
// colon followed by a string value, space separated, all lower case
func SubSystem(def, usage string) *Line {
	var p *string
	return &Line{def, func(s string) bool {
		// TODO: scan clog registry and verify valid log levels
		*p = s
		return true
	}, usage, p}
}

func Network(def, usage string) *Line {
	p := new(string)

	networkValidate := func(s string) bool {
		for _, x := range Networks {
			if x == s {
				*p = s
				return true
			}
		}
		return false
	}

	if !networkValidate(def) {
		panic("default network was not in available set")
	}
	return &Line{def, networkValidate, usage, p}
}

// NetAddr is for a single network address ie scheme://host:port
func NetAddr(def, usage string) *Line {

	return &Line{}
}

// NetAddrs is for a multiple network addresses ie scheme://host:port, separated by spaces
func NetAddrs(def, usage string) *Line {

	return &Line{}
}

type Lines map[string]*Line

// Config is the declaration of our set of application configuration variables.
// Custom functions are written per type that generate a Line struct and contain
// a validator/setter function that checks the input
var Config = Lines{
	"datadir":   Path("~/.", "base directory containing configuration and data"),
	"loglevel":  LogLevel("info", "sets the base default log level"),
	"subsystem": SubSystem("", "[subsystem:loglevel ]+ listsubsystems to see available"),
	"network":   Network("mainnet", "network to connect to"),
	"addpeers":  NetAddrs("", "permanent p2p network peers"),
	"connect":   NetAddrs("", "whitelisted peers"),
	"rpc":       NetAddr("localhost:11048", "node rpc to connect to"),
}
