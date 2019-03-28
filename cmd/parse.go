package cmd

import "fmt"

func Parse(args []string) int {
	// parse commandline
	cmd, tokens, cmds := parseCLI(args)
	if cmd == nil {
		help := commands[HELP]
		cmd = &help
	}
	// read configuration
	dd, ok := Config["app.datadir"]
	dsp := dd.Value.(*string)
	datadir := *dsp
	if ok {
		if t, ok := tokens["datadir"]; ok {
			datadir = t.Value
		}
	}
	fmt.Println("loading config from:", datadir)

	// run as configured
	cmd.Handler(
		args,
		tokens,
		cmds,
		commands)
	return 0
}
