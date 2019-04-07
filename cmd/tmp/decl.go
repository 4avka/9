package config

func NewApp(name string, g ...AppGenerator) (out *App) {
	gen := AppGenerators(g)
	out = &App{}
	gen.RunAll(out)
	return
}

func Version(ver string) AppGenerator {
	return func(ctx *App) {
		ctx.Version = func() string {
			return ver
		}
	}
}

func Group(name string, g ...CatGenerator) AppGenerator {
	return func(ctx *App) {
		// TODO make new cat run generators on it
	}
}
