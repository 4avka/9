package config

type Row struct {
	Name     string
	Value    interface{}
	Init     func()
	Get      func() interface{}
	Put      func(interface{}) bool
	Validate func(*Row, interface{}) bool
	Usage    string
}

type RowGenerator func(ctx *Row)

type RowGenerators []RowGenerator

func (r *RowGenerators) RunAll(row *Row) {
	var init func()
	for _, x := range *r {
		// TODO construct Init() and then run after this block
		init = row.Init
		row.Init = func() {
			x(row)
			init()
		}
	}
	init()
	return
}
