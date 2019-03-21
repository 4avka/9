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

type Lines map[string]*Line

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

	// TODO: make this check and use random port if port is used

	return &Line{}
}

// NetAddrs is for a multiple network addresses ie scheme://host:port, separated by spaces
func NetAddrs(def, usage string) *Line {

	return &Line{}
}

// Int is for a single 64 bit integer. We see no point in complicating things,
// so this is golang `int` with no special meanings
func Int(def, usage string) *Line {

	return &Line{}
}

// IntBounded is an integer whose value must be between a min and max
func IntBounded(def, usage string, min, max int) *Line {

	return &Line{}
}

// Enable is a boolean value
func Enable(usage string) *Line {

	return &Line{}
}

// Disable is a boolean value
func Disable(usage string) *Line {

	return &Line{}
}

// Duration is a time value in golang 24h60m60s format
func Duration(def, usage string) *Line {

	return &Line{}
}

// String is just a boring old string
func String(def, usage string) *Line {

	return &Line{}
}
