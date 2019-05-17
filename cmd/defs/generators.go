package defs

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"git.parallelcoin.io/dev/9/cmd/nine"
	"git.parallelcoin.io/dev/9/pkg/util"
	"git.parallelcoin.io/dev/9/pkg/util/cl"
)

// DataDir is the folder all servers and apps in this suite use to store
// configuration and working data
var DataDir string = filepath.Dir(util.AppDataDir("9", false))

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
func NewApp(name string, g ...AppGenerator) (out *App) {
	gen := AppGenerators(g)
	out = &App{
		Name:     name,
		Cats:     make(Cats),
		Commands: make(Commands),
	}
	gen.RunAll(out)
	// set ref to App in each Row
	for _, x := range out.Cats {
		for _, y := range x {
			y.App = out
		}
	}
	return
}

// which is made from

// Version fills the Version field of an App
func Version(ver string) AppGenerator {
	return func(ctx *App) {
		ctx.Version = func() string {
			return ver
		}
	}
}

// Tagline is a short text describing the application
func Tagline(ver string) AppGenerator {
	return func(ctx *App) {
		ctx.Tagline = ver
	}
}

// About is a longer text describing the application
func About(ver string) AppGenerator {
	return func(ctx *App) {
		ctx.About = ver
	}
}

// DefaultRunner is the command that runs when no parameters are given
func DefaultRunner(fn func(ctx *App) int) AppGenerator {
	return func(ctx *App) {
		ctx.Default = fn
	}
}

// Group is a collection of categories and bundles each category
func Group(name string, g ...CatGenerator) AppGenerator {
	G := CatGenerators(g)
	return func(ctx *App) {
		ctx.Cats[name] = make(Cat)
		G.RunAll(ctx.Cats[name])
	}
}

// Cmd is a collection of subcommands
func Cmd(name string, g ...CommandGenerator) AppGenerator {
	G := CommandGenerators(g)
	return func(ctx *App) {
		ctx.Commands[name] = G.RunAll()
	}
}

// Command Item Generators

// Pattern is the regular expression pattern that matches the CLI args for each
// Command item
func Pattern(patt string) CommandGenerator {
	return func(ctx *Command) {
		ctx.Pattern = patt
		ctx.RE = regexp.MustCompile(ctx.Pattern)
	}
}

// Short is the short help text for a Command
func Short(usage string) CommandGenerator {
	return func(ctx *Command) {
		ctx.Short = usage
	}
}

// Detail is the long help text for a Command
func Detail(usage string) CommandGenerator {
	return func(ctx *Command) {
		ctx.Detail = usage
	}
}

// Opts is the collection of valid options for a Command
func Opts(opts ...string) CommandGenerator {
	return func(ctx *Command) {
		ctx.Opts = opts
	}
}

// Precs is the collection of tags for items that are preferentially selected when
// an two or more items are specified (for example, help always overrides everything)
func Precs(precs ...string) CommandGenerator {
	return func(ctx *Command) {
		ctx.Precedent = precs
	}
}

// Handler is the function that is called when a command is selected
func Handler(hnd func(args []string, tokens Tokens, app *App) int) CommandGenerator {
	return func(ctx *Command) {
		ctx.Handler = hnd
	}
}

// Group Item Generators

