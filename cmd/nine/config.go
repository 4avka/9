package nine
import (
	"net"
	"strings"
	"time"
	chaincfg "git.parallelcoin.io/dev/9/pkg/chain/config"
	"git.parallelcoin.io/dev/9/pkg/util"
)
type Mapstringstring map[string]*string
func (m Mapstringstring) String() (out string) {
	for i, x := range m {
		out += i + ":" + *x + " "
	}
	return strings.TrimSpace(out)
}
type Config struct {
	ConfigFile               *string
	AppDataDir               *string
	DataDir                  *string
	LogDir                   *string
	LogLevel                 *string
	Subsystems               *Mapstringstring
	Network                  *string
	AddPeers                 *[]string
	ConnectPeers             *[]string
	MaxPeers                 *int
	Listeners                *[]string
	DisableListen            *bool
	DisableBanning           *bool
	BanDuration              *time.Duration
	BanThreshold             *int
	Whitelists               *[]string
	Username                 *string
	Password                 *string
	ServerUser               *string
	ServerPass               *string
	LimitUser                *string
	LimitPass                *string
	RPCConnect               *string
	RPCListeners             *[]string
	RPCCert                  *string
	RPCKey                   *string
	RPCMaxClients            *int
	RPCMaxWebsockets         *int
	RPCMaxConcurrentReqs     *int
	RPCQuirks                *bool
	DisableRPC               *bool
	NoTLS                    *bool
	DisableDNSSeed           *bool
	ExternalIPs              *[]string
	Proxy                    *string
	ProxyUser                *string
	ProxyPass                *string
	OnionProxy               *string
	OnionProxyUser           *string
	OnionProxyPass           *string
	Onion                    *bool
	TorIsolation             *bool
	TestNet3                 *bool
	RegressionTest           *bool
	SimNet                   *bool
	AddCheckpoints           *[]string
	DisableCheckpoints       *bool
	DbType                   *string
	Profile                  *int
	CPUProfile               *string
	Upnp                     *bool
	MinRelayTxFee            *float64
	FreeTxRelayLimit         *float64
	NoRelayPriority          *bool
	TrickleInterval          *time.Duration
	MaxOrphanTxs             *int
	Algo                     *string
	Generate                 *bool
	GenThreads               *int
	MiningAddrs              *[]string
	MinerListener            *string
	MinerPass                *string
	BlockMinSize             *int
	BlockMaxSize             *int
	BlockMinWeight           *int
	BlockMaxWeight           *int
	BlockPrioritySize        *int
	UserAgentComments        *[]string
	NoPeerBloomFilters       *bool
	NoCFilters               *bool
	SigCacheMaxSize          *int
	BlocksOnly               *bool
	TxIndex                  *bool
	AddrIndex                *bool
	RelayNonStd              *bool
	RejectNonStd             *bool
	TLSSkipVerify            *bool
	Wallet                   *bool
	NoInitialLoad            *bool
	WalletPass               *string
	WalletServer             *string
	CAFile                   *string
	OneTimeTLSKey            *bool
	ServerTLS                *bool
	LegacyRPCListeners       *[]string
	LegacyRPCMaxClients      *int
	LegacyRPCMaxWebsockets   *int
	ExperimentalRPCListeners *[]string
	State                    *StateConfig
	ActiveNetParams          *Params
}
// StateConfig stores current state of the node
type StateConfig struct {
	Lookup              func(string) ([]net.IP, error)
	Oniondial           func(string, string, time.Duration) (net.Conn, error)
	Dial                func(string, string, time.Duration) (net.Conn, error)
	AddedCheckpoints    []chaincfg.Checkpoint
	ActiveMiningAddrs   []util.Address
	ActiveMinerKey      []byte
	ActiveMinRelayTxFee util.Amount
	ActiveWhitelists    []*net.IPNet
	DropAddrIndex       bool
	DropTxIndex         bool
	DropCfIndex         bool
	Save                bool
}
// Params is used to group parameters for various networks such as the main network and test networks.
type Params struct {
	*chaincfg.Params
	RPCPort string
}
// MainNetParams contains parameters specific to the main network (wire.MainNet).  NOTE: The RPC port is intentionally different than the reference implementation because pod does not handle wallet requests.  The separate wallet process listens on the well-known port and forwards requests it does not handle on to pod.  This approach allows the wallet process to emulate the full reference implementation RPC API.
var MainNetParams = Params{
	Params:  &chaincfg.MainNetParams,
	RPCPort: "11048",
}
// RegressionNetParams contains parameters specific to the regression test network (wire.TestNet).  NOTE: The RPC port is intentionally different than the reference implementation - see the MainNetParams comment for details.
var RegressionNetParams = Params{
	Params:  &chaincfg.RegressionNetParams,
	RPCPort: "31048",
}
// SimNetParams contains parameters specific to the simulation test network (wire.SimNet).
var SimNetParams = Params{
	Params:  &chaincfg.SimNetParams,
	RPCPort: "41048",
}
// TestNet3Params contains parameters specific to the test network (version 3) (wire.TestNet3).  NOTE: The RPC port is intentionally different than the reference implementation - see the MainNetParams comment for details.
var TestNet3Params = Params{
	Params:  &chaincfg.TestNet3Params,
	RPCPort: "21048",
}
var ActiveNetParams = &MainNetParams
