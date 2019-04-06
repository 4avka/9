package config

type RowGenerator func(ctx *Row)
type RowGenerators []RowGenerator

func Default(in interface{}) RowGenerator {
	return func(ctx *Row) {}
}

func Min(in interface{}) RowGenerator {
	return func(ctx *Row) {}
}
func Max(in interface{}) RowGenerator {
	return func(ctx *Row) {}
}
func Usage(usage string) RowGenerator {
	return func(ctx *Row) {}
}
func RandomString(length int) RowGenerator {
	return func(ctx *Row) {}
}
func DefaultPort(port int) RowGenerator {
	return func(ctx *Row) {}
}
