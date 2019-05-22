package app
import (
	"fmt"
	"math/rand"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"time"
	"git.parallelcoin.io/dev/9/cmd/def"
	"git.parallelcoin.io/dev/9/cmd/nine"
	"git.parallelcoin.io/dev/9/pkg/ifc"
	"git.parallelcoin.io/dev/9/pkg/util"
	"git.parallelcoin.io/dev/9/pkg/util/cl"
)
// DataDir is the folder all servers and apps in this suite use to store
// configuration and working data
var DataDir = filepath.Dir(util.AppDataDir("9", false))
// Networks is the list of network types the node and wallet can connect to
var Networks = []string{"mainnet", "testnet", "simnet", "regtestnet"}
// NetParams stores the information required to set the parameters for the network
var NetParams = map[string]*nine.Params{
	"mainnet":    &nine.MainNetParams,
	"testnet":    &nine.TestNet3Params,
	"simnet":     &nine.SimNetParams,
	"regtestnet": &nine.RegressionNetParams,
}
// NewApp generates a new App using a collection of generator functions passed to it
func NewApp(name string, g ...def.AppGenerator) (out *def.App) {
	gen := def.AppGenerators(g)
	out = &def.App{
		Name:     name,
		Cats:     make(def.Cats),
		Commands: make(def.Commands),
	}
	gen.RunAll(out)
	// set ref to App in each def.Row
	for _, x := range out.Cats {
		for _, y := range x {
			y.App = out
		}
	}
	return
}
// which is made from
// Version fills the Version field of an App
func Version(ver string) def.AppGenerator {
	return func(ctx *def.App) {
		ctx.Version = func() string {
			return ver
		}
	}
}
// Tagline is a short text describing the application
func Tagline(ver string) def.AppGenerator {
	return func(ctx *def.App) {
		ctx.Tagline = ver
	}
}
// About is a longer text describing the application
func About(ver string) def.AppGenerator {
	return func(ctx *def.App) {
		ctx.About = ver
	}
}
// DefaultRunner is the command that runs when no parameters are given
func DefaultRunner(fn func(ctx *def.App) int) def.AppGenerator {
	return func(ctx *def.App) {
		ctx.Default = fn
	}
}
// Group is a collection of categories and bundles each category
func Group(name string, g ...def.CatGenerator) def.AppGenerator {
	G := def.CatGenerators(g)
	return func(ctx *def.App) {
		ctx.Cats[name] = make(def.Cat)
		G.RunAll(ctx.Cats[name])
	}
}
// Cmd is a collection of subcommands
func Cmd(name string, g ...def.CommandGenerator) def.AppGenerator {
	G := def.CommandGenerators(g)
	return func(ctx *def.App) {
		ctx.Commands[name] = G.RunAll()
	}
}
// def.Command Item Generators
// Pattern is the regular expression pattern that matches the CLI args for each
// def.Command item
func Pattern(patt string) def.CommandGenerator {
	return func(ctx *def.Command) {
		ctx.Pattern = patt
		// Panic if RE is not correct - only the programmer affects these and they must work
		ctx.RE = regexp.MustCompile(ctx.Pattern)
	}
}
// Short is the short help text for a def.Command
func Short(usage string) def.CommandGenerator {
	return func(ctx *def.Command) {
		ctx.Short = usage
	}
}
// Detail is the long help text for a def.Command
func Detail(usage string) def.CommandGenerator {
	return func(ctx *def.Command) {
		ctx.Detail = usage
	}
}
// Opts is the collection of valid options for a def.Command
func Opts(opts ...string) def.CommandGenerator {
	return func(ctx *def.Command) {
		ctx.Opts = opts
	}
}
// Precs is the collection of tags for items that are preferentially selected when
// an two or more items are specified (for example, help always overrides everything)
func Precs(precs ...string) def.CommandGenerator {
	return func(ctx *def.Command) {
		ctx.Precedent = precs
	}
}
// Handler is the function that is called when a command is selected
func Handler(hnd func(args []string, tokens def.Tokens, app *def.App) int) def.CommandGenerator {
	return func(ctx *def.Command) {
		ctx.Handler = hnd
	}
}
// Group Item Generators
// File is an item storing a filename
func File(name string, g ...def.RowGenerator) def.CatGenerator {
	G := def.RowGenerators(g)
	return func(ctx *def.Cat) {
		c := &def.Row{}
		c.Init = func(cc *def.Row) {
			cc.Name = name
			cc.Type = "string"
			cc.Validate = Valid.File
			cc.Value = ifc.NewIface()
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Put = func(in interface{}) bool {
				valid := cc.Validate(cc, in)
				if valid {
					// cc.Value =
					cc.Value.Put(in)
					return true
				}
				return false
			}
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = c
	}
}
// Dir is an item storing a directory specification
func Dir(name string, g ...def.RowGenerator) def.CatGenerator {
	G := def.RowGenerators(g)
	return func(ctx *def.Cat) {
		c := &def.Row{}
		c.Init = func(cc *def.Row) {
			cc.Name = name
			cc.Type = "string"
			cc.Validate = Valid.Dir
			cc.Value = ifc.NewIface()
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Put = func(in interface{}) bool {
				valid := cc.Validate(cc, in)
				if valid {
					// cc.Value =
					cc.Value.Put(in)
					return true
				}
				return false
			}
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = c
	}
}
// Port is a 16 bit sized number that represents a network port number
func Port(name string, g ...def.RowGenerator) def.CatGenerator {
	G := def.RowGenerators(g)
	return func(ctx *def.Cat) {
		c := &def.Row{}
		c.Init = func(cc *def.Row) {
			cc.Name = name
			cc.Type = "port"
			cc.Validate = Valid.Port
			cc.Value = ifc.NewIface()
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Put = func(in interface{}) bool {
				valid := cc.Validate(cc, in)
				if valid {
					cc.Value = cc.Value.Put(in)
					return true
				}
				return false
			}
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = c
	}
}
// Enable is a boolean item that is false by default
func Enable(name string, g ...def.RowGenerator) def.CatGenerator {
	G := def.RowGenerators(g)
	return func(ctx *def.Cat) {
		c := &def.Row{}
		c.Init = func(cc *def.Row) {
			cc.Name = name
			cc.Type = "bool"
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Validate = Valid.Bool
			cc.Value = ifc.NewIface().Put(false)
			cc.Default = ifc.NewIface().Put(false)
			cc.Put = func(in interface{}) bool {
				valid := cc.Validate(cc, in)
				if valid {
					cc.Value = cc.Value.Put(in)
					return true
				}
				return false
			}
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = c
	}
}
// Enabled is an boolean item that is true by default
func Enabled(name string, g ...def.RowGenerator) def.CatGenerator {
	G := def.RowGenerators(g)
	return func(ctx *def.Cat) {
		c := &def.Row{}
		c.Init = func(cc *def.Row) {
			cc.Name = name
			cc.Type = "bool"
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Validate = Valid.Bool
			cc.Value = ifc.NewIface().Put(true)
			cc.Default = ifc.NewIface().Put(true)
			cc.Put = func(in interface{}) bool {
				valid := cc.Validate(cc, in)
				if valid {
					cc.Value = cc.Value.Put(in)
					return true
				}
				return false
			}
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = c
	}
}
// Int stores an integer number (signed integer width of current platform's processor)
func Int(name string, g ...def.RowGenerator) def.CatGenerator {
	G := def.RowGenerators(g)
	return func(ctx *def.Cat) {
		c := &def.Row{}
		c.Init = func(cc *def.Row) {
			cc.Name = name
			cc.Type = "int"
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Validate = Valid.Int
			cc.Value = ifc.NewIface()
			cc.Put = func(in interface{}) bool {
				valid := cc.Validate(cc, in)
				if valid {
					cc.Value = cc.Value.Put(in)
				}
				return valid
			}
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = c
	}
}
// Tag is basically just a string that can store any string value
func Tag(name string, g ...def.RowGenerator) def.CatGenerator {
	G := def.RowGenerators(g)
	return func(ctx *def.Cat) {
		c := &def.Row{}
		c.Init = func(cc *def.Row) {
			cc.Name = name
			cc.Type = "string"
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Validate = Valid.Tag
			cc.Value = ifc.NewIface()
			cc.Put = func(in interface{}) bool {
				valid := cc.Validate(cc, in)
				if valid {
					cc.Value = cc.Value.Put(in)
				}
				return valid
			}
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = c
	}
}
// Tags is a collection of strings containing arbitrary content
func Tags(name string, g ...def.RowGenerator) def.CatGenerator {
	G := def.RowGenerators(g)
	return func(ctx *def.Cat) {
		c := &def.Row{}
		c.Init = func(cc *def.Row) {
			cc.Name = name
			cc.Type = "stringslice"
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Validate = Valid.Tags
			cc.Value = ifc.NewIface()
			cc.Put = func(in interface{}) bool {
				valid := cc.Validate(cc, in)
				if valid {
					cc.Value = cc.Value.Put(in)
				}
				return valid
			}
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = c
	}
}
// Addr is a network address specification for dialing
func Addr(name string, defPort int, g ...def.RowGenerator) def.CatGenerator {
	G := def.RowGenerators(g)
	return func(ctx *def.Cat) {
		c := &def.Row{}
		c.Init = func(cc *def.Row) {
			cc.Name = name
			cc.Type = "string"
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Validate = GenAddr(name, defPort)
			cc.Value = ifc.NewIface()
			cc.Put = func(in interface{}) bool {
				valid := cc.Validate(cc, in)
				if valid {
					cc.Value = cc.Value.Put(in)
				}
				return valid
			}
			cc.Usage = fmt.Sprintf(
				"\n\nNOTE: port must be between 1025-65535, port %d will be assumed if no port is given",
				defPort)
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = c
	}
}
// Addrs is a collection of addresses
func Addrs(name string, defPort int, g ...def.RowGenerator) def.CatGenerator {
	G := def.RowGenerators(g)
	return func(ctx *def.Cat) {
		c := &def.Row{}
		c.Init = func(cc *def.Row) {
			cc.Name = name
			cc.Type = "stringslice"
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Validate = GenAddrs(name, defPort)
			cc.Value = ifc.NewIface()
			cc.Put = func(in interface{}) bool {
				valid := cc.Validate(cc, in)
				if valid {
					cc.Value = cc.Value.Put(in)
				}
				return valid
			}
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = c
	}
}
// Level is debug logging level specification one of a set of predefined values
func Level(g ...def.RowGenerator) def.CatGenerator {
	G := def.RowGenerators(g)
	const lvl = "level"
	return func(ctx *def.Cat) {
		c := &def.Row{}
		c.Init = func(cc *def.Row) {
			cc.Name = lvl
			cc.Type = "options"
			cc.Opts = cl.GetLevelOpts()
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Validate = Valid.Level
			cc.Value = ifc.NewIface()
			cc.Put = func(in interface{}) bool {
				valid := cc.Validate(cc, in)
				if valid {
					cc.Value = cc.Value.Put(in)
				}
				return valid
			}
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[lvl] = c
	}
}
// Algo is a multi-item select that contains all of the different block-algorithms
// available to mine with, as well as algorithmic selectors
func Algo(name string, g ...def.RowGenerator) def.CatGenerator {
	G := def.RowGenerators(g)
	return func(ctx *def.Cat) {
		c := &def.Row{}
		c.Init = func(cc *def.Row) {
			cc.Name = name
			cc.Type = "options"
			cc.Opts = getAlgoOptions()
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Validate = Valid.Algo
			cc.Value = ifc.NewIface()
			cc.Put = func(in interface{}) bool {
				valid := cc.Validate(cc, in)
				if valid {
					cc.Value = cc.Value.Put(in)
				}
				return valid
			}
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = c
	}
}
// Float is a floating point number, 64 bits by default (same as JSON spec)
func Float(name string, g ...def.RowGenerator) def.CatGenerator {
	G := def.RowGenerators(g)
	return func(ctx *def.Cat) {
		c := &def.Row{}
		c.Init = func(cc *def.Row) {
			cc.Name = name
			cc.Type = "float"
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Validate = Valid.Float
			cc.Value = ifc.NewIface()
			cc.Put = func(in interface{}) bool {
				valid := cc.Validate(cc, in)
				if valid {
					cc.Value = cc.Value.Put(in)
				}
				return valid
			}
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = c
	}
}
// Duration is a time library duration specification for an amount of time.
// The value is a 64 bit integer being the number of nanoseconds for a period of time
func Duration(name string, g ...def.RowGenerator) def.CatGenerator {
	G := def.RowGenerators(g)
	return func(ctx *def.Cat) {
		c := &def.Row{}
		c.Init = func(cc *def.Row) {
			cc.Name = name
			cc.Type = "duration"
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Validate = Valid.Duration
			cc.Value = ifc.NewIface()
			cc.Put = func(in interface{}) bool {
				valid := cc.Validate(cc, in)
				if valid {
					cc.Value = cc.Value.Put(in)
				}
				return valid
			}
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = c
	}
}
// Net is the various types of network a blockchain server can connect to - test, main
// and so forth
func Net(name string, g ...def.RowGenerator) def.CatGenerator {
	G := def.RowGenerators(g)
	return func(ctx *def.Cat) {
		c := &def.Row{}
		c.Init = func(cc *def.Row) {
			cc.Name = name
			cc.Type = "options"
			cc.Opts = Networks
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Validate = Valid.Net
			cc.Value = ifc.NewIface()
			cc.Put = func(in interface{}) bool {
				valid := cc.Validate(cc, in)
				if valid {
					cc.Value = cc.Value.Put(in)
				}
				return valid
			}
			G.RunAll(cc)
		}
		c.Init(c)
		(*ctx)[name] = c
	}
}
// which is populated by
// Usage populates the usage field for information about a config item
func Usage(usage string) def.RowGenerator {
	return func(ctx *def.Row) {
		ctx.Usage = usage + " " + ctx.Usage
	}
}
// Default sets the default value for a config item
func Default(in interface{}) def.RowGenerator {
	return func(ctx *def.Row) {
		ctx.Default = ifc.NewIface()
		switch I := in.(type) {
		case string:
			if ctx.Validate(ctx, I) {
				ctx.Default.Put(I)
			}
		case []string:
			if ctx.Validate(ctx, I) {
				ctx.Default.Put(I)
			}
		case int:
			if ctx.Validate(ctx, I) {
				ctx.Default.Put(I)
			}
		case float64:
			if ctx.Validate(ctx, I) {
				ctx.Default.Put(I)
			}
		case bool:
			if ctx.Validate(ctx, I) {
				ctx.Default.Put(I)
			}
		case time.Duration:
			if ctx.Validate(ctx, I) {
				ctx.Default.Put(I)
			}
		case *string:
			if ctx.Validate(ctx, I) {
				ctx.Default.Put(I)
			}
		case *[]string:
			if ctx.Validate(ctx, I) {
				ctx.Default.Put(I)
			}
		case *int:
			if ctx.Validate(ctx, I) {
				ctx.Default.Put(I)
			}
		case *float64:
			if ctx.Validate(ctx, I) {
				ctx.Default.Put(I)
			}
		case *bool:
			if ctx.Validate(ctx, I) {
				ctx.Default.Put(I)
			}
		case *time.Duration:
			if ctx.Validate(ctx, I) {
				ctx.Default.Put(I)
			}
		case nil:
		default:
			fmt.Println("type not found", ctx.Name, reflect.TypeOf(in))
			return
		}
		// ctx.Value.Put(nil)
	}
}
// Min attaches to the validator a test that enforces a minimum
func Min(min int) def.RowGenerator {
	return func(ctx *def.Row) {
		ctx.Min = ctx.Min.Put(min)
		v := ctx.Validate
		var e error
		ctx.Validate = func(r *def.Row, in interface{}) bool {
			n := min
			switch I := in.(type) {
			case int:
				n = I
			case *int:
				n = *I
			case string:
				n, e = strconv.Atoi(I)
				if e != nil {
					return false
				}
			case *string:
				n, e = strconv.Atoi(*I)
				if e != nil {
					return false
				}
			}
			if n < min {
				return false
				// in = min
			}
			// none of the above will affect if this wasn't an int
			return v(r, in)
		}
	}
}
// Max attaches to the validator a test that enforces a maximum
func Max(max int) def.RowGenerator {
	return func(ctx *def.Row) {
		ctx.Max = ctx.Max.Put(max)
		v := ctx.Validate
		var e error
		ctx.Validate = func(r *def.Row, in interface{}) bool {
			n := max
			switch I := in.(type) {
			case int:
				n = I
			case *int:
				n = *I
			case string:
				n, e = strconv.Atoi(I)
				if e != nil {
					return false
				}
			case *string:
				n, e = strconv.Atoi(*I)
				if e != nil {
					return false
				}
			}
			if n > max {
				return false
				// in = max
			}
			// none of the above will affect if this wasn't an int
			return v(r, in)
		}
	}
}
// RandomString generates a random number and converts to base32 for
// a default random password of some number of characters
func RandomString(n int) def.RowGenerator {
	const (
		letterBytes   = "abcdefghijklmnopqrstuvwxyz234567"
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)
	var src = rand.NewSource(time.Now().UnixNano())
	return func(ctx *def.Row) {
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
		ctx.Value = ctx.Value.Put(sb)
	}
}
