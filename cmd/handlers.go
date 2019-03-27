package cmd

import (
	"fmt"
	"sort"

	"git.parallelcoin.io/dev/9/pkg/util/cl"
	"github.com/davecgh/go-spew/spew"
)

func optTagList(s []string) (S string) {
	if len(s) > 1 {

		S = "[<"
		for i, x := range s {
			S += x
			if i < len(s)-1 {
				S += ">|<"
			} else {
				S += ">]"
			}
		}
	}
	return
}

func getCommands(cmds Commands) (s []string) {
	for i := range cmds {
		s = append(s, i)
	}
	return
}

func Help(args []string, cmds, tokens, all Commands) int {
	log <- cl.Debug{"HELP\n", "Help", args, getCommands(tokens)}
	fmt.Println(APPNAME, "-", APPDESC)
	fmt.Println()
	// fmt.Println("args received:", args[1:])
	// fmt.Println("cmds received:", cmds)
	// fmt.Println()
	fmt.Println("help with", APPNAME)
	fmt.Println()
	if len(cmds) == 1 {
		// help was invoked
		var tags []string
		for i := range all {
			tags = append(tags, i)
		}
		sort.Strings(tags)
		for _, x := range tags {
			if all[x].Handler != nil {
				fmt.Printf("\t%s '%s' %s\n\t\t%s\n\n",
					x, all[x].pattern,
					optTagList(all[x].Optional),
					all[x].Usage)
				// fmt.Println(x.Detail)
				// fmt.Println()
			}
		}
	} else {
		// some number of other commands were mentioned
		var tags []string
		for i := range cmds {
			tags = append(tags, i)
		}
		sort.Strings(tags)
		for _, x := range tags {
			if x != "help" {
				fmt.Printf("%s '%s' %s\n\n\t%s\n",
					x, cmds[x].pattern,
					optTagList(all[x].Optional),
					cmds[x].Usage)
				fmt.Println("\n", cmds[x].Detail)
				fmt.Println()
			}
		}
	}
	return 0
}

func Conf(args []string, cmds, tokens, all Commands) int {
	fmt.Println("running Conf", args, getCommands(tokens))
	spew.Dump(tokens)
	return 0
}

func New(args []string, cmds, tokens, all Commands) int {
	fmt.Println("running New", args, getCommands(tokens))
	return 0
}

func Copy(args []string, cmds, tokens, all Commands) int {
	fmt.Println("running Copy", args, getCommands(tokens))
	return 0
}

func List(args []string, cmds, tokens, all Commands) int {
	fmt.Println("running List", args, getCommands(tokens))
	return 0
}

func Ctl(args []string, cmds, tokens, all Commands) int {
	fmt.Println("running Ctl", args, getCommands(tokens))
	return 0
}

func Node(args []string, cmds, tokens, all Commands) int {
	fmt.Println("running Node", args, getCommands(tokens))
	return 0
}

func Wallet(args []string, cmds, tokens, all Commands) int {
	fmt.Println("running Wallet", args, getCommands(tokens))
	return 0
}

func Shell(args []string, cmds, tokens, all Commands) int {
	fmt.Println("running Shell", args, getCommands(tokens))
	return 0
}

func Test(args []string, cmds, tokens, all Commands) int {
	fmt.Println("running Test", args, getCommands(tokens))
	return 0
}

func Create(args []string, cmds, tokens, all Commands) int {
	fmt.Println("running Create", args, getCommands(tokens))
	return 0
}
