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

func runNode(args []string, tokens Tokens, cmds, all Commands) int {
	if *config.RejectNonStd && *config.RelayNonStd {
		fmt.Println("cannot both relay and reject nonstandard transactions")
		return 1
	}
	// set AppDataDir for running as node
	*config.AppDataDir =
		CleanAndExpandPath(filepath.Join(*config.DataDir, "node"))
	*config.LogDir = *config.AppDataDir
	// Validate any given whitelisted IP addresses and networks.
	log <- cl.Debug{"validating whitelists"}
	if len(*config.Whitelists) > 0 {
		var ip net.IP
		stateconfig.ActiveWhitelists =
			make([]*net.IPNet, 0, len(*config.Whitelists))
		for _, addr := range *config.Whitelists {
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
	// TODO: simplify this to a boolean and one slice for config fercryinoutloud
	if len(*config.AddPeers) > 0 && len(*config.ConnectPeers) > 0 {
		fmt.Println("ERROR:", cl.Ine(),
			"cannot have addpeers at the same time as connectpeers")
	}
	// if proxy is not enabled, empty the proxy field as node sees presence as a
	// on switch
	if !*Config["proxy.enable"].Value.(*bool) {
		*config.Proxy = ""
	}
	// if proxy is enabled or listeners list is empty, disable p2p listener
	if *config.Proxy != "" || len(*config.ConnectPeers) > 0 &&
		len(*config.Listeners) < 1 {
		*config.DisableListen = true
	}
	if *config.DisableListen && len(*config.Listeners) < 1 {
		*config.Listeners = []string{
			net.JoinHostPort("127.0.0.1", node.DefaultPort),
		}
	}
	// Check to make sure limited and admin users don't have the same username
	log <- cl.Debug{"checking admin and limited username is different"}
	if *config.Username != "" && *config.Username == *config.LimitUser {
		str := "%s: --username and --limituser must not specify the same username"
		err := fmt.Errorf(str, "runNode")
		log <- cl.Error{err}
		return 1
	}
	// Check to make sure limited and admin users don't have the same password
	log <- cl.Debug{"checking limited and admin passwords are not the same"}
	if *config.Password != "" &&
		*config.Password == *config.LimitPass {
		str := "%s: --password and --limitpass must not specify the same password"
		err := fmt.Errorf(str, "runNode")
		log <- cl.Error{err}
		// fmt.Fprintln(os.Stderr, usageMessage)
		return 1
	}
	// The RPC server is disabled if no username or password is provided.
	log <- cl.Debug{"checking rpc server has a login enabled"}
	if (*config.Username == "" || *config.Password == "") &&
		(*config.LimitUser == "" || *config.LimitPass == "") {
		*config.DisableRPC = true
	}
	if *config.DisableRPC {
		log <- cl.Inf("RPC service is disabled")
	}
	log <- cl.Debug{"checking rpc server has listeners set"}
	if !*config.DisableRPC && len(*config.RPCListeners) == 0 {
		log <- cl.Debug{"looking up default listener"}
		addrs, err := net.LookupHost(node.DefaultRPCListener)
		if err != nil {
			log <- cl.Error{err}
			return 1
		}
		*config.RPCListeners = make([]string, 0, len(addrs))
		log <- cl.Debug{"setting listeners"}
		for _, addr := range addrs {
			addr = net.JoinHostPort(addr, activenetparams.RPCClientPort)
			*config.RPCListeners = append(*config.RPCListeners, addr)
		}
	}
	// Validate the the minrelaytxfee.
	log <- cl.Debug{"checking min relay tx fee"}
	var err error
	stateconfig.ActiveMinRelayTxFee, err = util.NewAmount(*config.MinRelayTxFee)
	if err != nil {
		str := "%s: invalid minrelaytxfee: %v"
		err := fmt.Errorf(str, "runNode", err)
		log <- cl.Error{err}
		return 1
	}
	// Limit the block priority and minimum block sizes to max block size.
	log <- cl.Debug{"checking validating block priority and minimium size/weight"}
	*config.BlockPrioritySize = int(minUint32(
		uint32(*config.BlockPrioritySize),
		uint32(*config.BlockMaxSize)))
	*config.BlockMinSize = int(minUint32(
		uint32(*config.BlockMinSize),
		uint32(*config.BlockMaxSize)))
	*config.BlockMinWeight = int(minUint32(
		uint32(*config.BlockMinWeight),
		uint32(*config.BlockMaxWeight)))
	switch {
	// If the max block size isn't set, but the max weight is, then we'll set the limit for the max block size to a safe limit so weight takes precedence.
	case *config.BlockMaxSize == node.DefaultBlockMaxSize &&
		*config.BlockMaxWeight != node.DefaultBlockMaxWeight:
		*config.BlockMaxSize = blockchain.MaxBlockBaseSize - 1000
	// If the max block weight isn't set, but the block size is, then we'll scale the set weight accordingly based on the max block size value.
	case *config.BlockMaxSize != node.DefaultBlockMaxSize &&
		*config.BlockMaxWeight == node.DefaultBlockMaxWeight:
		*config.BlockMaxWeight = *config.BlockMaxSize * blockchain.WitnessScaleFactor
	}
	// Look for illegal characters in the user agent comments.
	log <- cl.Debug{"checking user agent comments"}
	for _, uaComment := range *config.UserAgentComments {
		if strings.ContainsAny(uaComment, "/:()") {
			err := fmt.Errorf("%s: The following characters must not "+
				"appear in user agent comments: '/', ':', '(', ')'",
				"runNode")
			log <- cl.Err(err.Error())
			return 1
		}
	}
	// Check mining addresses are valid and saved parsed versions.
	log <- cl.Debug{"checking mining addresses"}
	stateconfig.ActiveMiningAddrs =
		make([]util.Address, 0, len(*config.MiningAddrs))
	if len(*config.MiningAddrs) > 0 {
		for _, strAddr := range *config.MiningAddrs {
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
				*config.MiningAddrs = []string{}
			}
		}
	}
	// Ensure there is at least one mining address when the generate flag
	// is set.
	if (*config.Generate || len(*config.MinerListener) > 1) && len(*config.MiningAddrs) == 0 {
		str := "%s: the generate flag is set, but there are no mining addresses specified "
		err := fmt.Errorf(str, "runNode")
		log <- cl.Err(err.Error())
		return 1
	}
	if *config.MinerPass != "" {
		stateconfig.ActiveMinerKey = fork.Argon2i([]byte(*config.MinerPass))
	}
	// Add default port to all rpc listener addresses if needed and remove duplicate addresses.
	log <- cl.Debug{"checking rpc listener addresses"}
	*config.RPCListeners = node.NormalizeAddresses(*config.RPCListeners,
		activenetparams.RPCClientPort)
	// Add default port to all listener addresses if needed and remove duplicate addresses.
	*config.Listeners = node.NormalizeAddresses(*config.Listeners,
		activenetparams.DefaultPort)
	// Add default port to all added peer addresses if needed and remove duplicate addresses.
	*config.AddPeers = node.NormalizeAddresses(*config.AddPeers,
		activenetparams.DefaultPort)
	*config.ConnectPeers = node.NormalizeAddresses(*config.ConnectPeers,
		activenetparams.DefaultPort)
	// --onionproxy and not --onion are contradictory (TODO: this is kinda stupid hm? switch *and* toggle by presence of flag value, one should be enough)
	if !*config.Onion && *config.OnionProxy != "" {
		err := fmt.Errorf("%s: the --onionproxy and --onion options may not be activated at the same time", "runNode")
		log <- cl.Error{err}
		return 1
	}
	// Check the checkpoints for syntax errors.
	log <- cl.Debug{"checking the checkpoints"}
	stateconfig.AddedCheckpoints, err =
		node.ParseCheckpoints(*config.AddCheckpoints)
	if err != nil {
		str := "%s: Error parsing checkpoints: %v"
		err := fmt.Errorf(str, "runNode", err)
		log <- cl.Err(err.Error())
		return 1
	}
	// Tor stream isolation requires either proxy or onion proxy to be set.
	if *config.TorIsolation &&
		*config.Proxy == "" &&
		*config.OnionProxy == "" {
		str := "%s: Tor stream isolation requires either proxy or onionproxy to be set"
		err := fmt.Errorf(str, "runNode")
		log <- cl.Error{err}
		return 1
	}
	// Setup dial and DNS resolution (lookup) functions depending on the specified options.  The default is to use the standard net.DialTimeout function as well as the system DNS resolver.  When a proxy is specified, the dial function is set to the proxy specific dial function and the lookup is set to use tor (unless --noonion is specified in which case the system DNS resolver is used).
	log <- cl.Debug{"setting network dialer and lookup"}
	stateconfig.Dial = net.DialTimeout
	stateconfig.Lookup = net.LookupIP
	if *config.Proxy != "" {
		log <- cl.Debug{"we are loading a proxy!"}
		_, _, err := net.SplitHostPort(*config.Proxy)
		if err != nil {
			str := "%s: Proxy address '%s' is invalid: %v"
			err := fmt.Errorf(str, "runNode", *config.Proxy, err)
			log <- cl.Error{err}
			return 1
		}
		// Tor isolation flag means proxy credentials will be overridden unless there is also an onion proxy configured in which case that one will be overridden.
		torIsolation := false
		if *config.TorIsolation &&
			*config.OnionProxy == "" &&
			(*config.ProxyUser != "" ||
				*config.ProxyPass != "") {
			torIsolation = true
			log <- cl.Warn{
				"Tor isolation set -- overriding specified proxy user credentials"}
		}
		proxy := &socks.Proxy{
			Addr:         *config.Proxy,
			Username:     *config.ProxyUser,
			Password:     *config.ProxyPass,
			TorIsolation: torIsolation,
		}
		stateconfig.Dial = proxy.DialTimeout
		// Treat the proxy as tor and perform DNS resolution through it unless the --noonion flag is set or there is an onion-specific proxy configured.
		if *config.Onion &&
			*config.OnionProxy == "" {
			stateconfig.Lookup = func(host string) ([]net.IP, error) {
				return connmgr.TorLookupIP(host, *config.Proxy)
			}
		}
	}
	// Setup onion address dial function depending on the specified options. The default is to use the same dial function selected above.  However, when an onion-specific proxy is specified, the onion address dial function is set to use the onion-specific proxy while leaving the normal dial function as selected above.  This allows .onion address traffic to be routed through a different proxy than normal traffic.
	log <- cl.Debug{"setting up tor proxy if enabled"}
	if *config.OnionProxy != "" {
		_, _, err := net.SplitHostPort(*config.OnionProxy)
		if err != nil {
			str := "%s: Onion proxy address '%s' is invalid: %v"
			err := fmt.Errorf(str, "runNode", *config.OnionProxy, err)
			log <- cl.Error{err}
			return 1
		}
		// Tor isolation flag means onion proxy credentials will be overriddenode.
		if *config.TorIsolation &&
			(*config.OnionProxyUser != "" || *config.OnionProxyPass != "") {
			log <- cl.Warn{
				"Tor isolation set - overriding specified onionproxy user credentials "}
		}
		stateconfig.Oniondial =
			func(network, addr string, timeout time.Duration) (net.Conn, error) {
				proxy := &socks.Proxy{
					Addr:         *config.OnionProxy,
					Username:     *config.OnionProxyUser,
					Password:     *config.OnionProxyPass,
					TorIsolation: *config.TorIsolation,
				}
				return proxy.DialTimeout(network, addr, timeout)
			}
		// When configured in bridge mode (both --onion and --proxy are configured), it means that the proxy configured by --proxy is not a tor proxy, so override the DNS resolution to use the onion-specific proxy.
		if *config.Proxy != "" {
			stateconfig.Lookup = func(host string) ([]net.IP, error) {
				return connmgr.TorLookupIP(host, *config.OnionProxy)
			}
		}
	} else {
		stateconfig.Oniondial = stateconfig.Dial
	}
	// Specifying --noonion means the onion address dial function results in an error.
	if !*config.Onion {
		stateconfig.Oniondial = func(a, b string, t time.Duration) (net.Conn, error) {
			return nil, errors.New("tor has been disabled")
		}
	}
	// run the node!
	return 0
}
