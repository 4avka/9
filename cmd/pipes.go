package cmd

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"git.parallelcoin.io/dev/9/cmd/node"
	"git.parallelcoin.io/dev/9/pkg/chain/fork"
	"git.parallelcoin.io/dev/9/pkg/util/cl"
)

func (l *Lines) Group(s string, items ...*Line) (out *Lines) {
	ll := make(Lines)
	out = &ll
	for i, x := range items {
		if x == nil || x.Name == "" {
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
	return
}

func Int(s string) *Line {
	return &Line{
		Name: s,
		Validate: func(l *Line, si interface{}) bool {
			s := si.(string)
			n, e := strconv.Atoi(s)
			if e != nil {
				l.INT(n)
			}
			return true
		},
	}
}

func Float(s string) *Line {
	return &Line{Name: s}
}

func Duration(s string) *Line {
	return &Line{Name: s}
}

func Log(s string) *Line {
	return &Line{
		Name: s,
		Validate: func(l *Line, si interface{}) bool {
			s := si.(string)
			for x := range cl.Levels {
				if x == s {
					l.STRING(s)
					return true
				}
			}
			return false
		},
	}
}

func Tags(s string) *Line {
	return &Line{
		Name: s,
		Validate: func(l *Line, si interface{}) bool {
			if si != nil {
				s := strings.TrimSpace(si.(string))
				if len(s) < 1 {
					l.SLICE([]string{})
				} else {
					values := strings.Split(s, "`")
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
}

func Tag(s string) *Line {
	return &Line{
		Name: s,
		Validate: func(l *Line, si interface{}) bool {
			if si == nil {
				return false
			}
			l.STRING(strings.TrimSpace(si.(string)))
			return true
		},
	}
}

func File(s string) *Line {
	return &Line{
		Name: s,
		Validate: func(l *Line, si interface{}) bool {
			if si == nil {
				return false
			}
			s := si.(string)
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
}

func Dir(s string) *Line {
	return &Line{
		Name: s,
		Validate: func(l *Line, si interface{}) bool {
			if si == nil {
				return false
			}
			s := si.(string)
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
}

func Port(s string) *Line {
	return &Line{}
}

func Addr(s string) *Line {
	return &Line{}
}

func Addrs(s string) *Line {
	return &Line{}
}

func Algo(s string) *Line {
	return &Line{}
}

func Net(s string) *Line {
	return &Line{
		Name: s,
		Validate: func(l *Line, si interface{}) bool {
			s := si.(string)
			for _, x := range Networks {
				if x == s {
					fork.IsTestnet = false
					switch s {
					case "testnet":
						tn, sn, rn = true, false, false
						activenetparams = &node.TestNet3Params
						fork.IsTestnet = true
					case "simnet":
						tn, sn, rn = false, true, false
						activenetparams = &node.SimNetParams
					case "regtestnet":
						tn, sn, rn = false, false, true
						activenetparams = &node.RegressionNetParams
					default:
						s = "mainnet"
						tn, sn, rn = false, false, false
						activenetparams = &node.MainNetParams
					}
					log <- cl.Info{"running on", s}
					return true
				}
			}
			return false
		},
	}
}

func bv(l *Line, si interface{}) bool {
	if si == nil {
		return false
	}
	switch strings.ToLower(si.(string)) {
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
func Enable(s string) (o *Line) {
	o = &Line{Name: s, Validate: bv}
	o.BOOL(false)
	return o
}

// Enabled is a boolean that defaults to true
func Enabled(s string) (o *Line) {
	o = &Line{Name: s, Validate: bv}
	o.BOOL(true)
	return &Line{Name: s}
}

// Default sets a default value for the Line
func (l *Line) Default(d interface{}) (out *Line) {
	if l == nil || l.Value == nil {
		return &Line{}
	}
	_ = l.Validate(l, d)
	return
}

// Usage is the short text explaining a configuration option
func (l *Line) Usage(s string) *Line {
	if l != nil && l.Value != nil {
		// All lines *should* have a Usage and it *should* be last so validate!
		_ = l.Validate(l, l.Initial)
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
