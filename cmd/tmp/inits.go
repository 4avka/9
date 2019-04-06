package config

type Init struct {
	Init      func()
	Validator func(*Row, interface{}) bool
	Getter    func() interface{}
	Putter    func(interface{}) bool
}

var Inits = map[string]Init{
	"file": Init{
		Validator: func(*Row, interface{}) bool {
			return true
		},
		Getter: func() interface{} { return nil },
		Putter: func(interface{}) bool { return false },
	},
}
