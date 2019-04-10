package config

import (
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"git.parallelcoin.io/dev/9/pkg/chain/fork"
	"git.parallelcoin.io/dev/9/pkg/util/cl"
)

var DataDir string
var Networks = []string{"mainnet", "testnet", "simnet", "regtestnet"}

// GenAddr returns a validator with a set default port assumed if one is not present
func GenAddr(name string, port int) func(r *Row, in interface{}) bool {
	return func(r *Row, in interface{}) bool {
		var s *string
		switch I := in.(type) {
		case string:
			s = &I
		case *string:
			s = I
		default:
			return false
		}
		_, _, err := net.SplitHostPort(*s)
		if err != nil {
			*s = net.JoinHostPort(*s, fmt.Sprint(port))
		}
		if r != nil {
			if r.Value != nil {
				*r.Value.(*string) = *s
			} else {
				r.Value = s
			}
		}
		return true
	}
}

// GenAddrs returns a validator with a set default port assumed if one is not present
func GenAddrs(name string, port int) func(r *Row, in interface{}) bool {
	return func(r *Row, in interface{}) bool {
		var s string
		ss := &[]string{}
		isString := false
		switch I := in.(type) {
		case string:
			s = I
			isString = true
		case *string:
			s = *I
			isString = true
		case []string:
			ss = &I
		case *[]string:
			ss = I
		default:
			return false
		}
		if isString {
			sss := strings.Split(s, " ")
			if len(sss) < 1 {
				return false
			}
			for _, x := range sss {
				if len(x) > 0 {
					*ss = append(*ss, x)
				}
			}
		}
		if ss != nil {
			if r != nil {
				if r.Value != nil {
					*r.Value.(*[]string) = *ss
				} else {
					r.Value = ss
				}
			}
			return true
		}
		return false
	}
}

func getAlgoOptions() (options []string) {
	var modernd = "random"
	for _, x := range fork.P9AlgoVers {
		options = append(options, x)
	}
	options = append(options, modernd)
	return
}

