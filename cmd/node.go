package cmd

import (
	"fmt"
	"net"
	"path/filepath"

	"git.parallelcoin.io/dev/9/cmd/node"

	"git.parallelcoin.io/dev/9/pkg/util/cl"
	"github.com/davecgh/go-spew/spew"
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

	spew.Dump(*config)

	return 0
}
