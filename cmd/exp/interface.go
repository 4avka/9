package exp

type Variable interface {
	Name()
	Init()
	Get() interface{}
	Put(interface{}) bool
	Validate(interface{}) bool
}

type Row interface {
}
