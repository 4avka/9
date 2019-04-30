package cmd
import (
	"fmt"
	"sort"
	"git.parallelcoin.io/dev/9/cmd/node"
	"git.parallelcoin.io/dev/9/cmd/walletmain"
	"git.parallelcoin.io/dev/9/cmd/ctl"
	"git.parallelcoin.io/dev/9/pkg/util/cl"
)
func optTagList(s []string) (ss string) {
	if len(ss) > 1 {
		ss = "[<"
		for i, x := range s {
			ss += x
			if i < len(s)-1 {
				ss += ">|<"
			} else {
				ss += ">]"
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
func getTokens(cmds Tokens) (s []string) {
	for _, x := range cmds {
		s = append(s, x.Value)
	}
	return
}
func Help(args []string, tokens Tokens, cmds, all Commands) int {
	log <- cl.Debug{"HELP\n", "Help", args, getTokens(tokens)}
	fmt.Println(APPNAME, APPVERSION, "-", APPDESC)
	fmt.Println()
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
func Conf(args []string, tokens Tokens, cmds, all Commands) int {
	var r int
	for r = 2; r == 2; {
		// r = RunConf(args, tokens, cmds, all)
	}
	return r
}
func New(args []string, tokens Tokens, cmds, all Commands) int {
	fmt.Println("running New", args, getTokens(tokens))
	return 0
}
func Copy(args []string, tokens Tokens, cmds, all Commands) int {
	fmt.Println("running Copy", args, getTokens(tokens))
	return 0
}
func List(args []string, tokens Tokens, cmds, all Commands) int {
	if j := validateProxyListeners(); j != 0 {
		return j
	}
	if _, ok := tokens[WALLET]; ok {
		*Config.Wallet = true
	}
	ctl.ListCommands()
	return 0
}
func Ctl(args []string, tokens Tokens, cmds, all Commands) int {
	cl.Register.SetAllLevels(*Config.LogLevel)
	setAppDataDir("node")
	if j := validateProxyListeners(); j != 0 {
		return j
	}
	if _, ok := tokens[WALLET]; ok {
		*Config.Wallet = true
	}
	var i int
	var x string
	for i, x = range args {
		if cmds[CTL].RE.Match([]byte(x)) {
			i++
			break
		}
	}
	ctl.Main(args[i:], Config)
	return 0
}
func Node(args []string, tokens Tokens, cmds, all Commands) int {
	cl.Register.SetAllLevels(*Config.LogLevel)
	setAppDataDir("node")
	if validateWhitelists() != 0 ||
		validateProxyListeners() != 0 ||
		validatePasswords() != 0 ||
		validateRPCCredentials() != 0 ||
		validateBlockLimits() != 0 ||
		validateUAComments() != 0 ||
		validateMiner() != 0 ||
		validateCheckpoints() != 0 ||
		validateAddresses() != 0 ||
		validateDialers() != 0 {
		return 1
	}
	// run the node!
	node.StateCfg = stateconfig
	node.Cfg = Config
	node.ActiveNetParams = activenetparams
	if node.Main(nil) != nil {
		return 1
	}
	return 0
}
func Wallet(args []string, tokens Tokens, cmds, all Commands) int {
	// spew.Dump(*config)
	cl.Register.SetAllLevels(*Config.LogLevel)
	setAppDataDir("wallet")
	walletmain.CreateWallet(Config, activenetparams)
	walletmain.Main(Config, activenetparams)
	return 0
}
func Shell(args []string, tokens Tokens, cmds, all Commands) int {
	cl.Register.SetAllLevels(*Config.LogLevel)
	fmt.Println("running Shell", args, getTokens(tokens))
	return 0
}
func Test(args []string, tokens Tokens, cmds, all Commands) int {
	cl.Register.SetAllLevels(*Config.LogLevel)
	fmt.Println("running Test", args, getTokens(tokens))
	return 0
}
func Create(args []string, tokens Tokens, cmds, all Commands) int {
	cl.Register.SetAllLevels(*Config.LogLevel)
	fmt.Println("running Create", args, getTokens(tokens))
	return 0
}