// File is an item storing a filename
func File(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Type = "string"
			cc.Validate = Valid.File
			cc.Value = NewIface()
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
func Dir(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Type = "string"
			cc.Validate = Valid.Dir
			cc.Value = NewIface()
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
func Port(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Type = "port"
			cc.Validate = Valid.Port
			cc.Value = NewIface()
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
func Enable(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Type = "bool"
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Validate = Valid.Bool
			cc.Value = NewIface().Put(false)
			cc.Default = NewIface().Put(false)
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
func Enabled(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Type = "bool"
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Validate = Valid.Bool
			cc.Value = NewIface().Put(true)
			cc.Default = NewIface().Put(true)
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
func Int(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Type = "int"
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Validate = Valid.Int
			cc.Value = NewIface()
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
func Tag(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Type = "string"
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Validate = Valid.Tag
			cc.Value = NewIface()
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
func Tags(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Type = "stringslice"
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Validate = Valid.Tags
			cc.Value = NewIface()
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
func Addr(name string, defPort int, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Type = "string"
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Validate = GenAddr(name, defPort)
			cc.Value = NewIface()
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
func Addrs(name string, defPort int, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Type = "stringslice"
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Validate = GenAddrs(name, defPort)
			cc.Value = NewIface()
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
				return cc.Value.Get()
			}
			cc.Validate = Valid.Level
			cc.Value = NewIface()
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
func Algo(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Type = "options"
			cc.Opts = getAlgoOptions()
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Validate = Valid.Algo
			cc.Value = NewIface()
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
func Float(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Type = "float"
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Validate = Valid.Float
			cc.Value = NewIface()
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
func Duration(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Type = "duration"
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Validate = Valid.Duration
			cc.Value = NewIface()
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
func Net(name string, g ...RowGenerator) CatGenerator {
	G := RowGenerators(g)
	return func(ctx *Cat) {
		c := &Row{}
		c.Init = func(cc *Row) {
			cc.Name = name
			cc.Type = "options"
			cc.Opts = Networks
			cc.Get = func() interface{} {
				return cc.Value.Get()
			}
			cc.Validate = Valid.Net
			cc.Value = NewIface()
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
func Usage(usage string) RowGenerator {
	return func(ctx *Row) {
		ctx.Usage = usage + " " + ctx.Usage
	}
}

// Default sets the default value for a config item
func Default(in interface{}) RowGenerator {
	return func(ctx *Row) {
		ctx.Default = NewIface()
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
func Min(min int) RowGenerator {
	return func(ctx *Row) {
		ctx.Min = ctx.Min.Put(min)
		v := ctx.Validate
		var e error
		ctx.Validate = func(r *Row, in interface{}) bool {
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
func Max(max int) RowGenerator {
	return func(ctx *Row) {
		ctx.Max = ctx.Max.Put(max)
		v := ctx.Validate
		var e error
		ctx.Validate = func(r *Row, in interface{}) bool {
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
		ctx.Value = ctx.Value.Put(sb)
	}
}

// SaveConfig writes all the data in Cats the config file at the root of DataDir
func (app *App) SaveConfig() {
	if app == nil {
		return
	}
	datadir, ok := app.Cats["app"]["datadir"].Value.Get().(string)
	if !ok {
		return
	}
	configFile := CleanAndExpandPath(filepath.Join(
		datadir, "config"), "")
	if EnsureDir(configFile) {
	}
	fh, err := os.Create(configFile)
	if err != nil {
		panic(err)
	}
	j, e := json.MarshalIndent(app, "", "\t")
	if e != nil {
		panic(e)
	}
	_, err = fmt.Fprint(fh, string(j))
	if err != nil {
		panic(err)
	}
}

// MarshalJSON cherrypicks Cats for the values needed to correctly configure it
// and some extra information to make the JSON output friendly to human editors
func (r *App) MarshalJSON() ([]byte, error) {
	out := make(CatsJSON)
	for i, x := range r.Cats {
		out[i] = make(CatJSON)
		for j, y := range x {
			min, _ := y.Min.Get().(int)
			max, _ := y.Max.Get().(int)
			out[i][j] = Line{
				Value:   y.Value.Get(),
				Default: y.Default.Get(),
				Min:     min,
				Max:     max,
				Usage:   y.Usage,
			}
		}
	}
	return json.Marshal(out)
}

// UnmarshalJSON takes the cherrypicked JSON output of Marshal and puts it back into
// an App
func (r *App) UnmarshalJSON(data []byte) error {
	out := make(CatsJSON)
	e := json.Unmarshal(data, &out)
	if e != nil {
		return e
	}
	for i, x := range out {
		for j, y := range x {
			R := r.Cats[i][j]
			if y.Value != nil {
				switch R.Type {
				case "int", "port":
					y.Value = int(y.Value.(float64))
				case "duration":
					y.Value = time.Duration(int(y.Value.(float64)))
				case "stringslice":
					rt, ok := y.Value.([]string)
					ro := []string{}
					if ok {
						for _, z := range rt {
							R.Validate(R, z)
							ro = append(ro, z)
						}
						// R.Value.Put(ro)
					}
					// break
				case "float":
				}
			}
			R.Validate(R, y.Value)
			// R.Value.Put(y.Value)
		}
	}
	return nil
}

// RunAll triggers AppGenerators to configure an App
func (r *AppGenerators) RunAll(app *App) {
	for _, x := range *r {
		x(app)
	}
	return
}

// RunAll runs a collection of CatGenerators on a Cat
func (r *CatGenerators) RunAll(cat Cat) {
	for _, x := range *r {
		x(&cat)
	}
	return
}

func (r *RowGenerators) RunAll(row *Row) {
	for _, x := range *r {
		x(row)
	}
}

func (r *CommandGenerators) RunAll() *Command {
	c := &Command{}
	for _, x := range *r {
		x(c)
	}
	return c
}
