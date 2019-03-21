package cmd

// Config is the declaration of our set of application configuration variables.
// Custom functions are written per type that generate a Line struct and contain
// a validator/setter function that checks the input
var Config = Lines{
	"app.datadir":             Path("~", "base directory containing configuration and data"),
	"app.profile":             IntBounded("1100", "http profiling on specified port", 1025, 65534),
	"app.cpuprofile":          Path("cpu.prof", "write cpu profile"),
	"app.upnp":                Enable("enable port forwarding via UPNP"),
	"log.level":               LogLevel("info", "sets the base default log level"),
	"log.subsystem":           SubSystem("", "[subsystem:loglevel ]+"),
	"p2p.network":             Network("mainnet", "network to connect to"),
	"p2p.addpeer":             NetAddrs("11047", "add permanent p2p peer"),
	"p2p.connect":             NetAddrs("11047", "connect only to these outbound peers"),
	"p2p.listen":              NetAddrs("127.0.0.1:11047", "address to listen on for p2p connections"),
	"p2p.whitelist":           NetAddrs("11047", "peers who are never banned"),
	"p2p.maxpeers":            IntBounded("125", "maximum number of peers to connect to", 2, 1024),
	"p2p.nodns":               Disable("disable DNS seeding"),
	"p2p.banduration":         Duration("24h", "how long a ban lasts"),
	"p2p.nobanning":           Disable("disable banning of peers"),
	"p2p.externalips":         NetAddrs("11047", "additional external IP addresses to bind to"),
	"p2p.minrelaytxfee":       Float("0.0001000", "minimum relay tx fee, baseline considered to be zero for relay"),
	"p2p.freetxrelaylimit":    Float("15.0", "limit of 'free' relay in thousand bytes per minute"),
	"p2p.norelaypriority":     Enable("disables prioritisation of relayed transactions"),
	"p2p.trickleinterval":     Duration("27s", "minimum time between attempts to send new inventory to a connected peer"),
	"p2p.maxorphanstxs":       IntBounded("100", "maximum number of orphan transactions to keep", 10, 10000),
	"p2p.useragentcomments":   String("", "comment to add to version identifier for node"),
	"chain.rpc":               NetAddr("127.0.0.1:11048", "address of chain rpc to connect to"),
	"chain.addcheckpoints":    StringSlice("", "add checkpoints [height:hash ]*"),
	"chain.dbtype":            String("ffldb", "set database backend to use for chain"),
	"rpc.wallet":              NetAddr("127.0.0.1:11046", "address of wallet rpc to connect to"),
	"rpc.listen":              NetAddrs("127.0.0.11048", "address to listen for node rpc clients"),
	"rpc.user":                String("user", "username for rpc services"),
	"rpc.pass":                String("pa55word", "password for rpc services"),
	"rpc.maxclients":          IntBounded("8", "max clients for rpc", 2, 1024),
	"rpc.maxwebsockets":       IntBounded("8", "maximum websockets clients", 2, 1024),
	"rpc.maxconcurrentreqs":   IntBounded("128", "maximum concurrent requests to handle", 2, 1024),
	"rpc.quirks":              Enable("enable json rpc quirks matching bitcoin"),
	"rpc.disable":             Disable("disable rpc server"),
	"limit.user":              String("limit", "username with limited privileges"),
	"limit.pass":              String("pa55word", "password for limited user"),
	"tls.disable":             Disable("disable SSL on RPC connections"),
	"proxy.enable":            Enable("enable socks proxy"),
	"proxy.tor":               Enable("proxy is a tor proxy"),
	"proxy.user":              String("user", "username for proxy"),
	"proxy.pass":              String("pa55word", "username for proxy"),
	"proxy.address":           NetAddr("127.0.0.1:9050", "address of socks proxy"),
	"proxy.isolation":         Enable("enable randomisation of tor login to separate streams"),
	"mining.algo":             Algos("", "select from available mining algorithms"),
	"mining.generate":         Enable("enable builtin CPU miner"),
	"mining.genthreads":       IntBounded("-1", "set number of threads, -1 = all", -1, 4096),
	"mining.addresses":        StringSlice("", "set mining addresses, space separated"),
	"mining.listener":         NetAddr("127.0.0.1:11049", "set listener address for mining dispatcher"),
	"mining.pass":             String("", "password to secure mining dispatch connections"),
	"mining.bias":             Float("-0.5", "bias for difficulties -1 = always easy, 1 always hardest"),
	"mining.switch":           Duration("1s", "maximum time to mine per round"),
	"block.minsize":           IntBounded("240", "min block size", 240, 2<<32),
	"block.maxsize":           IntBounded("900900", "max block size", 250000, 10000000),
	"block.minweight":         IntBounded("2000", "min block weight", 2000, 100000),
	"block.maxweight":         IntBounded("3006000", "max block weight", 10000, 10000000),
	"block.prioritysize":      IntBounded("50000", "the default size for high priority low fee transactions", 1000, 200000),
	"p2p.nobloomfilters":      Disable("disable bloom filters"),
	"p2p.nocfilters":          Disable("disable cfilters"),
	"chain.sigcachemaxsize":   IntBounded("100000", "max number of signatures to keep in memory", 1000, 10000000),
	"p2p.blocksonly":          Enable("relay only blocks"),
	"chain.notxindex":         Disable("disable transaction index"),
	"chain.noaddrindex":       Disable("disable address index (disables also transaction index)"),
	"chain.relaynonstd":       Enable("relay nonstandard transactions even if net parameters don't allow it"),
	"chain.rejectnonstandard": Enable("reject nonstandard transactions even if net parameters allow it"),
	"wallet":                  Enable("use configured wallet rpc instead of full node"),
	"wallet.noinitialload":    Enable("disable automatic opening of the wallet at startup"),
	"wallet.pass":             String("", "password for the non-own transaction data in the wallet"),
	"tls.skipverify":          Enable("skip verifying tls certificates with CAFile"),
	"tls.cafile":              String("cafile", "set the certificate authority file to use for verifying rpc connections"),
	"tls.onetime":             Enable("creates a key pair but does not write the secret for future runs"),
	"tls.server":              Enable("enable tls for RPC servers"),
}

