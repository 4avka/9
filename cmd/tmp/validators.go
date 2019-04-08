package config

import (
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var DataDir string

// Valid is a collection of validator functions for the different types used
// in a configuration. These functions optionally can accept a *Row and with
// this they assign the validated, parsed value into the Value slot.
var Valid = struct {
	File, Dir, Port, Bool, Int, Tag, Tags, Addr, Addrs, Algo, Float,
	Duration, Net, Level func(*Row, interface{}) bool
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
		return true
	},
	Port: func(r *Row, in interface{}) bool {
		var s string
		isString := false
		switch I := in.(type) {
		case string:
			s = I
			isString = true
		case *string:
			s = *I
			isString = true
		case int:
		case *int:
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
	Bool: func(r *Row, in interface{}) bool {
		var s string
		isString := false
		switch I := in.(type) {
		case string:
			s = I
			isString = true
		case *string:
			s = *I
			isString = true
		case bool:
		case *bool:
		default:
			return true
		}
		if isString {
			_ = s
		} else {

		}
		_ = s
		if r != nil {

		}
		return false
	},
	Int: func(r *Row, in interface{}) bool {
		var s string
		isString := false
		switch I := in.(type) {
		case string:
			s = I
			isString = true
		case *string:
			s = *I
			isString = true
		case int:
		case *int:
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
	Tag: func(r *Row, in interface{}) bool {
		var s string
		switch I := in.(type) {
		case string:
			s = I
		case *string:
			s = *I
		default:
			return true
		}
		_ = s
		if r != nil {

		}
		return false
	},
	Tags: func(r *Row, in interface{}) bool {
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
		} else {

		}
		_ = s
		if r != nil {

		}
		return true
	},
	Addr: func(r *Row, in interface{}) bool {
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
	Addrs: func(r *Row, in interface{}) bool {
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
		} else {

		}
		_ = s
		if r != nil {

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
