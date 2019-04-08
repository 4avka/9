package config

func NewApp(name string, g ...AppGenerator) (out *App) {
	gen := AppGenerators(g)
	out = &App{
		Name: name,
		Cats: make(Cats),
	}
	gen.RunAll(out)
	return
}

// which is made from

func Version(ver string) AppGenerator {
	return func(ctx *App) {
		ctx.Version = func() string {
			return ver
		}
	}
}

func Group(name string, g ...CatGenerator) AppGenerator {
	return func(ctx *App) {
		ctx.Cats[name] = make(Cat)
		G := CatGenerators(g)
		G.RunAll(ctx.Cats[name])
	}
}

// which is made from

func File(name string, g ...RowGenerator) CatGenerator {
	return func(ctx *Cat) {
		C := *ctx
		c := &Row{Name: name}
		G := RowGenerators(g)
		G.RunAll(c)
		C[name] = *c
	}
}

// which is populated by

func Usage(usage string) RowGenerator {
	return func(ctx *Row) {
		ctx.Usage = usage
	}
}
