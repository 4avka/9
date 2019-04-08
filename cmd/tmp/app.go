package config

type Cats map[string]Cat

type App struct {
	Name    string
	Usage   string
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
