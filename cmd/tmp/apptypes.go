package config

type App struct {
	Name    string
	Version func() string
	Cats    Cats
}
type AppGenerator func(ctx *App)
type AppGenerators []AppGenerator

func (r *AppGenerators) RunAll(app *App) {
	for _, x := range *r {
		x(app)
	}
	return
}

type Cats map[string]Cat

type Cat map[string]Row
type CatGenerator func(ctx *Cat)
type CatGenerators []CatGenerator

func (r *CatGenerators) RunAll(cat Cat) {
	for _, x := range *r {
		x(&cat)
	}
	return
}

type Row struct {
	Name     string
	Value    interface{}
	Init     func(*Row)
	Get      func() interface{}
	Put      func(interface{}) bool
	Validate func(*Row, interface{}) bool
	Usage    string
}
type RowGenerator func(ctx *Row)
type RowGenerators []RowGenerator

func (r *RowGenerators) RunAll(row *Row) {
	for _, x := range *r {
		x(row)
	}
}
