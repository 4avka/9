package def

import (
	"regexp"

	"git.parallelcoin.io/dev/9/cmd/nine"
	"git.parallelcoin.io/dev/9/pkg/ifc"
)

// App contains all the configuration and subcommand definitions for an app
type App struct {
	Name     string
	Tagline  string
	About    string
	Version  func() string
	Default  func(ctx *App) int
	Cats     Cats
	Commands Commands
	Config   *nine.Config
	Started  chan struct{}
}

// AppGenerator is a function that configures an App
type AppGenerator func(ctx *App)

// AppGenerators is a collection of AppGenerators
type AppGenerators []AppGenerator

// Row is a configuration variable
type Row struct {
	Name     string
	Type     string
	Opts     []string
	Value    *ifc.Iface
	Default  *ifc.Iface
	Min      *ifc.Iface
	Max      *ifc.Iface
	Init     func(*Row)
	Get      func() interface{}
	Put      func(interface{}) bool
	Validate func(*Row, interface{}) bool
	String   string
	Usage    string
	App      *App
}

// RowGenerator configures a Row
type RowGenerator func(ctx *Row)

// RowGenerators is a collection of Rows
type RowGenerators []RowGenerator

// Cat is a collection of Rows with tag labels
type Cat map[string]*Row

// CatGenerator is a function that configures a Cat
type CatGenerator func(ctx *Cat)

// CatGenerators is a collection of Cat's
type CatGenerators []CatGenerator

// Line is the JSON formatted version of a Cat
type Line struct {
	Value   interface{} `json:"value"`
	Default interface{} `json:"default,omitempty"`
	Min     int         `json:"min,omitempty"`
	Max     int         `json:"max,omitempty"`
	Usage   string      `json:"usage"`
}

// CatJSON is a collection of lines with their tag
type CatJSON map[string]Line

// CatsJSON is a collection of collections of lines with grouping tags
type CatsJSON map[string]CatJSON

// Cats are a collection of Rows with a string tag
type Cats map[string]Cat

// CommandHandler is the launcher that runs a command
type CommandHandler func(args []string, tokens Tokens, app *App) int

// Command is the collection of metadata and handler for a subcommand
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

// CommandGenerator is a function that configures a Command
type CommandGenerator func(ctx *Command)

// CommandGenerators is a collection of CommandGenerators
type CommandGenerators []CommandGenerator

// Commands is a tagged collection of Commands
type Commands map[string]*Command

// Token is a struct that ties together CLI invocation to the Command it
// relates to
type Token struct {
	Value string
	Cmd   Command
}

// Tokens is a collection of Tokens
type Tokens map[string]Token

// Optional is a set of possible valid items accompanying a Token
type Optional []string

// Precedent is a set of possible valid items that match preferentially
// to the item in a Command
type Precedent []string
