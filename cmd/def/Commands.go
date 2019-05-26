package def

import (
	"regexp"
	"sort"
)

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

// GetSortedKeys returns the keys of a slice of Commands in
// lexicographic order
func (r *Commands) GetSortedKeys() (out []string) {
	for i := range *r {
		out = append(out, i)
	}
	sort.Strings(out)
	return
}

// RunAll executes all the generators in a CommandGenerators slice
func (r *CommandGenerators) RunAll() *Command {
	c := &Command{}
	for _, x := range *r {
		x(c)
	}
	return c
}
