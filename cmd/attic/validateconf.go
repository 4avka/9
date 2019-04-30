package cmd
import (
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"time"
	"git.parallelcoin.io/dev/9/cmd/node"
	blockchain "git.parallelcoin.io/dev/9/pkg/chain"
	"git.parallelcoin.io/dev/9/pkg/chain/fork"
	"git.parallelcoin.io/dev/9/pkg/peer/connmgr"
	"git.parallelcoin.io/dev/9/pkg/util"
	"git.parallelcoin.io/dev/9/pkg/util/cl"
	"github.com/btcsuite/go-socks/socks"
)
func setAppDataDir(name string) {
	if Config != nil {
		if Config.AppDataDir != nil {
			// set AppDataDir for running as node
			*Config.AppDataDir =
				CleanAndExpandPath(filepath.Join(*Config.DataDir, name))
		}
		if Config.LogDir != nil {
			*Config.LogDir = *Config.AppDataDir
		}
	}
}
func validateWhitelists() int {
	// Validate any given whitelisted IP addresses and networks.
	log <- cl.Debug{"validating whitelists"}
	if len(*Config.Whitelists) > 0 {
		var ip net.IP
		stateconfig.ActiveWhitelists =
			make([]*net.IPNet, 0, len(*Config.Whitelists))
		for _, addr := range *Config.Whitelists {
			_, ipnet, err := net.ParseCIDR(addr)
			if err != nil {
				err = fmt.Errorf("%s '%s'", cl.Ine(), err.Error())
				ip = net.ParseIP(addr)
				if ip == nil {
					str := err.Error() + " %s: the whitelist value of '%s' is invalid"
					err = fmt.Errorf(str, "runNode", addr)
					log <- cl.Err(err.Error())
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
			stateconfig.ActiveWhitelists =
				append(stateconfig.ActiveWhitelists, ipnet)
		}
	}
	return 0
}
func validateProxyListeners() int {
	// if proxy is not enabled, empty the proxy field as node sees presence as a
	// on switch
	if !*(*config)["proxy.enable"].Value.(*bool) {
		*Config.Proxy = ""
	}
	// if proxy is enabled or listeners list is empty, disable p2p listener
	if *Config.Proxy != "" || len(*Config.ConnectPeers) > 0 &&
		len(*Config.Listeners) < 1 {
		*Config.DisableListen = true
	}
	if *Config.DisableListen && len(*Config.Listeners) < 1 {
		*Config.Listeners = []string{
			net.JoinHostPort("127.0.0.1", node.DefaultPort),
		}
	}
	return 0
}
func validatePasswords() int {
	// Check to make sure limited and admin users don't have the same username
	log <- cl.Debug{"checking admin and limited username is different"}
	if *Config.Username != "" && *Config.Username == *Config.LimitUser {
		str := "%s: --username and --limituser must not specify the same username"
		err := fmt.Errorf(str, "runNode")
		log <- cl.Error{err}
		return 1
	}
	// Check to make sure limited and admin users don't have the same password
	log <- cl.Debug{"checking limited and admin passwords are not the same"}
	if *Config.Password != "" &&
		*Config.Password == *Config.LimitPass {
		str := "%s: --password and --limitpass must not specify the same password"
		err := fmt.Errorf(str, "runNode")
		log <- cl.Error{err}
		// fmt.Fprintln(os.Stderr, usageMessage)
		return 1
	}
	return 0
}
func validateRPCCredentials() int {
	// The RPC server is disabled if no username or password is provided.
	log <- cl.Debug{"checking rpc server has a login enabled"}
	if (*Config.Username == "" || *Config.Password == "") &&
		(*Config.LimitUser == "" || *Config.LimitPass == "") {
		*Config.DisableRPC = true
	}
	if *Config.DisableRPC {
		log <- cl.Inf("RPC service is disabled")
	}
	log <- cl.Debug{"checking rpc server has listeners set"}
	if !*Config.DisableRPC && len(*Config.RPCListeners) == 0 {
		log <- cl.Debug{"looking up default listener"}
		addrs, err := net.LookupHost(node.DefaultRPCListener)
		if err != nil {
			log <- cl.Error{err}
			return 1
		}
		*Config.RPCListeners = make([]string, 0, len(addrs))
		log <- cl.Debug{"setting listeners"}
		for _, addr := range addrs {
			addr = net.JoinHostPort(addr, activenetparams.RPCPort)
			*Config.RPCListeners = append(*Config.RPCListeners, addr)
		}
	}
	return 0
}
func validateBlockLimits() int {
	// Validate the the minrelaytxfee.
	log <- cl.Debug{"checking min relay tx fee"}
	var err error
	stateconfig.ActiveMinRelayTxFee, err = util.NewAmount(*Config.MinRelayTxFee)
	if err != nil {
		str := "%s: invalid minrelaytxfee: %v"
		err := fmt.Errorf(str, "runNode", err)
		log <- cl.Error{err}
		return 1
	}
	// Limit the block priority and minimum block sizes to max block size.
	log <- cl.Debug{
		"checking validating block priority and minimium size/weight"}
	*Config.BlockPrioritySize = int(minUint32(
		uint32(*Config.BlockPrioritySize),
		uint32(*Config.BlockMaxSize)))
	*Config.BlockMinSize = int(minUint32(
		uint32(*Config.BlockMinSize),
		uint32(*Config.BlockMaxSize)))
	*Config.BlockMinWeight = int(minUint32(
		uint32(*Config.BlockMinWeight),
		uint32(*Config.BlockMaxWeight)))
	switch {
	// If the max block size isn't set, but the max weight is, then we'll set the limit for the max block size to a safe limit so weight takes precedence.
	case *Config.BlockMaxSize == node.DefaultBlockMaxSize &&
		*Config.BlockMaxWeight != node.DefaultBlockMaxWeight:
		*Config.BlockMaxSize = blockchain.MaxBlockBaseSize - 1000
	// If the max block weight isn't set, but the block size is, then we'll scale the set weight accordingly based on the max block size value.
	case *Config.BlockMaxSize != node.DefaultBlockMaxSize &&
		*Config.BlockMaxWeight == node.DefaultBlockMaxWeight:
		*Config.BlockMaxWeight = *Config.BlockMaxSize * blockchain.WitnessScaleFactor
	}
	if *Config.RejectNonStd && *Config.RelayNonStd {
		fmt.Println("cannot both relay and reject nonstandard transactions")
		return 1
	}
	return 0
}
func validateUAComments() int {
	// Look for illegal characters in the user agent comments.
	log <- cl.Debug{"checking user agent comments"}
	for _, uaComment := range *Config.UserAgentComments {
		if strings.ContainsAny(uaComment, "/:()") {
			err := fmt.Errorf("%s: The following characters must not "+
				"appear in user agent comments: '/', ':', '(', ')'",
				"runNode")
			log <- cl.Err(err.Error())
			return 1
		}
	}
	return 0
}
func validateMiner() int {
	// Check mining addresses are valid and saved parsed versions.
	log <- cl.Debug{"checking mining addresses"}
	stateconfig.ActiveMiningAddrs =
		make([]util.Address, 0, len(*Config.MiningAddrs))
	if len(*Config.MiningAddrs) > 0 {
		for _, strAddr := range *Config.MiningAddrs {
			if len(strAddr) > 1 {
				addr, err := util.DecodeAddress(strAddr, activenetparams.Params)
				if err != nil {
					str := "%s: mining address '%s' failed to decode: %v"
					err := fmt.Errorf(str, "runNode", strAddr, err)
					log <- cl.Err(err.Error())
					return 1
				}
				if !addr.IsForNet(activenetparams.Params) {
					str := "%s: mining address '%s' is on the wrong network"
					err := fmt.Errorf(str, "runNode", strAddr)
					log <- cl.Error{err}
					return 1
				}
				stateconfig.ActiveMiningAddrs =
					append(stateconfig.ActiveMiningAddrs, addr)
			} else {
				*Config.MiningAddrs = []string{}
			}
		}
	}
	// Ensure there is at least one mining address when the generate flag
	// is set.
	if (*Config.Generate || len(*Config.MinerListener) > 1) && len(*Config.MiningAddrs) == 0 {
		str := "%s: the generate flag is set, but there are no mining addresses specified "
		err := fmt.Errorf(str, "runNode")
		log <- cl.Err(err.Error())
		return 1
	}
	if *Config.MinerPass != "" {
		stateconfig.ActiveMinerKey = fork.Argon2i([]byte(*Config.MinerPass))
	}
	return 0
}
func validateCheckpoints() int {
	var err error
	// Check the checkpoints for syntax errors.
	log <- cl.Debug{"checking the checkpoints"}
	stateconfig.AddedCheckpoints, err =
		node.ParseCheckpoints(*Config.AddCheckpoints)
	if err != nil {
		str := "%s: Error parsing checkpoints: %v"
		err := fmt.Errorf(str, "runNode", err)
		log <- cl.Err(err.Error())
		return 1
	}
	return 0
}
func validateDialers() int {
	if !*Config.Onion && *Config.OnionProxy != "" {
		log <- cl.Error{"cannot enable tor proxy without an address specified"}
		return 1
	}
	// Tor stream isolation requires either proxy or onion proxy to be set.
	if *Config.TorIsolation &&
		*Config.Proxy == "" &&
		*Config.OnionProxy == "" {
		str := "%s: Tor stream isolation requires either proxy or onionproxy to be set"
		err := fmt.Errorf(str, "runNode")
		log <- cl.Error{err}
		return 1
	}
	// Setup dial and DNS resolution (lookup) functions depending on the specified options.  The default is to use the standard net.DialTimeout function as well as the system DNS resolver.  When a proxy is specified, the dial function is set to the proxy specific dial function and the lookup is set to use tor (unless --noonion is specified in which case the system DNS resolver is used).
	log <- cl.Debug{"setting network dialer and lookup"}
	stateconfig.Dial = net.DialTimeout
	stateconfig.Lookup = net.LookupIP
	if *Config.Proxy != "" {
		fmt.Println("loading proxy")
		log <- cl.Debug{"we are loading a proxy!"}
		_, _, err := net.SplitHostPort(*Config.Proxy)
		if err != nil {
			str := "%s: Proxy address '%s' is invalid: %v"
			err := fmt.Errorf(str, "runNode", *Config.Proxy, err)
			log <- cl.Error{err}
			return 1
		}
		// Tor isolation flag means proxy credentials will be overridden unless there is also an onion proxy configured in which case that one will be overridden.
		torIsolation := false
		if *Config.TorIsolation &&
			*Config.OnionProxy == "" &&
			(*Config.ProxyUser != "" ||
				*Config.ProxyPass != "") {
			torIsolation = true
			log <- cl.Warn{
				"Tor isolation set -- overriding specified proxy user credentials"}
		}
		proxy := &socks.Proxy{
			Addr:         *Config.Proxy,
			Username:     *Config.ProxyUser,
			Password:     *Config.ProxyPass,
			TorIsolation: torIsolation,
		}
		stateconfig.Dial = proxy.DialTimeout
		// Treat the proxy as tor and perform DNS resolution through it unless the --noonion flag is set or there is an onion-specific proxy configured.
		if *Config.Onion &&
			*Config.OnionProxy != "" {
			stateconfig.Lookup = func(host string) ([]net.IP, error) {
				return connmgr.TorLookupIP(host, *Config.Proxy)
			}
		}
	}
	// Setup onion address dial function depending on the specified options. The default is to use the same dial function selected above.  However, when an onion-specific proxy is specified, the onion address dial function is set to use the onion-specific proxy while leaving the normal dial function as selected above.  This allows .onion address traffic to be routed through a different proxy than normal traffic.
	log <- cl.Debug{"setting up tor proxy if enabled"}
	if *Config.OnionProxy != "" {
		_, _, err := net.SplitHostPort(*Config.OnionProxy)
		if err != nil {
			str := "%s: Onion proxy address '%s' is invalid: %v"
			err := fmt.Errorf(str, "runNode", *Config.OnionProxy, err)
			log <- cl.Error{err}
			return 1
		}
		// Tor isolation flag means onion proxy credentials will be overriddenode.
		if *Config.TorIsolation &&
			(*Config.OnionProxyUser != "" || *Config.OnionProxyPass != "") {
			log <- cl.Warn{
				"Tor isolation set - overriding specified onionproxy user credentials "}
		}
		stateconfig.Oniondial =
			func(network, addr string, timeout time.Duration) (net.Conn, error) {
				proxy := &socks.Proxy{
					Addr:         *Config.OnionProxy,
					Username:     *Config.OnionProxyUser,
					Password:     *Config.OnionProxyPass,
					TorIsolation: *Config.TorIsolation,
				}
				return proxy.DialTimeout(network, addr, timeout)
			}
		// When configured in bridge mode (both --onion and --proxy are configured), it means that the proxy configured by --proxy is not a tor proxy, so override the DNS resolution to use the onion-specific proxy.
		if *Config.Proxy != "" {
			stateconfig.Lookup = func(host string) ([]net.IP, error) {
				return connmgr.TorLookupIP(host, *Config.OnionProxy)
			}
		}
	} else {
		stateconfig.Oniondial = stateconfig.Dial
	}
	// Specifying --noonion means the onion address dial function results in an error.
	if !*Config.Onion {
		stateconfig.Oniondial = func(a, b string, t time.Duration) (net.Conn, error) {
			return nil, errors.New("tor has been disabled")
		}
	}
	return 0
}
func validateAddresses() int {
	// TODO: simplify this to a boolean and one slice for config fercryinoutloud
	if len(*Config.AddPeers) > 0 && len(*Config.ConnectPeers) > 0 {
		fmt.Println("ERROR:", cl.Ine(),
			"cannot have addpeers at the same time as connectpeers")
		return 1
	}
	// Add default port to all rpc listener addresses if needed and remove duplicate addresses.
	log <- cl.Debug{"checking rpc listener addresses"}
	*Config.RPCListeners =
		node.NormalizeAddresses(*Config.RPCListeners,
			activenetparams.RPCPort)
	// Add default port to all listener addresses if needed and remove duplicate addresses.
	*Config.Listeners =
		node.NormalizeAddresses(*Config.Listeners,
			activenetparams.DefaultPort)
	// Add default port to all added peer addresses if needed and remove duplicate addresses.
	*Config.AddPeers =
		node.NormalizeAddresses(*Config.AddPeers,
			activenetparams.DefaultPort)
	*Config.ConnectPeers =
		node.NormalizeAddresses(*Config.ConnectPeers,
			activenetparams.DefaultPort)
	// --onionproxy and not --onion are contradictory (TODO: this is kinda stupid hm? switch *and* toggle by presence of flag value, one should be enough)
	return 0
}
