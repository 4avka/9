package connmgr
import (
	"fmt"
	mrand "math/rand"
	"net"
	"strconv"
	"time"
	chaincfg "git.parallelcoin.io/dev/9/pkg/chain/config"
	"git.parallelcoin.io/dev/9/pkg/chain/wire"
	"git.parallelcoin.io/dev/9/pkg/util/cl"
)
const (
	// These constants are used by the DNS seed code to pick a random last seen time.
	secondsIn3Days int32 = 24 * 60 * 60 * 3
	secondsIn4Days int32 = 24 * 60 * 60 * 4
)
// OnSeed is the signature of the callback function which is invoked when DNS seeding is succesfull.
type OnSeed func(addrs []*wire.NetAddress)
// LookupFunc is the signature of the DNS lookup function.
type LookupFunc func(
	string) ([]net.IP, error)
// SeedFromDNS uses DNS seeding to populate the address manager with peers.
func SeedFromDNS(
	chainParams *chaincfg.Params, reqServices wire.ServiceFlag,
	lookupFn LookupFunc, seedFn OnSeed) {
	for _, dnsseed := range chainParams.DNSSeeds {
		var host string
		if !dnsseed.HasFiltering || reqServices == wire.SFNodeNetwork {
			host = dnsseed.Host
		} else {
			host = fmt.Sprintf("x%x.%s", uint64(reqServices), dnsseed.Host)
		}
		go func(host string) {
			randSource := mrand.New(mrand.NewSource(time.Now().UnixNano()))
			seedpeers, err := lookupFn(host)
			if err != nil {
				log <- cl.Infof{"DNS discovery failed on seed %s: %v", host, err}
				return
			}
			numPeers := len(seedpeers)
			log <- cl.Infof{"%d addresses found from DNS seed %s", numPeers, host}
			if numPeers == 0 {
				return
			}
			addresses := make([]*wire.NetAddress, len(seedpeers))
			// if this errors then we have *real* problems
			intPort, _ := strconv.Atoi(chainParams.DefaultPort)
			for i, peer := range seedpeers {
				addresses[i] = wire.NewNetAddressTimestamp(
					// bitcoind seeds with addresses from a time randomly selected between 3 and 7 days ago.
					time.Now().Add(-1*time.Second*time.Duration(secondsIn3Days+
						randSource.Int31n(secondsIn4Days))),
					0, peer, uint16(intPort))
			}
			seedFn(addresses)
		}(host)
	}
}
