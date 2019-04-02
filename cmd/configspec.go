package cmd

const defaultDatadir = "~/." + APPNAME

// Config is the declaration of our set of application configuration variables.
// Custom functions are written per type that generate a Line struct and contain
// a validator/setter function that checks the input
func getConfig() *Lines {
	return &Lines{
		"app.cpuprofile": Path(
			"", "write cpu profile"),
		"app.appdatadir": Path(
			"", "subcommand data dir, automatically set by subcommand name"),
		"app.datadir": Path(
			CleanAndExpandPath(defaultDatadir), "base directory containing configuration and data"),
		"app.logdir": Path(
			"", "where logs are written"),
		"app.profileenable": Enable(
			"enable http profiling"),
		"app.profile": IntBounded(
			1100, "http profiling on specified port", 1025, 65534),
		"app.upnp": Enable(
			"enable port forwarding via UPNP"),
		"block.maxsize": IntBounded(
			900900, "max block size", 250000, 10000000),
		"block.maxweight": IntBounded(
			3006000, "max block weight", 10000, 10000000),
		"block.minsize": IntBounded(
			240, "min block size", 240, 2<<32),
		"block.minweight": IntBounded(
			2000, "min block weight", 2000, 100000),
		"block.prioritysize": IntBounded(
			50000,
			"the default size for high priority low fee transactions",
			1000, 200000),
		"chain.addcheckpoints": StringSlice(
			"", "add checkpoints [height:hash ]*"),
		"chain.disablecheckpoints": Enable(
			"disables checkpoints (danger!)"),
		"chain.dbtype": String(
			"ffldb", "set database backend to use for chain"),
		"chain.addrindex": Disable(
			"enable address index (disables also transaction index)"),
		"chain.txindex": Disable(
			"enable transaction index"),
		"chain.rejectnonstd": Enable(
			"reject nonstandard transactions even if net parameters allow it"),
		"chain.relaynonstd": Enable(
			"relay nonstandard transactions even if net parameters disallow it"),
		"chain.rpc": NetAddr(
			"127.0.0.1:11048", "address of chain rpc to connect to"),
		"chain.sigcachemaxsize": IntBounded(
			100000, "max number of signatures to keep in memory", 1000, 10000000),
		"limit.pass": String(
			"pa55word", "password for limited user"),
		"limit.user": String(
			"limit", "username with limited privileges"),
		"log.level": LogLevel(
			"info", "sets the base default log level"),
		"log.subsystem": SubSystem(
			"", "[subsystem:loglevel ]+"),
		"log.nowrite": Enable(
			"disable writing to log file"),
		"mining.addresses": StringSlice(
			"", "set mining addresses, space separated"),
		"mining.algo": Algos(
			"random", "select from available mining algorithms"),
		"mining.bias": Float(
			"-0.5", "bias for difficulties -1 = always easy, 1 always hardest"),
		"mining.generate": Enable(
			"enable builtin CPU miner"),
		"mining.genthreads": IntBounded(
			-1, "set number of threads, -1 = all", -1, 4096),
		"mining.listener": NetAddr(
			"127.0.0.1:11049", "set listener address for mining dispatcher"),
		"mining.pass": String(
			// TODO: generate random pass
			"", "password to secure mining dispatch connections"),
		"mining.switch": Duration(
			"1s", "maximum time to mine per round"),
		"p2p.addpeer": NetAddrs(
			"11047", "add permanent p2p peer"),
		"p2p.banthreshold": Int(
			100, "how many ban units triggers a ban"),
		"p2p.banduration": Duration(
			"24h", "how long a ban lasts"),
		"p2p.disableban": Enable(
			"disables banning peers"),
		"p2p.blocksonly": Enable(
			"relay only blocks"),
		"p2p.connect": NetAddrs(
			"11047", "connect only to these outbound peers"),
		"p2p.nolisten": Enable(
			"disable p2p listener"),
		"p2p.externalips": NetAddrs(
			"11047", "additional external IP addresses to bind to"),
		"p2p.freetxrelaylimit": Float(
			"15.0", "limit of 'free' relay in thousand bytes per minute"),
		"p2p.listen": NetAddrs(
			"127.0.0.1:11047", "address to listen on for p2p connections"),
		"p2p.maxorphantxs": IntBounded(
			100, "maximum number of orphan transactions to keep", 10, 10000),
		"p2p.maxpeers": IntBounded(
			125, "maximum number of peers to connect to", 2, 1024),
		"p2p.minrelaytxfee": Float(
			"0.0001000",
			"minimum relay tx fee, baseline considered to be zero for relay"),
		"p2p.network": Network(
			"mainnet", "network to connect to"),
		"p2p.nobanning": Disable(
			"disable banning of peers"),
		"p2p.nobloomfilters": Enable(
			"disable bloom filters"),
		"p2p.nocfilters": Enable(
			"disable cfilters"),
		"p2p.nodns": Enable(
			"disable DNS seeding"),
		"p2p.norelaypriority": Enable(
			"disables prioritisation of relayed transactions"),
		"p2p.trickleinterval": Duration(
			"27s",
			"minimum time between attempts to send new inventory to a connected peer"),
		"p2p.useragentcomments": StringSlice(
			"", "comment to add to version identifier for node"),
		"p2p.whitelist": NetAddrs(
			"11047", "peers who are never banned"),
		"proxy.address": NetAddr(
			"127.0.0.1:9050", "address of socks proxy"),
		"proxy.enable": Enable(
			"enable socks proxy"),
		"proxy.isolation": Enable(
			"enable randomisation of tor login to separate streams"),
		"proxy.pass": String(
			"pa55word", "password for proxy"),
		"proxy.tor": Enable(
			"proxy is a tor proxy"),
		"proxy.user": String(
			"user", "username for proxy"),
		"rpc.connect": NetAddr(
			"127.0.0.1:11048", "connect to this node RPC endpoint"),
		"rpc.disable": Enable(
			"disable rpc server"),
		"rpc.listen": NetAddrs(
			"127.0.0.1:11048", "address to listen for node rpc clients"),
		"rpc.maxclients": IntBounded(
			8, "max clients for rpc", 2, 1024),
		"rpc.maxconcurrentreqs": IntBounded(
			128, "maximum concurrent requests to handle", 2, 1024),
		"rpc.maxwebsockets": IntBounded(
			8, "maximum websockets clients", 2, 1024),
		"rpc.pass": String(
			"pa55word", "password for rpc services"),
		"rpc.quirks": Enable(
			"enable json rpc quirks matching bitcoin core"),
		"rpc.user": String(
			"user", "username for rpc services"),
		"rpc.wallet": NetAddr(
			"127.0.0.1:11046", "address of wallet rpc to connect to"),
		"tls.key": Path(
			"tls.key", "file containing tls key"),
		"tls.cert": Path(
			"tls.cert", "file containing tls certificate"),
		"tls.cafile": Path(
			"tls.cafile",
			"set the certificate authority file to use for verifying rpc connections"),
		"tls.disable": Disable(
			"disable SSL on RPC connections"),
		"tls.onetime": Enable(
			"creates a key pair but does not write the secret for future runs"),
		"tls.server": Enable(
			"enable tls for RPC servers"),
		"tls.skipverify": Enable(
			"skip verifying tls certificates with CAFile"),
		"wallet.noinitialload": Enable(
			"disable automatic opening of the wallet at startup"),
		"wallet.pass": String(
			"", "password for the non-own transaction data in the wallet"),
		"wallet": Enable(
			"use configured wallet rpc instead of full node"),
	}
}
