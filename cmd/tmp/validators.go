package config

import (
	"fmt"
	"net"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var DataDir string

// GenAddr returns a validator with a set default port assumed if one is not present
func GenAddr(name string, port int) func(r *Row, in interface{}) bool {
	return func(r *Row, in interface{}) bool {
		var s string
		switch I := in.(type) {
		case string:
			s = I
		case *string:
			s = *I
		default:
			return false
		}
		_, _, err := net.SplitHostPort(s)
		if err != nil {
			s = net.JoinHostPort(s, fmt.Sprint(port))
		}
		if r != nil {
			r.Value = s
		}
		return true
	}
}

// GenAddrs returns a validator with a set default port assumed if one is not present
func GenAddrs(name string, port int) func(r *Row, in interface{}) bool {
	return func(r *Row, in interface{}) bool {
		var s string
		isString := false
		switch I := in.(type) {
		case string:
			s = I
			isString = true
		case *string:
			s = *I
			isString = true
		case []string:
		case *[]string:
		default:
			return false
		}
		if isString {
			_ = s
		}
		_ = s
		if r != nil {

		}
		return true
	}
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
				runtime.GOOS != "windows" {
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
				runtime.GOOS != "windows" {
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
			r.Value = &ii
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
			r.Value = &b
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
			r.Value = ii
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
			r.Value = &s
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
			r.Value = ss
		}
		return true
	},
	// TODO: write generators for these to set default port
	// Addr: func(r *Row, in interface{}) bool {
	// 	var s string
	// 	switch I := in.(type) {
	// 	case string:
	// 		s = I
	// 	case *string:
	// 		s = *I
	// 	default:
	// 		return false
	// 	}
	// 	_, _, err := net.SplitHostPort(s)
	// 	if err != nil {
	// 		s = net.JoinHostPort(s, fmt.Sprint(defPort))
	// 	}
	// 	if r != nil {
	// 		r.Value = s
	// 	}
	// 	return true
	// },
	// Addrs: func(r *Row, in interface{}) bool {
	// 	var s string
	// 	isString := false
	// 	switch I := in.(type) {
	// 	case string:
	// 		s = I
	// 		isString = true
	// 	case *string:
	// 		s = *I
	// 		isString = true
	// 	case []string:
	// 	case *[]string:
	// 	default:
	// 		return false
	// 	}
	// 	if isString {
	// 		_ = s
	// 	} else {

	// 	}
	// 	_ = s
	// 	if r != nil {

	// 	}
	// 	return true
	// },
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
		_ = s
		if r != nil {

		}
		return true
	},
	Float: func(r *Row, in interface{}) bool {
		var s string
		isString := false
		switch I := in.(type) {
		case string:
			s = I
			isString = true
		case *string:
			s = *I
			isString = true
		case float64:
		case *float64:
		default:
			return false
		}
		if isString {
			_ = s
		} else {

		}
		_ = s
		if r != nil {

		}
		return true
	},
	Duration: func(r *Row, in interface{}) bool {
		var s string
		isString := false
		switch I := in.(type) {
		case string:
			s = I
			isString = true
		case *string:
			s = *I
			isString = true
		case time.Duration:
		case *time.Duration:
		default:
			return false
		}
		if isString {
			_ = s
		} else {

		}
		_ = s
		if r != nil {

		}
		return true
	},
	Net: func(r *Row, in interface{}) bool {
		var s string
		switch I := in.(type) {
		case string:
			s = I
		case *string:
			s = *I
		default:
			return false
		}
		_ = s
		if r != nil {

		}
		return true
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
		_ = s
		if r != nil {

		}
		return true
	},
}
