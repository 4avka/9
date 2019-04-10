package config

import (
	"encoding/json"
	"regexp"
	"time"

	"git.parallelcoin.io/dev/9/cmd/nine"
)

type App struct {
	Name     string
	Tagline  string
	About    string
	Version  func() string
	Default  func(ctx *App) int
	Cats     Cats
	Commands Commands
	Config   *nine.Config
}

type Line struct {
	Value   interface{} `json:"value"`
	Default interface{} `json:"default,omitempty"`
	Min     interface{} `json:"min,omitempty"`
	Max     interface{} `json:"max,omitempty"`
	Usage   string      `json:"usage"`
}

type CatJSON map[string]Line

type CatsJSON map[string]CatJSON

func (r *App) MarshalJSON() ([]byte, error) {
	out := make(CatsJSON)
	for i, x := range r.Cats {
		out[i] = make(CatJSON)
		for j, y := range x {
			out[i][j] = Line{
				Value:   y.Value,
				Default: y.Default,
				Min:     y.Min,
				Max:     y.Max,
				Usage:   y.Usage,
			}
		}
	}
	return json.Marshal(out)
}

func (r *App) UnmarshalJSON(data []byte) error {
	out := make(CatsJSON)
	e := json.Unmarshal(data, &out)
	if e != nil {
		return e
	}
	for i, x := range out {
		for j, y := range x {
			R := r.Cats[i][j]
			R.Put(y.Value)
		}
	}
	return nil
}

type AppGenerator func(ctx *App)
type AppGenerators []AppGenerator

func (r *AppGenerators) RunAll(app *App) {
	for _, x := range *r {
		x(app)
	}
	return
}

type Cats map[string]Cat

// Str returns the pointer to a value in the category map
func (r *Cats) Str(cat, item string) (out *string) {
	C := *r
	if C[cat][item].Value != nil {
		return C[cat][item].Value.(*string)
	}
	return
}

// Tags returns the pointer to a value in the category map
func (r *Cats) Tags(cat, item string) (out *[]string) {
	C := *r
	if C[cat][item].Value != nil {
		return C[cat][item].Value.(*[]string)
	}
	return
}

// Map returns the pointer to a value in the category map
func (r *Cats) Map(cat, item string) (out *nine.Mapstringstring) {
	C := *r
	if C[cat][item].Value != nil {
		return C[cat][item].Value.(*nine.Mapstringstring)
	}
	return
}

// Int returns the pointer to a value in the category map
func (r *Cats) Int(cat, item string) (out *int) {
	C := *r
	if C[cat][item].Value != nil {
		return C[cat][item].Value.(*int)
	}
	return
}

// Bool returns the pointer to a value in the category map
func (r *Cats) Bool(cat, item string) (out *bool) {
	C := *r
	if C[cat][item].Value != nil {
		return C[cat][item].Value.(*bool)
	}
	return
}

// Float returns the pointer to a value in the category map
func (r *Cats) Float(cat, item string) (out *float64) {
	C := *r
	if C[cat][item].Value != nil {
		return C[cat][item].Value.(*float64)
	}
	return
}

// Duration returns the pointer to a value in the category map
func (r *Cats) Duration(cat, item string) (out *time.Duration) {
	C := *r
	if C[cat][item].Value != nil {
		return C[cat][item].Value.(*time.Duration)
	}
	return
}

type Cat map[string]Row
type CatGenerator func(ctx *Cat)
type CatGenerators []CatGenerator

func (r *CatGenerators) RunAll(cat Cat) {
	for _, x := range *r {
		x(&cat)
	}
	return
}

type Row struct {
	Name     string
	Value    interface{}
	Default  interface{}
	Min      interface{}
	Max      interface{}
	Init     func(*Row)
	Get      func() interface{}
	Put      func(interface{}) bool
	Validate func(*Row, interface{}) bool
	Usage    string
}
type RowGenerator func(ctx *Row)
type RowGenerators []RowGenerator

func (r *RowGenerators) RunAll(row *Row) {
	for _, x := range *r {
		x(row)
	}
}

// Token is a struct that ties together CLI invocation to the Command it
// relates to
type Token struct {
	Value string
	Cmd   Command
}
type Tokens map[string]Token

type Optional []string
type Precedent []string

type CommandHandler func(args []string, tokens Tokens, app *App) int

type Command struct {
	Name      string
	Pattern   string
	RE        *regexp.Regexp
	Short     string
	Detail    string
	Opts      Optional
	Precedent Precedent
	Handler   CommandHandler
}
type CommandGenerator func(ctx *Command)
type CommandGenerators []CommandGenerator
type Commands map[string]*Command

func (r *CommandGenerators) RunAll() *Command {
	c := &Command{}
	for _, x := range *r {
		x(c)
	}
	return c
}
