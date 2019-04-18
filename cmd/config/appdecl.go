package config

import (
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"time"

	"git.parallelcoin.io/dev/9/pkg/util/cl"
)

func NewApp(name string, g ...AppGenerator) (out *App) {
	gen := AppGenerators(g)
	out = &App{
		Name:     name,
		Cats:     make(Cats),
		Commands: make(Commands),
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

func Tagline(ver string) AppGenerator {
	return func(ctx *App) {
		ctx.Tagline = ver
	}
}

func About(ver string) AppGenerator {
	return func(ctx *App) {
		ctx.About = ver
	}
}

func DefaultRunner(fn func(ctx *App) int) AppGenerator {
	return func(ctx *App) {
		ctx.Default = fn
	}
}

func Group(name string, g ...CatGenerator) AppGenerator {
	G := CatGenerators(g)
	return func(ctx *App) {
		ctx.Cats[name] = make(Cat)
		G.RunAll(ctx.Cats[name])
	}
}

func Cmd(name string, g ...CommandGenerator) AppGenerator {
	G := CommandGenerators(g)
	return func(ctx *App) {
		ctx.Commands[name] = G.RunAll()
	}
}

// Command Item Generators

func Pattern(patt string) CommandGenerator {
	return func(ctx *Command) {
		ctx.Pattern = patt
		ctx.RE = regexp.MustCompile(ctx.Pattern)
	}
}

func Short(usage string) CommandGenerator {
	return func(ctx *Command) {
		ctx.Short = usage
	}
}

func Detail(usage string) CommandGenerator {
	return func(ctx *Command) {
		ctx.Detail = usage
	}
}

func Opts(opts ...string) CommandGenerator {
	return func(ctx *Command) {
		ctx.Opts = opts
	}
}

func Precs(precs ...string) CommandGenerator {
	return func(ctx *Command) {
		ctx.Precedent = precs
	}
}

func Handler(hnd func(args []string, tokens Tokens, app *App) int) CommandGenerator {
	return func(ctx *Command) {
		ctx.Handler = hnd
	}
}

// Group Item Generators

func File(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Type = "string"
			cc.Validate = Valid.File
			cc.Value = new(interface{})
			cc.Get = func() interface{} {
				return cc.Value
			}
			cc.Put = func(in interface{}) bool {
				return cc.Validate(cc, in)
			}
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
			cc.Type = "string"
			cc.Validate = Valid.Dir
			cc.Value = new(interface{})
			cc.Get = func() interface{} {
				return cc.Value
			}
			cc.Put = func(in interface{}) bool {
				return cc.Validate(cc, in)
			}
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
			cc.Type = "port"
			cc.Validate = Valid.Port
			cc.Value = new(interface{})
			cc.Get = func() interface{} {
				return cc.Value
			}
			cc.Put = func(in interface{}) bool {
				return cc.Validate(cc, in)
			}
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = *c
	}
}

func Enable(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Type = "bool"
			cc.Get = func() interface{} {
				return cc.Value
			}
			cc.Validate = Valid.Bool
			cc.Value = new(interface{})
			*cc.Value = false
			cc.Default = new(interface{})
			*cc.Default = false
			cc.Put = func(in interface{}) bool {
				return cc.Validate(cc, in)
			}
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = *c
	}
}

func Enabled(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Type = "bool"
			cc.Get = func() interface{} {
				return cc.Value
			}
			cc.Validate = Valid.Bool
			cc.Value = new(interface{})
			*cc.Value = true
			cc.Default = new(interface{})
			*cc.Default = true
			cc.Put = func(in interface{}) bool {
				return cc.Validate(cc, in)
			}
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = *c
	}
}

func Int(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Type = "int"
			cc.Get = func() interface{} {
				return cc.Value
			}
			cc.Validate = Valid.Int
			cc.Value = new(interface{})
			cc.Put = func(in interface{}) bool {
				return cc.Validate(cc, in)
			}
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
			cc.Type = "string"
			cc.Get = func() interface{} {
				return cc.Value
			}
			cc.Validate = Valid.Tag
			cc.Value = new(interface{})
			cc.Put = func(in interface{}) bool {
				return cc.Validate(cc, in)
			}
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
			cc.Type = "stringslice"
			cc.Get = func() interface{} {
				return cc.Value
			}
			cc.Validate = Valid.Tags
			cc.Value = new(interface{})
			cc.Put = func(in interface{}) bool {
				return cc.Validate(cc, in)
			}
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = *c
	}
}

func Addr(name string, defPort int, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Type = "string"
			cc.Get = func() interface{} {
				return cc.Value
			}
			cc.Validate = GenAddr(name, defPort)
			cc.Put = func(in interface{}) bool {
				return cc.Validate(cc, in)
			}
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = *c
	}
}

