package config

type Configger interface {
	Get(path string) interface{}
	Put(data interface{}, path string) bool
}

type Lines map[string]*Line

type Configuration struct {
	Lines
	Name    string
	Version func() string
}

func (r *Configuration) Get(path string) interface{} {
	if _, ok := r.Lines[path]; ok {
		return r.Lines[path].Getter()
	}
	return nil
}

func (r *Configuration) Put(in interface{}, path string) bool {
	if _, ok := r.Lines[path]; ok {
		return r.Lines[path].Putter(in)
	}
	return false
}

type Line struct {
	Name     string
	Init     func()
	Getter   func() interface{}
	Putter   func(interface{}) bool
	Validate func(*Line, interface{}) bool
	Usage    string
}

type Generator func(ctx *Configuration)

type Generators []Generator

func (r *Generators) RunAll(ctx *Configuration) {
	R := *r
	for _, x := range R {
		x(ctx)
	}
}

func App(name string, g ...Generator) (c *Configuration) {
	c = &Configuration{
		Name: name,
	}
	G := Generators(g)
	G.RunAll(c)
	return
}

func Group(name string, g ...Generator) Generator {
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

func File(name string, g ...Generator) Generator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {
		ctx.Lines[name] = &Line{Name: name}
		G := Generators(g)
		G.RunAll(ctx)
	}
}

func Dir(name string, g ...Generator) Generator {
	return func(ctx *Configuration) {
		// TODO: validator and accessors
		ctx.Lines[name] = &Line{Name: name}
		G := Generators(g)
		G.RunAll(ctx)
	}
}

func Enable(name string, g ...Generator) Generator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {}
}

func Enabled(name string, g ...Generator) Generator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {}
}

func Int(name string, g ...Generator) Generator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {}
}

func Float(name string, g ...Generator) Generator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {}
}

func Tag(name string, g ...Generator) Generator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {}
}

func Tags(name string, g ...Generator) Generator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {}
}

func Addr(name string, g ...Generator) Generator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {}
}

func Addrs(name string, g ...Generator) Generator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {}
}

func Level(g ...Generator) Generator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {}
}

func Algo(g ...Generator) Generator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {}
}

func Duration(name string, g ...Generator) Generator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {}
}

func Net(name string, g ...Generator) Generator {
	// TODO: validator and accessors
	return func(ctx *Configuration) {}
}
