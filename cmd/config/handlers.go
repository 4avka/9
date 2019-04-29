package config

import (
	"fmt"
	"git.parallelcoin.io/dev/9/cmd/nine"
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
	sort.Strings(s)
	return
}

func getTokens(cmds Tokens) (s []string) {
	for _, x := range cmds {
		s = append(s, x.Value)
	}
	sort.Strings(s)
	return
}

func Help(args []string, tokens Tokens, app *App) int {
	fmt.Println(app.Name, app.Version(), "-", app.Tagline)
	fmt.Println()
	fmt.Println("help with", app.Name)
	fmt.Println()
	if len(tokens) == 1 {
		// help was invoked
		var tags []string
		for i := range app.Commands {
			tags = append(tags, i)
		}
		sort.Strings(tags)
		for _, x := range tags {
			// if ac := app.Commands[x]; ac.Handler != nil {
			ac := app.Commands[x]
			fmt.Printf("\t%s '%s' %s\n\t\t%s\n\n",
				x, ac.Pattern,
				optTagList(ac.Opts),
				ac.Short)
			// }
		}
	} else {
		// some number of other commands were mentioned
		fmt.Println(
			"showing items mentioned alongside help in commandline:",
			tokens.GetSortedKeys(),
		)
		fmt.Println()
		var tags []string
		for i := range tokens {
			tags = append(tags, i)
		}
		sort.Strings(tags)
		for _, x := range tags {
			if x != "help" {
				fmt.Printf("%s '%s' %s\n\n\t%s\n",
					x, app.Commands[x].Pattern,
					optTagList(app.Commands[x].Opts),
					app.Commands[x].Short)
				fmt.Println("\n", app.Commands[x].Detail)
				fmt.Println()
			}
		}
	}
	return 0
}

func Conf(args []string, tokens Tokens, app *App) int {
	var r int
	for r = 2; r == 2; {
		r = Run(args, tokens, app)
	}
	return r
}

func New(args []string, tokens Tokens, app *App) int {
	fmt.Println("running New", args, getTokens(tokens))
	return 0
}

func Copy(args []string, tokens Tokens, app *App) int {
	fmt.Println("running Copy", args, getTokens(tokens))
	return 0
}

func List(args []string, tokens Tokens, app *App) int {
	if j := validateProxyListeners(app); j != 0 {
		return j
	}
	if _, ok := tokens["wallet"]; ok {
		app.Cats["wallet"]["enable"].Put(true)
	}
	ctl.ListCommands()
	return 0
}

func Ctl(args []string, tokens Tokens, app *App) int {
	// spew.Dump(app.Cats["app"])
	cl.Register.SetAllLevels(*app.Config.LogLevel)
	setAppDataDir(app, "ctl")
	if j := validateProxyListeners(app); j != 0 {
		return j
	}
	if _, ok := tokens["wallet"]; ok {
		*app.Config.Wallet = true
	}
	var i int
	var x string
	for i, x = range args {
		if app.Commands["ctl"].RE.Match([]byte(x)) {
			i++
			break
		}
	}
	ctl.Main(args[i:], app.Config)
	return 0
}

func Node(args []string, tokens Tokens, app *App) int {
	node.StateCfg = app.Config.State
	node.Cfg = app.Config
	cl.Register.SetAllLevels(*app.Config.LogLevel)
	// setAppDataDir(app, "node")
	_ = nine.ActiveNetParams //= activenetparams
	if validateWhitelists(app) != 0 ||
		validateProxyListeners(app) != 0 ||
		validatePasswords(app) != 0 ||
		validateRPCCredentials(app) != 0 ||
		validateBlockLimits(app) != 0 ||
		validateUAComments(app) != 0 ||
		validateMiner(app) != 0 ||
		validateCheckpoints(app) != 0 ||
		validateAddresses(app) != 0 ||
		validateDialers(app) != 0 {
		return 1
	}
	// run the node!
	if node.Main(nil) != nil {
		return 1
	}
	return 0
}

func Wallet(args []string, tokens Tokens, app *App) int {
	setAppDataDir(app, "wallet")

	// dbDir := walletmain.NetworkDir(*app.Config.AppDataDir, app.Config.ActiveNetParams.Params)
	netDir := walletmain.NetworkDir(*app.Config.AppDataDir, app.Config.ActiveNetParams.Params)
	wdb := netDir + "/wallet.db"
	log <- cl.Debug{"opening wallet:", wdb}
	if !FileExists(wdb) {
		if e := walletmain.CreateWallet(
			app.Config, app.Config.ActiveNetParams); e != nil {
			panic("could not create wallet " + e.Error())
		}
	} else {
		if e := walletmain.Main(app.Config, app.Config.ActiveNetParams); e != nil {
			return 1
		}
	}
	return 0
}

func Shell(args []string, tokens Tokens, app *App) int {
	fmt.Println("running Shell", args, getTokens(tokens))
	setAppDataDir(app, "wallet")
	// dbDir := walletmain.NetworkDir(*app.Config.AppDataDir, app.Config.ActiveNetParams.Params)
	netDir := walletmain.NetworkDir(*app.Config.AppDataDir, app.Config.ActiveNetParams.Params)
	wdb := netDir + "/wallet.db"
	log <- cl.Debug{"opening wallet:", wdb}
	if !FileExists(wdb) {
		if e := walletmain.CreateWallet(
			app.Config, app.Config.ActiveNetParams); e != nil {
			panic("could not create wallet " + e.Error())
		}
	} else {
		go Node(args, tokens, app)
		if e := walletmain.Main(app.Config, app.Config.ActiveNetParams); e != nil {
			return 1
		}
	}
	return 0
}

func Test(args []string, tokens Tokens, app *App) int {
	cl.Register.SetAllLevels(*app.Config.LogLevel)
	fmt.Println("running Test", args, getTokens(tokens))
	return 0
}

func Create(args []string, tokens Tokens, app *App) int {
	cl.Register.SetAllLevels(*app.Config.LogLevel)
	fmt.Println("running Create", args, getTokens(tokens))
	return 0
}

func TestHandler(args []string, tokens Tokens, app *App) int {
	return 0
}

func GUI(args []string, tokens Tokens, app *App) int {
	return 0
}

func Mine(args []string, tokens Tokens, app *App) int {
	return 0
}
func GenCerts(args []string, tokens Tokens, app *App) int {
	return 0
}
func GenCA(args []string, tokens Tokens, app *App) int {
	return 0
}
