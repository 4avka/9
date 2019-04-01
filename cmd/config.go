package cmd

import (
	"time"

	"git.parallelcoin.io/dev/9/cmd/nine"
)

var config = MakeConfig(&Config)

func MakeConfig(c *Lines) (out *nine.Config) {
	var tn, sn, rn = new(bool), new(bool), new(bool)
	cfg := *c
	String := func(path string) *string {
		return cfg[path].Value.(*string)
	}
	Tags := func(path string) *[]string {
		return cfg[path].Value.(*[]string)
	}
	Map := func(path string) *nine.Mapstringstring {
		return cfg[path].Value.(*nine.Mapstringstring)
	}
	Int := func(path string) *int {
		return cfg[path].Value.(*int)
	}
	Bool := func(path string) *bool {
		return cfg[path].Value.(*bool)
	}
	Float := func(path string) *float64 {
		return cfg[path].Value.(*float64)
	}
	Duration := func(path string) *time.Duration {
		return cfg[path].Value.(*time.Duration)
	}
	out = &nine.Config{
		ConfigFile:               nil,
		DataDir:                  String("app.datadir"),
		LogDir:                   String("app.datadir"),
		LogLevel:                 String("log.level"),
		Subsystems:               Map("log.subsystem"),
		Network:                  String("p2p.network"),
		AddPeers:                 Tags("p2p.addpeer"),
		ConnectPeers:             Tags("p2p.connect"),
		MaxPeers:                 Int("p2p.maxpeers"),
		Listeners:                Tags("p2p.listen"),
		DisableListen:            Bool("p2p.nolisten"),
		DisableBanning:           Bool("p2p.nobanning"),
		BanDuration:              Duration("p2p.banduration"),
		BanThreshold:             Int("p2p.banthreshold"),
		Whitelists:               Tags("p2p.whitelist"),
		Username:                 String("rpc.user"),
		Password:                 String("rpc.pass"),
		ServerUser:               String("rpc.user"),
		ServerPass:               String("rpc.pass"),
		LimitUser:                String("limit.user"),
		LimitPass:                String("limit.pass"),
		RPCConnect:               String("rpc.connect"),
		RPCListeners:             Tags("rpc.listen"),
		RPCCert:                  String("tls.cert"),
		RPCKey:                   String("tls.key"),
		RPCMaxClients:            Int("rpc.maxclients"),
		RPCMaxWebsockets:         Int("rpc.maxwebsockets"),
		RPCMaxConcurrentReqs:     Int("rpc.maxconcurrentreqs"),
		RPCQuirks:                Bool("rpc.quirks"),
		DisableRPC:               Bool("rpc.disable"),
		NoTLS:                    Bool("tls.disable"),
		DisableDNSSeed:           Bool("p2p.nodns"),
		ExternalIPs:              Tags("p2p.externalips"),
		Proxy:                    String("proxy.address"),
		ProxyUser:                String("proxy.user"),
		ProxyPass:                String("proxy.pass"),
		OnionProxy:               String("proxy.address"),
		OnionProxyUser:           String("proxy.user"),
		OnionProxyPass:           String("proxy.pass"),
		Onion:                    Bool("proxy.tor"),
		TorIsolation:             Bool("proxy.isolation"),
		TestNet3:                 tn,
		RegressionTest:           rn,
		SimNet:                   sn,
		AddCheckpoints:           Tags("chain.addcheckpoints"),
		DisableCheckpoints:       Bool("chain.disablecheckpoints"),
		DbType:                   String("chain.dbtype"),
		Profile:                  Int("app.profile"),
		CPUProfile:               String("app.cpuprofile"),
		Upnp:                     Bool("app.upnp"),
		MinRelayTxFee:            Float("p2p.minrelaytxfee"),
		FreeTxRelayLimit:         Float("p2p.freetxrelaylimit"),
		NoRelayPriority:          Bool("p2p.norelaypriority"),
		TrickleInterval:          Duration("p2p.trickleinterval"),
		MaxOrphanTxs:             Int("p2p.maxorphantxs"),
		Algo:                     String("mining.algo"),
		Generate:                 Bool("mining.generate"),
		GenThreads:               Int("mining.genthreads"),
		MiningAddrs:              Tags("mining.addresses"),
		MinerListener:            String("mining.listener"),
		MinerPass:                String("mining.pass"),
		BlockMinSize:             Int("block.minsize"),
		BlockMaxSize:             Int("block.maxsize"),
		BlockMinWeight:           Int("block.minweight"),
		BlockMaxWeight:           Int("block.maxweight"),
		BlockPrioritySize:        Int("block.prioritysize"),
		UserAgentComments:        Tags("p2p.useragentcomments"),
		NoPeerBloomFilters:       Bool("p2p.nobloomfilters"),
		NoCFilters:               Bool("p2p.nocfilters"),
		SigCacheMaxSize:          Int("chain.sigcachemaxsize"),
		BlocksOnly:               Bool("p2p.blocksonly"),
		TxIndex:                  Bool("chain.notxindex"),
		AddrIndex:                Bool("chain.noaddrindex"),
		RelayNonStd:              Bool("chain.relaynonstd"),
		RejectNonStd:             Bool("chain.rejectnonstd"),
		TLSSkipVerify:            Bool("tls.skipverify"),
		Wallet:                   Bool("wallet"),
		NoInitialLoad:            Bool("wallet.noinitialload"),
		WalletPass:               String("wallet.pass"),
		WalletServer:             String("rpc.wallet"),
		CAFile:                   String("tls.cafile"),
		OneTimeTLSKey:            Bool("tls.onetime"),
		ServerTLS:                Bool("tls.server"),
		LegacyRPCListeners:       Tags("rpc.listen"),
		LegacyRPCMaxClients:      Int("rpc.maxclients"),
		LegacyRPCMaxWebsockets:   Int("rpc.maxwebsockets"),
		ExperimentalRPCListeners: nil,
	}
	return
}
