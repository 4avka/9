package config

import (
	"fmt"
	"net"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"git.parallelcoin.io/dev/9/cmd/nine"

	"git.parallelcoin.io/dev/9/pkg/chain/fork"
	"git.parallelcoin.io/dev/9/pkg/util/cl"
)

var DataDir string
var Networks = []string{"mainnet", "testnet", "simnet", "regtestnet"}
var NetParams = map[string]*nine.Params{
	"mainnet":    &nine.MainNetParams,
	"testnet":    &nine.TestNet3Params,
	"simnet":     &nine.SimNetParams,
	"regtestnet": &nine.RegressionNetParams,
}

// GenAddr returns a validator with a set default port assumed if one is not present
func GenAddr(name string, port int) func(r *Row, in interface{}) bool {
	return func(r *Row, in interface{}) bool {
		var s *string
		switch I := in.(type) {
		case string:
			s = &I
		case *string:
			s = I
		case nil:
			r.Value.Put(nil)
			return true
		default:
			return false
		}
		if s == nil {
			r.Value.Put(nil)
			return true
		}
		if *s == "" {
			s = nil
			r.Value.Put(nil)
			return true
		}
		h, p, err := net.SplitHostPort(*s)
		if err != nil {
			*s = net.JoinHostPort(*s, fmt.Sprint(port))
		} else {
			n, e := strconv.Atoi(p)
			if e == nil {
				if n < 1025 || n > 65535 {
					return false
				}
			} else {
				return false
				// p = ""
			}
			if p == "" {
				p = fmt.Sprint(port)
				*s = net.JoinHostPort(h, p)
			} else {
				*s = net.JoinHostPort(h, p)
			}
		}
		if r != nil {
			r.Value = r.Value.Put(*s)
			r.String = *s
		}
		return true
	}
}

// GenAddrs returns a validator with a set default port assumed if one is not present
func GenAddrs(name string, port int) func(r *Row, in interface{}) bool {
	return func(r *Row, in interface{}) bool {
		var s string
		ss := []string{}
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
			sss := strings.Split(s, " ")
			if len(sss) < 1 {
				return false
			}
			for _, x := range sss {
				if len(x) > 0 {
					ss = append(ss, x)
				}
			}
		}
		if ss != nil {
			if r != nil {
				r.Value = r.Value.Put(ss)
			}
			r.String = fmt.Sprint(ss)
		}
		return true
	}
}

func getAlgoOptions() (options []string) {
	var modernd = "random"
	for _, x := range fork.P9AlgoVers {
		options = append(options, x)
	}
	options = append(options, modernd)
	sort.Strings(options)
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
			if r != nil && ss != "." {
				r.Value = r.Value.Put(ss)
				r.String = fmt.Sprint(ss)
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
				r.String = fmt.Sprint(ss)
				r.Value = r.Value.Put(ss)
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
			r.Value = r.Value.Put(ii)
			r.String = fmt.Sprint(ii)
		}
		return true
	},
	Bool: func(r *Row, in interface{}) bool {
		var sb string
		var b bool
		switch I := in.(type) {
		case string:
			sb = I
			if strings.ToUpper(sb) == "TRUE" {
				b = true
				goto boolout
			}
			if strings.ToUpper(sb) == "FALSE" {
				b = false
				goto boolout
			}
		case *string:
			sb = *I
			if strings.ToUpper(sb) == "TRUE" {
				b = true
				goto boolout
			}
			if strings.ToUpper(sb) == "FALSE" {
				b = false
				goto boolout
			}
		case bool:
			b = I
		case *bool:
			b = *I
		default:
			return false
		}
	boolout:
		if r != nil {
			r.String = fmt.Sprint(b)
			r.Value = r.Value.Put(b)
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
			r.String = fmt.Sprint(ii)
			r.Value = r.Value.Put(ii)
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
			r.Value = r.Value.Put(s)
			r.String = fmt.Sprint(s)
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
			r.Value = r.Value.Put(ss)
			r.String = fmt.Sprint(ss)
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
		var o string
		options := getAlgoOptions()
		for _, x := range options {
			if s == x {
				o = s
			}
		}
		if o == "" {
			rnd := "random"
			o = rnd
		}
		if r != nil {
			r.String = fmt.Sprint(o)
			r.Value = r.Value.Put(o)
		}
		return true
	},
	Float: func(r *Row, in interface{}) bool {
		var s string
		var f float64
		isString := false
		switch I := in.(type) {
		case string:
			s = I
			isString = true
		case *string:
			s = *I
			isString = true
		case float64:
			f = I
		case *float64:
			f = *I
		default:
			return false
		}
		if isString {
			ff, e := strconv.ParseFloat(s, 64)
			if e != nil {
				return false
			}
			f = ff
		}
		if r != nil {
			r.Value = r.Value.Put(f)
			r.String = fmt.Sprint(f)
		}
		return true
	},
	Duration: func(r *Row, in interface{}) bool {
		var s string
		var t time.Duration
		isString := false
		switch I := in.(type) {
		case string:
			s = I
			isString = true
		case *string:
			s = *I
			isString = true
		case time.Duration:
			t = I
		case *time.Duration:
			t = *I
		default:
			return false
		}
		if isString {
			dd, e := time.ParseDuration(s)
			if e != nil {
				return false
			}
			t = dd
		}
		if r != nil {
			r.String = fmt.Sprint(t)
			r.Value = r.Value.Put(t)
		}
		return true
	},
	Net: func(r *Row, in interface{}) bool {
		var sn string
		switch I := in.(type) {
		case string:
			sn = I
		case *string:
			sn = *I
		default:
			return false
		}
		found := false
		for _, x := range Networks {
			if x == sn {
				found = true
				*nine.ActiveNetParams = *NetParams[x]
			}
		}
		if r != nil && found {
			r.String = fmt.Sprint(sn)
			r.Value = r.Value.Put(sn)
		}
		return found
	},
	Level: func(r *Row, in interface{}) bool {
		var sl string
		switch I := in.(type) {
		case string:
			sl = I
		case *string:
			sl = *I
		default:
			return false
		}
		found := false
		for x := range cl.Levels {
			if x == sl {
				found = true
			}
		}
		if r != nil && found {
			r.String = fmt.Sprint(sl)
			r.Value = r.Value.Put(sl)
		}
		return found
	},
}
