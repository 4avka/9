package conf

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"git.parallelcoin.io/dev/9/cmd/ctl"
	"git.parallelcoin.io/dev/9/cmd/defs"
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

// Help prints out help information based on the contents of the commandline
func Help(args []string, tokens defs.Tokens, app *defs.App) int {
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

// Conf runs the configuration menu system
func Conf(args []string, tokens Tokens, app *App) int {
	var r int
	for r = 2; r == 2; {
		r = conf.Run(args, tokens, app)
	}
	return r
}

// // New ???
// func New(args []string, tokens Tokens, app *App) int {
// 	fmt.Println("running New", args, getTokens(tokens))
// 	return 0
// }

// // Copy duplicates a configuration to create new one(s) based on it
// func Copy(args []string, tokens Tokens, app *App) int {
// 	fmt.Println("running Copy", args, getTokens(tokens))
// 	return 0
// }

// List prints the available commands for ctl
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

// Ctl sends RPC commands input in the command line arguments and prints the result
// back to stdout
func Ctl(args []string, tokens Tokens, app *App) int {
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

// Node launches the full node
func Node(args []string, tokens Tokens, app *App) int {
	node.StateCfg = app.Config.State
	node.Cfg = app.Config
	cl.Register.SetAllLevels(*app.Config.LogLevel)
	setAppDataDir(app, "node")
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

// Wallet launches the wallet server
func Wallet(args []string, tokens Tokens, app *App) int {
	setAppDataDir(app, "wallet")
	netDir := walletmain.NetworkDir(*app.Config.AppDataDir, app.Config.ActiveNetParams.Params)
	wdb := netDir // + "/wallet.db"
	log <- cl.Debug{"opening wallet:", wdb}
	if !FileExists(wdb) {
		if e := walletmain.CreateWallet(
			app.Config, app.Config.ActiveNetParams, wdb); e != nil {
			panic("could not create wallet " + e.Error())
		}
	} else {
		setAppDataDir(app, "node")
		if e := walletmain.Main(app.Config, app.Config.ActiveNetParams, netDir); e != nil {
			return 1
		}
	}
	return 0
}

// Shell runs a combined full node and wallet server for use in the common standard
// configuration provided by many bitcoin and bitcoin fork servers
func Shell(args []string, tokens Tokens, app *App) int {
	setAppDataDir(app, "node")
	netDir := walletmain.NetworkDir(filepath.Join(*app.Config.DataDir, "wallet"), app.Config.ActiveNetParams.Params)
	wdb := netDir // + "/wallet.db"
	log <- cl.Debug{"opening wallet:", wdb}
	if !FileExists(wdb) {
		if e := walletmain.CreateWallet(
			app.Config, app.Config.ActiveNetParams, wdb); e != nil {
			panic("could not create wallet " + e.Error())
		}
	} else {
		go Node(args, tokens, app)
		if e := walletmain.Main(app.Config, app.Config.ActiveNetParams, netDir); e != nil {
			return 1
		}
	}
	return 0
}

// Test runs a testnet based on a set of configuration directories
func Test(args []string, tokens Tokens, app *App) int {
	cl.Register.SetAllLevels(*app.Config.LogLevel)
	fmt.Println("running Test", args, getTokens(tokens))
	return 0
}

// Create generates a set of configurations that are set to connect to each other
// in a testnet
func Create(args []string, tokens Tokens, app *App) int {
	netDir := walletmain.NetworkDir(filepath.Join(*app.Config.DataDir, "wallet"), app.Config.ActiveNetParams.Params)
	wdb := netDir // + "/wallet.db"
	if !FileExists(wdb) {
		if e := walletmain.CreateWallet(
			app.Config, app.Config.ActiveNetParams, wdb); e != nil {
			panic("could not create wallet " + e.Error())
		}
	} else {
		fmt.Println("wallet already exists in", wdb+"/wallet.db", "refusing to overwrite")
		return 1
	}
	return 0
}

// // TestHandler ???
// func TestHandler(args []string, tokens Tokens, app *App) int {
// 	return 0
// }

// GUI runs a shell in the background and a GUI interface for wallet and node
func GUI(args []string, tokens Tokens, app *App) int {
	return 0
}

// Mine runs the standalone miner
func Mine(args []string, tokens Tokens, app *App) int {
	return 0
}

// GenCerts generates TLS certificates
func GenCerts(args []string, tokens Tokens, app *App) int {
	return 0
}

// GenCA creates a signing key that GenCerts will use if present to sign keys that
// it can be used to certify for multiple nodes connected to each other
// (wallet/node and RPC)
func GenCA(args []string, tokens Tokens, app *App) int {
	return 0
}

// MinUint32 is a helper function to return the minimum of two uint32s. This avoids a math import and the need to cast to floats.
func MinUint32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

func isWindows() bool {
	return runtime.GOOS == "windows"
}

// EnsureDir checks a file could be written to a path, creates the directories as needed
func EnsureDir(fileName string) bool {
	dirName := filepath.Dir(fileName)
	if _, serr := os.Stat(dirName); serr != nil {
		merr := os.MkdirAll(dirName, os.ModePerm)
		if merr != nil {
			panic(merr)
		}
		return true
	}
	return false
}

// FileExists reports whether the named file or directory exists.
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// CleanAndExpandPath expands environment variables and leading ~ in the passed path, cleans the result, and returns it.
func CleanAndExpandPath(path, datadir string) string {
	// Expand initial ~ to OS specific home directory.
	homeDir := filepath.Dir(util.AppDataDir("9", false))
	if strings.HasPrefix(path, "~") {
		return strings.Replace(path, "~", homeDir, 1)

	}
	if strings.HasPrefix(path, "./") {
		// explicitly prefix is this must be a relative path
		pwd, _ := os.Getwd()
		return filepath.Join(pwd, path)
	} else if !strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "\\") {
		if path != datadir {
			return filepath.Join(datadir, path)
		}
	}
	// NOTE: The os.ExpandEnv doesn't work with Windows-style %VARIABLE%, but they variables can still be expanded via POSIX-style $VARIABLE.
	path = filepath.Clean(os.ExpandEnv(path))
	return path
}

// NormalizeAddress returns addr with the passed default port appended if there is not already a port specified.
func NormalizeAddress(addr, defaultPort string) string {
	_, _, err := net.SplitHostPort(addr)
	if err != nil {
		return net.JoinHostPort(addr, defaultPort)
	}
	return addr
}

// NormalizeAddresses returns a new slice with all the passed peer addresses normalized with the given default port, and all duplicates removed.
func NormalizeAddresses(addrs []string, defaultPort string) []string {
	for i, addr := range addrs {
		addrs[i] = NormalizeAddress(addr, defaultPort)
	}
	return RemoveDuplicateAddresses(addrs)
}

// RemoveDuplicateAddresses returns a new slice with all duplicate entries in addrs removed.
func RemoveDuplicateAddresses(addrs []string) []string {
	result := make([]string, 0, len(addrs))
	seen := map[string]struct{}{}
	for _, val := range addrs {
		if _, ok := seen[val]; !ok {
			result = append(result, val)
			seen[val] = struct{}{}
		}
	}
	return result
}

func intersection(a, b []string) (out []string) {
	for _, x := range a {
		for _, y := range b {
			if x == y {
				out = append(out, x)
			}
		}
	}
	return
}

func uniq(elements []string) []string {
	// Use map to record duplicates as we find them.
	encountered := map[string]bool{}
	result := []string{}

	for v := range elements {
		if encountered[elements[v]] == true {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}

func setAppDataDir(app *App, name string) {
	if app != nil {
		if app.Config != nil {
			if app.Config.AppDataDir == nil {
				app.Config.AppDataDir = new(string)
				// set AppDataDir for running as node
				*app.Config.AppDataDir =
					CleanAndExpandPath(
						filepath.Join(*app.Config.DataDir, name),
						*app.Config.DataDir)
			}
			if app.Config.LogDir == nil {
				app.Config.LogDir = new(string)
				*app.Config.LogDir = *app.Config.AppDataDir
			}
		}
	}
}

func validateWhitelists(app *App) int {
	// Validate any given whitelisted IP addresses and networks.
	if app.Config.Whitelists != nil {
		var ip net.IP

		app.Config.State.ActiveWhitelists =
			make([]*net.IPNet, 0, len(*app.Config.Whitelists))
		for _, addr := range *app.Config.Whitelists {
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
			app.Config.State.ActiveWhitelists =
				append(app.Config.State.ActiveWhitelists, ipnet)
		}
	}
	return 0
}

func validateProxyListeners(app *App) int {
	// if proxy is not enabled, empty the proxy field as node sees presence as a
	// on switch
	if app.Config.Proxy != nil {
		*app.Config.Proxy = ""
	}
	// if proxy is enabled or listeners list is empty, or connect peers are set,
	// disable p2p listener
	if app.Config.Proxy != nil ||
		app.Config.ConnectPeers != nil ||
		app.Config.Listeners == nil {
		if app.Config.DisableListen == nil {
			acd := true
			app.Config.DisableListen = &acd
		} else {
			*app.Config.DisableListen = true
		}
	}
	if !*app.Config.DisableListen && len(*app.Config.Listeners) < 1 {
		*app.Config.Listeners = []string{
			net.JoinHostPort("127.0.0.1", node.DefaultPort),
		}
	}
	return 0
}

func validatePasswords(app *App) int {

	// Check to make sure limited and admin users don't have the same username
	if *app.Config.Username != "" && *app.Config.Username == *app.Config.LimitUser {
		str := "%s: --username and --limituser must not specify the same username"
		err := fmt.Errorf(str, "runNode")
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	// Check to make sure limited and admin users don't have the same password
	if *app.Config.Password != "" &&
		*app.Config.Password == *app.Config.LimitPass {
		str := "%s: --password and --limitpass must not specify the same password"
		err := fmt.Errorf(str, "runNode")
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func validateRPCCredentials(app *App) int {
	// The RPC server is disabled if no username or password is provided.
	if (*app.Config.Username == "" || *app.Config.Password == "") &&
		(*app.Config.LimitUser == "" || *app.Config.LimitPass == "") {
		*app.Config.DisableRPC = true
	}
	if *app.Config.DisableRPC {
	}
	if !*app.Config.DisableRPC && len(*app.Config.RPCListeners) == 0 {
		addrs, err := net.LookupHost(node.DefaultRPCListener)
		if err != nil {
			return 1
		}
		*app.Config.RPCListeners = make([]string, 0, len(addrs))
		for _, addr := range addrs {
			addr = net.JoinHostPort(addr, app.Config.ActiveNetParams.RPCPort)
			*app.Config.RPCListeners = append(*app.Config.RPCListeners, addr)
		}
	}
	return 0
}

func validateBlockLimits(app *App) int {
	// Validate the the minrelaytxfee.
	// log <- cl.Debug{"checking min relay tx fee"}
	var err error
	app.Config.State.ActiveMinRelayTxFee, err =
		util.NewAmount(*app.Config.MinRelayTxFee)
	if err != nil {
		str := "%s: invalid minrelaytxfee: %v"
		err := fmt.Errorf(str, "runNode", err)
		fmt.Println(err)
		return 1
	}
	// Limit the block priority and minimum block sizes to max block size.
	*app.Config.BlockPrioritySize = int(MinUint32(
		uint32(*app.Config.BlockPrioritySize),
		uint32(*app.Config.BlockMaxSize)))
	*app.Config.BlockMinSize = int(MinUint32(
		uint32(*app.Config.BlockMinSize),
		uint32(*app.Config.BlockMaxSize)))
	*app.Config.BlockMinWeight = int(MinUint32(
		uint32(*app.Config.BlockMinWeight),
		uint32(*app.Config.BlockMaxWeight)))
	switch {
	// If the max block size isn't set, but the max weight is, then we'll set the limit for the max block size to a safe limit so weight takes precedence.
	case *app.Config.BlockMaxSize == node.DefaultBlockMaxSize &&
		*app.Config.BlockMaxWeight != node.DefaultBlockMaxWeight:
		*app.Config.BlockMaxSize = blockchain.MaxBlockBaseSize - 1000
	// If the max block weight isn't set, but the block size is, then we'll scale the set weight accordingly based on the max block size value.
	case *app.Config.BlockMaxSize != node.DefaultBlockMaxSize &&
		*app.Config.BlockMaxWeight == node.DefaultBlockMaxWeight:
		*app.Config.BlockMaxWeight = *app.Config.BlockMaxSize * blockchain.WitnessScaleFactor
	}
	if *app.Config.RejectNonStd && *app.Config.RelayNonStd {
		fmt.Println("cannot both relay and reject nonstandard transactions")
		return 1
	}
	return 0
}

func validateUAComments(app *App) int {
	// Look for illegal characters in the user agent comments.
	// log <- cl.Debug{"checking user agent comments"}
	if app.Config.UserAgentComments != nil {
		for _, uaComment := range *app.Config.UserAgentComments {
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

func validateMiner(app *App) int {
	// Check mining addresses are valid and saved parsed versions.
	// log <- cl.Debug{"checking mining addresses"}
	if app.Config.MiningAddrs != nil {
		app.Config.State.ActiveMiningAddrs =
			make([]util.Address, 0, len(*app.Config.MiningAddrs))
		if len(*app.Config.MiningAddrs) > 0 {
			for _, strAddr := range *app.Config.MiningAddrs {
				if len(strAddr) > 1 {
					addr, err := util.DecodeAddress(strAddr, app.Config.ActiveNetParams.Params)
					if err != nil {
						str := "%s: mining address '%s' failed to decode: %v"
						err := fmt.Errorf(str, "runNode", strAddr, err)
						fmt.Fprintln(os.Stderr, err)
						return 1
					}
					if !addr.IsForNet(app.Config.ActiveNetParams.Params) {
						str := "%s: mining address '%s' is on the wrong network"
						err := fmt.Errorf(str, "runNode", strAddr)
						fmt.Fprintln(os.Stderr, err)
						return 1
					}
					app.Config.State.ActiveMiningAddrs =
						append(app.Config.State.ActiveMiningAddrs, addr)
				} else {
					*app.Config.MiningAddrs = []string{}
				}
			}
		}
	}
	// Ensure there is at least one mining address when the generate flag
	// is set.
	if (*app.Config.Generate ||
		app.Config.MinerListener != nil) &&
		app.Config.MiningAddrs != nil {
		str := "%s: the generate flag is set, but there are no mining addresses specified "
		err := fmt.Errorf(str, "runNode")
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if *app.Config.MinerPass != "" {
		app.Config.State.ActiveMinerKey = fork.Argon2i([]byte(*app.Config.MinerPass))
	}
	return 0
}

func validateCheckpoints(app *App) int {
	var err error
	// Check the checkpoints for syntax errors.
	// log <- cl.Debug{"checking the checkpoints"}
	if app.Config.AddCheckpoints != nil {
		app.Config.State.AddedCheckpoints, err =
			node.ParseCheckpoints(*app.Config.AddCheckpoints)
		if err != nil {
			str := "%s: Error parsing checkpoints: %v"
			err := fmt.Errorf(str, "runNode", err)
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	return 0
}

func validateDialers(app *App) int {
	// if !*app.Config.Onion && *app.Config.OnionProxy != "" {
	// 	// log <- cl.Error{"cannot enable tor proxy without an address specified"}
	// 	return 1
	// }

	// Tor stream isolation requires either proxy or onion proxy to be set.
	if *app.Config.TorIsolation &&
		app.Config.Proxy == nil {
		str := "%s: Tor stream isolation requires either proxy or onionproxy to be set"
		err := fmt.Errorf(str, "runNode")
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	// Setup dial and DNS resolution (lookup) functions depending on the specified options.  The default is to use the standard net.DialTimeout function as well as the system DNS resolver.  When a proxy is specified, the dial function is set to the proxy specific dial function and the lookup is set to use tor (unless --noonion is specified in which case the system DNS resolver is used).
	// log <- cl.Debug{"setting network dialer and lookup"}
	app.Config.State.Dial = net.DialTimeout
	app.Config.State.Lookup = net.LookupIP
	if app.Config.Proxy != nil {
		fmt.Println("loading proxy")
		// log <- cl.Debug{"we are loading a proxy!"}
		_, _, err := net.SplitHostPort(*app.Config.Proxy)
		if err != nil {
			str := "%s: Proxy address '%s' is invalid: %v"
			err := fmt.Errorf(str, "runNode", *app.Config.Proxy, err)
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		// Tor isolation flag means proxy credentials will be overridden unless
		// there is also an onion proxy configured in which case that one will be overridden.
		torIsolation := false
		if *app.Config.TorIsolation &&
			(app.Config.ProxyUser != nil ||
				app.Config.ProxyPass != nil) {
			torIsolation = true
			// log <- cl.Warn{
			// "Tor isolation set -- overriding specified proxy user credentials"}
		}
		proxy := &socks.Proxy{
			Addr:         *app.Config.Proxy,
			Username:     *app.Config.ProxyUser,
			Password:     *app.Config.ProxyPass,
			TorIsolation: torIsolation,
		}
		app.Config.State.Dial = proxy.DialTimeout
		// Treat the proxy as tor and perform DNS resolution through it unless the --noonion flag is set or there is an onion-specific proxy configured.
		if *app.Config.Onion &&
			*app.Config.OnionProxy != "" {
			app.Config.State.Lookup = func(host string) ([]net.IP, error) {
				return connmgr.TorLookupIP(host, *app.Config.Proxy)
			}
		}
	}
	// Setup onion address dial function depending on the specified options. The default is to use the same dial function selected above.  However, when an onion-specific proxy is specified, the onion address dial function is set to use the onion-specific proxy while leaving the normal dial function as selected above.  This allows .onion address traffic to be routed through a different proxy than normal traffic.
	// log <- cl.Debug{"setting up tor proxy if enabled"}
	if app.Config.OnionProxy != nil {
		_, _, err := net.SplitHostPort(*app.Config.OnionProxy)
		if err != nil {
			str := "%s: Onion proxy address '%s' is invalid: %v"
			err := fmt.Errorf(str, "runNode", *app.Config.OnionProxy, err)
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		// Tor isolation flag means onion proxy credentials will be overriddenode.
		if *app.Config.TorIsolation &&
			(*app.Config.OnionProxyUser != "" || *app.Config.OnionProxyPass != "") {
			// log <- cl.Warn{
			// "Tor isolation set - overriding specified onionproxy user credentials "}
		}
		app.Config.State.Oniondial =
			func(network, addr string, timeout time.Duration) (net.Conn, error) {
				proxy := &socks.Proxy{
					Addr:         *app.Config.OnionProxy,
					Username:     *app.Config.OnionProxyUser,
					Password:     *app.Config.OnionProxyPass,
					TorIsolation: *app.Config.TorIsolation,
				}
				return proxy.DialTimeout(network, addr, timeout)
			}
		// When configured in bridge mode (both --onion and --proxy are configured), it means that the proxy configured by --proxy is not a tor proxy, so override the DNS resolution to use the onion-specific proxy.
		if *app.Config.Proxy != "" {
			app.Config.State.Lookup = func(host string) ([]net.IP, error) {
				return connmgr.TorLookupIP(host, *app.Config.OnionProxy)
			}
		}
	} else {
		app.Config.State.Oniondial = app.Config.State.Dial
	}
	// Specifying --noonion means the onion address dial function results in an error.
	if !*app.Config.Onion {
		app.Config.State.Oniondial = func(a, b string, t time.Duration) (net.Conn, error) {
			return nil, errors.New("tor has been disabled")
		}
	}
	return 0
}

func validateAddresses(app *App) int {
	// TODO: simplify this to a boolean and one slice for config fercryinoutloud
	if app.Config.AddPeers != nil && app.Config.ConnectPeers != nil {
		fmt.Println("ERROR:", cl.Ine(),
			"cannot have addpeers at the same time as connectpeers")
		return 1
	}
	// Add default port to all rpc listener addresses if needed and remove duplicate addresses.
	// log <- cl.Debug{"checking rpc listener addresses"}
	*app.Config.RPCListeners =
		node.NormalizeAddresses(*app.Config.RPCListeners,
			app.Config.ActiveNetParams.RPCPort)
	// Add default port to all listener addresses if needed and remove duplicate addresses.
	if app.Config.Listeners != nil {
		*app.Config.Listeners =
			node.NormalizeAddresses(*app.Config.Listeners,
				app.Config.ActiveNetParams.DefaultPort)
	}
	// Add default port to all added peer addresses if needed and remove duplicate addresses.
	if app.Config.AddPeers != nil {
		*app.Config.AddPeers =
			node.NormalizeAddresses(*app.Config.AddPeers,
				app.Config.ActiveNetParams.DefaultPort)
	}
	if app.Config.ConnectPeers != nil {
		*app.Config.ConnectPeers =
			node.NormalizeAddresses(*app.Config.ConnectPeers,
				app.Config.ActiveNetParams.DefaultPort)
	}
	// --onionproxy and not --onion are contradictory (TODO: this is kinda stupid hm? switch *and* toggle by presence of flag value, one should be enough)
	return 0
}
