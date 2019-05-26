package def

import (
	"time"

	"git.parallelcoin.io/dev/9/pkg/ifc"
)

// Row is a configuration variable
type Row struct {
	Name     string
	Type     string
	Opts     []string
	Value    *ifc.Iface
	Default  *ifc.Iface
	Min      *ifc.Iface
	Max      *ifc.Iface
	Init     func(*Row)
	Get      func() interface{}
	Put      func(interface{}) bool
	Validate func(*Row, interface{}) bool
	String   string
	Usage    string
	App      *App
}

// RowGenerator configures a Row
type RowGenerator func(ctx *Row)

// RowGenerators is a collection of Rows
type RowGenerators []RowGenerator

// Bool returns the content of a Row that contains a Bool
func (r *Row) Bool() bool {
	return r.Value.Get().(bool)
}

// Int returns the content of a Row that contains a Int
func (r *Row) Int() int {
	return r.Value.Get().(int)
}

// Float returns the content of a Row that contains a Float
func (r *Row) Float() float64 {
	return r.Value.Get().(float64)
}

// Duration returns the content of a Row that contains a Duration
func (r *Row) Duration() time.Duration {
	return r.Value.Get().(time.Duration)
}

// Tag returns the content of a Row that contains a Tag
func (r *Row) Tag() string {
	return r.Value.Get().(string)
}

// Tags returns the content of a Row that contains a Tags
func (r *Row) Tags() []string {
	return r.Value.Get().([]string)
}

// RunAll executes the generators in a RowGenerators slice
func (r *RowGenerators) RunAll(row *Row) {
	for _, x := range *r {
		x(row)
	}
}