package config

import "github.com/davecgh/go-spew/spew"

func (r *ConfigGenerators) RunAll(ctx *Configuration) {
	R := *r
	for _, x := range R {
		x(ctx)
	}
}

func App(name string, a ...ConfigGenerator) *Configuration {
	A := ConfigGenerators(a)
	cfg := &Configuration{
		Name:       name,
		Categories: make(Categories),
	}
	A.RunAll(cfg)
	spew.Dump(A)
	return cfg
}

func Version(version string) ConfigGenerator {
	return func(ctx *Configuration) {
		ctx.Version =
			func() string {
				return version
			}
	}
}
