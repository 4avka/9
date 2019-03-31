package cmd

import (
	"fmt"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"git.parallelcoin.io/dev/9/cmd/nine"
	"git.parallelcoin.io/dev/9/pkg/chain/fork"
	"git.parallelcoin.io/dev/9/pkg/util/cl"
)

// Line is a configuration line, made into map becomes a
// configuration thingy that has set-like properties.
type Line struct {

	// Default is the default for this value
	Default interface{}

	// Validator returns true if the string is properly formed for the type
	Validator func(string) bool

	// The help string that will be shown by the interactive config
	// system and cli help
	Comment string

	// Value is where this value is actually stored
	Value interface{}
}

func (l *Line) String() string {
	return fmt.Sprint(*l.Value.(*string))
}

type Lines map[string]*Line

type Stringslice []string

func (s Stringslice) String() (out string) {
	for i, x := range s {
		out += x
		if i < len(s)-1 {
			out += " "
		}
	}
	return
}

func (l Lines) String() (out string) {
	tags := make([]string, 0)
	for i := range l {
		tags = append(tags, i)
	}
	sort.Strings(tags)
	for _, x := range tags {
		out += fmt.Sprint("NAME ", x)
		out += fmt.Sprint(" VALUE ", l[x])
		out += fmt.Sprint(" DEFAULT ", l[x].Default)
		out += fmt.Sprint(" COMMENT ", l[x].Comment, "\n")
	}
	return
}

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
	var p string
	if !logLevelValidate(def) {
		panic("log level was not in available set")
	}
	p = def
	options := []string{}
	for i := range cl.Levels {
		options = append(options, i)
	}
	avail := fmt.Sprint(" { ", Stringslice(options), " }")

	var l Line
	l = Line{
		def, func(s string) bool {
			s = strings.TrimSpace(s)
			if len(s) < 1 {
				return true
			}
			for x := range cl.Levels {
				if x == s {
					l.Value = &s
					return true
				}
			}
			return false
		}, usage + avail, &p,
	}
	return &l
}

func Path(def, usage string) *Line {
	p := CleanAndExpandPath(def)
	var l Line
	l = Line{def, func(s string) bool {
		ss := CleanAndExpandPath(s)
		l.Value = &ss
		return true
	}, usage, &p}
	return &l
}

// SubSystem is just a list of alphanumeric names followed by a
// colon followed by a string value, space separated, all lower case.
func SubSystem(def, usage string) *Line {
	p := make(nine.Mapstringstring)
	return &Line{def, func(s string) bool {
		s = strings.TrimSpace(s)
		if len(s) < 1 {
			return true
		}
		ss := strings.Split(s, " ")
		for _, y := range ss {
			sss := strings.Split(y, ":")
			for _, x := range cl.Register.List() {
				if x == sss[0] {
					if _, ok := p[x]; !ok {
						cl.Register.Get(x).SetLevel(sss[1])
						p[x] = &sss[1]
					}
					return true
				}
			}
		}
		return false
	}, usage, &p}
}

func Network(def, usage string) *Line {
	var p *string
	networkValidate := func(s string) bool {
		for _, x := range Networks {
			if x == s {
				p = &s
				return true
			}
		}
		return false
	}
	if !networkValidate(def) {
		panic("default network was not in available set")
	}
	nets := fmt.Sprint(Networks)
	nets = nets[1 : len(nets)-1]
	nets = " { " + nets + " }"
	return &Line{
		def, networkValidate, usage + nets, p,
	}
}

// NetAddr is for a single network address ie scheme://host:port
func NetAddr(def, usage string) *Line {
	p := def
	defaultPort, _, _ := net.SplitHostPort(def)
	var l Line
	l = Line{def, func(s string) bool {
		_, _, err := net.SplitHostPort(s)
		if err != nil {
			a := net.JoinHostPort(s, defaultPort)
			l.Value = &a
			return true
		} else {
			l.Value = &s
		}
		return true
	}, usage, &p}
	return &l
}

// NetAddrs is for a multiple network addresses ie scheme://host:port, separated by spaces.
// If a default is given, its port is taken as the default port. If only a number is present,
// it is used as the defaultPort
func NetAddrs(def, usage string) *Line {
	o := []string{}
	var defaultPort string
	if len(def) < 1 {
		def = ":11047"
	} else {
		o = append(o, def)
	}
	n, e := strconv.Atoi(def)
	if e == nil {
		defaultPort = fmt.Sprint(n)
	} else if len(def) > 1 {
		defaultPort, _, _ = net.SplitHostPort(def)
	}
	var l Line
	l = Line{
		def, func(ss string) bool {
			ss = strings.TrimSpace(ss)
			if len(ss) < 1 {
				return true
			}
			s := strings.Split(ss, " ")
			for _, x := range s {
				_, _, err := net.SplitHostPort(x)
				if err != nil {
					a := net.JoinHostPort(x, defaultPort)
					ss := append(l.Value.([]string), a)
					l.Value = &ss
					return true
				}
				sss := append(l.Value.([]string), x)
				l.Value = &sss
			}
			return true
		}, usage, &o,
	}
	return &l
}

