package cmd

import "git.parallelcoin.io/dev/9/pkg/util/cl"

func Start(args []string) int {
	log <- cl.Inf("starting 9")
	return Parse(args)
}
