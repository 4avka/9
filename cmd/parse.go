package cmd

import (
	"errors"

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
	log <- cl.Info{args}
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

	withHandlers := make(Commands)
	for i := range commandsFound {
		if commands[i].Handler != nil {
			log <- cl.Info{"found", i}
			withHandlers[i] = commands[i]
		}
	}
	// search the precedents of each in the case of multiple
	// with handlers and delete the one that has another in the
	// list of matching handlers. If one is left we can run it,
	// otherwise return an error.
	var withHandlersNames []string
	if len(withHandlers) > 1 {
		for i := range withHandlers {
			withHandlersNames = append(withHandlersNames, i)
		}
	}
	log <- cl.Info{"handlersnames", withHandlersNames}
	for i, x := range withHandlers {
		log <- cl.Info{"precedent", x.Precedent}
		for _, y := range x.Precedent {

			for _, z := range withHandlersNames {
				log <- cl.Info{"handlers", i, z, y}
				if y == z {
					log <- cl.Info{"deleting", z}
					delete(withHandlers, z)
					goto out
				}
			}
		}
	out:
	}
	for i := range withHandlers {
		log <- cl.Warn{">>> resolved", i}
	}
	if len(withHandlers) > 1 {
		err := errors.New("unable to resolve which command to run")
		log <- cl.Error{err}
		return err
	}

	return nil
}
