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
		*config.DataDir = datadir
	} else {
		*config.DataDir = Config["app.datadir"].Default.(string)
		datadir = *config.DataDir
	}
	setAppDataDir(cmd.name)

	if *Config["tls.cert"].Value.(*string) == "" {
		rpccert := CleanAndExpandPath(
			filepath.Join(datadir, Config["tls.cert"].Default.(string)))
		*Config["tls.cert"].Value.(*string) = rpccert
	}
	if *Config["tls.key"].Value.(*string) == "" {
		rpckey := CleanAndExpandPath(
			filepath.Join(datadir, Config["tls.key"].Default.(string)))
		*Config["tls.key"].Value.(*string) = rpckey
	}
	if *Config["tls.cafile"].Value.(*string) == "" {
		cafile := CleanAndExpandPath(
			filepath.Join(datadir, Config["tls.cafile"].Default.(string)))
		*Config["tls.cafile"].Value.(*string) = cafile
	}
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
			_, ok := Config[out[2]]
			if !ok {
				// fmt.Println("unknown key found", out[2], "ignoring")
			} else {
				valid := Config[out[2]].Validator(out[4])
				if !valid {
					fmt.Println(
						"config parsing error on line", i,
						"name", out[2], "erroneous value:", out[4])
				}
			}
		}
	}
	*config.ConfigFile = configFile
	fmt.Println("setting debug level to", *config.LogLevel)
	cl.Register.SetAllLevels(*config.LogLevel)
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