// Valid is a collection of validator functions for the different types used
// in a configuration. These functions optionally can accept a *Row and with
// this they assign the validated, parsed value into the Value slot.
var Valid = struct {
	File, Dir, Port, Bool, Int, Tag, Tags, Algo, Float, Duration, Net,
	Level func(*Row, interface{}) bool
}{
	File: func(r *Row, in interface{}) bool {
		var s string
		switch I := in.(type) {
		case string:
			s = I
		case *string:
			s = *I
		default:
			return false
		}
		if len(s) > 0 {
			if !strings.HasPrefix(s, "/") && !strings.HasPrefix(s, ".") &&
				isWindows() {
				s = filepath.Join(DataDir, s)
			}
			ss := CleanAndExpandPath(s)
			if ss == "." {
				ss = ""
			}
			if r != nil {
				r.Value = &ss
			}
			return true
		}
		return false
	},
	Dir: func(r *Row, in interface{}) bool {
		var s string
		switch I := in.(type) {
		case string:
			s = I
		case *string:
			s = *I
		default:
			return false
		}
		if len(s) > 0 {
			if !strings.HasPrefix(s, "/") && !strings.HasPrefix(s, ".") &&
				isWindows() {
				s = filepath.Join(DataDir, s)
			}
			ss := CleanAndExpandPath(s)
			if ss == "." {
				ss = ""
			}
			if r != nil {
				if r.Value != nil {
					*r.Value.(*string) = ss
				} else {
					r.Value = &ss
				}
			}
			return true
		}
		return false
	},
	Port: func(r *Row, in interface{}) bool {
		var s string
		var ii int
		isString := false
		switch I := in.(type) {
		case string:
			s = I
			isString = true
		case *string:
			s = *I
			isString = true
		case int:
			ii = I
		case *int:
			ii = *I
		default:
			return false
		}
		if isString {
			n, e := strconv.Atoi(s)
			if e != nil {
				return false
			}
			ii = n
		}
		if ii < 1025 || ii > 65535 {
			return false
		}
		if r != nil {
			if r.Value != nil {
				*r.Value.(*int) = ii
			} else {
				r.Value = &ii
			}
		}
		return false
	},
	Bool: func(r *Row, in interface{}) bool {
		var s string
		var b bool
		isString := false
		switch I := in.(type) {
		case string:
			s = I
			isString = true
		case *string:
			s = *I
			isString = true
		case bool:
			b = I
		case *bool:
			b = *I
		default:
			return false
		}
		if isString {
			if strings.ToUpper(s) == "TRUE" {
				b = true
				goto boolout
			}
			if strings.ToUpper(s) == "FALSE" {
				b = false
				goto boolout
			}
			return false
		}
	boolout:
		if r != nil {
			if r.Value != nil {
				*r.Value.(*bool) = b
			} else {
				r.Value = &b
			}
		}
		return true
	},
	Int: func(r *Row, in interface{}) bool {
		var s string
		var ii int
		isString := false
		switch I := in.(type) {
		case string:
			s = I
			isString = true
		case *string:
			s = *I
			isString = true
		case int:
			ii = I
		case *int:
			ii = *I
		default:
			return false
		}
		if isString {
			n, e := strconv.Atoi(s)
			if e != nil {
				return false
			}
			ii = n
		}
		if r != nil {
			if r.Value != nil {
				*r.Value.(*int) = ii
			} else {
				r.Value = &ii
			}
		}
		return true
	},
	Tag: func(r *Row, in interface{}) bool {
		var s string
		switch I := in.(type) {
		case string:
			s = I
		case *string:
			s = *I
		default:
			return false
		}
		s = strings.TrimSpace(s)
		if len(s) < 1 {
			return false
		}
		if r != nil {
			if r.Value != nil {
				*r.Value.(*string) = s
			} else {
				r.Value = &s
			}
		}
		return true
	},
	Tags: func(r *Row, in interface{}) bool {
		var s string
		var ss []string
		isString := false
		switch I := in.(type) {
		case string:
			s = I
			isString = true
		case *string:
			s = *I
			isString = true
		case []string:
			ss = I
		case *[]string:
			ss = *I
		default:
			return false
		}
		if isString {
			s = strings.TrimSpace(s)
			sss := strings.Split(s, " ")
			var ssss []string
			for _, x := range sss {
				if len(x) > 0 {
					ssss = append(ssss, x)
				}
			}
			if len(ssss) < 1 {
				return false
			}
			ss = ssss
		}
		if r != nil {
			if r.Value != nil {
				*r.Value.(*[]string) = ss
			} else {
				r.Value = ss
			}
		}
		return true
	},
	Algo: func(r *Row, in interface{}) bool {
		var s string
		switch I := in.(type) {
		case string:
			s = I
		case *string:
			s = *I
		default:
			return false
		}
		var o *string
		options := getAlgoOptions()
		for _, x := range options {
			if s == x {
				o = &s
			}
		}
		if o == nil {
			rnd := "random"
			o = &rnd
		}
		if r != nil {
			if r.Value != nil {
				*r.Value.(*string) = *o
			} else {
				r.Value = o
			}
		}
		return true
	},
	Float: func(r *Row, in interface{}) bool {
		var s string
		var f *float64
		isString := false
		switch I := in.(type) {
		case string:
			s = I
			isString = true
		case *string:
			s = *I
			isString = true
		case float64:
			f = &I
		case *float64:
			f = I
		default:
			return false
		}
		if isString {
			ff, e := strconv.ParseFloat(s, 64)
			if e != nil {
				return false
			}
			f = &ff
		}
		_ = s
		if r != nil {
			if r.Value != nil {
				*r.Value.(*float64) = *f
			} else {
				r.Value = f
			}
		}
		return true
	},
	Duration: func(r *Row, in interface{}) bool {
		var s string
		var t *time.Duration
		isString := false
		switch I := in.(type) {
		case string:
			s = I
			isString = true
		case *string:
			s = *I
			isString = true
		case time.Duration:
			t = &I
		case *time.Duration:
			t = I
		default:
			return false
		}
		if isString {
			dd, e := time.ParseDuration(s)
			if e != nil {
				return false
			}
			t = &dd
		}
		if r != nil {
			if r.Value != nil {
				*r.Value.(*time.Duration) = *t
			} else {
				r.Value = t
			}
		}
		return true
	},
	Net: func(r *Row, in interface{}) bool {
		var s *string
		switch I := in.(type) {
		case string:
			s = &I
		case *string:
			s = I
		default:
			return false
		}
		found := false
		for _, x := range Networks {
			if x == *s {
				found = true
			}
		}
		if r != nil && found {
			if r.Value != nil {
				*r.Value.(*string) = *s
			} else {
				r.Value = s
			}
		}
		return found
	},
	Level: func(r *Row, in interface{}) bool {
		var s string
		switch I := in.(type) {
		case string:
			s = I
		case *string:
			s = *I
		default:
			return false
		}
		found := false
		for x := range cl.Levels {
			if x == s {
				found = true
			}
		}
		if r != nil && found {
			if r.Value != nil {
				*r.Value.(*string) = s
			} else {
				r.Value = &s
			}
		}
		return found
	},
}
