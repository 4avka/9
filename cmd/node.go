package cmd

import "github.com/davecgh/go-spew/spew"

func runNode(args []string, tokens Tokens, cmds, all Commands) int {
	spew.Dump(*config)
	return 0
}