var Subcommands = Commands{
	"default": {
		"launch the GUI",
		Lines{
			"datadir": Path("~", "specify a data directory"),
		},
		func(args ...string) error {
			return nil
		},
	},
	"droptxindex": {
		"drop the transaction index",
		Lines{
			"datadir": Path("~", "specify a data directory"),
		},
		func(args ...string) error {
			return nil
		},
	},
	"dropaddrindex": {
		"drop the address index",
		Lines{
			"datadir": Path("~", "specify a data directory"),
		},
		func(args ...string) error {
			return nil
		},
	},
	"dropcfindex": {
		"drop the compact filters index",
		Lines{
			"datadir": Path("~", "specify a data directory"),
		},
		func(args ...string) error {
			return nil
		},
	},
	"node": {
		"run a full node",
		Lines{
			"datadir": Path("~", "specify a data directory"),
		},
		func(args ...string) error {
			return nil
		}},
	"wallet": {
		"run a wallet node",
		Lines{
			"datadir": Path("~", "specify a data directory"),
		},
		func(args ...string) error {
			return nil
		},
	},
	"shell": {
		"run a combined wallet/full node",
		Lines{
			"datadir": Path("~", "specify a data directory"),
		},
		func(args ...string) error {
			return nil
		},
	},
	"ctl": {
		"send rpc queries to a node",
		Lines{
			"datadir": Path("~", "specify a data directory"),
		},
		func(args ...string) error {
			return nil
		},
	},
	"cli": {
		"send rpc queries to a wallet",
		Lines{
			"datadir": Path("~", "specify a data directory"),
		},
		func(args ...string) error {
			return nil
		},
	},
	"mine": {
		"run the standalone miner",
		Lines{},
		func(args ...string) error {
			return nil
		},
	},
	"gen.certs": {
		"generate TLS key and certificate",
		Lines{
			"datadir": Path("~", "specify a data directory"),
		},
		func(args ...string) error {
			return nil
		},
	},
	"gen.cafile": {
		"generate a TLS Certificate Authority",
		Lines{
			"datadir": Path("~", "specify a data directory"),
		},
		func(args ...string) error {
			return nil
		},
	},
	"copy": {
		"copies a profile (many) new one(s)",
		Lines{
			"datadir":  Path("~", "specify a data directory"),
			"basename": String("test", "base name for testnet data directories"),
			"number":   IntBounded("1", "number of data directories to create", 1, 100),
		},
		func(args ...string) error {
			return nil
		},
	},
	"new": {
		"creates new testnet profile directories from defaults",
		Lines{
			"datadir":  Path("~", "specify a data directory"),
			"basename": String("test", "base name for testnet data directories"),
			"number":   IntBounded("1", "number of data directories to create", 1, 100),
		},
		func(args ...string) error {
			return nil
		},
	},
	"conf": {
		"run a visual CLI configuration editor",
		Lines{
			"datadir": Path("~", "specify a data directory"),
		},
		func(args ...string) error {
			return nil
		}},
}
