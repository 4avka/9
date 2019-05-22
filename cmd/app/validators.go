package app
import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
	"git.parallelcoin.io/dev/9/pkg/chain"
	"git.parallelcoin.io/dev/9/cmd/def"
	"git.parallelcoin.io/dev/9/cmd/nine"
	"git.parallelcoin.io/dev/9/cmd/node"
	"git.parallelcoin.io/dev/9/pkg/chain/fork"
	"git.parallelcoin.io/dev/9/pkg/ifc"
	"git.parallelcoin.io/dev/9/pkg/peer/connmgr"
	"git.parallelcoin.io/dev/9/pkg/util"
	"git.parallelcoin.io/dev/9/pkg/util/cl"
	"github.com/btcsuite/go-socks/socks"
)
// GenAddr returns a validator with a set default port assumed if one is not present
func GenAddr(name string, port int) func(r *def.Row, in interface{}) bool {
	return func(r *def.Row, in interface{}) bool {
		var s *string
		switch I := in.(type) {
		case string:
			s = &I
		case *string:
			s = I
		case nil:
			r.Value.Put(nil)
			return true
		default:
			return false
		}
		if s == nil {
			r.Value.Put(nil)
			return true
		}
		if *s == "" {
			s = nil
			r.Value.Put(nil)
			return true
		}
		h, p, err := net.SplitHostPort(*s)
		if err != nil {
			*s = net.JoinHostPort(*s, fmt.Sprint(port))
		} else {
			n, e := strconv.Atoi(p)
			if e == nil {
				if n < 1025 || n > 65535 {
					return false
				}
			} else {
				return false
				// p = ""
			}
			if p == "" {
				p = fmt.Sprint(port)
				*s = net.JoinHostPort(h, p)
			} else {
				*s = net.JoinHostPort(h, p)
			}
		}
		if r != nil {
			r.Value.Put(*s)
			r.String = *s
			r.App.SaveConfig()
		}
		return true
	}
}
// GenAddrs returns a validator with a set default port assumed if one is not present
func GenAddrs(name string, port int) func(r *def.Row, in interface{}) bool {
	return func(r *def.Row, in interface{}) bool {
		var s []string
		existing, ok := r.Value.Get().([]string)
		if !ok {
			existing = []string{}
		}
		switch I := in.(type) {
		case string:
			s = append(s, I)
		case *string:
			s = append(s, *I)
		case []string:
			s = I
		case *[]string:
			s = *I
		case []interface{}:
			for _, x := range I {
				so, ok := x.(string)
				if ok {
					s = append(s, so)
				}
			}
		case nil:
			return false
		default:
			fmt.Println(name, port, "invalid type", in, reflect.TypeOf(in))
			return false
		}
		for _, sse := range s {
			h, p, e := net.SplitHostPort(sse)
			if e != nil {
				sse = net.JoinHostPort(sse, fmt.Sprint(port))
			} else {
				n, e := strconv.Atoi(p)
				if e == nil {
					if n < 1025 || n > 65535 {
						fmt.Println(name, port, "port out of range")
						return false
					}
				} else {
					fmt.Println(name, port, "port not an integer")
					return false
				}
				if p == "" {
					p = fmt.Sprint(port)
				}
				sse = net.JoinHostPort(h, p)
			}
			existing = append(existing, sse)
		}
		if r != nil {
			// eliminate duplicates
			tmpMap := make(map[string]struct{})
			for _, x := range existing {
				tmpMap[x] = struct{}{}
			}
			existing = []string{}
			for i := range tmpMap {
				existing = append(existing, i)
			}
			sort.Strings(existing)
			r.Value.Put(existing)
			r.String = fmt.Sprint(existing)
			r.App.SaveConfig()
		}
		return true
	}
}
func getAlgoOptions() (options []string) {
	var modernd = "random"
	for _, x := range fork.P9AlgoVers {
		options = append(options, x)
	}
	options = append(options, modernd)
	sort.Strings(options)
	return
}
// Valid is a collection of validator functions for the different types used
// in a configuration. These functions optionally can accept a *def.Row and with
// this they assign the validated, parsed value into the Value slot.
var Valid = struct {
	File, Dir, Port, Bool, Int, Tag, Tags, Algo, Float, Duration, Net,
	Level func(*def.Row, interface{}) bool
}{}
func init() {
	Valid.File = func(r *def.Row, in interface{}) bool {
		var s *string
		switch I := in.(type) {
		case string:
			s = &I
		case *string:
			s = I
		default:
			return false
		}
		if len(*s) > 0 {
			ss := util.CleanAndExpandPath(*s, *datadir)
			if r != nil {
				r.String = fmt.Sprint(ss)
				if r.Value == nil {
					r.Value = ifc.NewIface()
				}
				r.Value.Put(ss)
				r.App.SaveConfig()
				return true
			}
			return false
		}
		return false
	}
	Valid.Dir = func(r *def.Row, in interface{}) bool {
		var s *string
		switch I := in.(type) {
		case string:
			s = &I
		case *string:
			s = I
		default:
			return false
		}
		if len(*s) > 0 {
			ss := util.CleanAndExpandPath(*s, *datadir)
			if r != nil {
				r.String = fmt.Sprint(ss)
				if r.Value == nil {
					r.Value = ifc.NewIface()
				}
				r.Value.Put(ss)
				r.App.SaveConfig()
				return true
			}
			return false
		}
		return false
	}
	Valid.Port = func(r *def.Row, in interface{}) bool {
		var s string
		var ii int
		isString := false
		switch I := in.(type) {
		case string:
			s = I
			isString = true
		case *string:
			s = *I
			isString = true
		case int:
			ii = I
		case *int:
			ii = *I
		default:
			return false
		}
		if isString {
			n, e := strconv.Atoi(s)
			if e != nil {
				return false
			}
			ii = n
		}
		if ii < 1025 || ii > 65535 {
			return false
		}
		if r != nil {
			r.Value.Put(ii)
			r.String = fmt.Sprint(ii)
			r.App.SaveConfig()
		}
		return true
	}
	Valid.Bool = func(r *def.Row, in interface{}) bool {
		var sb string
		var b bool
		switch I := in.(type) {
		case string:
			sb = I
			if strings.ToUpper(sb) == "TRUE" {
				b = true
				goto boolout
			}
			if strings.ToUpper(sb) == "FALSE" {
				b = false
				goto boolout
			}
		case *string:
			sb = *I
			if strings.ToUpper(sb) == "TRUE" {
				b = true
				goto boolout
			}
			if strings.ToUpper(sb) == "FALSE" {
				b = false
				goto boolout
			}
		case bool:
			b = I
		case *bool:
			b = *I
		default:
			return false
		}
	boolout:
		if r != nil {
			r.String = fmt.Sprint(b)
			r.Value.Put(b)
			r.App.SaveConfig()
		}
		return true
	}
	Valid.Int = func(r *def.Row, in interface{}) bool {
		var s string
		var ii int
		isString := false
		switch I := in.(type) {
		case string:
			s = I
			isString = true
		case *string:
			s = *I
			isString = true
		case int:
			ii = I
		case *int:
			ii = *I
		default:
			return false
		}
		if isString {
			n, e := strconv.Atoi(s)
			if e != nil {
				return false
			}
			ii = n
		}
		if r != nil {
			r.String = fmt.Sprint(ii)
			//r.Value =
			r.Value.Put(ii)
			r.App.SaveConfig()
		}
		return true
	}
	Valid.Tag = func(r *def.Row, in interface{}) bool {
		var s string
		switch I := in.(type) {
		case string:
			s = I
		case *string:
			s = *I
		default:
			return false
		}
		s = strings.TrimSpace(s)
		if len(s) < 1 {
			return false
		}
		if r != nil {
			r.Value.Put(s)
			r.String = fmt.Sprint(s)
			r.App.SaveConfig()
		}
		return true
	}
	Valid.Tags = func(r *def.Row, in interface{}) bool {
		var s []string
		existing, ok := r.Value.Get().([]string)
		if !ok {
			existing = []string{}
		}
		switch I := in.(type) {
		case string:
			s = append(s, I)
		case *string:
			s = append(s, *I)
		case []string:
			s = I
		case *[]string:
			s = *I
		case []interface{}:
			for _, x := range I {
				so, ok := x.(string)
				if ok {
					s = append(s, so)
				}
			}
		case nil:
			return false
		default:
			fmt.Println("invalid type", in, reflect.TypeOf(in))
			return false
		}
		for _, sse := range s {
			existing = append(existing, sse)
		}
		if r != nil {
			// eliminate duplicates
			tmpMap := make(map[string]struct{})
			for _, x := range existing {
				tmpMap[x] = struct{}{}
			}
			existing = []string{}
			for i := range tmpMap {
				existing = append(existing, i)
			}
			sort.Strings(existing)
			r.Value.Put(existing)
			r.String = fmt.Sprint(existing)
			r.App.SaveConfig()
		}
		return true
	}
	Valid.Algo = func(r *def.Row, in interface{}) bool {
		var s string
		switch I := in.(type) {
		case string:
			s = I
		case *string:
			s = *I
		default:
			return false
		}
		var o string
		options := getAlgoOptions()
		for _, x := range options {
			if s == x {
				o = s
			}
		}
		if o == "" {
			rnd := "random"
			o = rnd
		}
		if r != nil {
			r.String = fmt.Sprint(o)
			r.Value.Put(o)
			r.App.SaveConfig()
		}
		return true
	}
	Valid.Float = func(r *def.Row, in interface{}) bool {
		var s string
		var f float64
		isString := false
		switch I := in.(type) {
		case string:
			s = I
			isString = true
		case *string:
			s = *I
			isString = true
		case float64:
			f = I
		case *float64:
			f = *I
		default:
			return false
		}
		if isString {
			ff, e := strconv.ParseFloat(s, 64)
			if e != nil {
				return false
			}
			f = ff
		}
		if r != nil {
			r.Value.Put(f)
			r.String = fmt.Sprint(f)
			r.App.SaveConfig()
		}
		return true
	}
	Valid.Duration = func(r *def.Row, in interface{}) bool {
		var s string
		var t time.Duration
		isString := false
		switch I := in.(type) {
		case string:
			s = I
			isString = true
		case *string:
			s = *I
			isString = true
		case time.Duration:
			t = I
		case *time.Duration:
			t = *I
		default:
			return false
		}
		if isString {
			dd, e := time.ParseDuration(s)
			if e != nil {
				return false
			}
			t = dd
		}
		if r != nil {
			r.String = fmt.Sprint(t)
			r.Value.Put(t)
			r.App.SaveConfig()
		}
		return true
	}
	Valid.Net = func(r *def.Row, in interface{}) bool {
		var sn string
		switch I := in.(type) {
		case string:
			sn = I
		case *string:
			sn = *I
		default:
			return false
		}
		found := false
		for _, x := range Networks {
			if x == sn {
				found = true
				*nine.ActiveNetParams = *NetParams[x]
			}
		}
		if r != nil && found {
			r.String = fmt.Sprint(sn)
			r.Value.Put(sn)
			r.App.SaveConfig()
		}
		return found
	}
	Valid.Level = func(r *def.Row, in interface{}) bool {
		var sl string
		switch I := in.(type) {
		case string:
			sl = I
		case *string:
			sl = *I
		default:
			return false
		}
		found := false
		for x := range cl.Levels {
			if x == sl {
				found = true
			}
		}
		if r != nil && found {
			r.String = fmt.Sprint(sl)
			r.Value.Put(sl)
			r.App.SaveConfig()
		}
		return found
	}
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
		*ap.Config.BlockMaxSize = chain.MaxBlockBaseSize - 1000
	// If the max block weight isn't set, but the block size is, then we'll scale the set weight accordingly based on the max block size value.
	case *ap.Config.BlockMaxSize != node.DefaultBlockMaxSize &&
		*ap.Config.BlockMaxWeight == node.DefaultBlockMaxWeight:
		*ap.Config.BlockMaxWeight = *ap.Config.BlockMaxSize *
			chain.WitnessScaleFactor
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
