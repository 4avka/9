package app

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"git.parallelcoin.io/dev/9/cmd/def"
	"git.parallelcoin.io/dev/9/cmd/node"
	"git.parallelcoin.io/dev/9/pkg/util"
	"git.parallelcoin.io/dev/9/pkg/util/cl"
)

var datadir = new(string)

// Parse commandline
func Parse(ap *def.App, args []string) int {
	cmd, tokens := ParseCLI(ap, args)
	if cmd == nil {
		cmd = ap.Commands["help"]
	}
	// get datadir from cli args if given
	if dd, ok := tokens["datadir"]; ok {
		datadir = &dd.Value
		pwd, _ := os.Getwd()
		*datadir = filepath.Join(pwd, *datadir)
		dd.Value = *datadir
		ap.Cats["app"]["datadir"].Value.Put(*datadir)
		DataDir = *datadir
	} else {
		ddd := util.AppDataDir("9", false)
		ap.Cats["app"]["datadir"].Put(ddd)
		datadir = &ddd
		DataDir = *datadir
	}
	// for i, x := range ap.Cats {
	// 	for j := range x {
	// 		// if i == "app" && j == "datadir" {
	// 		// 	break
	// 		// }
	// 		ap.Cats[i][j].Init(ap.Cats[i][j])
	// 	}
	// }

	// // set AppDataDir for running as node
	// aa := util.CleanAndExpandPath(filepath.Join(
	// 	*datadir,
	// 	cmd.Name),
	// 	*datadir)
	// ap.Config.AppDataDir, ap.Config.LogDir = &aa, &aa

	configFile := util.CleanAndExpandPath(filepath.Join(
		*datadir, "config"), *datadir)
	// *ap.Config.ConfigFile = configFile
	if !util.FileExists(configFile) {
		if util.EnsureDir(configFile) {
		}
		fh, err := os.Create(configFile)
		if err != nil {
			panic(err)
		}
		j, e := json.MarshalIndent(ap, "", "\t")
		if e != nil {
			panic(e)
		}
		_, err = fmt.Fprint(fh, string(j))
		if err != nil {
			panic(err)
		}
	}
	conf, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(err)
	}
	e := json.Unmarshal(conf, ap)
	if e != nil {
		panic(e)
	}
	// now we can initialise the App
	for i, x := range ap.Cats {
		for j := range x {
			temp := ap.Cats[i][j]
			temp.App = ap
			ap.Cats[i][j] = temp
		}
	}
	ap.Config = MakeConfig(ap)
	ap.Config.ActiveNetParams = node.ActiveNetParams
	if ap.Config.LogLevel != nil {
		cl.Register.SetAllLevels(*ap.Config.LogLevel)
	}
	// run as configured
	r := cmd.Handler(
		args,
		tokens,
		ap)
	return r
}

func ParseCLI(ap *def.App, args []string) (cmd *def.Command, tokens def.Tokens) {
	cmd = new(def.Command)
	// collect set of items in commandline
	if len(args) < 2 {
		fmt.Print("No args given, printing help:\n\n")
		args = append(args, "h")
	}
	commandsFound := make(map[string]int)
	tokens = make(def.Tokens)
	for _, x := range args[1:] {
		for i, y := range ap.Commands {
			if y.RE.MatchString(x) {
				if _, ok := commandsFound[i]; ok {
					tokens[i] = def.Token{x, *y}
					commandsFound[i]++
					break
				} else {
					tokens[i] = def.Token{x, *y}
					commandsFound[i] = 1
					break
				}
			}
		}
	}
	var withHandlersNames []string
	withHandlers := make(def.Commands)
	for i := range commandsFound {
		if ap.Commands[i].Handler != nil {
			withHandlers[i] = ap.Commands[i]
			withHandlersNames = append(withHandlersNames, i)
		}
	}
	invoked := make(def.Commands)
	for i, x := range withHandlers {
		invoked[i] = x
	}
	// search the precedents of each in the case of multiple
	// with handlers and delete the one that has another in the
	// list of matching handlers. If one is left we can run it,
	// otherwise return an error.
	var resolved []string
	if len(withHandlersNames) > 1 {
		var common [][]string
		for _, x := range withHandlersNames {
			i := util.Intersection(withHandlersNames, withHandlers[x].Precedent)
			common = append(common, i)
		}
		for _, x := range common {
			for _, y := range x {
				if y != "" {
					resolved = append(resolved, y)
				}
			}
		}
		resolved = util.Uniq(resolved)
		if len(resolved) > 1 {
			withHandlers = make(def.Commands)
			common = [][]string{}
			withHandlersNames = resolved
			resolved = []string{}
			for _, x := range withHandlersNames {
				withHandlers[x] = ap.Commands[x]
			}
			for _, x := range withHandlersNames {
				i := util.Intersection(withHandlersNames, withHandlers[x].Precedent)
				common = append(common, i)
			}
			for _, x := range common {
				for _, y := range x {
					if y != "" {
						resolved = append(resolved, y)
					}
				}
			}
			resolved = util.Uniq(resolved)
		}
	} else if len(withHandlersNames) == 1 {
		resolved = []string{withHandlersNames[0]}
	}
	// fmt.Println(resolved)
	if len(resolved) < 1 {
		err := fmt.Errorf(
			"\nunable to resolve which command to run:\n\tinput: '%s'",
			withHandlersNames)
		fmt.Println(err)
		return nil, tokens
	}
	*cmd = *ap.Commands[resolved[0]]
	return cmd, tokens
}
