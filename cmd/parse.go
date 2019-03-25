package cmd

import (
	"fmt"
	"regexp"
	"strings"
)

var commands = map[string]*regexp.Regexp{
	"help":    regexp.MustCompile("^(h|help)$"),
	"conf":    regexp.MustCompile("^(C|conf)$"),
	"new":     regexp.MustCompile("^(N|new)$"),
	"copy":    regexp.MustCompile("^(cp|copy)$"),
	"list":    regexp.MustCompile("^(l|list|listcommands)$"),
	"ctl":     regexp.MustCompile("^(c|ctl)$"),
	"node":    regexp.MustCompile("^(n|node)$"),
	"wallet":  regexp.MustCompile("^(w|wallet)$"),
	"test":    regexp.MustCompile("^(t|test)$"),
	"create":  regexp.MustCompile("^(create)$"),
	"datadir": regexp.MustCompile("^(.*/)$"),
	"integer": regexp.MustCompile("^([0-9]+)$"),
	"word":    regexp.MustCompile("^([a-zA-Z0-9._-]*)$"),
}

var allcommands = func() (s []string) {
	for i := range commands {
		s = append(s, i)
	}
	return
}()

type opts []string
type precedent []string

// CommandSet is a collection of items centering around a primary item, the Main
//
// The Main is a command, such as node or wallet or help
//
// The Precedent indicates that if this set also contains something that it
// prefers to match that set instead.
//
// By using sets, we drastically reduce the complexity of command line parsing
// and leave complex things to configuration files.
//
// Nil in opts indicate no preference.
type CommandSet struct {
	// Main is the name of the command
	Main string
	// Optional keywords that may appear with the command
	Optional opts
	// Precedent indicates other commands that will preferentially match
	Precedent precedent
	// Handler
	Handler func(CommandSet) int
}

var CommandSets = []CommandSet{
	{
		"help",
		nil,
		nil,
		nil,
	},
	{
		"list",
		opts{"datadir", "ctl"},
		precedent{"help"},
		nil,
	},
	{
		"ctl",
		opts{"datadir", "list"},
		precedent{"help", "list"},
		nil,
	},
	{
		"conf",
		opts{"datadir"},
		precedent{"help"},
		nil,
	},
	{
		"node",
		opts{"datadir"},
		precedent{"help"},
		nil,
	},
	{
		"wallet",
		opts{"datadir", "create"},
		precedent{"help"},
		nil,
	},
	{
		"shell",
		opts{"datadir", "create"},
		precedent{"help"},
		nil,
	},
	{
		"test",
		opts{"word", "integer"},
		precedent{"help"},
		nil,
	},
	{
		"copy",
		opts{"datadir", "word", "integer"},
		precedent{"help"},
		nil,
	},
	{
		"new",
		opts{"word", "integer"},
		precedent{"help"},
		nil,
	},
}

func Parse(args []string) int {
	// GenerateLines(lines)
	set := make(map[string]int)
	for i, x := range positivetests {
		for j, y := range x[1:] {
			Y := []byte(strings.ToLower(y))
			foundnonword := false
			if commands["help"].Match(Y) {
				fmt.Println(i, x, "HELP")
				set["help"] = j
				foundnonword = true
			}
			if commands["node"].Match(Y) {
				fmt.Println(i, x, "NODE")
				set["node"] = j
				foundnonword = true
			}
			if commands["ctl"].Match(Y) {
				fmt.Println(i, x, "CTL")
				set["ctl"] = j
				foundnonword = true
			}
			if commands["conf"].Match(Y) {
				fmt.Println(i, x, "CONF")
				set["conf"] = j
				foundnonword = true
			}
			if commands["wallet"].Match(Y) {
				fmt.Println(i, x, "WALLET")
				set["wallet"] = j
				foundnonword = true
			}
			if commands["test"].Match(Y) {
				fmt.Println(i, x, "TEST")
				set["test"] = j
				foundnonword = true
			}
			if commands["create"].Match(Y) {
				fmt.Println(i, x, "CREATE")
				set["create"] = j
				foundnonword = true
			}
			if commands["datadir"].Match(Y) {
				fmt.Println(i, x, "DATADIR")
				set["datadir"] = j
				foundnonword = true
			}
			if commands["integer"].Match(Y) {
				fmt.Println(i, x, "INTEGER")
				set["integer"] = j
				foundnonword = true
			}
			if commands["word"].Match(Y) {
				if !foundnonword {
					fmt.Println(i, x, "WORD")
					set["word"] = j
				}
			}
		}
	}
	return 0
}

var positivetests = [][]string{
	{"9", "h"},
	{"9", "help"},
	{"9", "h", "node"},
	{"9", "help", "conf"},
	{"9", "test/", "c"},
	{"9", "node", "h"},
	{"9", "C", "basename", "9"},
	{"9", "n", "test/"},
	{"9", "t", "basename"},
	{"9", "create", "testnet", "9"},
}
