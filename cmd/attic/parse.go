package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"git.parallelcoin.io/dev/9/pkg/util/cl"
)

var configRE = regexp.MustCompile(
	"(NAME )(.*)( VALUE )(.*)( DEFAULT )(.*)( COMMENT )(.*)")

func Parse(args []string) int {
	// parse commandline
	cmd, tokens, cmds := parseCLI(args)
	if cmd == nil {
		help := commands[HELP]
		cmd = &help
	}
	var datadir string
	if dd, ok := tokens[DATADIR]; ok {
		datadir = dd.Value
		*Config.DataDir = datadir
	} else {
		*Config.DataDir = (*config)["app.datadir"].Initial.(string)
		datadir = *Config.DataDir
	}
	setAppDataDir(cmd.name)

	setDefaultTLSPaths(datadir)

	configFile := CleanAndExpandPath(filepath.Join(datadir, "config"))
	log <- cl.Debug{"loading config from:", configFile}
	if !FileExists(configFile) {
		fmt.Println("config file not found: creating new one at ", configFile)
		if EnsureDir(configFile) {
			fmt.Println("created new directory to store data", datadir)
		}
		fh, err := os.Create(configFile)
		if err != nil {
			panic(err)
		}
		_, err = fmt.Fprint(fh, Config)
		if err != nil {
			panic(err)
		}

	} else {
		conf, err := ioutil.ReadFile(configFile)
		if err != nil {
			panic(err)
		}
		confstring := string(conf)
		splitted := strings.Split(confstring, "\n")
		for i, x := range splitted {
			out := configRE.FindStringSubmatch(x)
			if len(out) < 1 {
				continue
			}
			_, ok := (*config)[out[2]]
			if !ok {
				// fmt.Println("unknown key found", out[2], "ignoring")
			} else {
				valid := (*config)[out[2]].Validate((*config)[out[2]], out[4])
				if !valid {
					fmt.Println(
						"config parsing error on line", i,
						"name", out[2], "erroneous value:", out[4])
				}
			}
		}
	}
	if Config.Network != nil {
		// switchDefaultAddrs(*Config.Network)
	}
	*Config.ConfigFile = configFile
	if Config.LogLevel != nil {
		fmt.Println("setting debug level to", *Config.LogLevel)
		cl.Register.SetAllLevels(*Config.LogLevel)
	}
	log <- cl.Debug{"yay"}
	// run as configured
	_ = cmds
	r := cmd.Handler(
		args,
		tokens,
		cmds,
		commands)
	return r
}
