package config

type RowGenerator func(ctx *Row)
type RowGenerators []RowGenerator

func (r *RowGenerators) RunAll(ctx *Row) {
	R := *r
	for _, x := range R {
		x(ctx)
	}
}

func Default(in interface{}) RowGenerator {
	return func(ctx *Row) {
		ctx.Value = in
	}
}

func Min(in interface{}) RowGenerator {
	return func(ctx *Row) {}
}
func Max(in interface{}) RowGenerator {
	return func(ctx *Row) {}
}
func Usage(usage string) RowGenerator {
	return func(ctx *Row) {
		ctx.Usage = usage
	}
}
func RandomString(length int) RowGenerator {
	return func(ctx *Row) {}
}
func DefaultPort(port int) RowGenerator {
	return func(ctx *Row) {}
}
