package node
import (
	"git.parallelcoin.io/dev/9/cmd/nine"
	"git.parallelcoin.io/dev/9/pkg/chain/wire"
)
// ActiveNetParams is a pointer to the parameters specific to the currently active bitcoin network.
var ActiveNetParams = &nine.MainNetParams
// NetName returns the name used when referring to a bitcoin network.  At the time of writing, pod currently places blocks for testnet version 3 in the data and log directory "testnet", which does not match the Name field of the chaincfg parameters.  This function can be used to override this directory name as "testnet" when the passed active network matches wire.TestNet3. A proper upgrade to move the data and log directories for this network to "testnet3" is planned for the future, at which point this function can be removed and the network parameter's name used instead.
func NetName(chainParams *nine.Params) string {
	switch chainParams.Net {
	case wire.TestNet3:
		return "testnet"
	default:
		return chainParams.Name
	}
}
