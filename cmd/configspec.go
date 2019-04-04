package cmd

import (
	"fmt"

	"git.parallelcoin.io/dev/9/cmd/node"
	"git.parallelcoin.io/dev/9/cmd/node/mempool"
)

var defaultDatadir = CleanAndExpandPath("~/." + APPNAME)

// Config is the declaration of our set of application configuration variables.
// Custom functions are written per type that generate a Line struct and contain
// a validator/setter function that checks the input
func getConfig() *Lines {
	l := make(Lines)
	return l.Group("p2p",
		// This is first because it sets the port bases on the rest of this set
		Net("network").
			Default("mainnet").
			Usage("network to connect to"),
	).Group("app",
		File("cpuprofile").Usage("write cpu profile to this file, empty disables cpu profiling"),
		Dir("datadir").Default(defaultDatadir).Usage("base folder to keep data for an instance of 9"),
		Dir("appdatadir").Usage("subcommand data directory, sets to datadir/appname if unset"),
		Dir("logdir").Usage("where logs are written, defaults to the appdatadir if unset"),
		Port("profile").Default("1104").Usage("http profiling on specified port"),
		Enable("upnp").Usage("enable port forwarding via UPNP"),
	).Group("block",
		Int("maxsize").Default(node.DefaultBlockMaxSize).Min(node.BlockWeightMin).Max(node.BlockSizeMax).Usage("max block size in bytes"),
		Int("maxweight").Default(node.DefaultBlockMaxWeight).Min(node.DefaultBlockMinWeight).Max(node.BlockWeightMax).Usage("max block weight"),
		Int("minsize").Default(node.DefaultBlockMinSize).Min(node.DefaultBlockMinSize).Max(node.BlockSizeMax).Usage("min block size"),
		Int("minweight").Default(node.DefaultBlockMinWeight).Min(node.DefaultBlockMinWeight).Max(node.BlockWeightMax).Usage("min block weight"),
		Int("prioritysize").Default(mempool.DefaultBlockPrioritySize).Min(1000).Max(node.BlockSizeMax).Usage("the default size for high priority low fee transactions"),
	).Group("chain",
		Tags("addcheckpoints").Usage("add checkpoints [height:hash ]*"),
		Enable("disablecheckpoints").Usage("disables checkpoints (danger!)"),
		Tag("dbtype").Default("ffldb").Usage("set database backend to use for chain"),
		Enabled("addrindex").Usage("enable address index (disables also transaction index)"),
		Enabled("txindex").Usage("enable transaction index"),
		Enable("rejectnonstd").Usage("reject nonstandard transactions even if net parameters allow it"),
		Enable("relaynonstd").Usage("relay nonstandard transactions even if net parameters disallow it"),
		Addr("rpc").Default("127.0.0.1:11048").Usage("address of chain rpc to connect to"),
		Int("sigcachemaxsize").Default(node.DefaultSigCacheMaxSize).Min(1000).Max(10000000).Default("max number of signatures to keep in memory"),
	).Group("limit",
		Tag("pass").Default("pa55word1").Usage("password for limited user"),
		Tag("user").Default("limit").Usage("username with limited privileges"),
	).Group("log",
		Log("level").Default("info").Usage("sets the base default log level"),
		Tags("subsystem").Usage("[subsystem:loglevel ]+"),
		Enable("nowrite").Usage("disable writing to log file"),
	).Group("mining",
		Tags("addresses").Usage("set mining addresses, space separated"),
		Algo("algo").Default("random").Usage("select from available mining algorithms"),
		Float("bias").Default(-0.5).Usage("bias for difficulties -1 = always easy, 1 always hardest"),
		Enable("generate").Usage("enable builtin CPU miner"),
		Int("genthreads").Default(node.DefaultGenThreads).Min(-1).Max(4096).Usage("set number of threads, -1 = all"),
		Addr("listener").Usage("set listener address for mining dispatcher"),
		Tag("pass").Usage("password to secure mining dispatch connections"), // TODO: generate random pass
		Duration("switch").Default("1s").Usage("maximum time to mine per round"),
		Addrs("p2p.addpeer").Default("11047").Usage("add permanent p2p peer"),
	).Group("p2p",
		Int("banthreshold").Default(node.DefaultBanThreshold).Usage("how many ban units triggers a ban"),
		Duration("banduration").Default(fmt.Sprint(node.DefaultBanDuration)).Usage("how long a ban lasts"),
		Enable("disableban").Usage("disables banning peers"),
		Enable("blocksonly").Usage("relay only blocks"),
		Addrs("connect").Default("11047").Usage("connect only to these outbound peers"),
		Enable("nolisten").Usage("disable p2p listener"),
		Addrs("externalips").Default("11047").Usage("additional external IP addresses to bind to"),
		Float("freetxrelaylimit").Default(15.0).Usage("limit of 'free' relay in thousand bytes per minute"),
		Addrs("listen").Default("127.0.0.1:11047").Usage("address to listen on for p2p connections"),
		Int("maxorphantxs").Default(node.DefaultMaxOrphanTransactions).Min(0).Max(10000).Usage("maximum number of orphan transactions to keep"),
		Int("maxpeers").Default(node.DefaultMaxPeers).Min(2).Max(1024).Usage("maximum number of peers to connect to"),
		Float("minrelaytxfee").Default(0.0001).Usage("minimum relay tx fee, baseline considered to be zero for relay"),
		Enable("nobanning").Usage("disable banning of peers"),
		Enable("nobloomfilters").Usage("disable bloom filters"),
		Enable("nocfilters").Usage("disable cfilters"),
		Enable("nodns").Usage("disable DNS seeding"),
		Enable("norelaypriority").Usage("disables prioritisation of relayed transactions"),
		Duration("trickleinterval").Default("27s").Usage("minimum time between attempts to send new inventory to a connected peer"),
		Tags("useragentcomments").Usage("comment to add to version identifier for node"),
		Addrs("whitelist").Default("11047").Usage("peers who are never banned"),
	).Group("proxy",
		Addr("proxy.address").Default("127.0.0.1:9050").Usage("address of socks proxy"),
		Enable("isolation").Usage("enable randomisation of tor login to separate streams"),
		Tag("pass").Default("pa55word").Usage("password for proxy"),
		Enable("tor").Usage("proxy is a tor proxy"),
		Tag("user").Default("user").Usage("username for proxy"),
	).Group("rpc",
		Addr("rpc.connect").Default("127.0.0.1:11048").Usage("connect to this node RPC endpoint"),
		Enable("rpc.disable").Usage("disable rpc server"),
		Addrs("rpc.listen").Default("127.0.0.1:11048").Usage("address to listen for node rpc clients"),
		Int("rpc.maxclients").Default(node.DefaultMaxRPCClients).Min(2).Max(1024).Usage("max clients for rpc"),
		Int("rpc.maxconcurrentreqs").Default(node.DefaultMaxRPCConcurrentReqs).Min(2).Max(1024).Usage("maximum concurrent requests to handle"),
		Int("rpc.maxwebsockets").Default(node.DefaultMaxRPCWebsockets).Max(1024).Usage("maximum websockets clients"),
		Tag("rpc.pass").Default("pa55word").Usage("password for rpc services"),
		Enable("rpc.quirks").Usage("enable json rpc quirks matching bitcoin core"),
		Tag("rpc.user").Default("user").Usage("username for rpc services"),
		Addr("rpc.wallet").Default("127.0.0.1:11046").Usage("address of wallet rpc to connect to"),
	).Group("tls",
		File("tls.key").Default("tls.key").Usage("file containing tls key"),
		File("tls.cert").Default("tls.cert").Usage("file containing tls certificate"),
		File("tls.cafile").Default("tls.cafile").Usage("set the certificate authority file to use for verifying rpc connections"),
		Enable("tls.disable").Usage("disable SSL on RPC connections"),
		Enable("tls.onetime").Usage("creates a key pair but does not write the secret for future runs"),
		Enabled("tls.server").Usage("enable tls for RPC servers"),
		Enable("tls.skipverify").Usage("skip verifying tls certificates with CAFile"),
	).Group("wallet",
		Enable("wallet.noinitialload").Usage("disable automatic opening of the wallet at startup"),
		Tag("wallet.pass").Usage("password for the non-own transaction data in the wallet"),
		Tag("wallet.enable").Usage("use configured wallet rpc instead of full node"),
	)
}
