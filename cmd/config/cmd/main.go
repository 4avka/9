package main

import (
	"encoding/json"
	"fmt"

	. "git.parallelcoin.io/dev/9/cmd/config"
	"git.parallelcoin.io/dev/9/cmd/node"
	"git.parallelcoin.io/dev/9/cmd/node/mempool"
)

func main() {
	a := NewApp("9",
		Version("v1.9.9"),
		Group("app",
			File("cpuprofile",
				Usage("write cpu profile to this file, empty disables cpu profiling"),
			),
			Dir("datadir",
				Default("~/.9"),
				Usage("base folder to keep data for an instance of 9"),
			),
			Dir("appdatadir",
				Usage("subcommand data directory, sets to datadir/appname if unset"),
			),
			Dir("logdir",
				Usage("where logs are written, defaults to the appdatadir if unset"),
			),
			Port("profile",
				Usage("http profiling on specified port"),
			),
			Enable("upnp",
				Usage("enable port forwarding via UPNP"),
			),
		), Group("block",
			Int("maxsize",
				Default(node.DefaultBlockMaxSize),
				Min(node.BlockWeightMin),
				Max(node.BlockSizeMax),
				Usage("max block size in bytes"),
			),
			Int("maxweight",
				Default(node.DefaultBlockMaxWeight),
				Min(node.DefaultBlockMinWeight),
				Max(node.BlockWeightMax),
				Usage("max block weight"),
			),
			Int("minsize",
				Default(node.DefaultBlockMinSize),
				Min(node.DefaultBlockMinSize),
				Max(node.BlockSizeMax),
				Usage("min block size"),
			),
			Int("minweight",
				Default(node.DefaultBlockMinWeight),
				Min(node.DefaultBlockMinWeight),
				Max(node.BlockWeightMax),
				Usage("min block weight"),
			),
			Int("prioritysize",
				Default(mempool.DefaultBlockPrioritySize),
				Min(1000),
				Max(node.BlockSizeMax),
				Usage("the default size for high priority low fee transactions"),
			),
		), Group("chain",
			Tags("addcheckpoints",
				Usage("add checkpoints [height:hash ]*"),
			),
			Enable("disablecheckpoints",
				Usage("disables checkpoints (danger!)"),
			),
			Tag("dbtype",
				Default("ffldb"),
				Usage("set database backend to use for chain"),
			),
			Enabled("addrindex",
				Usage("enable address index (disables also transaction index)"),
			),
			Enabled("txindex",
				Usage("enable transaction index"),
			),
			Enable("rejectnonstd",
				Usage("reject nonstandard transactions even if net parameters allow it"),
			),
			Enable("relaynonstd",
				Usage("relay nonstandard transactions even if net parameters disallow it"),
			),
			Addr("rpc", 11048,
				Default("127.0.0.1:11048"),
				Usage("address of chain rpc to connect to"),
			),
			Int("sigcachemaxsize",
				Default(node.DefaultSigCacheMaxSize),
				Min(1000),
				Max(10000000),
				Usage("max number of signatures to keep in memory"),
			),
		), Group("limit",
			Tag("pass",
				RandomString(32),
				Usage("password for limited user"),
			),
			Tag("user",
				Default("limit"),
				Usage("username with limited privileges"),
			),
		), Group("log",
			Level(
				Default("info"),
				Usage("sets the base default log level"),
			),
			Tags("subsystem",
				Usage("[subsystem:loglevel ]+"),
			),
			Enable("nowrite",
				Usage("disable writing to log file"),
			),
		), Group("mining",
			Tags("addresses",
				Usage("set mining addresses, space separated"),
			),
			Algo("algo",
				Default("random"),
				Usage("select from available mining algorithms"),
			),
			Float("bias",
				Default(-0.5),
				Usage("bias for difficulties -1 = always easy, 1 always hardest"),
			),
			Enable("generate",
				Usage("enable builtin CPU miner"),
			),
			Int("genthreads",
				Default(node.DefaultGenThreads),
				Min(-1),
				Max(4096),
				Usage("set number of threads, -1 = all"),
			),
			Addr("listener", 11045,
				Usage("set listener address for mining dispatcher"),
			),
			Tag("pass",
				RandomString(32),
				Usage("password to secure mining dispatch connections"),
			),
			Duration("switch",
				Default("1s"),
				Usage("maximum time to mine per round"),
			),
		), Group("p2p",
			Addrs("addpeer", 11047,
				Usage("add permanent p2p peer"),
			),
			Int("banthreshold",
				Default(node.DefaultBanThreshold),
				Usage("how many ban units triggers a ban"),
			),
			Duration("banduration",
				Default(node.DefaultBanDuration),
				Usage("how long a ban lasts"),
			),
			Enable("disableban",
				Usage("disables banning peers"),
			),
			Enable("blocksonly",
				Usage("relay only blocks"),
			),
			Addrs("connect", 11047,
				Usage("connect only to these outbound peers"),
			),
			Enable("nolisten",
				Usage("disable p2p listener"),
			),
			Addrs("externalips", 11047,
				Usage("additional external IP addresses to bind to"),
			),
			Float("freetxrelaylimit",
				Default(15.0),
				Usage("limit of 'free' relay in thousand bytes per minute"),
			),
			Addrs("listen", 11047,
				Default("127.0.0.1:11047"),
				Usage("address to listen on for p2p connections"),
			),
			Int("maxorphantxs",
				Default(node.DefaultMaxOrphanTransactions),
				Min(0),
				Max(10000),
				Usage("maximum number of orphan transactions to keep in memory"),
			),
			Int("maxpeers",
				Default(node.DefaultMaxPeers),
				Min(2),
				Max(1024),
				Usage("maximum number of peers to connect to"),
			),
			Float("minrelaytxfee",
				Default(0.0001),
				Usage("minimum relay tx fee, baseline considered to be zero for relay"),
			),
			Net("network",
				Default("mainnet"),
				Usage("network to connect to"),
			),
			Enable("nobanning",
				Usage("disable banning of peers"),
			),
			Enable("nobloomfilters",
				Usage("disable bloom filters"),
			),
			Enable("nocfilters",
				Usage("disable cfilters"),
			),
			Enable("nodns",
				Usage("disable DNS seeding"),
			),
			Enable("norelaypriority",
				Usage("disables prioritisation of relayed transactions"),
			),
			Duration("trickleinterval",
				Default("27s"),
				Usage("minimum time between attempts to send new inventory to a connected peer"),
			),
			Tags("useragentcomments",
				Usage("comment to add to version identifier for node"),
			),
			Addrs("whitelist", 11047,
				Usage("peers who are never banned"),
			),
		),
		Group("proxy",
			Addr("address", 9050,
				Usage("address of socks proxy"),
			),
			Enable("isolation",
				Usage("enable randomisation of tor login to separate streams"),
			),
			Tag("pass",
				RandomString(32),
				Usage("password for proxy"),
			),
			Enable("tor",
				Usage("proxy is a tor proxy"),
			),
			Tag("user",
				Default("user"),
				Usage("username for proxy"),
			),
		),
		Group("rpc",
			Addr("connect", 11048,
				Default("127.0.0.1:11048"),
				Usage("connect to this node RPC endpoint"),
			),
			Enable("disable",
				Usage("disable rpc server"),
			),
			Addrs("listen", 11048,
				Default("127.0.0.1:11048"),
				Usage("address to listen for node rpc clients"),
			),
			Int("maxclients",
				Default(node.DefaultMaxRPCClients),
				Min(2),
				Max(1024),
				Usage("max clients for rpc"),
			),
			Int("maxconcurrentreqs",
				Default(node.DefaultMaxRPCConcurrentReqs),
				Min(2),
				Max(1024),
				Usage("maximum concurrent requests to handle"),
			),
			Int("maxwebsockets",
				Default(node.DefaultMaxRPCWebsockets),
				Max(1024),
				Usage("maximum websockets clients"),
			),
			Tag("pass",
				RandomString(32),
				Usage("password for rpc services"),
			),
			Enable("quirks",
				Usage("enable json rpc quirks matching bitcoin core"),
			),
			Tag("user",
				Default("user"),
				Usage("username for rpc services"),
			),
		),
		Group("tls",
			File("key",
				Default("tls.key"),
				Usage("file containing tls key"),
			),
			File("cert",
				Default("tls.cert"),
				Usage("file containing tls certificate"),
			),
			File("cafile",
				Default("tls.cafile"),
				Usage("set the certificate authority file to use for verifying rpc connections"),
			),
			Enable("disable",
				Usage("disable SSL on RPC connections"),
			),
			Enable("onetime",
				Usage("creates a key pair but does not write the secret for future runs"),
			),
			Enabled("server",
				Usage("enable tls for RPC servers"),
			),
			Enable("skipverify",
				Usage("skip verifying tls certificates with CAFile"),
			),
		),
		Group("wallet",
			Addr("server", 11046,
				Default("127.0.0.1:11046"),
				Usage("address of wallet rpc to connect to"),
			),
			Enable("noinitialload",
				Usage("disable automatic opening of the wallet at startup"),
			),
			Tag("pass",
				RandomString(32),
				Usage("password for the non-own transaction data in the wallet"),
			),
			Enable("enable",
				Usage("use configured wallet rpc instead of full node"),
			),
		),
	)
	// cfg := MakeConfig(a)
	j, e := json.MarshalIndent(a, "", "\t")
	if e != nil {
		panic(e)
	}
	fmt.Println(string(j))

	_ = a
}