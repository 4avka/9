package config

func Group(name string, g ...RowGenerator) ConfigGenerator {
	return func(ctx *Configuration) {
		C := *ctx
		if c, ok := C.Categories[name]; ok {
			panic("category " + name + "already exists")
		} else {
			G := RowGenerators(g)
			G.RunAll(c)
			C.Categories[name] = c
		}
	}
}
