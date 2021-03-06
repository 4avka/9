package netsync
import (
	"git.parallelcoin.io/dev/9/cmd/node/mempool"
	blockchain "git.parallelcoin.io/dev/9/pkg/chain"
	chaincfg "git.parallelcoin.io/dev/9/pkg/chain/config"
	chainhash "git.parallelcoin.io/dev/9/pkg/chain/hash"
	"git.parallelcoin.io/dev/9/pkg/chain/wire"
	"git.parallelcoin.io/dev/9/pkg/peer"
	"git.parallelcoin.io/dev/9/pkg/util"
)
// PeerNotifier exposes methods to notify peers of status changes to transactions, blocks, etc. Currently server (in the main package) implements this interface.
type PeerNotifier interface {
	AnnounceNewTransactions(newTxs []*mempool.TxDesc)
	UpdatePeerHeights(latestBlkHash *chainhash.Hash, latestHeight int32, updateSource *peer.Peer)
	RelayInventory(invVect *wire.InvVect, data interface{})
	TransactionConfirmed(tx *util.Tx)
}
// Config is a configuration struct used to initialize a new SyncManager.
type Config struct {
	PeerNotifier       PeerNotifier
	Chain              *blockchain.BlockChain
	TxMemPool          *mempool.TxPool
	ChainParams        *chaincfg.Params
	DisableCheckpoints bool
	MaxPeers           int
	FeeEstimator       *mempool.FeeEstimator
}
