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

type Rows map[string]*Row

type Categories map[string]*Rows
type CategoryGenerator func(*Categories)

type Configuration struct {
	Categories
	Name    string
	Version func() string
}

type ConfigGenerator func(ctx *Configuration)

type ConfigGenerators []ConfigGenerator
