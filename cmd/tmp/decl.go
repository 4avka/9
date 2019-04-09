package config

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"
)

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

func File(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Validate = Valid.File
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
			cc.Validate = Valid.Dir
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
			cc.Validate = Valid.Port
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

func boolRow(name string, enabled bool, g RowGenerators) CatGenerator {
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Get = func() interface{} {
				return cc.Value
			}
			cc.Validate = Valid.Bool
			cc.Value = &enabled
			cc.Put = func(in interface{}) bool {
				return cc.Validate(cc, in)
			}
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
			cc.Get = func() interface{} {
				return cc.Value
			}
			cc.Validate = Valid.Int
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
			cc.Get = func() interface{} {
				return cc.Value
			}
			cc.Validate = Valid.Tag
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
			cc.Get = func() interface{} {
				return cc.Value
			}
			cc.Validate = Valid.Tags
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
			cc.Get = func() interface{} {
				return cc.Value
			}
			cc.Validate = Valid.Level
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
			cc.Get = func() interface{} {
				return cc.Value
			}
			cc.Validate = Valid.Algo
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
			cc.Get = func() interface{} {
				return cc.Value
			}
			cc.Validate = Valid.Float
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
			cc.Get = func() interface{} {
				return cc.Value
			}
			cc.Validate = Valid.Duration
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
			cc.Get = func() interface{} {
				return cc.Value
			}
			cc.Validate = Valid.Net
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
	var ii interface{}
	return func(ctx *Row) {
		switch I := in.(type) {
		case string:
			ii = &I
		case []string:
			ii = &I
		case int:
			ii = &I
		case float64:
			ii = &I
		case bool:
			ii = &I
		case time.Duration:
			ii = &I
		case *string:
			ii = I
		case *[]string:
			ii = I
		case *int:
			ii = I
		case *float64:
			ii = I
		case *bool:
			ii = I
		case *time.Duration:
			ii = I
		default:
			fmt.Println(ctx.Name, reflect.TypeOf(in))
			return
		}
		if !ctx.Validate(ctx, ii) {
			fmt.Println(ctx.Name, "fail validate", reflect.TypeOf(ii), ii)
		}
	}
}

// Min attaches to the validator a test that enforces a minimum
func Min(min int) RowGenerator {
	return func(ctx *Row) {
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
		v := ctx.Validate
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
		ctx.Value = &sb
	}
}
