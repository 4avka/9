package config

type ConfigGenerator func(ctx *Configuration)
type ConfigGenerators []ConfigGenerator

func (r *ConfigGenerators) RunAll(ctx *Configuration) {
	R := *r
	for _, x := range R {
		x(ctx)
	}
}

func App(name string, g ...ConfigGenerator) (c *Configuration) {
	c = &Configuration{
		Name: name,
		Rows: make(Rows),
	}
	G := ConfigGenerators(g)
	G.RunAll(c)
	c.InitAll()
	return
}

func Group(name string, g ...ConfigGenerator) ConfigGenerator {
	return func(ctx *Configuration) {
		for _, x := range g {
			x(ctx)
		}
	}
}

func Version(version string) func(*Configuration) {
	return func(ctx *Configuration) {
		ctx.Version =
			func() string {
				return version
			}
	}
}

func File(name string, g ...RowGenerator) ConfigGenerator {
	return func(ctx *Configuration) {
		f := Inits["file"]
		ctx.Rows[name] = &Row{
			Name:     name,
			Validate: f.Validator,
			Get:      f.Getter,
			Put:      f.Putter,
		}
		G := RowGenerators(g)
		ctx.Rows[name].Init = func() {
			G.RunAll(ctx.Rows[name])
		}
	}
}

func Dir(name string, g ...ConfigGenerator) ConfigGenerator {
	return func(ctx *Configuration) {
		// TODO: validator and accessors
		ctx.Rows[name] = &Row{Name: name}
		G := ConfigGenerators(g)
		G.RunAll(ctx)
	}
}

func Enable(name string, g ...ConfigGenerator) ConfigGenerator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {}
}

func Enabled(name string, g ...ConfigGenerator) ConfigGenerator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {}
}

func Int(name string, g ...ConfigGenerator) ConfigGenerator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {}
}

func Float(name string, g ...ConfigGenerator) ConfigGenerator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {}
}

func Tag(name string, g ...ConfigGenerator) ConfigGenerator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {}
}

func Tags(name string, g ...ConfigGenerator) ConfigGenerator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {}
}

func Addr(name string, g ...ConfigGenerator) ConfigGenerator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {}
}

func Addrs(name string, g ...ConfigGenerator) ConfigGenerator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {}
}

func Level(g ...ConfigGenerator) ConfigGenerator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {}
}

func Algo(g ...ConfigGenerator) ConfigGenerator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {}
}

func Duration(name string, g ...ConfigGenerator) ConfigGenerator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {}
}

func Net(name string, g ...ConfigGenerator) ConfigGenerator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {}
}
