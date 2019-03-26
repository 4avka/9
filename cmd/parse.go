package cmd

import (
	"git.parallelcoin.io/dev/9/pkg/util/cl"
)

func Parse(args []string) int {
	// parse commandline
	err := parseCLI(args)
	if err != nil {
		log <- cl.Error{err}
	}
	// read configuration

	return 0
}

func parseCLI(args []string) error {
	// collect set of items in commandline
	if len(args) < 2 {
		log <- cl.Info{"no args given, printing help"}
	}
	commandsFound := make(map[string]int)
	for _, x := range args[1:] {
		for _, y := range commandsList {
			if commands[y].RE.Match([]byte(x)) {
				if _, ok := commandsFound[y]; ok {
					// log <- cl.Info{"found", i, y, x}
					commandsFound[y]++
					break
				} else {
					// log <- cl.Info{"found", i, y, x}
					commandsFound[y] = 1
					break
				}
			}
		}
	}

	log <- cl.Warn{args, ":", commandsFound}

	return nil
}
