package cmd

import (
	"strings"

	"github.com/davecgh/go-spew/spew"
)

// Declaration syntax used for defining a set of lines makes for a
// much more readable specification, which also doubles as help
// information text itself.
//  - no tab and one string defines a path prefix opening
//  - items have first line with name, type and a : (for humans),
//    and several parameters that is defined by the type
//	- integers can have a second parameter either the minimum or a
//    range specification as 1<100 - one or other number can be omitted to
//    omit the test
// whitespace only matters for the beginnings of lines, errors in omissions
// of items will cause a panic as this should be corrected at program-time
// variable specifications always start with one tab and comments after it
// must have two, and the parser will panic from any of these.
//
// strings denoting a mutual exclusive list of items have their own type.
//  - they may not contain spaces as these are used to separate them
//  - first item is default, each following is one of the options
//  - there can be only one, and the flag will not register if the data
//    does not match it
//
// boolean enabler/disablers just have one line as help text would most
// likely always start with 'enable' and 'disable' it might as well be
// the pattern filter.
//
// in string (and netaddr and variants) if the comment starts with 'enable'
// this means the default is empty and putting something in it enables
// the associated toggle, and is not stored in configuration
//
//  - strings with the comment starting with 'password' are interpreted
// to be passwords and generate a random password if no value is provided
// for convenience.

// getGroups splits an array of line specifications into its sections
// It will skip past all lines that are not just one word at the beginning
// which might be useful to using this as part of a generator
func getGroups(lines []string) (out [][]string) {
	cursor := -1
	for _, x := range lines {
		if len(x) < 1 {
			continue
		}
		if !strings.Contains(x, " ") || !strings.Contains(x, "\t") {
			cursor++
			out = append(out, []string{})
		}
		out[cursor] = append(out[cursor], x)
	}
	return
}

func GenerateLines(input string) (out *Lines) {
	o := make(Lines)
	out = &o
	splitted := strings.Split(input, "\n")
	groups := getGroups(splitted)
	spew.Dump(groups)
	return
}

var lines = `
app
	cpuprofile path: cpu.prof
		enable cpuprofiling to specified file
	datadir path: ~
		base directory containing configuration and datadir
	profile int: 1100 1025<65534
		enable port to listen on for memory profiling
	upnp 
		enable port forwarding via UPnP
block
	maxsize int: 900900 250000<10000000
		max block size
	maxweight int: 3006000 10000<10000000
		max block weight
	minsize int: 240 240<
		min block size
	minweight int: 2000 2000<100000
		min block weight
	prioritysize int: 50000 1000<200000
		the default size for high priority low fee transactions
chain
	addcheckpoints stringslice:
		add checkpoints "height:hash " space separated
	dbtype mutex: ffldb
		set database backend to use for chain
	notxindex
		disable transaction index
	noaddrindex 
		disable	address index (disables also transaction index)
	rejectnonstandard 
		enable rejection of nonstandard transactions even if net parameters allow them
	relaynonstd 
		enable relaying nonstandard transactions even if net parameters don't allow it
	rpc netaddr: 127.0.0.1:11048
		enable connecting to a node rpc server
	sigcachemaxsize int: 100000 1000-10000000
		max number of signatures to keep in memory
limit
	user string: limit
		username with limited privileges
	pass string: pa55word
		password for limited user
log
	level mutex: info warning error critical debug trace
		sets the base default log level
	subsystem stringslice:
		list of comma separated height:hash values
mining
	addresses stringslice:
		set mining addresses, space separated
	algo mutex: random blake2b blake14lr blake2s keccak scrypt sha256d skein stribog x11
		select from available mining algorithms
	bias float: -0.5 -1<1
		bias for difficulties on random mode -1 = always easy, 1 always hardest
	generate 
		enable builtin CPU miner
	genthreads int: -1 -1<1024
		set number of threads, -1 = all
	listener netaddr: 127.0.0.1:11049
		enable mining dispatcher
	pass string:
		password to secure mining dispatch connections 
	switchtime duration: 1s
		minimum time to mine one algorithm per round before switching in random mode
p2p
	addpeer netaddrs: 11047
		add permanent p2p peer (default is port if missing from address)
	banduration duration: 24h
		how long a ban lasts
	blocksonly
		enable relay of only blocks - don't relay transactions
	connect netaddrs: 11047
		enable connecting outbound only to the peers specified
	externalips netaddrs: 11047
		enable binding to additional external IP addresses for virtual interfaces
	freetxrelaylimit float: 15.0
		limit of 'free' relay in thousand bytes per minute
	listen netaddrs: 11047
		enable listening for connections from the p2p network
	maxorphanstxs int: 100 10<10000
		maximum number of orphan transactions to keep
	maxpeers int: 125 2<1024
		maximum number of peers to connect to
	minrelaytxfee float 0.0001000
		minimum relay tx fee, baseline threshold considered to be zero for relay
	network mutex: mainnet testnet simnet regtest
		network to connect to
	nobanning
		disable banning of peers
	nobloomfilters
		disable bloom filters
	nocfilters
		disable cfilters
	nodns
		disable DNS seeding
	norelaypriority
		disable prioritisation of relayed transactions
	trickleinterval duration: 27s 1s<
		minimum time between attempts to send new inventory to a connected peer
	useragentcomments string:
		comment to add to version identifier for node
	whitelist netaddrs: 11047
		peers who are never banned
proxy
	address netaddr: 127.0.0.1:9050
		enable socks proxy at address
	isolation
		enable randomisation of tor logins to separate streams
	user string: user
		username for proxy
	pass string: 
		password for proxy
	tor
		enable extended functionality provided by a tor proxy
rpc
	listen netaddrs: 127.0.0.1:11048
		enable rpc server (can enable multiple listeners for different interfaces)
	maxclients int:8 2<1024
		max clients for rpc
	maxconcurrentreqs int: 128 2<1024
		maximum concurrent requests to handle
	maxwebsockets int: 8 2<1024
		maximum websockets clients
	pass string: 
		password for rpc services",
	quirks
		enable json rpc quirks matching bitcoin
	user string: user
		username for rpc services
	wallet netaddr: 11046
		address of wallet rpc to connect to
tls
	cafile string: cafile
		set the certificate authority file to use for verifying rpc connections
	disable
		disable SSL on RPC connections
	oneshot
		creates a key pair but does not write the secret for future runs
	server
		enable tls for RPC servers
	skipverify
		enable the skipping of verifying tls certificates with CAFile
wallet
	noinitialload
		disable automatic opening of the wallet at startup
	pass string:
		password for the non-own transaction data in the wallet
	wallet
		enable the use of configured wallet rpc instead of full node
`
