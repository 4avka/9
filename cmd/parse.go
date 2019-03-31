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

var configRE = regexp.MustCompile("(NAME )(.*)( VALUE )(.*)( DEFAULT )(.*)( COMMENT )(.*)")

func Parse(args []string) int {
	// parse commandline
	cmd, tokens, cmds := parseCLI(args)
	if cmd == nil {
		help := commands[HELP]
		cmd = &help
	}
	var datadir string
	// read configuration
	dd, ok := Config["app.datadir"]
	if ok {
		datadir = *dd.Value.(*string)
		if t, ok := tokens["datadir"]; ok {
			datadir = t.Value
			Config["app.datadir"].Value = datadir
		}
	}
	log <- cl.Debug{"loading config from:", datadir}
	configFile := CleanAndExpandPath(filepath.Join(datadir, "config"))
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
		// fmt.Println("loading config from", configFile)
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
			valid := Config[out[2]].Validator(out[4])
			if !valid {
				fmt.Println(
					"config parsing error on line", i,
					"name", out[2], "erroneous value:", out[4])
			}
		}
		// fmt.Println("loaded config")
		// fmt.Println(Config)
	}

	// spew.Dump(Config)
	// run as configured
	_ = cmds
	r := cmd.Handler(
		args,
		tokens,
		cmds,
		commands)
	return r
}
