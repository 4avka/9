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

// TODO: these need to make and use inits and attach validator/accossors

func File(name string, g ...RowGenerator) CatGenerator {
	return func(ctx *Cat) {
		C := *ctx
		c := &Row{Name: name}
		G := RowGenerators(g)
		G.RunAll(c)
		C[name] = *c
	}
}

func Dir(name string, g ...RowGenerator) CatGenerator {
	return func(ctx *Cat) {
		C := *ctx
		c := &Row{Name: name}
		G := RowGenerators(g)
		G.RunAll(c)
		C[name] = *c
	}
}

func Port(name string, g ...RowGenerator) CatGenerator {
	return func(ctx *Cat) {
		C := *ctx
		c := &Row{Name: name}
		G := RowGenerators(g)
		G.RunAll(c)
		C[name] = *c
	}
}

func Enable(name string, g ...RowGenerator) CatGenerator {
	enabled := false
	return func(ctx *Cat) {
		C := *ctx
		c := &Row{Name: name, Value: &enabled}
		G := RowGenerators(g)
		G.RunAll(c)
		C[name] = *c
	}
}

func Enabled(name string, g ...RowGenerator) CatGenerator {
	enabled := true
	return func(ctx *Cat) {
		C := *ctx
		c := &Row{Name: name, Value: &enabled}
		G := RowGenerators(g)
		G.RunAll(c)
		C[name] = *c
	}
}

func Int(name string, g ...RowGenerator) CatGenerator {
	return func(ctx *Cat) {
		C := *ctx
		c := &Row{Name: name}
		G := RowGenerators(g)
		G.RunAll(c)
		C[name] = *c
	}
}

func Tag(name string, g ...RowGenerator) CatGenerator {
	return func(ctx *Cat) {
		C := *ctx
		c := &Row{Name: name}
		G := RowGenerators(g)
		G.RunAll(c)
		C[name] = *c
	}
}

func Tags(name string, g ...RowGenerator) CatGenerator {
	return func(ctx *Cat) {
		C := *ctx
		c := &Row{Name: name}
		G := RowGenerators(g)
		G.RunAll(c)
		C[name] = *c
	}
}

func Addr(name string, g ...RowGenerator) CatGenerator {
	return func(ctx *Cat) {
		C := *ctx
		c := &Row{Name: name}
		G := RowGenerators(g)
		G.RunAll(c)
		C[name] = *c
	}
}

func Level(g ...RowGenerator) CatGenerator {
	const lvl = "level"
	return func(ctx *Cat) {
		C := *ctx
		c := &Row{Name: lvl}
		G := RowGenerators(g)
		G.RunAll(c)
		C[lvl] = *c
	}
}

func Algo(name string, g ...RowGenerator) CatGenerator {
	const lvl = "level"
	return func(ctx *Cat) {
		C := *ctx
		c := &Row{Name: name}
		G := RowGenerators(g)
		G.RunAll(c)
		C[lvl] = *c
	}
}

func Float(name string, g ...RowGenerator) CatGenerator {
	const lvl = "level"
	return func(ctx *Cat) {
		C := *ctx
		c := &Row{Name: name}
		G := RowGenerators(g)
		G.RunAll(c)
		C[lvl] = *c
	}
}

func Duration(name string, g ...RowGenerator) CatGenerator {
	const lvl = "level"
	return func(ctx *Cat) {
		C := *ctx
		c := &Row{Name: name}
		G := RowGenerators(g)
		G.RunAll(c)
		C[lvl] = *c
	}
}
func Addrs(name string, g ...RowGenerator) CatGenerator {
	return func(ctx *Cat) {
		C := *ctx
		c := &Row{Name: name}
		G := RowGenerators(g)
		G.RunAll(c)
		C[name] = *c
	}
}

func Net(name string, g ...RowGenerator) CatGenerator {
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
