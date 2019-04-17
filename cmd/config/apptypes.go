package config

import (
	"encoding/json"
	"regexp"
	"sort"
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
			R.Put(&y.Value)
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

func (r *Cats) getValue(cat, item string) (out *interface{}) {
	if r == nil {
		return
	} else if C, ok := (*r)[cat]; !ok {
		return
	} else if cc, ok := C[item]; !ok {
		return
	} else {
		return cc.Value
	}
}

// Str returns the pointer to a value in the category map
func (r *Cats) Str(cat, item string) (out *string) {
	cv := r.getValue(cat, item)
	if cv == nil {
		return
	}
	CC := *cv
	if ci, ok := CC.(string); !ok {
		return
	} else {
		return &ci
	}
}

// Tags returns the pointer to a value in the category map
func (r *Cats) Tags(cat, item string) (out *[]string) {
	cv := r.getValue(cat, item)
	if cv == nil {
		return
	}
	CC := *cv
	if ci, ok := CC.([]string); !ok {
		return
	} else {
		return &ci
	}
}

// Map returns the pointer to a value in the category map
func (r *Cats) Map(cat, item string) (out *nine.Mapstringstring) {
	cv := r.getValue(cat, item)
	if cv == nil {
		return
	}
	CC := *cv
	if ci, ok := CC.(nine.Mapstringstring); !ok {
		return
	} else {
		return &ci
	}
}

// Int returns the pointer to a value in the category map
func (r *Cats) Int(cat, item string) (out *int) {
	cv := r.getValue(cat, item)
	if cv == nil {
		return
	}
	CC := *cv
	if ci, ok := CC.(int); !ok {
		return
	} else {
		return &ci
	}
}

// Bool returns the pointer to a value in the category map
func (r *Cats) Bool(cat, item string) (out *bool) {
	cv := r.getValue(cat, item)
	if cv == nil {
		return
	}
	CC := *cv
	if ci, ok := CC.(bool); !ok {
		return
	} else {
		return &ci
	}
}

// Float returns the pointer to a value in the category map
func (r *Cats) Float(cat, item string) (out *float64) {
	cv := r.getValue(cat, item)
	if cv == nil {
		return
	}
	CC := *cv
	if ci, ok := CC.(float64); !ok {
		return
	} else {
		return &ci
	}
}

// Duration returns the pointer to a value in the category map
func (r *Cats) Duration(cat, item string) (out *time.Duration) {
	cv := r.getValue(cat, item)
	if cv == nil {
		return
	}
	CC := *cv
	if ci, ok := CC.(time.Duration); !ok {
		return
	} else {
		return &ci
	}
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
	Type     string
	Opts     []string
	Value    *interface{}
	Default  *interface{}
	Min      *interface{}
	Max      *interface{}
	Init     func(*Row)
	Get      func() interface{}
	Put      func(interface{}) bool
	Validate func(*Row, interface{}) bool
	String   string
	Usage    string
}

func (r *Row) Bool() bool {
	if *r.Value == nil {
		return false
	}
	return (*r.Value).(bool)
}

func (r *Row) Int() int {
	if *r.Value == nil {
		return -1
	}
	return (*r.Value).(int)
}

func (r *Row) Float() float64 {
	if *r.Value == nil {
		return -1.0
	}
	return (*r.Value).(float64)
}

func (r *Row) Duration() time.Duration {
	if *r.Value == nil {
		return 0
	}
	return (*r.Value).(time.Duration)
}

func (r *Row) Tag() string {
	if *r.Value == nil {
		return ""
	}
	return (*r.Value).(string)
}

func (r *Row) Tags() []string {
	if *r.Value == nil {
		return nil
	}
	return (*r.Value).([]string)
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

// GetSortedKeys returns the keys of a map in alphabetical order
func (r *CatJSON) GetSortedKeys() (out []string) {
	for i := range *r {
		out = append(out, i)
	}
	sort.Strings(out)
	return
}

func (r *CatsJSON) GetSortedKeys() (out []string) {
	for i := range *r {
		out = append(out, i)
	}
	sort.Strings(out)
	return
}

func (r *Cats) GetSortedKeys() (out []string) {
	for i := range *r {
		out = append(out, i)
	}
	sort.Strings(out)
	return
}

func (r Cat) GetSortedKeys() (out []string) {
	for i := range r {
		out = append(out, i)
	}
	sort.Strings(out)
	return
}

func (r *Tokens) GetSortedKeys() (out []string) {
	for i := range *r {
		out = append(out, i)
	}
	sort.Strings(out)
	return
}

func (r *Commands) GetSortedKeys() (out []string) {
	for i := range *r {
		out = append(out, i)
	}
	sort.Strings(out)
	return
}
