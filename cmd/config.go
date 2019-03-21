package cmd

// Config is the declaration of our set of application configuration variables.
// Custom functions are written per type that generate a Line struct and contain
// a validator/setter function that checks the input
var Config = Lines{
	"datadir":               Path("~/.", "base directory containing configuration and data"),
	"log.level":             LogLevel("info", "sets the base default log level"),
	"log.subsystem":         SubSystem("", "[subsystem:loglevel ]+"),
	"network":               Network("mainnet", "network to connect to"),
	"p2p.addpeer":           NetAddrs("", "add permanent p2p peer"),
	"p2p.connect":           NetAddrs("", "connect only to these outbound peers"),
	"p2p.listen":            NetAddrs("127.0.0.1:11047", "address to listen on for p2p connections"),
	"ban.disable":           Disable("disable banning of peers"),
	"ban.duration":          Duration("24h", "how long a ban lasts"),
	"whitelist":             NetAddrs("", "peers who are never banned"),
	"rpc.chain":             NetAddr("127.0.0.1:11048", "address of chain rpc to connect to"),
	"rpc.wallet":            NetAddr("127.0.0.1:11046", "address of wallet rpc to connect to"),
	"rpc.listen":            NetAddrs("127.0.0.11048", "address to listen for node rpc clients"),
	"rpc.user":              String("user", "username for rpc services"),
	"rpc.pass":              String("pa55word", "password for rpc services"),
	"limit.user":            String("limit", "username with limited privileges"),
	"limit.pass":            String("pa55word", "password for limited user"),
	"maxpeers":              IntBounded("125", "maximum number of peers to connect to", 2, 1024),
	"rpc.maxclients":        IntBounded("8", "max clients for rpc", 2, 1024),
	"rpc.maxwebsockets":     IntBounded("8", "maximum websockets clients", 2, 1024),
	"rpc.maxconcurrentreqs": IntBounded("128", "maximum concurrent requests to handle", 2, 1024),
	"rpc.quirks":            Enable("enable json rpc quirks matching bitcoin"),
	"rpc.disable":           Disable("disable rpc server"),
	"notls":                 Disable("disable SSL on RPC server"),
}

var Subcommands = Commands{}
