package app

import (
	"fmt"
	"path/filepath"
	"sort"

	"git.parallelcoin.io/dev/9/cmd/conf"
	"git.parallelcoin.io/dev/9/cmd/ctl"
	"git.parallelcoin.io/dev/9/cmd/def"
	"git.parallelcoin.io/dev/9/cmd/ll"
	"git.parallelcoin.io/dev/9/cmd/nine"
	"git.parallelcoin.io/dev/9/cmd/node"
	"git.parallelcoin.io/dev/9/cmd/walletmain"
	"git.parallelcoin.io/dev/9/pkg/util"
	"git.parallelcoin.io/dev/9/pkg/util/cl"
)

// Log is the logger for node
var Log = cl.NewSubSystem("cmd/config", ll.DEFAULT)
var log = Log.Ch

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

func getCommands(cmds def.Commands) (s []string) {
	for i := range cmds {
		s = append(s, i)
	}
	sort.Strings(s)
	return
}

func getTokens(cmds def.Tokens) (s []string) {
	for _, x := range cmds {
		s = append(s, x.Value)
	}
	sort.Strings(s)
	return
}

// Help prints out help information based on the contents of the commandline
func Help(args []string, tokens def.Tokens, ap *def.App) int {
	fmt.Println(ap.Name, ap.Version(), "-", ap.Tagline)
	fmt.Println()
	fmt.Println("help with", ap.Name)
	fmt.Println()
	if len(tokens) == 1 {
		// help was invoked
		var tags []string
		for i := range ap.Commands {
			tags = append(tags, i)
		}
		sort.Strings(tags)
		for _, x := range tags {
			// if ac := ap.Commands[x]; ac.Handler != nil {
			ac := ap.Commands[x]
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
					x, ap.Commands[x].Pattern,
					optTagList(ap.Commands[x].Opts),
					ap.Commands[x].Short)
				fmt.Println("\n", ap.Commands[x].Detail)
				fmt.Println()
			}
		}
	}
	return 0
}

// Conf runs the configuration menu system
func Conf(args []string, tokens def.Tokens, ap *def.App) int {
	var r int
	for r = 2; r == 2; {
		r = conf.Run(args, tokens, ap)
	}
	return r
}

// // New ???
// func New(args []string, tokens def.Tokens, ap *def.App) int {
// 	fmt.Println("running New", args, getTokens(tokens))
// 	return 0
// }

// // Copy duplicates a configuration to create new one(s) based on it
// func Copy(args []string, tokens def.Tokens, ap *def.App) int {
// 	fmt.Println("running Copy", args, getTokens(tokens))
// 	return 0
// }

// List prints the available commands for ctl
func List(args []string, tokens def.Tokens, ap *def.App) int {
	if j := validateProxyListeners(ap); j != 0 {
		return j
	}
	if _, ok := tokens["wallet"]; ok {
		ap.Cats["wallet"]["enable"].Put(true)
	}
	ctl.ListCommands()
	return 0
}

// Ctl sends RPC commands input in the command line arguments and prints the result
// back to stdout
func Ctl(args []string, tokens def.Tokens, ap *def.App) int {
	cl.Register.SetAllLevels(*ap.Config.LogLevel)
	setAppDataDir(ap, "ctl")
	if j := validateProxyListeners(ap); j != 0 {
		return j
	}
	if _, ok := tokens["wallet"]; ok {
		*ap.Config.Wallet = true
	}
	var i int
	var x string
	for i, x = range args {
		if ap.Commands["ctl"].RE.Match([]byte(x)) {
			i++
			break
		}
	}
	ctl.Main(args[i:], ap.Config)
	return 0
}

