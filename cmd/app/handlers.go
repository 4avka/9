package app

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"git.parallelcoin.io/dev/9/cmd/conf"
	"git.parallelcoin.io/dev/9/cmd/ctl"
	"git.parallelcoin.io/dev/9/cmd/def"
	"git.parallelcoin.io/dev/9/cmd/ll"
	"git.parallelcoin.io/dev/9/cmd/nine"
	"git.parallelcoin.io/dev/9/cmd/node"
	"git.parallelcoin.io/dev/9/cmd/walletmain"
	blockchain "git.parallelcoin.io/dev/9/pkg/chain"
	"git.parallelcoin.io/dev/9/pkg/chain/fork"
	"git.parallelcoin.io/dev/9/pkg/peer/connmgr"
	"git.parallelcoin.io/dev/9/pkg/util"
	"git.parallelcoin.io/dev/9/pkg/util/cl"
	"github.com/btcsuite/go-socks/socks"
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
	if node.Main(nil) != nil {
		return 1
	}
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
		go Node(args, tokens, ap)
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

func setAppDataDir(ap *def.App, name string) {
	if ap != nil {
		if ap.Config != nil {
			if ap.Config.AppDataDir == nil {
				ap.Config.AppDataDir = new(string)
				// set AppDataDir for running as node
				*ap.Config.AppDataDir =
					util.CleanAndExpandPath(
						filepath.Join(*ap.Config.DataDir, name),
						*ap.Config.DataDir)
			}
			if ap.Config.LogDir == nil {
				ap.Config.LogDir = new(string)
				*ap.Config.LogDir = *ap.Config.AppDataDir
			}
		}
	}
}

func validateWhitelists(ap *def.App) int {
	// Validate any given whitelisted IP addresses and networks.
	if ap.Config.Whitelists != nil {
		var ip net.IP

		ap.Config.State.ActiveWhitelists =
			make([]*net.IPNet, 0, len(*ap.Config.Whitelists))
		for _, addr := range *ap.Config.Whitelists {
			_, ipnet, err := net.ParseCIDR(addr)
			if err != nil {
				err = fmt.Errorf("%s '%s'", cl.Ine(), err.Error())
				ip = net.ParseIP(addr)
				if ip == nil {
					str := err.Error() + " %s: the whitelist value of '%s' is invalid"
					err = fmt.Errorf(str, "runNode", addr)
					return 1
				}
				var bits int
				if ip.To4() == nil {
					// IPv6
					bits = 128
				} else {
					bits = 32
				}
				ipnet = &net.IPNet{
					IP:   ip,
					Mask: net.CIDRMask(bits, bits),
				}
			}
			ap.Config.State.ActiveWhitelists =
				append(ap.Config.State.ActiveWhitelists, ipnet)
		}
	}
	return 0
}

func validateProxyListeners(ap *def.App) int {
	// if proxy is not enabled, empty the proxy field as node sees presence as a
	// on switch
	if ap.Config.Proxy != nil {
		*ap.Config.Proxy = ""
	}
	// if proxy is enabled or listeners list is empty, or connect peers are set,
	// disable p2p listener
	if ap.Config.Proxy != nil ||
		ap.Config.ConnectPeers != nil ||
		ap.Config.Listeners == nil {
		if ap.Config.DisableListen == nil {
			acd := true
			ap.Config.DisableListen = &acd
		} else {
			*ap.Config.DisableListen = true
		}
	}
	if !*ap.Config.DisableListen && len(*ap.Config.Listeners) < 1 {
		*ap.Config.Listeners = []string{
			net.JoinHostPort("127.0.0.1", node.DefaultPort),
		}
	}
	return 0
}