// Int is for a single 64 bit integer. We see no point in
// complicating things, so this is golang `int` with no
// special meanings
func Int(def int, usage string) *Line {
	o := def
	var l Line
	l = Line{def, func(s string) bool {
		n, e := strconv.Atoi(s)
		if e == nil {
			l.Value = &n
		} else {
			return false
		}
		return true
	}, usage, &o}
	return &l
}

// IntBounded is an integer whose value must be between a min
// and max
func IntBounded(def int, usage string, min, max int) *Line {
	o := def
	var l Line
	l = Line{def, func(s string) bool {
		n, e := strconv.Atoi(s)
		if e == nil {
			if n < min || n > max {
				return false
			}
			l.Value = &n
		} else {
			return false
		}
		return true
	}, usage + fmt.Sprintf(" { %d < %d }", min, max), &o}
	return &l
}

// Enable is a boolean value
func Enable(usage string) *Line {
	o := false
	var l Line
	l = Line{o, func(s string) bool {
		ss := strings.ToLower(s) == "true"
		l.Value = &ss
		return true
	}, usage, &o}
	return &l
}

// Disable is a boolean value
func Disable(usage string) *Line {
	o := false
	var l Line
	l = Line{o, func(s string) bool {
		ss := strings.ToLower(s) == "true"
		l.Value = &ss
		return true
	}, usage, &o}
	return &l
}

// Duration is a time value in golang 24h60m60s format. If it fails to
// parse it will return zero duration (as well as if it was zero
// duration)
func Duration(def, usage string) *Line {
	o, e := time.ParseDuration(def)
	if e != nil {
		o = time.Second * 0
	}
	var l Line
	l = Line{def, func(s string) bool {
		var dd time.Duration
		dd, e = time.ParseDuration(s)
		if e != nil {
			ddd := time.Second * 0
			l.Value = &ddd
		} else {
			l.Value = &dd
		}
		return true
	}, usage, &o}
	return &l
}

// String is just a boring old string. There is no limitations on what
// a string can contain, it will have any leading or trailing
// whitespace trimmed.
func String(def, usage string) *Line {
	o := strings.TrimSpace(def)
	var l Line
	l = Line{def, func(s string) bool {
		ss := strings.TrimSpace(s)
		l.Value = &ss
		return true
	}, usage, &o}
	return &l
}

// StringSlice is an array of strings, encoded as a series of strings
// separated by backticks `
func StringSlice(def, usage string) *Line {
	s := strings.TrimSpace(def)
	ss := strings.Split(s, "`")
	var l Line
	l = Line{def, func(s string) bool {
		s = strings.TrimSpace(s)
		values := strings.Split(s, "`")
		if len(values) > 1 {
			l.Value = &values
		}
		return true
	}, usage, &ss}
	return &l
}

// Float is a 64 bit floating point number. Returns zero if nothing
// parsed out.
func Float(def, usage string) *Line {
	f, e := strconv.ParseFloat(def, 64)
	if e != nil {
		f = float64(0.0)
	}
	var l Line
	l = Line{def, func(s string) bool {
		var ff float64
		ff, e = strconv.ParseFloat(s, 64)
		if e != nil {
			fff := float64(0.0)
			l.Value = &fff
		} else {
			l.Value = &ff
		}
		return true
	}, usage, &f}
	return &l
}

// Algos is the available mining algorithms, read out of the fork
// package
func Algos(def, usage string) *Line {
	const modernd = "random"
	o := modernd
	options := []string{}
	for _, x := range fork.P9AlgoVers {
		if x == def {
			o = def
		}
		options = append(options, x)
	}
	avail := fmt.Sprint(options)
	avail = avail[1 : len(avail)-1]
	avail = fmt.Sprint(" { "+modernd+" ", avail, " }")
	var l Line
	l = Line{def, func(s string) bool {
		s = strings.TrimSpace(s)
		if len(s) < 1 || s == modernd {
			l.Value = modernd
			return true
		}
		for _, x := range fork.P9AlgoVers {
			if x == s {
				l.Value = s
				return true
			}
		}
		return false
	}, usage + avail, &o}
	return &l
}

// ValidName checks to see a name is a valid name - first letter
// alphabetical, last alpha/numeric, all between also . and -
func ValidName(s string) bool {
	re := regexp.MustCompile("[a-z][a-z0-9-.][a-z0-9]+")
	return re.Match([]byte(s))
}
