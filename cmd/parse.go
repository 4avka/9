package cmd

func Parse(args []string) int {
	// parse commandline
	cmd, cmds, tokens := parseCLI(args)
	if cmd == nil {
		help := commands[HELP]
		cmd = &help
	}
	// read configuration

	// run as configured
	cmd.Handler(
		args,
		cmds,
		tokens,
		commands)
	return 0
}
