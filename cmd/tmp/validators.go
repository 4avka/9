package config

import "time"

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
		}
		_ = s
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
		}
		_ = s
		return false
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
		}
		if isString {
			_ = s
		} else {

		}
		_ = s
		return false
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
		}
		if isString {
			_ = s
		} else {

		}
		_ = s
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
		}
		if isString {
			_ = s
		} else {

		}
		_ = s
		return false
	},
	Tag: func(r *Row, in interface{}) bool {
		var s string
		switch I := in.(type) {
		case string:
			s = I
		case *string:
			s = *I
		default:
		}
		_ = s
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
		}
		if isString {
			_ = s
		} else {

		}
		_ = s
		return false
	},
	Addr: func(r *Row, in interface{}) bool {
		var s string
		switch I := in.(type) {
		case string:
			s = I
		case *string:
			s = *I
		default:
		}
		_ = s
		return false
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
		}
		if isString {
			_ = s
		} else {

		}
		_ = s
		return false
	},
	Algo: func(r *Row, in interface{}) bool {
		var s string
		switch I := in.(type) {
		case string:
			s = I
		case *string:
			s = *I
		default:
		}
		_ = s
		return false
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
		}
		if isString {
			_ = s
		} else {

		}
		_ = s
		return false
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
		}
		if isString {
			_ = s
		} else {

		}
		_ = s
		return false
	},
	Net: func(r *Row, in interface{}) bool {
		var s string
		switch I := in.(type) {
		case string:
			s = I
		case *string:
			s = *I
		default:
		}
		_ = s
		return false
	},
	Level: func(r *Row, in interface{}) bool {
		var s string
		switch I := in.(type) {
		case string:
			s = I
		case *string:
			s = *I
		default:
		}
		_ = s
		return false
	},
}
