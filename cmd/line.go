package cmd

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"git.parallelcoin.io/dev/9/pkg/chain/fork"
	"git.parallelcoin.io/dev/9/pkg/util/cl"
)

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

var Networks = []string{"mainnet", "testnet", "simnet", "regtestnet"}

func logLevelValidate(s string) bool {
	for x := range cl.Levels {
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
// colon followed by a string value, space separated, all lower case.
func SubSystem(def, usage string) *Line {
	p := make(map[string]string)
	return &Line{def, func(s string) bool {
		ss := strings.Split(strings.TrimSpace(s), " ")
		for _, y := range ss {
			sss := strings.Split(y, ":")
			for _, x := range cl.Register.List() {
				if x == sss[0] {
					if _, ok := p[x]; !ok {
						cl.Register.Get(x).SetLevel(sss[1])
						p[x] = sss[1]
					}
					return true
				}
			}
		}
		return false
	}, usage, &p}
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

var netAddr func(s string) bool

// NetAddr is for a single network address ie scheme://host:port
func NetAddr(def, usage string) *Line {
	o := new(string)
	defaultPort, _, _ := net.SplitHostPort(def)
	netAddr = func(s string) bool {
		_, _, err := net.SplitHostPort(s)
		if err != nil {
			a := net.JoinHostPort(s, defaultPort)
			o = &a
			return true
		}
		o = &s
		return true
	}
	return &Line{def, netAddr, usage, o}
}

// NetAddrs is for a multiple network addresses ie scheme://host:port, separated by spaces. If a default is given, its port is taken as the default port. If only a number is present, it is used as the defaultPort
func NetAddrs(def, usage string) *Line {
	o := new([]string)
	var defaultPort string
	n, e := strconv.Atoi(def)
	if e == nil {
		defaultPort = fmt.Sprint(n)
	} else if len(def) > 1 {
		defaultPort, _, _ = net.SplitHostPort(def)
	}
	netAddrs := func(ss string) bool {
		s := strings.Split(ss, " ")
		for _, x := range s {
			_, _, err := net.SplitHostPort(x)
			if err != nil {
				a := net.JoinHostPort(x, defaultPort)
				*o = append(*o, a)
				return true
			}
			*o = append(*o, x)
		}
		return true
	}
	return &Line{def, netAddrs, usage, o}
}

// Int is for a single 64 bit integer. We see no point in complicating things,
// so this is golang `int` with no special meanings
func Int(def, usage string) *Line {
	var o *int
	n, e := strconv.Atoi(def)
	if e == nil {
		*o = n
	}
	return &Line{def, func(s string) bool {
		n, e := strconv.Atoi(def)
		if e == nil {
			*o = n
		} else {
			return false
		}
		return true
	}, usage, o}
}

// IntBounded is an integer whose value must be between a min and max
func IntBounded(def, usage string, min, max int) *Line {
	o := new(int)
	n, e := strconv.Atoi(def)
	if e == nil {
		*o = n
	}
	return &Line{def, func(s string) bool {
		n, e := strconv.Atoi(def)
		if e == nil {
			*o = n
		} else {
			return false
		}
		if n < min || n > max {
			return false
		}
		return true
	}, usage, o}
}

// Enable is a boolean value
func Enable(usage string) *Line {
	o := new(bool)
	*o = false
	return &Line{false, func(s string) bool {
		return true
	}, usage, o}
}

// Disable is a boolean value
func Disable(usage string) *Line {
	o := new(bool)
	*o = true
	return &Line{false, func(s string) bool {
		return true
	}, usage, o}
}

// Duration is a time value in golang 24h60m60s format. If it fails to parse it will return zero duration (as well as if it was zero duration)
func Duration(def, usage string) *Line {
	o, e := time.ParseDuration(def)
	if e != nil {
		o = time.Second * 0
	}
	return &Line{def, func(s string) bool {
		o, e = time.ParseDuration(s)
		if e != nil {
			o = time.Second * 0
		}
		return true
	}, usage, &o}
}

// String is just a boring old string. There is no limitations on what a string can contain, it will have any leading or trailing whitespace trimmed.
func String(def, usage string) *Line {
	o := strings.TrimSpace(def)
	return &Line{def, func(s string) bool {
		o = strings.TrimSpace(def)
		return true
	}, usage, &o}
}

// StringSlice is an array of strings, encoded as a series of strings separated by backticks `
func StringSlice(def, usage string) *Line {
	s := strings.TrimSpace(def)
	ss := strings.Split(s, "`")
	return &Line{def, func(s string) bool {
		s = strings.TrimSpace(s)
		ss = strings.Split(s, "`")
		return true
	}, usage, &ss}
}

// Float is a 64 bit floating point number. Returns zero if nothing parsed out.
func Float(def, usage string) *Line {
	f, e := strconv.ParseFloat(def, 64)
	if e != nil {
		f = float64(0.0)
	}
	return &Line{def, func(s string) bool {
		f, e = strconv.ParseFloat(s, 64)
		if e != nil {
			f = float64(0.0)
		}
		return true
	}, usage, &f}
}

// Algos is the available mining algorithms, read out of the fork package
func Algos(def, usage string) *Line {
	o := new(string)
	for _, x := range fork.P9AlgoVers {
		if x == def {
			*o = def
		}
	}
	return &Line{def, func(s string) bool {
		for _, x := range fork.P9AlgoVers {
			if x == def {
				*o = def
				return true
			}
		}
		return false
	}, usage, o}
}

// ValidName checks to see a name is a valid name - first letter alphabetical, last alpha/numeric, all between also . and -
func ValidName(s string) bool {
	re := regexp.MustCompile("[a-z][a-z0-9-.][a-z0-9]+")
	return re.Match([]byte(s))
}
