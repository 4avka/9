package config

var Valid = struct {
	File, Dir, Port, Bool, Int, Tag, Tags, Addr, Addrs, Algo, Float,
	Duration, Net, Level func(*Row, interface{}) bool
}{
	File: func(*Row, interface{}) bool {
		return false
	},
	Dir: func(*Row, interface{}) bool {
		return false
	},
	Port: func(*Row, interface{}) bool {
		return false
	},
	Bool: func(*Row, interface{}) bool {
		return false
	},
	Int: func(*Row, interface{}) bool {
		return false
	},
	Tag: func(*Row, interface{}) bool {
		return false
	},
	Tags: func(*Row, interface{}) bool {
		return false
	},
	Addr: func(*Row, interface{}) bool {
		return false
	},
	Addrs: func(*Row, interface{}) bool {
		return false
	},
	Algo: func(*Row, interface{}) bool {
		return false
	},
	Float: func(*Row, interface{}) bool {
		return false
	},
	Duration: func(*Row, interface{}) bool {
		return false
	},
	Net: func(*Row, interface{}) bool {
		return false
	},
	Level: func(*Row, interface{}) bool {
		return false
	},
}
