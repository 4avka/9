package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

type Line struct {
	Value   interface{} `json:"value"`
	Default interface{} `json:"default,omitempty"`
	Min     int         `json:"min,omitempty"`
	Max     int         `json:"max,omitempty"`
	Usage   string      `json:"usage"`
}

type CatJSON map[string]Line

type CatsJSON map[string]CatJSON

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
		o := cc.Value.Get()
		return &o
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

type Cat map[string]*Row
type CatGenerator func(ctx *Cat)
type CatGenerators []CatGenerator

func (r *CatGenerators) RunAll(cat Cat) {
	for _, x := range *r {
		x(&cat)
	}
	return
}

type Iface struct {
	Data *interface{}
}

func NewIface() *Iface {
	return &Iface{Data: new(interface{})}
}

func (i *Iface) Get() interface{} {
	if i == nil {
		return nil
	}
	if i.Data == nil {
		return nil
	}
	return *i.Data
}

func (i *Iface) Put(in interface{}) *Iface {
	if i == nil {
		i = NewIface()
	}
	if i.Data == nil {
		i.Data = new(interface{})
	}
	*i.Data = in
	return i
}

type Row struct {
	Name     string
	Type     string
	Opts     []string
	Value    *Iface
	Default  *Iface
	Min      *Iface
	Max      *Iface
	Init     func(*Row)
	Get      func() interface{}
	Put      func(interface{}) bool
	Validate func(*Row, interface{}) bool
	String   string
	Usage    string
	App      *App
}

func (r *Row) Bool() bool {
	return r.Value.Get().(bool)
}

func (r *Row) Int() int {
	return r.Value.Get().(int)
}

func (r *Row) Float() float64 {
	return r.Value.Get().(float64)
}

func (r *Row) Duration() time.Duration {
	return r.Value.Get().(time.Duration)
}

func (r *Row) Tag() string {
	return r.Value.Get().(string)
}

func (r *Row) Tags() []string {
	return r.Value.Get().([]string)
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
