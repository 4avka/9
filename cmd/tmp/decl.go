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
	G := CatGenerators(g)
	return func(ctx *App) {
		ctx.Cats[name] = make(Cat)
		G.RunAll(ctx.Cats[name])
	}
}

// which is made from

// TODO: these need to attach validator/accessors

func File(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Validate = Valid.File
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = *c
	}
}

func Dir(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Validate = Valid.Dir
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = *c
	}
}

func Port(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Validate = Valid.Port
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = *c
	}
}

func boolRow(name string, enabled bool, g RowGenerators) CatGenerator {
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Validate = Valid.Bool
			cc.Value = &enabled
			g.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = *c
	}
}

func Enable(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return boolRow(name, false, G)
}

func Enabled(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return boolRow(name, true, G)
}

func Int(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Validate = Valid.Int
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = *c
	}
}

func Tag(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Validate = Valid.Tag
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = *c
	}
}

func Tags(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Validate = Valid.Tags
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = *c
	}
}

func Addr(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Validate = Valid.Addr
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = *c
	}
}

func Level(g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	const lvl = "level"
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = lvl
			cc.Validate = Valid.Level
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[lvl] = *c
	}
}

func Algo(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Validate = Valid.Algo
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = *c
	}
}

func Float(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Validate = Valid.Float
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = *c
	}
}

func Duration(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Validate = Valid.Duration
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = *c
	}
}
func Addrs(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Validate = Valid.Addrs
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = *c
	}
}

func Net(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Validate = Valid.Net
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = *c
	}
}

// which is populated by

func Usage(usage string) RowGenerator {
	return func(ctx *Row) {
		ctx.Usage = usage
	}
}