func validatePasswords(ap *def.App) int {

	// Check to make sure limited and admin users don't have the same username
	if *ap.Config.Username != "" && *ap.Config.Username == *ap.Config.LimitUser {
		str := "%s: --username and --limituser must not specify the same username"
		err := fmt.Errorf(str, "runNode")
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	// Check to make sure limited and admin users don't have the same password
	if *ap.Config.Password != "" &&
		*ap.Config.Password == *ap.Config.LimitPass {
		str := "%s: --password and --limitpass must not specify the same password"
		err := fmt.Errorf(str, "runNode")
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func validateRPCCredentials(ap *def.App) int {
	// The RPC server is disabled if no username or password is provided.
	if (*ap.Config.Username == "" || *ap.Config.Password == "") &&
		(*ap.Config.LimitUser == "" || *ap.Config.LimitPass == "") {
		*ap.Config.DisableRPC = true
	}
	if *ap.Config.DisableRPC {
	}
	if !*ap.Config.DisableRPC && len(*ap.Config.RPCListeners) == 0 {
		addrs, err := net.LookupHost(node.DefaultRPCListener)
		if err != nil {
			return 1
		}
		*ap.Config.RPCListeners = make([]string, 0, len(addrs))
		for _, addr := range addrs {
			addr = net.JoinHostPort(addr, ap.Config.ActiveNetParams.RPCPort)
			*ap.Config.RPCListeners = append(*ap.Config.RPCListeners, addr)
		}
	}
	return 0
}

func validateBlockLimits(ap *def.App) int {
	// Validate the the minrelaytxfee.
	// log <- cl.Debug{"checking min relay tx fee"}
	var err error
	ap.Config.State.ActiveMinRelayTxFee, err =
		util.NewAmount(*ap.Config.MinRelayTxFee)
	if err != nil {
		str := "%s: invalid minrelaytxfee: %v"
		err := fmt.Errorf(str, "runNode", err)
		fmt.Println(err)
		return 1
	}
	// Limit the block priority and minimum block sizes to max block size.
	*ap.Config.BlockPrioritySize = int(util.MinUint32(
		uint32(*ap.Config.BlockPrioritySize),
		uint32(*ap.Config.BlockMaxSize)))
	*ap.Config.BlockMinSize = int(util.MinUint32(
		uint32(*ap.Config.BlockMinSize),
		uint32(*ap.Config.BlockMaxSize)))
	*ap.Config.BlockMinWeight = int(util.MinUint32(
		uint32(*ap.Config.BlockMinWeight),
		uint32(*ap.Config.BlockMaxWeight)))
	switch {
	// If the max block size isn't set, but the max weight is, then we'll set the limit for the max block size to a safe limit so weight takes precedence.
	case *ap.Config.BlockMaxSize == node.DefaultBlockMaxSize &&
		*ap.Config.BlockMaxWeight != node.DefaultBlockMaxWeight:
		*ap.Config.BlockMaxSize = blockchain.MaxBlockBaseSize - 1000
	// If the max block weight isn't set, but the block size is, then we'll scale the set weight accordingly based on the max block size value.
	case *ap.Config.BlockMaxSize != node.DefaultBlockMaxSize &&
		*ap.Config.BlockMaxWeight == node.DefaultBlockMaxWeight:
		*ap.Config.BlockMaxWeight = *ap.Config.BlockMaxSize *
			blockchain.WitnessScaleFactor
	}
	if *ap.Config.RejectNonStd && *ap.Config.RelayNonStd {
		fmt.Println("cannot both relay and reject nonstandard transactions")
		return 1
	}
	return 0
}

func validateUAComments(ap *def.App) int {
	// Look for illegal characters in the user agent comments.
	// log <- cl.Debug{"checking user agent comments"}
	if ap.Config.UserAgentComments != nil {
		for _, uaComment := range *ap.Config.UserAgentComments {
			if strings.ContainsAny(uaComment, "/:()") {
				err := fmt.Errorf("%s: The following characters must not "+
					"appear in user agent comments: '/', ':', '(', ')'",
					"runNode")
				fmt.Fprintln(os.Stderr, err)
				return 1
			}
		}
	}
	return 0
}

func validateMiner(ap *def.App) int {
	// Check mining addresses are valid and saved parsed versions.
	// log <- cl.Debug{"checking mining addresses"}
	if ap.Config.MiningAddrs != nil {
		ap.Config.State.ActiveMiningAddrs =
			make([]util.Address, 0, len(*ap.Config.MiningAddrs))
		if len(*ap.Config.MiningAddrs) > 0 {
			for _, strAddr := range *ap.Config.MiningAddrs {
				if len(strAddr) > 1 {
					addr, err := util.DecodeAddress(strAddr,
						ap.Config.ActiveNetParams.Params)
					if err != nil {
						str := "%s: mining address '%s' failed to decode: %v"
						err := fmt.Errorf(str, "runNode", strAddr, err)
						fmt.Fprintln(os.Stderr, err)
						return 1
					}
					if !addr.IsForNet(ap.Config.ActiveNetParams.Params) {
						str := "%s: mining address '%s' is on the wrong network"
						err := fmt.Errorf(str, "runNode", strAddr)
						fmt.Fprintln(os.Stderr, err)
						return 1
					}
					ap.Config.State.ActiveMiningAddrs =
						append(ap.Config.State.ActiveMiningAddrs, addr)
				} else {
					*ap.Config.MiningAddrs = []string{}
				}
			}
		}
	}
	// Ensure there is at least one mining address when the generate flag
	// is set.
	if (*ap.Config.Generate ||
		ap.Config.MinerListener != nil) &&
		ap.Config.MiningAddrs != nil {
		str := "%s: the generate flag is set, but there are no mining addresses specified "
		err := fmt.Errorf(str, "runNode")
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if *ap.Config.MinerPass != "" {
		ap.Config.State.ActiveMinerKey = fork.Argon2i([]byte(*ap.Config.MinerPass))
	}
	return 0
}

func validateCheckpoints(ap *def.App) int {
	var err error
	// Check the checkpoints for syntax errors.
	// log <- cl.Debug{"checking the checkpoints"}
	if ap.Config.AddCheckpoints != nil {
		ap.Config.State.AddedCheckpoints, err =
			node.ParseCheckpoints(*ap.Config.AddCheckpoints)
		if err != nil {
			str := "%s: Error parsing checkpoints: %v"
			err := fmt.Errorf(str, "runNode", err)
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	return 0
}

func validateDialers(ap *def.App) int {
	// if !*Config.Onion && *Config.OnionProxy != "" {
	// 	// log <- cl.Error{"cannot enable tor proxy without an address specified"}
	// 	return 1
	// }

	// Tor stream isolation requires either proxy or onion proxy to be set.
	if *ap.Config.TorIsolation &&
		ap.Config.Proxy == nil {
		str := "%s: Tor stream isolation requires either proxy or onionproxy to be set"
		err := fmt.Errorf(str, "runNode")
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	// Setup dial and DNS resolution (lookup) functions depending on the specified options.  The default is to use the standard net.DialTimeout function as well as the system DNS resolver.  When a proxy is specified, the dial function is set to the proxy specific dial function and the lookup is set to use tor (unless --noonion is specified in which case the system DNS resolver is used).
	// log <- cl.Debug{"setting network dialer and lookup"}
	ap.Config.State.Dial = net.DialTimeout
	ap.Config.State.Lookup = net.LookupIP
	if ap.Config.Proxy != nil {
		fmt.Println("loading proxy")
		// log <- cl.Debug{"we are loading a proxy!"}
		_, _, err := net.SplitHostPort(*ap.Config.Proxy)
		if err != nil {
			str := "%s: Proxy address '%s' is invalid: %v"
			err := fmt.Errorf(str, "runNode", *ap.Config.Proxy, err)
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		// Tor isolation flag means proxy credentials will be overridden unless
		// there is also an onion proxy configured in which case that one will be overridden.
		torIsolation := false
		if *ap.Config.TorIsolation &&
			(ap.Config.ProxyUser != nil ||
				ap.Config.ProxyPass != nil) {
			torIsolation = true
			// log <- cl.Warn{
			// "Tor isolation set -- overriding specified proxy user credentials"}
		}
		proxy := &socks.Proxy{
			Addr:         *ap.Config.Proxy,
			Username:     *ap.Config.ProxyUser,
			Password:     *ap.Config.ProxyPass,
			TorIsolation: torIsolation,
		}
		ap.Config.State.Dial = proxy.DialTimeout
		// Treat the proxy as tor and perform DNS resolution through it unless the --noonion flag is set or there is an onion-specific proxy configured.
		if *ap.Config.Onion &&
			*ap.Config.OnionProxy != "" {
			ap.Config.State.Lookup = func(host string) ([]net.IP, error) {
				return connmgr.TorLookupIP(host, *ap.Config.Proxy)
			}
		}
	}
	// Setup onion address dial function depending on the specified options. The default is to use the same dial function selected above.  However, when an onion-specific proxy is specified, the onion address dial function is set to use the onion-specific proxy while leaving the normal dial function as selected above.  This allows .onion address traffic to be routed through a different proxy than normal traffic.
	// log <- cl.Debug{"setting up tor proxy if enabled"}
	if ap.Config.OnionProxy != nil {
		_, _, err := net.SplitHostPort(*ap.Config.OnionProxy)
		if err != nil {
			str := "%s: Onion proxy address '%s' is invalid: %v"
			err := fmt.Errorf(str, "runNode", *ap.Config.OnionProxy, err)
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		// Tor isolation flag means onion proxy credentials will be overriddenode.
		if *ap.Config.TorIsolation &&
			(*ap.Config.OnionProxyUser != "" || *ap.Config.OnionProxyPass != "") {
			// log <- cl.Warn{
			// "Tor isolation set - overriding specified onionproxy user credentials "}
		}
		ap.Config.State.Oniondial =
			func(network, addr string, timeout time.Duration) (net.Conn, error) {
				proxy := &socks.Proxy{
					Addr:         *ap.Config.OnionProxy,
					Username:     *ap.Config.OnionProxyUser,
					Password:     *ap.Config.OnionProxyPass,
					TorIsolation: *ap.Config.TorIsolation,
				}
				return proxy.DialTimeout(network, addr, timeout)
			}
		// When configured in bridge mode (both --onion and --proxy are configured), it means that the proxy configured by --proxy is not a tor proxy, so override the DNS resolution to use the onion-specific proxy.
		if *ap.Config.Proxy != "" {
			ap.Config.State.Lookup = func(host string) ([]net.IP, error) {
				return connmgr.TorLookupIP(host, *ap.Config.OnionProxy)
			}
		}
	} else {
		ap.Config.State.Oniondial = ap.Config.State.Dial
	}
	// Specifying --noonion means the onion address dial function results in an error.
	if !*ap.Config.Onion {
		ap.Config.State.Oniondial = func(a, b string, t time.Duration) (net.Conn, error) {
			return nil, errors.New("tor has been disabled")
		}
	}
	return 0
}

func validateAddresses(ap *def.App) int {
	// TODO: simplify this to a boolean and one slice for config fercryinoutloud
	if ap.Config.AddPeers != nil && ap.Config.ConnectPeers != nil {
		fmt.Println("ERROR:", cl.Ine(),
			"cannot have addpeers at the same time as connectpeers")
		return 1
	}
	// Add default port to all rpc listener addresses if needed and remove duplicate addresses.
	// log <- cl.Debug{"checking rpc listener addresses"}
	*ap.Config.RPCListeners =
		node.NormalizeAddresses(*ap.Config.RPCListeners,
			ap.Config.ActiveNetParams.RPCPort)
	// Add default port to all listener addresses if needed and remove duplicate addresses.
	if ap.Config.Listeners != nil {
		*ap.Config.Listeners =
			node.NormalizeAddresses(*ap.Config.Listeners,
				ap.Config.ActiveNetParams.DefaultPort)
	}
	// Add default port to all added peer addresses if needed and remove duplicate addresses.
	if ap.Config.AddPeers != nil {
		*ap.Config.AddPeers =
			node.NormalizeAddresses(*ap.Config.AddPeers,
				ap.Config.ActiveNetParams.DefaultPort)
	}
	if ap.Config.ConnectPeers != nil {
		*ap.Config.ConnectPeers =
			node.NormalizeAddresses(*ap.Config.ConnectPeers,
				ap.Config.ActiveNetParams.DefaultPort)
	}
	// --onionproxy and not --onion are contradictory (TODO: this is kinda stupid hm? switch *and* toggle by presence of flag value, one should be enough)
	return 0
}