// Node launches the full node
func Node(args []string, tokens def.Tokens, ap *def.App) int {
	node.StateCfg = ap.Config.State
	node.Cfg = ap.Config
	cl.Register.SetAllLevels(*ap.Config.LogLevel)
	setAppDataDir(ap, "node")
	_ = nine.ActiveNetParams //= activenetparams
	if validateWhitelists(ap) != 0 ||
		validateProxyListeners(ap) != 0 ||
		validatePasswords(ap) != 0 ||
		validateRPCCredentials(ap) != 0 ||
		validateBlockLimits(ap) != 0 ||
		validateUAComments(ap) != 0 ||
		validateMiner(ap) != 0 ||
		validateCheckpoints(ap) != 0 ||
		validateAddresses(ap) != 0 ||
		validateDialers(ap) != 0 {
		return 1
	}
	// run the node!
	ap.Started = make(chan struct{})
	go node.Main(nil, ap.Started)
	return 0
}

// Wallet launches the wallet server
func Wallet(args []string, tokens def.Tokens, ap *def.App) int {
	setAppDataDir(ap, "wallet")
	netDir := walletmain.NetworkDir(*ap.Config.AppDataDir,
		ap.Config.ActiveNetParams.Params)
	wdb := netDir // + "/wallet.db"
	log <- cl.Debug{"opening wallet:", wdb}
	if !util.FileExists(wdb) {
		if e := walletmain.CreateWallet(
			ap.Config, ap.Config.ActiveNetParams, wdb); e != nil {
			panic("could not create wallet " + e.Error())
		}
	} else {
		setAppDataDir(ap, "node")
		if e := walletmain.Main(ap.Config, ap.Config.ActiveNetParams, netDir); e != nil {
			return 1
		}
	}
	return 0
}

// Shell runs a combined full node and wallet server for use in the common standard
// configuration provided by many bitcoin and bitcoin fork servers
func Shell(args []string, tokens def.Tokens, ap *def.App) int {
	setAppDataDir(ap, "node")
	netDir := walletmain.NetworkDir(
		filepath.Join(*ap.Config.DataDir, "wallet"),
		ap.Config.ActiveNetParams.Params)
	wdb := netDir // + "/wallet.db"
	log <- cl.Debug{"opening wallet:", wdb}
	if !util.FileExists(wdb) {
		if e := walletmain.CreateWallet(
			ap.Config, ap.Config.ActiveNetParams, wdb); e != nil {
			panic("could not create wallet " + e.Error())
		}
	} else {
		Node(args, tokens, ap)
		<-ap.Started
		log <- cl.Info{"starting wallet server"}
		if e := walletmain.Main(ap.Config, ap.Config.ActiveNetParams, netDir); e != nil {
			return 1
		}
	}
	return 0
}

// Test runs a testnet based on a set of configuration directories
func Test(args []string, tokens def.Tokens, ap *def.App) int {
	cl.Register.SetAllLevels(*ap.Config.LogLevel)
	fmt.Println("running Test", args, getTokens(tokens))
	return 0
}

// Create generates a set of configurations that are set to connect to each other
// in a testnet
func Create(args []string, tokens def.Tokens, ap *def.App) int {
	netDir := walletmain.NetworkDir(
		filepath.Join(*ap.Config.DataDir, "wallet"),
		ap.Config.ActiveNetParams.Params)
	wdb := netDir // + "/wallet.db"
	if !util.FileExists(wdb) {
		if e := walletmain.CreateWallet(
			ap.Config, ap.Config.ActiveNetParams, wdb); e != nil {
			panic("could not create wallet " + e.Error())
		}
	} else {
		fmt.Println("wallet already exists in", wdb+"/wallet.db", "refusing to overwrite")
		return 1
	}
	return 0
}

// // TestHandler ???
// func TestHandler(args []string, tokens def.Tokens, ap *def.App) int {
// 	return 0
// }

// GUI runs a shell in the background and a GUI interface for wallet and node
func GUI(args []string, tokens def.Tokens, ap *def.App) int {
	return 0
}

// Mine runs the standalone miner
func Mine(args []string, tokens def.Tokens, ap *def.App) int {
	return 0
}

// GenCerts generates TLS certificates
func GenCerts(args []string, tokens def.Tokens, ap *def.App) int {
	return 0
}

// GenCA creates a signing key that GenCerts will use if present to sign keys that
// it can be used to certify for multiple nodes connected to each other
// (wallet/node and RPC)
func GenCA(args []string, tokens def.Tokens, ap *def.App) int {
	return 0
}