func Addrs(name string, defPort int, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Type = "stringslice"
			cc.Get = func() interface{} {
				return cc.Value
			}
			cc.Validate = GenAddrs(name, defPort)
			cc.Put = func(in interface{}) bool {
				return cc.Validate(cc, in)
			}
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
			cc.Type = "options"
			cc.Opts = cl.GetLevelOpts()
			cc.Get = func() interface{} {
				return cc.Value
			}
			cc.Validate = Valid.Level
			cc.Value = new(interface{})
			cc.Put = func(in interface{}) bool {
				return cc.Validate(cc, in)
			}
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
			cc.Type = "options"
			cc.Opts = getAlgoOptions()
			cc.Get = func() interface{} {
				return cc.Value
			}
			cc.Validate = Valid.Algo
			cc.Value = new(interface{})
			cc.Put = func(in interface{}) bool {
				return cc.Validate(cc, in)
			}
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
			cc.Type = "float"
			cc.Get = func() interface{} {
				return cc.Value
			}
			cc.Validate = Valid.Float
			cc.Value = new(interface{})
			cc.Put = func(in interface{}) bool {
				return cc.Validate(cc, in)
			}
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
			cc.Type = "duration"
			cc.Get = func() interface{} {
				return cc.Value
			}
			cc.Validate = Valid.Duration
			cc.Value = new(interface{})
			cc.Put = func(in interface{}) bool {
				return cc.Validate(cc, in)
			}
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
			cc.Type = "options"
			cc.Opts = Networks
			cc.Get = func() interface{} {
				return cc.Value
			}
			cc.Validate = Valid.Net
			cc.Value = new(interface{})
			cc.Put = func(in interface{}) bool {
				return cc.Validate(cc, in)
			}
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = *c
	}
}

// which is populated by

// Usage populates the usage field for information about a config item
func Usage(usage string) RowGenerator {
	return func(ctx *Row) {
		ctx.Usage = usage
	}
}

// Default sets the default value for a config item
func Default(in interface{}) RowGenerator {
	return func(ctx *Row) {
		var iii interface{}
		ctx.Default = &iii
		switch I := in.(type) {
		case string:
			*ctx.Default = I
			ctx.Validate(ctx, I)
		case []string:
			*ctx.Default = I
			ctx.Validate(ctx, I)
		case int:
			*ctx.Default = I
			ctx.Validate(ctx, I)
		case float64:
			*ctx.Default = I
			ctx.Validate(ctx, I)
		case bool:
			*ctx.Default = I
			ctx.Validate(ctx, I)
		case time.Duration:
			*ctx.Default = I
			ctx.Validate(ctx, I)
		case *string:
			*ctx.Default = I
			ctx.Validate(ctx, I)
		case *[]string:
			*ctx.Default = I
			ctx.Validate(ctx, I)
		case *int:
			*ctx.Default = I
			ctx.Validate(ctx, I)
		case *float64:
			*ctx.Default = I
			ctx.Validate(ctx, I)
		case *bool:
			*ctx.Default = I
			ctx.Validate(ctx, I)
		case *time.Duration:
			*ctx.Default = I
			ctx.Validate(ctx, I)
		default:
			fmt.Println("type not found", ctx.Name, reflect.TypeOf(in))
			return
		}
	}
}

// Min attaches to the validator a test that enforces a minimum
func Min(min int) RowGenerator {
	return func(ctx *Row) {
		ctx.Min = new(interface{})
		*ctx.Min = min
		v := ctx.Validate
		ctx.Validate = func(r *Row, in interface{}) bool {
			n := min
			switch I := in.(type) {
			case int:
				n = I
			case *int:
				n = *I
			}
			if n < min {
				in = min
			}
			// none of the above will affect if this wasn't an int
			return v(r, in)
		}
	}
}

// Max attaches to the validator a test that enforces a maximum
func Max(max int) RowGenerator {
	return func(ctx *Row) {
		ctx.Max = new(interface{})
		v := ctx.Validate
		*ctx.Max = &max
		ctx.Validate = func(r *Row, in interface{}) bool {
			n := max
			switch I := in.(type) {
			case int:
				n = I
			case *int:
				n = *I
			}
			if n > max {
				in = max
			}
			// none of the above will affect if this wasn't an int
			return v(r, in)
		}
	}
}

// RandomsString generates a random number and converts to base32 for
// a default random password of some number of characters
func RandomString(n int) RowGenerator {
	const (
		letterBytes   = "abcdefghijklmnopqrstuvwxyz234567"
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)
	var src = rand.NewSource(time.Now().UnixNano())
	return func(ctx *Row) {
		b := make([]byte, n)
		l := len(letterBytes)
		// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
		for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
			if remain == 0 {
				cache, remain = src.Int63(), letterIdxMax
			}
			if idx := int(cache & letterIdxMask); idx < l {
				b[i] = letterBytes[idx]
				i--
			}
			cache >>= letterIdxBits
			remain--
		}

		sb := string(b)
		ctx.Value = new(interface{})
		*ctx.Value = sb
	}
}
