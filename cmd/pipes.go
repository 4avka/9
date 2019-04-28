package cmd

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"git.parallelcoin.io/dev/9/pkg/chain/fork"
	"git.parallelcoin.io/dev/9/pkg/util/cl"
	"github.com/davecgh/go-spew/spew"
)

func NewConfig() *Lines {
	return &Lines{}
}

func maybeStringPointer(si interface{}) string {
	switch st := si.(type) {
	case string:
		return st
	case *string:
		return *st
	default:
		return ""
	}
}

func (l *Lines) Group(s string, items ...*Line) (out *Lines) {
	ll := make(Lines)
	out = &ll
	fmt.Println(items)
	for i, x := range items {
		if x.Name == "" {
			// spew.Dump(x)
			fmt.Println("name should not be empty!")
			continue
		}
		// Prepend group name to item name
		x.Name = s + "." + x.Name
		if x.Validate != nil {
			// Store in map
			x.Validate(items[i], x.Initial)
		}
		(*out)[x.Name] = x
	}
	fmt.Println("\n", s)
	// spew.Dump(items)
	// spew.Dump(*out)
	return
}

func Int(s string) *Line {
	o := Line{
		Name: s,
		Validate: func(l *Line, si interface{}) bool {
			if si == nil {
				return false
			}
			ss := *si.(*string)
			n, e := strconv.Atoi(ss)
			if e != nil {
				l.INT(n)
			}
			return true
		},
	}
	return &o
}

func Float(s string) *Line {
	o := Line{Name: s,
		Validate: func(l *Line, si interface{}) bool {
			if si == nil {
				return false
			}
			ss := *si.(*string)
			n, e := strconv.ParseFloat(ss, 64)
			if e != nil {
				l.FLOAT(n)
			}
			return true
		},
	}
	return &o
}

func Duration(s string) *Line {
	o := Line{Name: s}
	return &o
}

func Log(s string) *Line {
	o := Line{
		Name: s,
		Validate: func(l *Line, si interface{}) bool {
			ss := *si.(*string)
			for x := range cl.Levels {
				if x == ss {
					l.STRING(ss)
					return true
				}
			}
			return false
		},
	}
	return &o
}

func Tags(s string) *Line {
	o := Line{
		Name: s,
		Validate: func(l *Line, si interface{}) bool {
			if si != nil {
				ss := strings.TrimSpace(*si.(*string))
				if len(ss) < 1 {
					l.SLICE([]string{})
					return false
				} else {
					values := strings.Split(ss, "`")
					if len(values) >= 1 {
						for _, x := range values {
							if len(x) > 1 {
								l.SLICE(append(*l.SLICE(), x))
							}
						}
					}
				}
				return true
			}
			return false
		},
	}
	return &o
}

func Tag(s string) *Line {
	o := Line{
		Name: s,
		Validate: func(l *Line, si interface{}) bool {
			s := maybeStringPointer(si)
			if si == nil {
				return false
			}
			l.STRING(strings.TrimSpace(s))
			return true
		},
	}
	return &o
}

func File(s string) *Line {
	out := Line{
		Name: s,
		Validate: func(l *Line, si interface{}) bool {
			if si == nil {
				return false
			}
			s := maybeStringPointer(si)
			if len(s) > 0 {
				if !strings.HasPrefix(s, "/") &&
					!strings.HasPrefix(s, ".") &&
					runtime.GOOS != "windows" {
					s = filepath.Join(DataDir, s)
				}
				s := CleanAndExpandPath(s)
				if s == "." {
					s = ""
				}
				l.STRING(s)
			}
			return true
		},
	}
	return &out
}

func Dir(s string) *Line {
	out := Line{
		Name: s,
		Validate: func(l *Line, si interface{}) bool {
			if si == nil {
				return false
			}
			s := maybeStringPointer(si)
			if len(s) > 0 {
				if !strings.HasPrefix(s, "/") &&
					!strings.HasPrefix(s, ".") &&
					runtime.GOOS != "windows" {
					s = filepath.Join(DataDir, s)
				}
				s := CleanAndExpandPath(s)
				if s == "." {
					s = ""
				}
				l.STRING(s)
				return true
			}
			return false
		},
	}
	return &out
}

func Port(s string) *Line {
	o := Line{Name: s}
	return &o
}

func Addr(s string) *Line {
	o := Line{Name: s}
	return &o
}

func Addrs(s string) *Line {
	o := Line{Name: s}
	return &o
}

func Algo(s string) *Line {
	o := Line{Name: s}
	return &o
}

func Net(s string) (o *Line) {
	o = &Line{
		Name: s,
		Validate: func(l *Line, si interface{}) bool {
			s := maybeStringPointer(si)
			for _, x := range Networks {
				if x == s {
					fork.IsTestnet = false
					switch s {
					case "testnet":
						tn, sn, rn = true, false, false
						// activenetparams = &node.TestNet3Params
						fork.IsTestnet = true
					case "simnet":
						tn, sn, rn = false, true, false
						// activenetparams = &node.SimNetParams
					case "regtestnet":
						tn, sn, rn = false, false, true
						// activenetparams = &node.RegressionNetParams
					default:
						s = "mainnet"
						tn, sn, rn = false, false, false
						// activenetparams = &node.MainNetParams
					}
					log <- cl.Info{"running on", s}
					l.STRING(s)
					return true
				}
			}
			fmt.Println("returning invalid")
			return false
		},
	}
	return o
}

func bv(l *Line, si interface{}) bool {
	if si == nil {
		return false
	}
	switch strings.ToLower(*si.(*string)) {
	case "true":
		l.BOOL(true)
	case "false":
		l.BOOL(false)
	default:
		return false
	}
	return true
}

// Enable is a boolean that defaults to false
func Enable(s string) *Line {
	g := false
	o := Line{Name: s, Validate: bv, Value: &g}
	return &o
}

// Enabled is a boolean that defaults to true
func Enabled(s string) (o *Line) {
	o = &Line{Name: s, Validate: bv}
	o.BOOL(true)
	return &Line{Name: s}
}

// Default sets a default value for the Line
func (l *Line) Default(d interface{}) (out *Line) {
	// spew.Dump(l)
	if l == nil {
		fmt.Println("empty *line")
		return &Line{}
	}
	fmt.Println(l.Validate(l, d))
	fmt.Println(*l.Value.(*string))
	return l
}

// Usage is the short text explaining a configuration option
func (l *Line) Usage(s string) *Line {
	spew.Dump(l)
	if l != nil {
		// All lines *should* have a Usage and it *should* be last so validate!
		l.Comment = s
	}
	return l
}

// Min is chained to validate at initialisation
func (l *Line) Min(i int) *Line {
	v := l.Validate
	if v == nil {
		v = func(*Line, interface{}) bool { return true }
	}
	l.Validate = func(*Line, interface{}) bool {
		if *l.INT() < i {
			l.INT(i)
		}
		return v(l, l.INT())
	}
	return l
}

// Max is chained to validate at initialisation
func (l *Line) Max(i int) *Line {
	v := l.Validate
	if v == nil {
		v = func(*Line, interface{}) bool { return true }
	}
	l.Validate = func(*Line, interface{}) bool {
		if l != nil {
			fmt.Println(l, l.INT)
			if *l.INT() < i {
				l.INT(i)
			}
			return v(l, l.INT())
		}
		return false
	}
	return l
}
