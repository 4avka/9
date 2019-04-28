package node

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	js "encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"git.parallelcoin.io/dev/9/cmd/node/mempool"
	blockchain "git.parallelcoin.io/dev/9/pkg/chain"
	chaincfg "git.parallelcoin.io/dev/9/pkg/chain/config"
	"git.parallelcoin.io/dev/9/pkg/chain/fork"
	chainhash "git.parallelcoin.io/dev/9/pkg/chain/hash"
	indexers "git.parallelcoin.io/dev/9/pkg/chain/index"
	"git.parallelcoin.io/dev/9/pkg/chain/mining"
	cpuminer "git.parallelcoin.io/dev/9/pkg/chain/mining/cpu"
	txscript "git.parallelcoin.io/dev/9/pkg/chain/tx/script"
	"git.parallelcoin.io/dev/9/pkg/chain/wire"
	database "git.parallelcoin.io/dev/9/pkg/db"
	p "git.parallelcoin.io/dev/9/pkg/peer"
	"git.parallelcoin.io/dev/9/pkg/rpc/json"
	"git.parallelcoin.io/dev/9/pkg/util"
	cl "git.parallelcoin.io/dev/9/pkg/util/cl"
	ec "git.parallelcoin.io/dev/9/pkg/util/elliptic"
	"github.com/btcsuite/websocket"
)

type commandHandler func(*rpcServer, interface{}, <-chan struct{}) (interface{}, error)

// gbtWorkState houses state that is used in between multiple RPC invocations to getblocktemplate.
type gbtWorkState struct {
	sync.Mutex
	lastTxUpdate  time.Time
	lastGenerated time.Time
	prevHash      *chainhash.Hash
	minTimestamp  time.Time
	template      *mining.BlockTemplate
	notifyMap     map[chainhash.Hash]map[int64]chan struct{}
	timeSource    blockchain.MedianTimeSource
	algo          string
}

// parsedRPCCmd represents a JSON-RPC request object that has been parsed into a known concrete command along with any error that might have happened while parsing it.
type parsedRPCCmd struct {
	id     interface{}
	method string
	cmd    interface{}
	err    *json.RPCError
}

// retrievedTx represents a transaction that was either loaded from the transaction memory pool or from the database.  When a transaction is loaded from the database, it is loaded with the raw serialized bytes while the mempool has the fully deserialized structure.  This structure therefore will have one of the two fields set depending on where is was retrieved from. This is mainly done for efficiency to avoid extra serialization steps when possible.
type retrievedTx struct {
	txBytes []byte
	blkHash *chainhash.Hash // Only set when transaction is in a block.
	tx      *util.Tx
}

// rpcServer provides a concurrent safe RPC server to a chain server.
type rpcServer struct {
	started                int32
	shutdown               int32
	Cfg                    rpcserverConfig
	authsha                [sha256.Size]byte
	limitauthsha           [sha256.Size]byte
	ntfnMgr                *wsNotificationManager
	numClients             int32
	statusLines            map[int]string
	statusLock             sync.RWMutex
	wg                     sync.WaitGroup
	gbtWorkState           *gbtWorkState
	helpCacher             *helpCacher
	requestProcessShutdown chan struct{}
	quit                   chan int
}

// rpcserverConfig is a descriptor containing the RPC server configuration.
type rpcserverConfig struct {
	// Listeners defines a slice of listeners for which the RPC server will take ownership of and accept connections.  Since the RPC server takes ownership of these listeners, they will be closed when the RPC server is stopped.
	Listeners []net.Listener
	// StartupTime is the unix timestamp for when the server that is hosting the RPC server started.
	StartupTime int64
	// ConnMgr defines the connection manager for the RPC server to use.  It provides the RPC server with a means to do things such as add, remove, connect, disconnect, and query peers as well as other connection-related data and tasks.
	ConnMgr rpcserverConnManager
	// SyncMgr defines the sync manager for the RPC server to use.
	SyncMgr rpcserverSyncManager
	// These fields allow the RPC server to interface with the local block chain data and state.
	TimeSource  blockchain.MedianTimeSource
	Chain       *blockchain.BlockChain
	ChainParams *chaincfg.Params
	DB          database.DB
	// TxMemPool defines the transaction memory pool to interact with.
	TxMemPool *mempool.TxPool
	// These fields allow the RPC server to interface with mining. Generator produces block templates and the CPUMiner solves them using the CPU.  CPU mining is typically only useful for test purposes when doing regression or simulation testing.
	Generator *mining.BlkTmplGenerator
	CPUMiner  *cpuminer.CPUMiner
	// These fields define any optional indexes the RPC server can make use of to provide additional data when queried.
	TxIndex   *indexers.TxIndex
	AddrIndex *indexers.AddrIndex
	CfIndex   *indexers.CfIndex
	// The fee estimator keeps track of how long transactions are left in the mempool before they are mined into blocks.
	FeeEstimator *mempool.FeeEstimator
	// Algo sets the algorithm expected from the RPC endpoint. This allows multiple ports to serve multiple types of miners with one main node per algorithm. Currently 514 for scrypt and anything else passes for sha256d. After hard fork 1 there is 9, and may be expanded in the future (equihash, cuckoo and cryptonight all require substantial block header/tx formatting changes)
	Algo string
}

// rpcserverConnManager represents a connection manager for use with the RPC server. The interface contract requires that all of these methods are safe for concurrent access.
type rpcserverConnManager interface {
	// Connect adds the provided address as a new outbound peer.  The permanent flag indicates whether or not to make the peer persistent and reconnect if the connection is lost.  Attempting to connect to an already existing peer will return an error.
	Connect(addr string, permanent bool) error
	// RemoveByID removes the peer associated with the provided id from the list of persistent peers.  Attempting to remove an id that does not exist will return an error.
	RemoveByID(id int32) error
	// RemoveByAddr removes the peer associated with the provided address from the list of persistent peers.  Attempting to remove an address that does not exist will return an error.
	RemoveByAddr(addr string) error
	// DisconnectByID disconnects the peer associated with the provided id. This applies to both inbound and outbound peers.  Attempting to remove an id that does not exist will return an error.
	DisconnectByID(id int32) error
	// DisconnectByAddr disconnects the peer associated with the provided address.  This applies to both inbound and outbound peers. Attempting to remove an address that does not exist will return an error.
	DisconnectByAddr(addr string) error
	// ConnectedCount returns the number of currently connected peers.
	ConnectedCount() int32
	// NetTotals returns the sum of all bytes received and sent across the network for all peers.
	NetTotals() (uint64, uint64)
	// ConnectedPeers returns an array consisting of all connected peers.
	ConnectedPeers() []rpcserverPeer
	// PersistentPeers returns an array consisting of all the persistent peers.
	PersistentPeers() []rpcserverPeer
	// BroadcastMessage sends the provided message to all currently connected peers.
	BroadcastMessage(msg wire.Message)
	// AddRebroadcastInventory adds the provided inventory to the list of inventories to be rebroadcast at random intervals until they show up in a block.
	AddRebroadcastInventory(iv *wire.InvVect, data interface{})
	// RelayTransactions generates and relays inventory vectors for all of the passed transactions to all connected peers.
	RelayTransactions(txns []*mempool.TxDesc)
}

// rpcserverPeer represents a peer for use with the RPC server. The interface contract requires that all of these methods are safe for concurrent access.
type rpcserverPeer interface {
	// ToPeer returns the underlying peer instance.
	ToPeer() *p.Peer
	// IsTxRelayDisabled returns whether or not the peer has disabled transaction relay.
	IsTxRelayDisabled() bool
	// BanScore returns the current integer value that represents how close the peer is to being banned.
	BanScore() uint32
	// FeeFilter returns the requested current minimum fee rate for which transactions should be announced.
	FeeFilter() int64
}

// rpcserverSyncManager represents a sync manager for use with the RPC server. The interface contract requires that all of these methods are safe for concurrent access.
type rpcserverSyncManager interface {
	// IsCurrent returns whether or not the sync manager believes the chain is current as compared to the rest of the network.
	IsCurrent() bool
	// SubmitBlock submits the provided block to the network after processing it locally.
	SubmitBlock(block *util.Block, flags blockchain.BehaviorFlags) (bool, error)
	// Pause pauses the sync manager until the returned channel is closed.
	Pause() chan<- struct{}
	// SyncPeerID returns the ID of the peer that is currently the peer being used to sync from or 0 if there is none.
	SyncPeerID() int32
	// LocateHeaders returns the headers of the blocks after the first known block in the provided locators until the provided stop hash or the current tip is reached, up to a max of wire.MaxBlockHeadersPerMsg hashes.
	LocateHeaders(locators []*chainhash.Hash, hashStop *chainhash.Hash) []wire.BlockHeader
}

// API version constants
const (
	jsonrpcSemverString = "1.3.0"
	jsonrpcSemverMajor  = 1
	jsonrpcSemverMinor  = 3
	jsonrpcSemverPatch  = 0
)
const (
	// rpcAuthTimeoutSeconds is the number of seconds a connection to the RPC server is allowed to stay open without authenticating before it is closed.
	rpcAuthTimeoutSeconds = 10
	// uint256Size is the number of bytes needed to represent an unsigned 256-bit integer.
	uint256Size = 32
	// gbtNonceRange is two 32-bit big-endian hexadecimal integers which represent the valid ranges of nonces returned by the getblocktemplate RPC.
	gbtNonceRange = "00000000ffffffff"
	// gbtRegenerateSeconds is the number of seconds that must pass before a new template is generated when the previous block hash has not changed and there have been changes to the available transactions in the memory pool.
	gbtRegenerateSeconds = 60
	// maxProtocolVersion is the max protocol version the server supports.
	maxProtocolVersion = 70002
)

// Errors
var (
	// ErrRPCNoWallet is an error returned to RPC clients when the provided command is recognized as a wallet command.
	ErrRPCNoWallet = &json.RPCError{
		Code:    json.ErrRPCNoWallet,
		Message: "This implementation does not implement wallet commands",
	}
)

// Errors
var (
	// ErrRPCUnimplemented is an error returned to RPC clients when the provided command is recognized, but not implemented.
	ErrRPCUnimplemented = &json.RPCError{
		Code:    json.ErrRPCUnimplemented,
		Message: "Command unimplemented",
	}
)
var (
	// gbtCapabilities describes additional capabilities returned with a block template generated by the getblocktemplate RPC. It is declared here to avoid the overhead of creating the slice on every invocation for constant data.
	gbtCapabilities = []string{"proposal"}
)
var (
	// gbtCoinbaseAux describes additional data that miners should include in the coinbase signature script.  It is declared here to avoid the overhead of creating a new object on every invocation for constant data.
	gbtCoinbaseAux = &json.GetBlockTemplateResultAux{
		Flags: hex.EncodeToString(builderScript(txscript.
			NewScriptBuilder().
			AddData([]byte(mining.CoinbaseFlags)))),
	}
)
var (
	// gbtMutableFields are the manipulations the server allows to be made to block templates generated by the getblocktemplate RPC.  It is declared here to avoid the overhead of creating the slice on every invocation for constant data.
	gbtMutableFields = []string{
		"time", "transactions/add", "prevblock", "coinbase/append",
	}
)

// list of commands that we recognize, but for which pod has no support because it lacks support for wallet functionality. For these commands the user should ask a connected instance of btcwallet.
var rpcAskWallet = map[string]struct{}{
	"addmultisigaddress":     {},
	"backupwallet":           {},
	"createencryptedwallet":  {},
	"createmultisig":         {},
	"dumpprivkey":            {},
	"dumpwallet":             {},
	"encryptwallet":          {},
	"getaccount":             {},
	"getaccountaddress":      {},
	"getaddressesbyaccount":  {},
	"getbalance":             {},
	"getnewaddress":          {},
	"getrawchangeaddress":    {},
	"getreceivedbyaccount":   {},
	"getreceivedbyaddress":   {},
	"gettransaction":         {},
	"gettxoutsetinfo":        {},
	"getunconfirmedbalance":  {},
	"getwalletinfo":          {},
	"importprivkey":          {},
	"importwallet":           {},
	"keypoolrefill":          {},
	"listaccounts":           {},
	"listaddressgroupings":   {},
	"listlockunspent":        {},
	"listreceivedbyaccount":  {},
	"listreceivedbyaddress":  {},
	"listsinceblock":         {},
	"listtransactions":       {},
	"listunspent":            {},
	"lockunspent":            {},
	"move":                   {},
	"sendfrom":               {},
	"sendmany":               {},
	"sendtoaddress":          {},
	"setaccount":             {},
	"settxfee":               {},
	"signmessage":            {},
	"signrawtransaction":     {},
	"walletlock":             {},
	"walletpassphrase":       {},
	"walletpassphrasechange": {},
}

// rpcHandlers maps RPC command strings to appropriate handler functions. This is set by init because help references rpcHandlers and thus causes a dependency loop.
var rpcHandlers map[string]commandHandler
var rpcHandlersBeforeInit = map[string]commandHandler{
	"addnode":              handleAddNode,
	"createrawtransaction": handleCreateRawTransaction,
	// "debuglevel":            handleDebugLevel,
	"decoderawtransaction":  handleDecodeRawTransaction,
	"decodescript":          handleDecodeScript,
	"estimatefee":           handleEstimateFee,
	"generate":              handleGenerate,
	"getaddednodeinfo":      handleGetAddedNodeInfo,
	"getbestblock":          handleGetBestBlock,
	"getbestblockhash":      handleGetBestBlockHash,
	"getblock":              handleGetBlock,
	"getblockchaininfo":     handleGetBlockChainInfo,
	"getblockcount":         handleGetBlockCount,
	"getblockhash":          handleGetBlockHash,
	"getblockheader":        handleGetBlockHeader,
	"getblocktemplate":      handleGetBlockTemplate,
	"getcfilter":            handleGetCFilter,
	"getcfilterheader":      handleGetCFilterHeader,
	"getconnectioncount":    handleGetConnectionCount,
	"getcurrentnet":         handleGetCurrentNet,
	"getdifficulty":         handleGetDifficulty,
	"getgenerate":           handleGetGenerate,
	"gethashespersec":       handleGetHashesPerSec,
	"getheaders":            handleGetHeaders,
	"getinfo":               handleGetInfo,
	"getmempoolinfo":        handleGetMempoolInfo,
	"getmininginfo":         handleGetMiningInfo,
	"getnettotals":          handleGetNetTotals,
	"getnetworkhashps":      handleGetNetworkHashPS,
	"getpeerinfo":           handleGetPeerInfo,
	"getrawmempool":         handleGetRawMempool,
	"getrawtransaction":     handleGetRawTransaction,
	"gettxout":              handleGetTxOut,
	"getwork":               handleGetWork,
	"help":                  handleHelp,
	"node":                  handleNode,
	"ping":                  handlePing,
	"searchrawtransactions": handleSearchRawTransactions,
	"sendrawtransaction":    handleSendRawTransaction,
	"setgenerate":           handleSetGenerate,
	"stop":                  handleStop,
	"submitblock":           handleSubmitBlock,
	"uptime":                handleUptime,
	"validateaddress":       handleValidateAddress,
	"verifychain":           handleVerifyChain,
	"verifymessage":         handleVerifyMessage,
	"version":               handleVersion,
}

// Commands that are available to a limited user
var rpcLimited = map[string]struct{}{
	// Websockets commands
	"loadtxfilter":          {},
	"notifyblocks":          {},
	"notifynewtransactions": {},
	"notifyreceived":        {},
	"notifyspent":           {},
	"rescan":                {},
	"rescanblocks":          {},
	"session":               {},
	// Websockets AND HTTP/S commands
	"help": {},
	// HTTP/S-only commands
	"createrawtransaction":  {},
	"decoderawtransaction":  {},
	"decodescript":          {},
	"estimatefee":           {},
	"getbestblock":          {},
	"getbestblockhash":      {},
	"getblock":              {},
	"getblockcount":         {},
	"getblockhash":          {},
	"getblockheader":        {},
	"getcfilter":            {},
	"getcfilterheader":      {},
	"getcurrentnet":         {},
	"getdifficulty":         {},
	"getheaders":            {},
	"getinfo":               {},
	"getnettotals":          {},
	"getnetworkhashps":      {},
	"getrawmempool":         {},
	"getrawtransaction":     {},
	"gettxout":              {},
	"searchrawtransactions": {},
	"sendrawtransaction":    {},
	"submitblock":           {},
	"uptime":                {},
	"validateaddress":       {},
	"verifymessage":         {},
	"version":               {},
}

// Commands that are currently unimplemented, but should ultimately be.
var rpcUnimplemented = map[string]struct{}{
	"estimatepriority": {},
	"getchaintips":     {},
	"getmempoolentry":  {},
	"getnetworkinfo":   {},
	"getwork":          {},
	"invalidateblock":  {},
	"preciousblock":    {},
	"reconsiderblock":  {},
}

// NotifyBlockConnected uses the newly-connected block to notify any long poll clients with a new block template when their existing block template is stale due to the newly connected block.
func (
	state *gbtWorkState,
) NotifyBlockConnected(
	blockHash *chainhash.Hash,
) {
	go func() {
		state.Lock()
		statelasttxupdate := state.lastTxUpdate
		state.Unlock()
		state.notifyLongPollers(blockHash, statelasttxupdate)
	}()
}

// NotifyMempoolTx uses the new last updated time for the transaction memory pool to notify any long poll clients with a new block template when their existing block template is stale due to enough time passing and the contents of the memory pool changing.
func (
	state *gbtWorkState,
) NotifyMempoolTx(
	lastUpdated time.Time,
) {
	go func() {
		state.Lock()
		defer state.Unlock()
		// No need to notify anything if no block templates have been generated yet.
		if state.prevHash == nil || state.lastGenerated.IsZero() {
			return
		}
		if time.Now().After(state.lastGenerated.Add(time.Second * gbtRegenerateSeconds)) {
			state.notifyLongPollers(state.prevHash, lastUpdated)
		}
	}()
}

// blockTemplateResult returns the current block template associated with the state as a json.GetBlockTemplateResult that is ready to be encoded to JSON and returned to the caller. This function MUST be called with the state locked.
func (
	state *gbtWorkState,
) blockTemplateResult(
	useCoinbaseValue bool,
	submitOld *bool,
) (
	*json.GetBlockTemplateResult,
	error,
) {
	// Ensure the timestamps are still in valid range for the template. This should really only ever happen if the local clock is changed after the template is generated, but it's important to avoid serving invalid block templates.
	template := state.template
	msgBlock := template.Block
	header := &msgBlock.Header
	adjustedTime := state.timeSource.AdjustedTime()
	maxTime := adjustedTime.Add(time.Second * blockchain.MaxTimeOffsetSeconds)
	if header.Timestamp.After(maxTime) {
		return nil, &json.RPCError{
			Code: json.ErrRPCOutOfRange,
			Message: fmt.Sprintf("The template time is after the "+
				"maximum allowed time for a block - template "+
				"time %v, maximum time %v", adjustedTime,
				maxTime),
		}
	}
	// Convert each transaction in the block template to a template result transaction.  The result does not include the coinbase, so notice the adjustments to the various lengths and indices.
	numTx := len(msgBlock.Transactions)
	transactions := make([]json.GetBlockTemplateResultTx, 0, numTx-1)
	txIndex := make(map[chainhash.Hash]int64, numTx)
	for i, tx := range msgBlock.Transactions {
		txHash := tx.TxHash()
		txIndex[txHash] = int64(i)
		// Skip the coinbase transaction.
		if i == 0 {
			continue
		}
		// Create an array of 1-based indices to transactions that come before this one in the transactions list which this one depends on.  This is necessary since the created block must ensure proper ordering of the dependencies.  A map is used before creating the final array to prevent duplicate entries when multiple inputs reference the same transaction.
		dependsMap := make(map[int64]struct{})
		for _, txIn := range tx.TxIn {
			if idx, ok := txIndex[txIn.PreviousOutPoint.Hash]; ok {
				dependsMap[idx] = struct{}{}
			}
		}
		depends := make([]int64, 0, len(dependsMap))
		for idx := range dependsMap {
			depends = append(depends, idx)
		}
		// Serialize the transaction for later conversion to hex.
		txBuf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
		if err := tx.Serialize(txBuf); err != nil {
			context := "Failed to serialize transaction"
			return nil, internalRPCError(err.Error(), context)
		}
		bTx := util.NewTx(tx)
		resultTx := json.GetBlockTemplateResultTx{
			Data:    hex.EncodeToString(txBuf.Bytes()),
			Hash:    txHash.String(),
			Depends: depends,
			Fee:     template.Fees[i],
			SigOps:  template.SigOpCosts[i],
			Weight:  blockchain.GetTransactionWeight(bTx),
		}
		transactions = append(transactions, resultTx)
	}
	// Generate the block template reply.  Note that following mutations are implied by the included or omission of fields:  Including MinTime -> time/decrement  Omitting CoinbaseTxn -> coinbase, generation
	targetDifficulty := fmt.Sprintf("%064x", blockchain.CompactToBig(header.Bits))
	templateID := encodeTemplateID(state.prevHash, state.lastGenerated)
	reply := json.GetBlockTemplateResult{
		Bits:         strconv.FormatInt(int64(header.Bits), 16),
		CurTime:      header.Timestamp.Unix(),
		Height:       int64(template.Height),
		PreviousHash: header.PrevBlock.String(),
		WeightLimit:  blockchain.MaxBlockWeight,
		SigOpLimit:   blockchain.MaxBlockSigOpsCost,
		SizeLimit:    wire.MaxBlockPayload,
		Transactions: transactions,
		Version:      header.Version,
		LongPollID:   templateID,
		SubmitOld:    submitOld,
		Target:       targetDifficulty,
		MinTime:      state.minTimestamp.Unix(),
		MaxTime:      maxTime.Unix(),
		Mutable:      gbtMutableFields,
		NonceRange:   gbtNonceRange,
		Capabilities: gbtCapabilities,
	}
	// If the generated block template includes transactions with witness data, then include the witness commitment in the GBT result.
	if template.WitnessCommitment != nil {
		reply.DefaultWitnessCommitment = hex.EncodeToString(template.WitnessCommitment)
	}
	if useCoinbaseValue {
		reply.CoinbaseAux = gbtCoinbaseAux
		reply.CoinbaseValue = &msgBlock.Transactions[0].TxOut[0].Value
	} else {
		// Ensure the template has a valid payment address associated with it when a full coinbase is requested.
		if !template.ValidPayAddress {
			return nil, &json.RPCError{
				Code: json.ErrRPCInternal.Code,
				Message: "A coinbase transaction has been " +
					"requested, but the server has not " +
					"been configured with any payment " +
					"addresses via --miningaddr",
			}
		}
		// Serialize the transaction for conversion to hex.
		tx := msgBlock.Transactions[0]
		txBuf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
		if err := tx.Serialize(txBuf); err != nil {
			context := "Failed to serialize transaction"
			return nil, internalRPCError(err.Error(), context)
		}
		resultTx := json.GetBlockTemplateResultTx{
			Data:    hex.EncodeToString(txBuf.Bytes()),
			Hash:    tx.TxHash().String(),
			Depends: []int64{},
			Fee:     template.Fees[0],
			SigOps:  template.SigOpCosts[0],
		}
		reply.CoinbaseTxn = &resultTx
	}
	return &reply, nil
}

// notifyLongPollers notifies any channels that have been registered to be notified when block templates are stale. This function MUST be called with the state locked.
func (
	state *gbtWorkState,
) notifyLongPollers(
	latestHash *chainhash.Hash,
	lastGenerated time.Time,
) {
	// Notify anything that is waiting for a block template update from a hash which is not the hash of the tip of the best chain since their work is now invalid.
	for hash, channels := range state.notifyMap {
		if !hash.IsEqual(latestHash) {
			for _, c := range channels {
				close(c)
			}
			delete(state.notifyMap, hash)
		}
	}
	// Return now if the provided last generated timestamp has not been initialized.
	if lastGenerated.IsZero() {
		return
	}
	// Return now if there is nothing registered for updates to the current best block hash.
	channels, ok := state.notifyMap[*latestHash]
	if !ok {
		return
	}
	// Notify anything that is waiting for a block template update from a block template generated before the most recently generated block template.
	lastGeneratedUnix := lastGenerated.Unix()
	for lastGen, c := range channels {
		if lastGen < lastGeneratedUnix {
			close(c)
			delete(channels, lastGen)
		}
	}
	// Remove the entry altogether if there are no more registered channels.
	if len(channels) == 0 {
		delete(state.notifyMap, *latestHash)
	}
}

// templateUpdateChan returns a channel that will be closed once the block template associated with the passed previous hash and last generated time is stale.  The function will return existing channels for duplicate parameters which allows  to wait for the same block template without requiring a different channel for each client. This function MUST be called with the state locked.
func (
	state *gbtWorkState,
) templateUpdateChan(
	prevHash *chainhash.Hash,
	lastGenerated int64,
) chan struct{} {
	// Either get the current list of channels waiting for updates about changes to block template for the previous hash or create a new one.
	channels, ok := state.notifyMap[*prevHash]
	if !ok {
		m := make(map[int64]chan struct{})
		state.notifyMap[*prevHash] = m
		channels = m
	}
	// Get the current channel associated with the time the block template was last generated or create a new one.
	c, ok := channels[lastGenerated]
	if !ok {
		c = make(chan struct{})
		channels[lastGenerated] = c
	}
	return c
}

// updateBlockTemplate creates or updates a block template for the work state. A new block template will be generated when the current best block has changed or the transactions in the memory pool have been updated and it has been long enough since the last template was generated.  Otherwise, the timestamp for the existing block template is updated (and possibly the difficulty on testnet per the consesus rules).  Finally, if the useCoinbaseValue flag is false and the existing block template does not already contain a valid payment address, the block template will be updated with a randomly selected payment address from the list of configured addresses. This function MUST be called with the state locked.
func (
	state *gbtWorkState,
) updateBlockTemplate(
	s *rpcServer,
	useCoinbaseValue bool,
) error {
	generator := s.Cfg.Generator
	lastTxUpdate := generator.TxSource().LastUpdated()
	if lastTxUpdate.IsZero() {
		lastTxUpdate = time.Now()
	}
	// Generate a new block template when the current best block has changed or the transactions in the memory pool have been updated and it has been at least gbtRegenerateSecond since the last template was generated.
	var msgBlock *wire.MsgBlock
	var targetDifficulty string
	latestHash := &s.Cfg.Chain.BestSnapshot().Hash
	template := state.template
	if template == nil || state.prevHash == nil ||
		!state.prevHash.IsEqual(latestHash) ||
		(state.lastTxUpdate != lastTxUpdate &&
			time.Now().After(state.lastGenerated.Add(time.Second*
				gbtRegenerateSeconds))) {
		// Reset the previous best hash the block template was generated against so any errors below cause the next invocation to try again.
		state.prevHash = nil
		// Choose a payment address at random if the caller requests a full coinbase as opposed to only the pertinent details needed to create their own coinbase.
		var payAddr util.Address
		if !useCoinbaseValue {
			payAddr = StateCfg.ActiveMiningAddrs[rand.Intn(len(StateCfg.ActiveMiningAddrs))]
		}
		// Create a new block template that has a coinbase which anyone can redeem.  This is only acceptable because the returned block template doesn't include the coinbase, so the caller will ultimately create their own coinbase which pays to the appropriate address(es).
		blkTemplate, err := generator.NewBlockTemplate(payAddr, state.algo)
		if err != nil {
			return internalRPCError("(rpcserver.go) Failed to create new block "+
				"template: "+err.Error(), "")
		}
		template = blkTemplate
		msgBlock = template.Block
		targetDifficulty = fmt.Sprintf("%064x",
			blockchain.CompactToBig(msgBlock.Header.Bits))
		// Get the minimum allowed timestamp for the block based on the median timestamp of the last several blocks per the chain consensus rules.
		best := s.Cfg.Chain.BestSnapshot()
		minTimestamp := mining.MinimumMedianTime(best)
		// Update work state to ensure another block template isn't generated until needed.
		state.template = template
		state.lastGenerated = time.Now()
		state.lastTxUpdate = lastTxUpdate
		state.prevHash = latestHash
		state.minTimestamp = minTimestamp
		log <- cl.Debugf{
			"generated block template (timestamp %v, target %s, merkle root %s)",
			msgBlock.Header.Timestamp,
			targetDifficulty,
			msgBlock.Header.MerkleRoot,
		}
		// Notify any clients that are long polling about the new template.
		state.notifyLongPollers(latestHash, lastTxUpdate)
	} else {
		// At this point, there is a saved block template and another request for a template was made, but either the available transactions haven't change or it hasn't been long enough to trigger a new block template to be generated.  So, update the existing block template. When the caller requires a full coinbase as opposed to only the pertinent details needed to create their own coinbase, add a payment address to the output of the coinbase of the template if it doesn't already have one.  Since this requires mining addresses to be specified via the config, an error is returned if none have been specified.
		if !useCoinbaseValue && !template.ValidPayAddress {
			// Choose a payment address at random.
			payToAddr := StateCfg.ActiveMiningAddrs[rand.Intn(len(StateCfg.ActiveMiningAddrs))]
			// Update the block coinbase output of the template to pay to the randomly selected payment address.
			pkScript, err := txscript.PayToAddrScript(payToAddr)
			if err != nil {
				context := "Failed to create pay-to-addr script"
				return internalRPCError(err.Error(), context)
			}
			template.Block.Transactions[0].TxOut[0].PkScript = pkScript
			template.ValidPayAddress = true
			// Update the merkle root.
			block := util.NewBlock(template.Block)
			merkles := blockchain.BuildMerkleTreeStore(block.Transactions(), false)
			template.Block.Header.MerkleRoot = *merkles[len(merkles)-1]
		}
		// Set locals for convenience.
		msgBlock = template.Block
		targetDifficulty = fmt.Sprintf("%064x",
			blockchain.CompactToBig(msgBlock.Header.Bits))
		// Update the time of the block template to the current time while accounting for the median time of the past several blocks per the chain consensus rules.
		generator.UpdateBlockTime(msgBlock)
		msgBlock.Header.Nonce = 0
		log <- cl.Debugf{
			"updated block template (timestamp %v, target %s)",
			msgBlock.Header.Timestamp,
			targetDifficulty,
		}
	}
	return nil
}

// NotifyNewTransactions notifies both websocket and getblocktemplate long poll clients of the passed transactions.  This function should be called whenever new transactions are added to the mempool.
func (
	s *rpcServer,
) NotifyNewTransactions(
	txns []*mempool.TxDesc,
) {
	for _, txD := range txns {
		// Notify websocket clients about mempool transactions.
		s.ntfnMgr.NotifyMempoolTx(txD.Tx, true)
		// Potentially notify any getblocktemplate long poll clients about stale block templates due to the new transaction.
		s.gbtWorkState.NotifyMempoolTx(s.Cfg.TxMemPool.LastUpdated())
	}
}

// RequestedProcessShutdown returns a channel that is sent to when an authorized RPC client requests the process to shutdown.  If the request can not be read immediately, it is dropped.
func (
	s *rpcServer,
) RequestedProcessShutdown() <-chan struct{} {
	return s.requestProcessShutdown
}

// Start is used by server.go to start the rpc listener.
func (
	s *rpcServer,
) Start() {
	if atomic.AddInt32(&s.started, 1) != 1 {
		return
	}
	rpcServeMux := http.NewServeMux()
	httpServer := &http.Server{
		Handler: rpcServeMux,
		// Timeout connections which don't complete the initial handshake within the allowed timeframe.
		ReadTimeout: time.Second * rpcAuthTimeoutSeconds,
	}
	rpcServeMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "close")
		w.Header().Set("Content-Type", "application/json")
		r.Close = true
		// Limit the number of connections to max allowed.
		if s.limitConnections(w, r.RemoteAddr) {
			return
		}
		// Keep track of the number of connected clients.
		s.incrementClients()
		defer s.decrementClients()
		_, isAdmin, err := s.checkAuth(r, true)
		if err != nil {
			jsonAuthFail(w)
			return
		}
		// Read and respond to the request.
		s.jsonRPCRead(w, r, isAdmin)
	})
	// Websocket endpoint.
	rpcServeMux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		authenticated, isAdmin, err := s.checkAuth(r, false)
		if err != nil {
			jsonAuthFail(w)
			return
		}
		// Attempt to upgrade the connection to a websocket connection using the default size for read/write buffers.
		ws, err := websocket.Upgrade(w, r, nil, 0, 0)
		if err != nil {
			if _, ok := err.(websocket.HandshakeError); !ok {
				log <- cl.Error{"unexpected websocket error:", err}
			}
			http.Error(w, "400 Bad Request.", http.StatusBadRequest)
			return
		}
		s.WebsocketHandler(ws, r.RemoteAddr, authenticated, isAdmin)
	})
	for _, listener := range s.Cfg.Listeners {
		s.wg.Add(1)
		go func(listener net.Listener) {
			log <- cl.Info{"RPC server listening on", listener.Addr()}
			httpServer.Serve(listener)
			log <- cl.Trace{"RPC listener done for", listener.Addr()}
			s.wg.Done()
		}(listener)
	}
	s.ntfnMgr.wg.Add(2)
	s.ntfnMgr.Start()
}

// Stop is used by server.go to stop the rpc listener.
func (
	s *rpcServer,
) Stop() error {
	if atomic.AddInt32(&s.shutdown, 1) != 1 {
		log <- cl.Inf("RPC server is already in the process of shutting down")
		return nil
	}
	log <- cl.Wrn("RPC server shutting down")
	for _, listener := range s.Cfg.Listeners {
		err := listener.Close()
		if err != nil {
			log <- cl.Error{"problem shutting down RPC:", err}
			return err
		}
	}
	s.ntfnMgr.Shutdown()
	s.ntfnMgr.WaitForShutdown()
	close(s.quit)
	s.wg.Wait()
	log <- cl.Inf("RPC server shutdown complete")
	return nil
}

// checkAuth checks the HTTP Basic authentication supplied by a wallet or RPC client in the HTTP request r.  If the supplied authentication does not match the username and password expected, a non-nil error is returned. This check is time-constant. The first bool return value signifies auth success (true if successful) and the second bool return value specifies whether the user can change the state of the server (true) or whether the user is limited (false). The second is always false if the first is.
func (
	s *rpcServer,
) checkAuth(
	r *http.Request,
	require bool,
) (
	bool,
	bool, error,
) {
	authhdr := r.Header["Authorization"]
	if len(authhdr) <= 0 {
		if require {
			log <- cl.Warn{"RPC authentication failure from", r.RemoteAddr}
			return false, false, errors.New("auth failure")
		}
		return false, false, nil
	}
	authsha := sha256.Sum256([]byte(authhdr[0]))
	// Check for limited auth first as in environments with limited users, those are probably expected to have a higher volume of calls
	limitcmp := subtle.ConstantTimeCompare(authsha[:], s.limitauthsha[:])
	if limitcmp == 1 {
		return true, false, nil
	}
	// Check for admin-level auth
	cmp := subtle.ConstantTimeCompare(authsha[:], s.authsha[:])
	if cmp == 1 {
		return true, true, nil
	}
	// Request's auth doesn't match either user
	log <- cl.Warn{"RPC authentication failure from", r.RemoteAddr}
	return false, false, errors.New("auth failure")
}

// decrementClients subtracts one from the number of connected RPC clients. Note this only applies to standard clients.  Websocket clients have their own limits and are tracked separately. This function is safe for concurrent access.
func (
	s *rpcServer,
) decrementClients() {
	atomic.AddInt32(&s.numClients, -1)
}

// Callback for notifications from blockchain.  It notifies clients that are long polling for changes or subscribed to websockets notifications.
func (
	s *rpcServer,
) handleBlockchainNotification(
	notification *blockchain.Notification,
) {
	switch notification.Type {
	case blockchain.NTBlockAccepted:
		block, ok := notification.Data.(*util.Block)
		if !ok {
			log <- cl.Wrn("chain accepted notification is not a block")
			break
		}
		// Allow any clients performing long polling via the getblocktemplate RPC to be notified when the new block causes their old block template to become stale.
		s.gbtWorkState.NotifyBlockConnected(block.Hash())
	case blockchain.NTBlockConnected:
		block, ok := notification.Data.(*util.Block)
		if !ok {
			log <- cl.Wrn("chain connected notification is not a block")
			break
		}
		// Notify registered websocket clients of incoming block.
		s.ntfnMgr.NotifyBlockConnected(block)
	case blockchain.NTBlockDisconnected:
		block, ok := notification.Data.(*util.Block)
		if !ok {
			log <- cl.Wrn("chain disconnected notification is not a block.")
			break
		}
		// Notify registered websocket clients.
		s.ntfnMgr.NotifyBlockDisconnected(block)
	}
}

// httpStatusLine returns a response Status-Line (RFC 2616 Section 6.1) for the given request and response status code.  This function was lifted and adapted from the standard library HTTP server code since it's not exported.
func (
	s *rpcServer,
) httpStatusLine(
	req *http.Request,
	code int,
) string {
	// Fast path:
	key := code
	proto11 := req.ProtoAtLeast(1, 1)
	if !proto11 {
		key = -key
	}
	s.statusLock.RLock()
	line, ok := s.statusLines[key]
	s.statusLock.RUnlock()
	if ok {
		return line
	}
	// Slow path:
	proto := "HTTP/1.0"
	if proto11 {
		proto = "HTTP/1.1"
	}
	codeStr := strconv.Itoa(code)
	text := http.StatusText(code)
	if text != "" {
		line = proto + " " + codeStr + " " + text + "\r\n"
		s.statusLock.Lock()
		s.statusLines[key] = line
		s.statusLock.Unlock()
	} else {
		text = "status code " + codeStr
		line = proto + " " + codeStr + " " + text + "\r\n"
	}
	return line
}

// incrementClients adds one to the number of connected RPC clients.  Note this only applies to standard clients.  Websocket clients have their own limits and are tracked separately. This function is safe for concurrent access.
func (
	s *rpcServer,
) incrementClients() {
	atomic.AddInt32(&s.numClients, 1)
}

// jsonRPCRead handles reading and responding to RPC messages.
func (
	s *rpcServer,
) jsonRPCRead(
	w http.ResponseWriter,
	r *http.Request,
	isAdmin bool,
) {
	if atomic.LoadInt32(&s.shutdown) != 0 {
		return
	}
	// Read and close the JSON-RPC request body from the caller.
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		errCode := http.StatusBadRequest
		http.Error(w, fmt.Sprintf("%d error reading JSON message: %v",
			errCode, err), errCode)
		return
	}
	// Unfortunately, the http server doesn't provide the ability to change the read deadline for the new connection and having one breaks long polling.  However, not having a read deadline on the initial connection would mean clients can connect and idle forever.  Thus, hijack the connecton from the HTTP server, clear the read deadline, and handle writing the response manually.
	hj, ok := w.(http.Hijacker)
	if !ok {
		errMsg := "webserver doesn't support hijacking"
		log <- cl.Warnf{errMsg}
		errCode := http.StatusInternalServerError
		http.Error(w, strconv.Itoa(errCode)+" "+errMsg, errCode)
		return
	}
	conn, buf, err := hj.Hijack()
	if err != nil {
		log <- cl.Warn{"failed to hijack HTTP connection:", err}
		errCode := http.StatusInternalServerError
		http.Error(w, strconv.Itoa(errCode)+" "+err.Error(), errCode)
		return
	}
	defer conn.Close()
	defer buf.Flush()
	conn.SetReadDeadline(timeZeroVal)
	// Attempt to parse the raw body into a JSON-RPC request.
	var responseID interface{}
	var jsonErr error
	var result interface{}
	var request json.Request
	if err := js.Unmarshal(body, &request); err != nil {
		jsonErr = &json.RPCError{
			Code:    json.ErrRPCParse.Code,
			Message: "Failed to parse request: " + err.Error(),
		}
	}
	if jsonErr == nil {
		/* The JSON-RPC 1.0 spec defines that notifications must have their "id" set to null and states that notifications do not have a response. A JSON-RPC 2.0 notification is a request with "json-rpc":"2.0", and without an "id" member.
		The specification states that notifications must not be responded to. JSON-RPC 2.0 permits the null value as a valid request id, therefore such requests are not notifications.
		Bitcoin Core serves requests with "id":null or even an absent "id", and responds to such requests with "id":null in the response. Pod does not respond to any request without and "id" or "id":null, regardless the indicated JSON-RPC protocol version unless RPC quirks are enabled.
		With RPC quirks enabled, such requests will be responded to if the reqeust does not indicate JSON-RPC version. RPC quirks can be enabled by the user to avoid compatibility issues with software relying on Core's behavior.
		*/
		if request.ID == nil && !(*Cfg.RPCQuirks && request.Jsonrpc == "") {
			return
		}
		// The parse was at least successful enough to have an ID so set it for the response.
		responseID = request.ID
		// Setup a close notifier.  Since the connection is hijacked, the CloseNotifer on the ResponseWriter is not available.
		closeChan := make(chan struct{}, 1)
		go func() {
			_, err := conn.Read(make([]byte, 1))
			if err != nil {
				close(closeChan)
			}
		}()
		// Check if the user is limited and set error if method unauthorized
		if !isAdmin {
			if _, ok := rpcLimited[request.Method]; !ok {
				jsonErr = &json.RPCError{
					Code:    json.ErrRPCInvalidParams.Code,
					Message: "limited user not authorized for this method",
				}
			}
		}
		if jsonErr == nil {
			// Attempt to parse the JSON-RPC request into a known concrete command.
			parsedCmd := parseCmd(&request)
			if parsedCmd.err != nil {
				jsonErr = parsedCmd.err
			} else {
				result, jsonErr = s.standardCmdResult(parsedCmd, closeChan)
			}
		}
	}
	// Marshal the response.
	msg, err := createMarshalledReply(responseID, result, jsonErr)
	if err != nil {
		log <- cl.Error{"failed to marshal reply:", err}
		return
	}
	// Write the response.
	err = s.writeHTTPResponseHeaders(r, w.Header(), http.StatusOK, buf)
	if err != nil {
		log <- cl.Error{err.Error()}
		return
	}
	if _, err := buf.Write(msg); err != nil {
		log <- cl.Error{"failed to write marshalled reply:", err}
	}
	// Terminate with newline to maintain compatibility with Bitcoin Core.
	if err := buf.WriteByte('\n'); err != nil {
		log <- cl.Error{"failed to append terminating newline to reply:", err}
	}
}

// limitConnections responds with a 503 service unavailable and returns true if adding another client would exceed the maximum allow RPC clients. This function is safe for concurrent access.
func (
	s *rpcServer,
) limitConnections(
	w http.ResponseWriter,
	remoteAddr string,
) bool {
	if int(atomic.LoadInt32(&s.numClients)+1) > *Cfg.RPCMaxClients {
		log <- cl.Infof{
			"max RPC clients exceeded [%d] - disconnecting client %s",
			Cfg.RPCMaxClients, remoteAddr}
		http.Error(w, "503 Too busy.  Try again later.",
			http.StatusServiceUnavailable)
		return true
	}
	return false
}

// standardCmdResult checks that a parsed command is a standard Bitcoin JSON-RPC command and runs the appropriate handler to reply to the command.  Any commands which are not recognized or not implemented will return an error suitable for use in replies.
func (
	s *rpcServer,
) standardCmdResult(
	cmd *parsedRPCCmd,
	closeChan <-chan struct{},
) (
	interface{},
	error,
) {
	handler, ok := rpcHandlers[cmd.method]
	if ok {
		goto handled
	}
	_, ok = rpcAskWallet[cmd.method]
	if ok {
		handler = handleAskWallet
		goto handled
	}
	_, ok = rpcUnimplemented[cmd.method]
	if ok {
		handler = handleUnimplemented
		goto handled
	}
	return nil, json.ErrRPCMethodNotFound
handled:
	return handler(s, cmd.cmd, closeChan)
}

// writeHTTPResponseHeaders writes the necessary response headers prior to writing an HTTP body given a request to use for protocol negotiation, headers to write, a status code, and a writer.
func (
	s *rpcServer,
) writeHTTPResponseHeaders(
	req *http.Request,
	headers http.Header,
	code int,
	w io.Writer,
) error {
	_, err := io.WriteString(w, s.httpStatusLine(req, code))
	if err != nil {
		return err
	}
	err = headers.Write(w)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, "\r\n")
	return err
}

// builderScript is a convenience function which is used for hard-coded scripts built with the script builder. Any errors are converted to a panic since it is only, and must only, be used with hard-coded, and therefore, known good, scripts.
func builderScript(
	builder *txscript.ScriptBuilder,
) []byte {
	script, err := builder.Script()
	if err != nil {
		panic(err)
	}
	return script
}

// chainErrToGBTErrString converts an error returned from btcchain to a string which matches the reasons and format described in BIP0022 for rejection reasons.
func chainErrToGBTErrString(
	err error,
) string {
	// When the passed error is not a RuleError, just return a generic rejected string with the error text.
	ruleErr, ok := err.(blockchain.RuleError)
	if !ok {
		return "rejected: " + err.Error()
	}
	switch ruleErr.ErrorCode {
	case blockchain.ErrDuplicateBlock:
		return "duplicate"
	case blockchain.ErrBlockTooBig:
		return "bad-blk-length"
	case blockchain.ErrBlockWeightTooHigh:
		return "bad-blk-weight"
	case blockchain.ErrBlockVersionTooOld:
		return "bad-version"
	case blockchain.ErrInvalidTime:
		return "bad-time"
	case blockchain.ErrTimeTooOld:
		return "time-too-old"
	case blockchain.ErrTimeTooNew:
		return "time-too-new"
	case blockchain.ErrDifficultyTooLow:
		return "bad-diffbits"
	case blockchain.ErrUnexpectedDifficulty:
		return "bad-diffbits"
	case blockchain.ErrHighHash:
		return "high-hash"
	case blockchain.ErrBadMerkleRoot:
		return "bad-txnmrklroot"
	case blockchain.ErrBadCheckpoint:
		return "bad-checkpoint"
	case blockchain.ErrForkTooOld:
		return "fork-too-old"
	case blockchain.ErrCheckpointTimeTooOld:
		return "checkpoint-time-too-old"
	case blockchain.ErrNoTransactions:
		return "bad-txns-none"
	case blockchain.ErrNoTxInputs:
		return "bad-txns-noinputs"
	case blockchain.ErrNoTxOutputs:
		return "bad-txns-nooutputs"
	case blockchain.ErrTxTooBig:
		return "bad-txns-size"
	case blockchain.ErrBadTxOutValue:
		return "bad-txns-outputvalue"
	case blockchain.ErrDuplicateTxInputs:
		return "bad-txns-dupinputs"
	case blockchain.ErrBadTxInput:
		return "bad-txns-badinput"
	case blockchain.ErrMissingTxOut:
		return "bad-txns-missinginput"
	case blockchain.ErrUnfinalizedTx:
		return "bad-txns-unfinalizedtx"
	case blockchain.ErrDuplicateTx:
		return "bad-txns-duplicate"
	case blockchain.ErrOverwriteTx:
		return "bad-txns-overwrite"
	case blockchain.ErrImmatureSpend:
		return "bad-txns-maturity"
	case blockchain.ErrSpendTooHigh:
		return "bad-txns-highspend"
	case blockchain.ErrBadFees:
		return "bad-txns-fees"
	case blockchain.ErrTooManySigOps:
		return "high-sigops"
	case blockchain.ErrFirstTxNotCoinbase:
		return "bad-txns-nocoinbase"
	case blockchain.ErrMultipleCoinbases:
		return "bad-txns-multicoinbase"
	case blockchain.ErrBadCoinbaseScriptLen:
		return "bad-cb-length"
	case blockchain.ErrBadCoinbaseValue:
		return "bad-cb-value"
	case blockchain.ErrMissingCoinbaseHeight:
		return "bad-cb-height"
	case blockchain.ErrBadCoinbaseHeight:
		return "bad-cb-height"
	case blockchain.ErrScriptMalformed:
		return "bad-script-malformed"
	case blockchain.ErrScriptValidation:
		return "bad-script-validate"
	case blockchain.ErrUnexpectedWitness:
		return "unexpected-witness"
	case blockchain.ErrInvalidWitnessCommitment:
		return "bad-witness-nonce-size"
	case blockchain.ErrWitnessCommitmentMismatch:
		return "bad-witness-merkle-match"
	case blockchain.ErrPreviousBlockUnknown:
		return "prev-blk-not-found"
	case blockchain.ErrInvalidAncestorBlock:
		return "bad-prevblk"
	case blockchain.ErrPrevBlockNotBest:
		return "inconclusive-not-best-prvblk"
	}
	return "rejected: " + err.Error()
}

// createMarshalledReply returns a new marshalled JSON-RPC response given the passed parameters.  It will automatically convert errors that are not of the type *json.RPCError to the appropriate type as needed.
func createMarshalledReply(
	id,
	result interface{},
	replyErr error,
) (
	[]byte,
	error,
) {
	var jsonErr *json.RPCError
	if replyErr != nil {
		if jErr, ok := replyErr.(*json.RPCError); ok {
			jsonErr = jErr
		} else {
			jsonErr = internalRPCError(replyErr.Error(), "")
		}
	}
	return json.MarshalResponse(id, result, jsonErr)
}

// createTxRawResult converts the passed transaction and associated parameters to a raw transaction JSON object.
func createTxRawResult(
	chainParams *chaincfg.Params,
	mtx *wire.MsgTx,
	txHash string,
	blkHeader *wire.BlockHeader,
	blkHash string,
	blkHeight int32,
	chainHeight int32,
) (
	*json.TxRawResult,
	error,
) {
	mtxHex, err := messageToHex(mtx)
	if err != nil {
		return nil, err
	}
	txReply := &json.TxRawResult{
		Hex:      mtxHex,
		Txid:     txHash,
		Hash:     mtx.WitnessHash().String(),
		Size:     int32(mtx.SerializeSize()),
		Vsize:    int32(mempool.GetTxVirtualSize(util.NewTx(mtx))),
		Vin:      createVinList(mtx),
		Vout:     createVoutList(mtx, chainParams, nil),
		Version:  mtx.Version,
		LockTime: mtx.LockTime,
	}
	if blkHeader != nil {
		// This is not a typo, they are identical in bitcoind as well.
		txReply.Time = blkHeader.Timestamp.Unix()
		txReply.Blocktime = blkHeader.Timestamp.Unix()
		txReply.BlockHash = blkHash
		txReply.Confirmations = uint64(1 + chainHeight - blkHeight)
	}
	return txReply, nil
}

// createVinList returns a slice of JSON objects for the inputs of the passed transaction.
func createVinList(
	mtx *wire.MsgTx,
) []json.Vin {
	// Coinbase transactions only have a single txin by definition.
	vinList := make([]json.Vin, len(mtx.TxIn))
	if blockchain.IsCoinBaseTx(mtx) {
		txIn := mtx.TxIn[0]
		vinList[0].Coinbase = hex.EncodeToString(txIn.SignatureScript)
		vinList[0].Sequence = txIn.Sequence
		vinList[0].Witness = witnessToHex(txIn.Witness)
		return vinList
	}
	for i, txIn := range mtx.TxIn {
		// The disassembled string will contain [error] inline if the script doesn't fully parse, so ignore the error here.
		disbuf, _ := txscript.DisasmString(txIn.SignatureScript)
		vinEntry := &vinList[i]
		vinEntry.Txid = txIn.PreviousOutPoint.Hash.String()
		vinEntry.Vout = txIn.PreviousOutPoint.Index
		vinEntry.Sequence = txIn.Sequence
		vinEntry.ScriptSig = &json.ScriptSig{
			Asm: disbuf,
			Hex: hex.EncodeToString(txIn.SignatureScript),
		}
		if mtx.HasWitness() {
			vinEntry.Witness = witnessToHex(txIn.Witness)
		}
	}
	return vinList
}

// createVinListPrevOut returns a slice of JSON objects for the inputs of the passed transaction.
func createVinListPrevOut(
	s *rpcServer,
	mtx *wire.MsgTx,
	chainParams *chaincfg.Params,
	vinExtra bool,
	filterAddrMap map[string]struct{},
) (
	[]json.VinPrevOut,
	error,
) {
	// Coinbase transactions only have a single txin by definition.
	if blockchain.IsCoinBaseTx(mtx) {
		// Only include the transaction if the filter map is empty because a coinbase input has no addresses and so would never match a non-empty filter.
		if len(filterAddrMap) != 0 {
			return nil, nil
		}
		txIn := mtx.TxIn[0]
		vinList := make([]json.VinPrevOut, 1)
		vinList[0].Coinbase = hex.EncodeToString(txIn.SignatureScript)
		vinList[0].Sequence = txIn.Sequence
		return vinList, nil
	}
	// Use a dynamically sized list to accommodate the address filter.
	vinList := make([]json.VinPrevOut, 0, len(mtx.TxIn))
	// Lookup all of the referenced transaction outputs needed to populate the previous output information if requested.
	var originOutputs map[wire.OutPoint]wire.TxOut
	if vinExtra || len(filterAddrMap) > 0 {
		var err error
		originOutputs, err = fetchInputTxos(s, mtx)
		if err != nil {
			return nil, err
		}
	}
	for _, txIn := range mtx.TxIn {
		// The disassembled string will contain [error] inline if the script doesn't fully parse, so ignore the error here.
		disbuf, _ := txscript.DisasmString(txIn.SignatureScript)
		// Create the basic input entry without the additional optional previous output details which will be added later if requested and available.
		prevOut := &txIn.PreviousOutPoint
		vinEntry := json.VinPrevOut{
			Txid:     prevOut.Hash.String(),
			Vout:     prevOut.Index,
			Sequence: txIn.Sequence,
			ScriptSig: &json.ScriptSig{
				Asm: disbuf,
				Hex: hex.EncodeToString(txIn.SignatureScript),
			},
		}
		if len(txIn.Witness) != 0 {
			vinEntry.Witness = witnessToHex(txIn.Witness)
		}
		// Add the entry to the list now if it already passed the filter since the previous output might not be available.
		passesFilter := len(filterAddrMap) == 0
		if passesFilter {
			vinList = append(vinList, vinEntry)
		}
		// Only populate previous output information if requested and available.
		if len(originOutputs) == 0 {
			continue
		}
		originTxOut, ok := originOutputs[*prevOut]
		if !ok {
			continue
		}
		// Ignore the error here since an error means the script couldn't parse and there is no additional information about it anyways.
		_, addrs, _, _ := txscript.ExtractPkScriptAddrs(originTxOut.PkScript, chainParams)
		// Encode the addresses while checking if the address passes the filter when needed.
		encodedAddrs := make([]string, len(addrs))
		for j, addr := range addrs {
			encodedAddr := addr.EncodeAddress()
			encodedAddrs[j] = encodedAddr
			// No need to check the map again if the filter already passes.
			if passesFilter {
				continue
			}
			if _, exists := filterAddrMap[encodedAddr]; exists {
				passesFilter = true
			}
		}
		// Ignore the entry if it doesn't pass the filter.
		if !passesFilter {
			continue
		}
		// Add entry to the list if it wasn't already done above.
		if len(filterAddrMap) != 0 {
			vinList = append(vinList, vinEntry)
		}
		// Update the entry with previous output information if requested.
		if vinExtra {
			vinListEntry := &vinList[len(vinList)-1]
			vinListEntry.PrevOut = &json.PrevOut{
				Addresses: encodedAddrs,
				Value:     util.Amount(originTxOut.Value).ToDUO(),
			}
		}
	}
	return vinList, nil
}

// createVoutList returns a slice of JSON objects for the outputs of the passed transaction.
func createVoutList(
	mtx *wire.MsgTx,
	chainParams *chaincfg.Params,
	filterAddrMap map[string]struct{},
) []json.Vout {
	voutList := make([]json.Vout, 0, len(mtx.TxOut))
	for i, v := range mtx.TxOut {
		// The disassembled string will contain [error] inline if the script doesn't fully parse, so ignore the error here.
		disbuf, _ := txscript.DisasmString(v.PkScript)
		// Ignore the error here since an error means the script couldn't parse and there is no additional information about it anyways.
		scriptClass, addrs, reqSigs, _ := txscript.ExtractPkScriptAddrs(v.PkScript, chainParams)
		// Encode the addresses while checking if the address passes the filter when needed.
		passesFilter := len(filterAddrMap) == 0
		encodedAddrs := make([]string, len(addrs))
		for j, addr := range addrs {
			encodedAddr := addr.EncodeAddress()
			encodedAddrs[j] = encodedAddr
			// No need to check the map again if the filter already passes.
			if passesFilter {
				continue
			}
			if _, exists := filterAddrMap[encodedAddr]; exists {
				passesFilter = true
			}
		}
		if !passesFilter {
			continue
		}
		var vout json.Vout
		vout.N = uint32(i)
		vout.Value = util.Amount(v.Value).ToDUO()
		vout.ScriptPubKey.Addresses = encodedAddrs
		vout.ScriptPubKey.Asm = disbuf
		vout.ScriptPubKey.Hex = hex.EncodeToString(v.PkScript)
		vout.ScriptPubKey.Type = scriptClass.String()
		vout.ScriptPubKey.ReqSigs = int32(reqSigs)
		voutList = append(voutList, vout)
	}
	return voutList
}

// decodeTemplateID decodes an ID that is used to uniquely identify a block template.  This is mainly used as a mechanism to track when to update clients that are using long polling for block templates.  The ID consists of the previous block hash for the associated template and the time the associated template was generated.
func decodeTemplateID(
	templateID string,
) (
	*chainhash.Hash,
	int64,
	error,
) {
	fields := strings.Split(templateID, "-")
	if len(fields) != 2 {
		return nil, 0, errors.New("invalid longpollid format")
	}
	prevHash, err := chainhash.NewHashFromStr(fields[0])
	if err != nil {
		return nil, 0, errors.New("invalid longpollid format")
	}
	lastGenerated, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return nil, 0, errors.New("invalid longpollid format")
	}
	return prevHash, lastGenerated, nil
}

// encodeTemplateID encodes the passed details into an ID that can be used to uniquely identify a block template.
func encodeTemplateID(
	prevHash *chainhash.Hash,
	lastGenerated time.Time,
) string {
	return fmt.Sprintf("%s-%d", prevHash.String(), lastGenerated.Unix())
}

// fetchInputTxos fetches the outpoints from all transactions referenced by the inputs to the passed transaction by checking the transaction mempool first then the transaction index for those already mined into blocks.
func fetchInputTxos(
	s *rpcServer,
	tx *wire.MsgTx,
) (
	map[wire.OutPoint]wire.TxOut,
	error,
) {
	mp := s.Cfg.TxMemPool
	originOutputs := make(map[wire.OutPoint]wire.TxOut)
	for txInIndex, txIn := range tx.TxIn {
		// Attempt to fetch and use the referenced transaction from the memory pool.
		origin := &txIn.PreviousOutPoint
		originTx, err := mp.FetchTransaction(&origin.Hash)
		if err == nil {
			txOuts := originTx.MsgTx().TxOut
			if origin.Index >= uint32(len(txOuts)) {
				errStr := fmt.Sprintf("unable to find output %v referenced from transaction %s:%d", origin, tx.TxHash(), txInIndex)
				return nil, internalRPCError(errStr, "")
			}
			originOutputs[*origin] = *txOuts[origin.Index]
			continue
		}
		// Look up the location of the transaction.
		blockRegion, err := s.Cfg.TxIndex.TxBlockRegion(&origin.Hash)
		if err != nil {
			context := "Failed to retrieve transaction location"
			return nil, internalRPCError(err.Error(), context)
		}
		if blockRegion == nil {
			return nil, rpcNoTxInfoError(&origin.Hash)
		}
		// Load the raw transaction bytes from the database.
		var txBytes []byte
		err = s.Cfg.DB.View(func(dbTx database.Tx) error {
			var err error
			txBytes, err = dbTx.FetchBlockRegion(blockRegion)
			return err
		})
		if err != nil {
			return nil, rpcNoTxInfoError(&origin.Hash)
		}
		// Deserialize the transaction
		var msgTx wire.MsgTx
		err = msgTx.Deserialize(bytes.NewReader(txBytes))
		if err != nil {
			context := "Failed to deserialize transaction"
			return nil, internalRPCError(err.Error(), context)
		}
		// Add the referenced output to the map.
		if origin.Index >= uint32(len(msgTx.TxOut)) {
			errStr := fmt.Sprintf("unable to find output %v "+
				"referenced from transaction %s:%d", origin,
				tx.TxHash(), txInIndex)
			return nil, internalRPCError(errStr, "")
		}
		originOutputs[*origin] = *msgTx.TxOut[origin.Index]
	}
	return originOutputs, nil
}

// fetchMempoolTxnsForAddress queries the address index for all unconfirmed transactions that involve the provided address.  The results will be limited by the number to skip and the number requested.
func fetchMempoolTxnsForAddress(
	s *rpcServer, addr util.Address,
	numToSkip,
	numRequested uint32,
) (
	[]*util.Tx,
	uint32,
) {
	// There are no entries to return when there are less available than the number being skipped.
	mpTxns := s.Cfg.AddrIndex.UnconfirmedTxnsForAddress(addr)
	numAvailable := uint32(len(mpTxns))
	if numToSkip > numAvailable {
		return nil, numAvailable
	}
	// Filter the available entries based on the number to skip and number requested.
	rangeEnd := numToSkip + numRequested
	if rangeEnd > numAvailable {
		rangeEnd = numAvailable
	}
	return mpTxns[numToSkip:rangeEnd], numToSkip
}

// genCertPair generates a key/cert pair to the paths provided.
func genCertPair(
	certFile,
	keyFile string,
) error {
	log <- cl.Inf("generating TLS certificates...")
	org := "pod autogenerated cert"
	validUntil := time.Now().Add(10 * 365 * 24 * time.Hour)
	cert, key, err := util.NewTLSCertPair(org, validUntil, nil)
	if err != nil {
		return err
	}
	// Write cert and key files.
	if err = ioutil.WriteFile(certFile, cert, 0666); err != nil {
		return err
	}
	if err = ioutil.WriteFile(keyFile, key, 0600); err != nil {
		os.Remove(certFile)
		return err
	}
	log <- cl.Inf("Done generating TLS certificates")
	return nil
}

// getDifficultyRatio returns the proof-of-work difficulty as a multiple of the minimum difficulty using the passed bits field from the header of a block.
func getDifficultyRatio(
	bits uint32,
	params *chaincfg.Params,
	algo int32,
) float64 {
	// The minimum difficulty is the max possible proof-of-work limit bits converted back to a number.  Note this is not the same as the proof of work limit directly because the block difficulty is encoded in a block with the compact form which loses precision.
	max := blockchain.CompactToBig(0x1d00ffff)
	target := blockchain.CompactToBig(bits)
	difficulty := new(big.Rat).SetFrac(max, target)
	outString := difficulty.FloatString(8)
	diff, err := strconv.ParseFloat(outString, 64)
	if err != nil {
		log <- cl.Error{"cannot get difficulty:", err}
		return 0
	}
	return diff
}

// handleAddNode handles addnode commands.
func handleAddNode(
	s *rpcServer,
	cmd interface{},
	closeChan <-chan struct{},
) (
	interface{},
	error,
) {
	c := cmd.(*json.AddNodeCmd)
	addr := NormalizeAddress(c.Addr, s.Cfg.ChainParams.DefaultPort)
	var err error
	switch c.SubCmd {
	case "add":
		err = s.Cfg.ConnMgr.Connect(addr, true)
	case "remove":
		err = s.Cfg.ConnMgr.RemoveByAddr(addr)
	case "onetry":
		err = s.Cfg.ConnMgr.Connect(addr, false)
	default:
		return nil, &json.RPCError{
			Code:    json.ErrRPCInvalidParameter,
			Message: "invalid subcommand for addnode",
		}
	}
	if err != nil {
		return nil, &json.RPCError{
			Code:    json.ErrRPCInvalidParameter,
			Message: err.Error(),
		}
	}
	// no data returned unless an error.
	return nil, nil
}

// handleAskWallet is the handler for commands that are recognized as valid, but are unable to answer correctly since it involves wallet state. These commands will be implemented in btcwallet.
func handleAskWallet(
	s *rpcServer,
	cmd interface{},
	closeChan <-chan struct{},
) (
	interface{},
	error,
) {
	return nil, ErrRPCNoWallet
}

// handleCreateRawTransaction handles createrawtransaction commands.
func handleCreateRawTransaction(
	s *rpcServer,
	cmd interface{},
	closeChan <-chan struct{},
) (
	interface{},
	error,
) {
	c := cmd.(*json.CreateRawTransactionCmd)
	// Validate the locktime, if given.
	if c.LockTime != nil &&
		(*c.LockTime < 0 || *c.LockTime > int64(wire.MaxTxInSequenceNum)) {
		return nil, &json.RPCError{
			Code:    json.ErrRPCInvalidParameter,
			Message: "Locktime out of range",
		}
	}
	// Add all transaction inputs to a new transaction after performing some validity checks.
	mtx := wire.NewMsgTx(wire.TxVersion)
	for _, input := range c.Inputs {
		txHash, err := chainhash.NewHashFromStr(input.Txid)
		if err != nil {
			return nil, rpcDecodeHexError(input.Txid)
		}
		prevOut := wire.NewOutPoint(txHash, input.Vout)
		txIn := wire.NewTxIn(prevOut, []byte{}, nil)
		if c.LockTime != nil && *c.LockTime != 0 {
			txIn.Sequence = wire.MaxTxInSequenceNum - 1
		}
		mtx.AddTxIn(txIn)
	}
	// Add all transaction outputs to the transaction after performing some validity checks.
	params := s.Cfg.ChainParams
	for encodedAddr, amount := range c.Amounts {
		// Ensure amount is in the valid range for monetary amounts.
		if amount <= 0 || amount > util.MaxSatoshi {
			return nil, &json.RPCError{
				Code:    json.ErrRPCType,
				Message: "Invalid amount",
			}
		}
		// Decode the provided address.
		addr, err := util.DecodeAddress(encodedAddr, params)
		if err != nil {
			return nil, &json.RPCError{
				Code:    json.ErrRPCInvalidAddressOrKey,
				Message: "Invalid address or key: " + err.Error(),
			}
		}
		// Ensure the address is one of the supported types and that the network encoded with the address matches the network the server is currently on.
		switch addr.(type) {
		case *util.AddressPubKeyHash:
		case *util.AddressScriptHash:
		default:
			return nil, &json.RPCError{
				Code:    json.ErrRPCInvalidAddressOrKey,
				Message: "Invalid address or key",
			}
		}
		if !addr.IsForNet(params) {
			return nil, &json.RPCError{
				Code: json.ErrRPCInvalidAddressOrKey,
				Message: "Invalid address: " + encodedAddr +
					" is for the wrong network",
			}
		}
		// Create a new script which pays to the provided address.
		pkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			context := "Failed to generate pay-to-address script"
			return nil, internalRPCError(err.Error(), context)
		}
		// Convert the amount to satoshi.
		satoshi, err := util.NewAmount(amount)
		if err != nil {
			context := "Failed to convert amount"
			return nil, internalRPCError(err.Error(), context)
		}
		txOut := wire.NewTxOut(int64(satoshi), pkScript)
		mtx.AddTxOut(txOut)
	}
	// Set the Locktime, if given.
	if c.LockTime != nil {
		mtx.LockTime = uint32(*c.LockTime)
	}
	// Return the serialized and hex-encoded transaction.  Note that this is intentionally not directly returning because the first return value is a string and it would result in returning an empty string to the client instead of nothing (nil) in the case of an error.
	mtxHex, err := messageToHex(mtx)
	if err != nil {
		return nil, err
	}
	return mtxHex, nil
}

// handleDecodeRawTransaction handles decoderawtransaction commands.
func handleDecodeRawTransaction(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*json.DecodeRawTransactionCmd)
	// Deserialize the transaction.
	hexStr := c.HexTx
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}
	serializedTx, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, rpcDecodeHexError(hexStr)
	}
	var mtx wire.MsgTx
	err = mtx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		return nil, &json.RPCError{
			Code:    json.ErrRPCDeserialization,
			Message: "TX decode failed: " + err.Error(),
		}
	}
	// Create and return the result.
	txReply := json.TxRawDecodeResult{
		Txid:     mtx.TxHash().String(),
		Version:  mtx.Version,
		Locktime: mtx.LockTime,
		Vin:      createVinList(&mtx),
		Vout:     createVoutList(&mtx, s.Cfg.ChainParams, nil),
	}
	return txReply, nil
}

// handleDecodeScript handles decodescript commands.
func handleDecodeScript(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*json.DecodeScriptCmd)
	// Convert the hex script to bytes.
	hexStr := c.HexScript
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}
	script, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, rpcDecodeHexError(hexStr)
	}
	// The disassembled string will contain [error] inline if the script doesn't fully parse, so ignore the error here.
	disbuf, _ := txscript.DisasmString(script)
	// Get information about the script. Ignore the error here since an error means the script couldn't parse and there is no additinal information about it anyways.
	scriptClass, addrs, reqSigs, _ := txscript.ExtractPkScriptAddrs(script, s.Cfg.ChainParams)
	addresses := make([]string, len(addrs))
	for i, addr := range addrs {
		addresses[i] = addr.EncodeAddress()
	}
	// Convert the script itself to a pay-to-script-hash address.
	p2sh, err := util.NewAddressScriptHash(script, s.Cfg.ChainParams)
	if err != nil {
		context := "Failed to convert script to pay-to-script-hash"
		return nil, internalRPCError(err.Error(), context)
	}
	// Generate and return the reply.
	reply := json.DecodeScriptResult{
		Asm:       disbuf,
		ReqSigs:   int32(reqSigs),
		Type:      scriptClass.String(),
		Addresses: addresses,
	}
	if scriptClass != txscript.ScriptHashTy {
		reply.P2sh = p2sh.EncodeAddress()
	}
	return reply, nil
}

// handleEstimateFee handles estimatefee commands.
func handleEstimateFee(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*json.EstimateFeeCmd)
	if s.Cfg.FeeEstimator == nil {
		return nil, errors.New("Fee estimation disabled")
	}
	if c.NumBlocks <= 0 {
		return -1.0, errors.New("Parameter NumBlocks must be positive")
	}
	feeRate, err := s.Cfg.FeeEstimator.EstimateFee(uint32(c.NumBlocks))
	if err != nil {
		return -1.0, err
	}
	// Convert to satoshis per kb.
	return float64(feeRate), nil
}

// handleGenerate handles generate commands.
func handleGenerate(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// Respond with an error if there are no addresses to pay the created blocks to.
	if len(StateCfg.ActiveMiningAddrs) == 0 {
		return nil, &json.RPCError{
			Code:    json.ErrRPCInternal.Code,
			Message: "No payment addresses specified via --miningaddr",
		}
	}
	// Respond with an error if there's virtually 0 chance of mining a block with the CPU.
	if !s.Cfg.ChainParams.GenerateSupported {
		return nil, &json.RPCError{
			Code:    json.ErrRPCDifficulty,
			Message: fmt.Sprintf("No support for `generate` on the current network, %s, as it's unlikely to be possible to mine a block with the CPU.", s.Cfg.ChainParams.Net),
		}
	}
	// Set the algorithm according to the port we were called on
	s.Cfg.CPUMiner.SetAlgo(s.Cfg.Algo)
	c := cmd.(*json.GenerateCmd)
	// Respond with an error if the client is requesting 0 blocks to be generated.
	if c.NumBlocks == 0 {
		return nil, &json.RPCError{
			Code:    json.ErrRPCInternal.Code,
			Message: "Please request a nonzero number of blocks to generate.",
		}
	}
	// Create a reply
	reply := make([]string, c.NumBlocks)
	fmt.Println(s.Cfg.Algo)
	blockHashes, err := s.Cfg.CPUMiner.GenerateNBlocks(c.NumBlocks, s.Cfg.Algo)
	if err != nil {
		return nil, &json.RPCError{
			Code:    json.ErrRPCInternal.Code,
			Message: err.Error(),
		}
	}
	// Mine the correct number of blocks, assigning the hex representation of the hash of each one to its place in the reply.
	for i, hash := range blockHashes {
		reply[i] = hash.String()
	}
	return reply, nil
}

// handleGetAddedNodeInfo handles getaddednodeinfo commands.
func handleGetAddedNodeInfo(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*json.GetAddedNodeInfoCmd)
	// Retrieve a list of persistent (added) peers from the server and filter the list of peers per the specified address (if any).
	peers := s.Cfg.ConnMgr.PersistentPeers()
	if c.Node != nil {
		node := *c.Node
		found := false
		for i, peer := range peers {
			if peer.ToPeer().Addr() == node {
				peers = peers[i : i+1]
				found = true
			}
		}
		if !found {
			return nil, &json.RPCError{
				Code:    json.ErrRPCClientNodeNotAdded,
				Message: "Node has not been added",
			}
		}
	}
	// Without the dns flag, the result is just a slice of the addresses as strings.
	if !c.DNS {
		results := make([]string, 0, len(peers))
		for _, peer := range peers {
			results = append(results, peer.ToPeer().Addr())
		}
		return results, nil
	}
	// With the dns flag, the result is an array of JSON objects which include the result of DNS lookups for each peer.
	results := make([]*json.GetAddedNodeInfoResult, 0, len(peers))
	for _, rpcPeer := range peers {
		// Set the "address" of the peer which could be an ip address or a domain name.
		peer := rpcPeer.ToPeer()
		var result json.GetAddedNodeInfoResult
		result.AddedNode = peer.Addr()
		result.Connected = json.Bool(peer.Connected())
		// Split the address into host and port portions so we can do a DNS lookup against the host.  When no port is specified in the address, just use the address as the host.
		host, _, err := net.SplitHostPort(peer.Addr())
		if err != nil {
			host = peer.Addr()
		}
		var ipList []string
		switch {
		case net.ParseIP(host) != nil, strings.HasSuffix(host, ".onion"):
			ipList = make([]string, 1)
			ipList[0] = host
		default:
			// Do a DNS lookup for the address.  If the lookup fails, just use the host.
			ips, err := podLookup(host)
			if err != nil {
				ipList = make([]string, 1)
				ipList[0] = host
				break
			}
			ipList = make([]string, 0, len(ips))
			for _, ip := range ips {
				ipList = append(ipList, ip.String())
			}
		}
		// Add the addresses and connection info to the result.
		addrs := make([]json.GetAddedNodeInfoResultAddr, 0, len(ipList))
		for _, ip := range ipList {
			var addr json.GetAddedNodeInfoResultAddr
			addr.Address = ip
			addr.Connected = "false"
			if ip == host && peer.Connected() {
				addr.Connected = directionString(peer.Inbound())
			}
			addrs = append(addrs, addr)
		}
		result.Addresses = &addrs
		results = append(results, &result)
	}
	return results, nil
}

// handleGetBestBlock implements the getbestblock command.
func handleGetBestBlock(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// All other "get block" commands give either the height, the hash, or both but require the block SHA.  This gets both for the best block.
	best := s.Cfg.Chain.BestSnapshot()
	result := &json.GetBestBlockResult{
		Hash:   best.Hash.String(),
		Height: best.Height,
	}
	return result, nil
}

// handleGetBestBlockHash implements the getbestblockhash command.
func handleGetBestBlockHash(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	best := s.Cfg.Chain.BestSnapshot()
	return best.Hash.String(), nil
}

// handleGetBlock implements the getblock command.
func handleGetBlock(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*json.GetBlockCmd)
	// Load the raw block bytes from the database.
	hash, err := chainhash.NewHashFromStr(c.Hash)
	if err != nil {
		return nil, rpcDecodeHexError(c.Hash)
	}
	var blkBytes []byte
	err = s.Cfg.DB.View(func(dbTx database.Tx) error {
		var err error
		blkBytes, err = dbTx.FetchBlock(hash)
		return err
	})
	if err != nil {
		return nil, &json.RPCError{
			Code:    json.ErrRPCBlockNotFound,
			Message: "Block not found",
		}
	}
	// When the verbose flag isn't set, simply return the serialized block as a hex-encoded string.
	if c.Verbose != nil && !*c.Verbose {
		return hex.EncodeToString(blkBytes), nil
	}
	// The verbose flag is set, so generate the JSON object and return it. Deserialize the block.
	blk, err := util.NewBlockFromBytes(blkBytes)
	if err != nil {
		context := "Failed to deserialize block"
		return nil, internalRPCError(err.Error(), context)
	}
	// Get the block height from chain.
	blockHeight, err := s.Cfg.Chain.BlockHeightByHash(hash)
	if err != nil {
		context := "Failed to obtain block height"
		return nil, internalRPCError(err.Error(), context)
	}
	blk.SetHeight(blockHeight)
	best := s.Cfg.Chain.BestSnapshot()
	// Get next block hash unless there are none.
	var nextHashString string
	if blockHeight < best.Height {
		nextHash, err := s.Cfg.Chain.BlockHashByHeight(blockHeight + 1)
		if err != nil {
			context := "No next block"
			return nil, internalRPCError(err.Error(), context)
		}
		nextHashString = nextHash.String()
	}
	params := s.Cfg.ChainParams
	blockHeader := &blk.MsgBlock().Header
	algoname := fork.GetAlgoName(blockHeader.Version, blockHeight)
	a := fork.GetAlgoVer(algoname, blockHeight)
	algoid := fork.GetAlgoID(algoname, blockHeight)
	blockReply := json.GetBlockVerboseResult{
		Hash:          c.Hash,
		Version:       blockHeader.Version,
		VersionHex:    fmt.Sprintf("%08x", blockHeader.Version),
		PowAlgoID:     algoid,
		PowAlgo:       algoname,
		PowHash:       blk.MsgBlock().BlockHashWithAlgos(blockHeight).String(),
		MerkleRoot:    blockHeader.MerkleRoot.String(),
		PreviousHash:  blockHeader.PrevBlock.String(),
		Nonce:         blockHeader.Nonce,
		Time:          blockHeader.Timestamp.Unix(),
		Confirmations: int64(1 + best.Height - blockHeight),
		Height:        int64(blockHeight),
		Size:          int32(len(blkBytes)),
		StrippedSize:  int32(blk.MsgBlock().SerializeSizeStripped()),
		Weight:        int32(blockchain.GetBlockWeight(blk)),
		Bits:          strconv.FormatInt(int64(blockHeader.Bits), 16),
		Difficulty:    getDifficultyRatio(blockHeader.Bits, params, a),
		NextHash:      nextHashString,
	}
	if c.VerboseTx == nil || !*c.VerboseTx {
		transactions := blk.Transactions()
		txNames := make([]string, len(transactions))
		for i, tx := range transactions {
			txNames[i] = tx.Hash().String()
		}
		blockReply.Tx = txNames
	} else {
		txns := blk.Transactions()
		rawTxns := make([]json.TxRawResult, len(txns))
		for i, tx := range txns {
			rawTxn, err := createTxRawResult(params, tx.MsgTx(),
				tx.Hash().String(), blockHeader, hash.String(),
				blockHeight, best.Height)
			if err != nil {
				return nil, err
			}
			rawTxns[i] = *rawTxn
		}
		blockReply.RawTx = rawTxns
	}
	return blockReply, nil
}

// handleGetBlockChainInfo implements the getblockchaininfo command.
func handleGetBlockChainInfo(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// Obtain a snapshot of the current best known blockchain state. We'll populate the response to this call primarily from this snapshot.
	params := s.Cfg.ChainParams
	chain := s.Cfg.Chain
	chainSnapshot := chain.BestSnapshot()
	chainInfo := &json.GetBlockChainInfoResult{
		Chain:         params.Name,
		Blocks:        chainSnapshot.Height,
		Headers:       chainSnapshot.Height,
		BestBlockHash: chainSnapshot.Hash.String(),
		Difficulty:    getDifficultyRatio(chainSnapshot.Bits, params, 2),
		MedianTime:    chainSnapshot.MedianTime.Unix(),
		Pruned:        false,
		Bip9SoftForks: make(map[string]*json.Bip9SoftForkDescription),
	}
	// Next, populate the response with information describing the current status of soft-forks deployed via the super-majority block signalling mechanism.
	height := chainSnapshot.Height
	chainInfo.SoftForks = []*json.SoftForkDescription{
		{
			ID:      "bip34",
			Version: 2,
			Reject: struct {
				Status bool `json:"status"`
			}{
				Status: height >= params.BIP0034Height,
			},
		},
		{
			ID:      "bip66",
			Version: 3,
			Reject: struct {
				Status bool `json:"status"`
			}{
				Status: height >= params.BIP0066Height,
			},
		},
		{
			ID:      "bip65",
			Version: 4,
			Reject: struct {
				Status bool `json:"status"`
			}{
				Status: height >= params.BIP0065Height,
			},
		},
	}
	// Finally, query the BIP0009 version bits state for all currently defined BIP0009 soft-fork deployments.
	for deployment, deploymentDetails := range params.Deployments {
		// Map the integer deployment ID into a human readable fork-name.
		var forkName string
		switch deployment {
		case chaincfg.DeploymentTestDummy:
			forkName = "dummy"
		case chaincfg.DeploymentCSV:
			forkName = "csv"
		case chaincfg.DeploymentSegwit:
			forkName = "segwit"
		default:
			return nil, &json.RPCError{
				Code: json.ErrRPCInternal.Code,
				Message: fmt.Sprintf("Unknown deployment %v "+
					"detected", deployment),
			}
		}
		// Query the chain for the current status of the deployment as identified by its deployment ID.
		deploymentStatus, err := chain.ThresholdState(uint32(deployment))
		if err != nil {
			context := "Failed to obtain deployment status"
			return nil, internalRPCError(err.Error(), context)
		}
		// Attempt to convert the current deployment status into a human readable string. If the status is unrecognized, then a non-nil error is returned.
		statusString, err := softForkStatus(deploymentStatus)
		if err != nil {
			return nil, &json.RPCError{
				Code: json.ErrRPCInternal.Code,
				Message: fmt.Sprintf("unknown deployment status: %v",
					deploymentStatus),
			}
		}
		// Finally, populate the soft-fork description with all the information gathered above.
		chainInfo.Bip9SoftForks[forkName] = &json.Bip9SoftForkDescription{
			Status:    strings.ToLower(statusString),
			Bit:       deploymentDetails.BitNumber,
			StartTime: int64(deploymentDetails.StartTime),
			Timeout:   int64(deploymentDetails.ExpireTime),
		}
	}
	return chainInfo, nil
}

// handleGetBlockCount implements the getblockcount command.
func handleGetBlockCount(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	best := s.Cfg.Chain.BestSnapshot()
	return int64(best.Height), nil
}

// handleGetBlockHash implements the getblockhash command.
func handleGetBlockHash(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*json.GetBlockHashCmd)
	hash, err := s.Cfg.Chain.BlockHashByHeight(int32(c.Index))
	if err != nil {
		return nil, &json.RPCError{
			Code:    json.ErrRPCOutOfRange,
			Message: "Block number out of range",
		}
	}
	return hash.String(), nil
}

// handleGetBlockHeader implements the getblockheader command.
func handleGetBlockHeader(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*json.GetBlockHeaderCmd)
	// Fetch the header from chain.
	hash, err := chainhash.NewHashFromStr(c.Hash)
	if err != nil {
		return nil, rpcDecodeHexError(c.Hash)
	}
	blockHeader, err := s.Cfg.Chain.HeaderByHash(hash)
	if err != nil {
		return nil, &json.RPCError{
			Code:    json.ErrRPCBlockNotFound,
			Message: "Block not found",
		}
	}
	// When the verbose flag isn't set, simply return the serialized block header as a hex-encoded string.
	if c.Verbose != nil && !*c.Verbose {
		var headerBuf bytes.Buffer
		err := blockHeader.Serialize(&headerBuf)
		if err != nil {
			context := "Failed to serialize block header"
			return nil, internalRPCError(err.Error(), context)
		}
		return hex.EncodeToString(headerBuf.Bytes()), nil
	}
	// The verbose flag is set, so generate the JSON object and return it. Get the block height from chain.
	blockHeight, err := s.Cfg.Chain.BlockHeightByHash(hash)
	if err != nil {
		context := "Failed to obtain block height"
		return nil, internalRPCError(err.Error(), context)
	}
	best := s.Cfg.Chain.BestSnapshot()
	// Get next block hash unless there are none.
	var nextHashString string
	if blockHeight < best.Height {
		nextHash, err := s.Cfg.Chain.BlockHashByHeight(blockHeight + 1)
		if err != nil {
			context := "No next block"
			return nil, internalRPCError(err.Error(), context)
		}
		nextHashString = nextHash.String()
	}
	var a int32 = 2
	if blockHeader.Version == 514 {
		a = 514
	}
	params := s.Cfg.ChainParams
	blockHeaderReply := json.GetBlockHeaderVerboseResult{
		Hash:          c.Hash,
		Confirmations: int64(1 + best.Height - blockHeight),
		Height:        blockHeight,
		Version:       blockHeader.Version,
		VersionHex:    fmt.Sprintf("%08x", blockHeader.Version),
		MerkleRoot:    blockHeader.MerkleRoot.String(),
		NextHash:      nextHashString,
		PreviousHash:  blockHeader.PrevBlock.String(),
		Nonce:         uint64(blockHeader.Nonce),
		Time:          blockHeader.Timestamp.Unix(),
		Bits:          strconv.FormatInt(int64(blockHeader.Bits), 16),
		Difficulty:    getDifficultyRatio(blockHeader.Bits, params, a),
	}
	return blockHeaderReply, nil
}

// handleGetBlockTemplate implements the getblocktemplate command. See https://en.bitcoin.it/wiki/BIP_0022 and https://en.bitcoin.it/wiki/BIP_0023 for more details.
func handleGetBlockTemplate(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*json.GetBlockTemplateCmd)
	request := c.Request
	// Set the default mode and override it if supplied.
	mode := "template"
	if request != nil && request.Mode != "" {
		mode = request.Mode
	}
	switch mode {
	case "template":
		return handleGetBlockTemplateRequest(s, request, closeChan)
	case "proposal":
		return handleGetBlockTemplateProposal(s, request)
	}
	return nil, &json.RPCError{
		Code:    json.ErrRPCInvalidParameter,
		Message: "Invalid mode",
	}
}

// handleGetBlockTemplateLongPoll is a helper for handleGetBlockTemplateRequest which deals with handling long polling for block templates.  When a caller sends a request with a long poll ID that was previously returned, a response is not sent until the caller should stop working on the previous block template in favor of the new one.  In particular, this is the case when the old block template is no longer valid due to a solution already being found and added to the block chain, or new transactions have shown up and some time has passed without finding a solution. See https://en.bitcoin.it/wiki/BIP_0022 for more details.
func handleGetBlockTemplateLongPoll(
	s *rpcServer, longPollID string, useCoinbaseValue bool, closeChan <-chan struct{}) (interface{}, error) {
	state := s.gbtWorkState
	state.Lock()
	// The state unlock is intentionally not deferred here since it needs to be manually unlocked before waiting for a notification about block template changes.
	if err := state.updateBlockTemplate(s, useCoinbaseValue); err != nil {
		state.Unlock()
		return nil, err
	}
	// Just return the current block template if the long poll ID provided by the caller is invalid.
	prevHash, lastGenerated, err := decodeTemplateID(longPollID)
	if err != nil {
		result, err := state.blockTemplateResult(useCoinbaseValue, nil)
		if err != nil {
			state.Unlock()
			return nil, err
		}
		state.Unlock()
		return result, nil
	}
	// Return the block template now if the specific block template/ identified by the long poll ID no longer matches the current block template as this means the provided template is stale.
	prevTemplateHash := &state.template.Block.Header.PrevBlock
	if !prevHash.IsEqual(prevTemplateHash) ||
		lastGenerated != state.lastGenerated.Unix() {
		// Include whether or not it is valid to submit work against the old block template depending on whether or not a solution has already been found and added to the block chain.
		submitOld := prevHash.IsEqual(prevTemplateHash)
		result, err := state.blockTemplateResult(useCoinbaseValue,
			&submitOld)
		if err != nil {
			state.Unlock()
			return nil, err
		}
		state.Unlock()
		return result, nil
	}
	// Register the previous hash and last generated time for notifications Get a channel that will be notified when the template associated with the provided ID is stale and a new block template should be returned to the caller.
	longPollChan := state.templateUpdateChan(prevHash, lastGenerated)
	state.Unlock()
	select {
	// When the client closes before it's time to send a reply, just return now so the goroutine doesn't hang around.
	case <-closeChan:
		// fmt.Println("chan:<-closeChan")
		return nil, ErrClientQuit
	// Wait until signal received to send the reply.
	case <-longPollChan:
		// fmt.Println("chan:<-longPollChan")
		// Fallthrough
	}
	// Get the lastest block template
	state.Lock()
	defer state.Unlock()
	if err := state.updateBlockTemplate(s, useCoinbaseValue); err != nil {
		return nil, err
	}
	// Include whether or not it is valid to submit work against the old block template depending on whether or not a solution has already been found and added to the block chain.
	submitOld := prevHash.IsEqual(&state.template.Block.Header.PrevBlock)
	result, err := state.blockTemplateResult(useCoinbaseValue, &submitOld)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// handleGetBlockTemplateProposal is a helper for handleGetBlockTemplate which deals with block proposals. See https://en.bitcoin.it/wiki/BIP_0023 for more details.
func handleGetBlockTemplateProposal(
	s *rpcServer, request *json.TemplateRequest) (interface{}, error) {
	hexData := request.Data
	if hexData == "" {
		return false, &json.RPCError{
			Code: json.ErrRPCType,
			Message: fmt.Sprintf("Data must contain the " +
				"hex-encoded serialized block that is being " +
				"proposed"),
		}
	}
	// Ensure the provided data is sane and deserialize the proposed block.
	if len(hexData)%2 != 0 {
		hexData = "0" + hexData
	}
	dataBytes, err := hex.DecodeString(hexData)
	if err != nil {
		return false, &json.RPCError{
			Code:    json.ErrRPCDeserialization,
			Message: fmt.Sprintf("data must be hexadecimal string (not %q)", hexData),
		}
	}
	var msgBlock wire.MsgBlock
	if err := msgBlock.Deserialize(bytes.NewReader(dataBytes)); err != nil {
		return nil, &json.RPCError{
			Code:    json.ErrRPCDeserialization,
			Message: "block decode failed: " + err.Error(),
		}
	}
	block := util.NewBlock(&msgBlock)
	// Ensure the block is building from the expected previous block.
	expectedPrevHash := s.Cfg.Chain.BestSnapshot().Hash
	prevHash := &block.MsgBlock().Header.PrevBlock
	if !expectedPrevHash.IsEqual(prevHash) {
		return "bad-prevblk", nil
	}
	if err := s.Cfg.Chain.CheckConnectBlockTemplate(block); err != nil {
		if _, ok := err.(blockchain.RuleError); !ok {
			errStr := fmt.Sprintf("failed to process block proposal: %v", err)
			log <- cl.Err(errStr)
			return nil, &json.RPCError{
				Code:    json.ErrRPCVerify,
				Message: errStr,
			}
		}
		log <- cl.Info{"rejected block proposal:", err}
		return chainErrToGBTErrString(err), nil
	}
	return nil, nil
}

// handleGetBlockTemplateRequest is a helper for handleGetBlockTemplate which deals with generating and returning block templates to the caller.  It handles both long poll requests as specified by BIP 0022 as well as regular requests.  In addition, it detects the capabilities reported by the caller in regards to whether or not it supports creating its own coinbase (the coinbasetxn and coinbasevalue capabilities) and modifies the returned block template accordingly.
func handleGetBlockTemplateRequest(
	s *rpcServer, request *json.TemplateRequest, closeChan <-chan struct{}) (interface{}, error) {
	// Extract the relevant passed capabilities and restrict the result to either a coinbase value or a coinbase transaction object depending on the request.  Default to only providing a coinbase value.
	useCoinbaseValue := true
	if request != nil {
		var hasCoinbaseValue, hasCoinbaseTxn bool
		for _, capability := range request.Capabilities {
			switch capability {
			case "coinbasetxn":
				hasCoinbaseTxn = true
			case "coinbasevalue":
				hasCoinbaseValue = true
			}
		}
		if hasCoinbaseTxn && !hasCoinbaseValue {
			useCoinbaseValue = false
		}
	}
	// When a coinbase transaction has been requested, respond with an error if there are no addresses to pay the created block template to.
	if !useCoinbaseValue && len(StateCfg.ActiveMiningAddrs) == 0 {
		return nil, &json.RPCError{
			Code: json.ErrRPCInternal.Code,
			Message: "A coinbase transaction has been requested, " +
				"but the server has not been configured with " +
				"any payment addresses via --miningaddr",
		}
	}
	// Return an error if there are no peers connected since there is no way to relay a found block or receive transactions to work on. However, allow this state when running in the regression test or simulation test mode.
	if !(*Cfg.RegressionTest || *Cfg.SimNet) &&
		s.Cfg.ConnMgr.ConnectedCount() == 0 {
		return nil, &json.RPCError{
			Code:    json.ErrRPCClientNotConnected,
			Message: "Pod is not connected to network",
		}
	}
	// No point in generating or accepting work before the chain is synced.
	currentHeight := s.Cfg.Chain.BestSnapshot().Height
	if currentHeight != 0 && !s.Cfg.SyncMgr.IsCurrent() {
		return nil, &json.RPCError{
			Code:    json.ErrRPCClientInInitialDownload,
			Message: "Pod is not yet synchronised...",
		}
	}
	// When a long poll ID was provided, this is a long poll request by the client to be notified when block template referenced by the ID should be replaced with a new one.
	if request != nil && request.LongPollID != "" {
		return handleGetBlockTemplateLongPoll(s, request.LongPollID,
			useCoinbaseValue, closeChan)
	}
	// Protect concurrent access when updating block templates.
	state := s.gbtWorkState
	state.Lock()
	defer state.Unlock()
	// Get and return a block template.  A new block template will be generated when the current best block has changed or the transactions in the memory pool have been updated and it has been at least five seconds since the last template was generated.  Otherwise, the timestamp for the existing block template is updated (and possibly the difficulty on testnet per the consesus rules).
	if err := state.updateBlockTemplate(s, useCoinbaseValue); err != nil {
		return nil, err
	}
	return state.blockTemplateResult(useCoinbaseValue, nil)
}

// handleGetCFilter implements the getcfilter command.
func handleGetCFilter(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if s.Cfg.CfIndex == nil {
		return nil, &json.RPCError{
			Code:    json.ErrRPCNoCFIndex,
			Message: "The CF index must be enabled for this command",
		}
	}
	c := cmd.(*json.GetCFilterCmd)
	hash, err := chainhash.NewHashFromStr(c.Hash)
	if err != nil {
		return nil, rpcDecodeHexError(c.Hash)
	}
	filterBytes, err := s.Cfg.CfIndex.FilterByBlockHash(hash, c.FilterType)
	if err != nil {
		log <- cl.Debugf{
			"could not find committed filter for %v: %v",
			hash,
			err,
		}
		return nil, &json.RPCError{
			Code:    json.ErrRPCBlockNotFound,
			Message: "block not found",
		}
	}
	log <- cl.Debug{"found committed filter for", hash}
	return hex.EncodeToString(filterBytes), nil
}

// handleGetCFilterHeader implements the getcfilterheader command.
func handleGetCFilterHeader(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	if s.Cfg.CfIndex == nil {
		return nil, &json.RPCError{
			Code:    json.ErrRPCNoCFIndex,
			Message: "The CF index must be enabled for this command",
		}
	}
	c := cmd.(*json.GetCFilterHeaderCmd)
	hash, err := chainhash.NewHashFromStr(c.Hash)
	if err != nil {
		return nil, rpcDecodeHexError(c.Hash)
	}
	headerBytes, err := s.Cfg.CfIndex.FilterHeaderByBlockHash(hash, c.FilterType)
	if len(headerBytes) > 0 {
		log <- cl.Debug{"found header of committed filter for", hash}
	} else {
		log <- cl.Debugf{
			"could not find header of committed filter for %v: %v",
			hash,
			err,
		}
		return nil, &json.RPCError{
			Code:    json.ErrRPCBlockNotFound,
			Message: "Block not found",
		}
	}
	hash.SetBytes(headerBytes)
	return hash.String(), nil
}

// handleGetConnectionCount implements the getconnectioncount command.
func handleGetConnectionCount(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return s.Cfg.ConnMgr.ConnectedCount(), nil
}

// handleGetCurrentNet implements the getcurrentnet command.
func handleGetCurrentNet(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return s.Cfg.ChainParams.Net, nil
}

// handleGetDifficulty implements the getdifficulty command. TODO: This command should default to the configured algo for cpu mining and take an optional parameter to query by algo
func handleGetDifficulty(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*json.GetDifficultyCmd)
	best := s.Cfg.Chain.BestSnapshot()
	prev, err := s.Cfg.Chain.BlockByHash(&best.Hash)
	if err != nil {
		fmt.Println("ERROR", err)
	}
	var algo = prev.MsgBlock().Header.Version
	if algo != 514 {
		algo = 2
	}
	bestbits := best.Bits
	if c.Algo == "scrypt" && algo != 514 {
		algo = 514
		for {
			if prev.MsgBlock().Header.Version != 514 {
				ph := prev.MsgBlock().Header.PrevBlock
				prev, err = s.Cfg.Chain.BlockByHash(&ph)
				if err != nil {
					fmt.Println("ERROR", err)
				}
				continue
			}
			bestbits = uint32(prev.MsgBlock().Header.Bits)
			break
		}
	}
	if c.Algo == "sha256d" && algo != 2 {
		algo = 2
		for {
			if prev.MsgBlock().Header.Version == 514 {
				ph := prev.MsgBlock().Header.PrevBlock
				prev, err = s.Cfg.Chain.BlockByHash(&ph)
				if err != nil {
					fmt.Println("ERROR", err)
				}
				continue
			}
			bestbits = uint32(prev.MsgBlock().Header.Bits)
			break
		}
	}
	return getDifficultyRatio(bestbits, s.Cfg.ChainParams, algo), nil
}

// handleGetGenerate implements the getgenerate command.
func handleGetGenerate(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return s.Cfg.CPUMiner.IsMining(), nil
}

// handleGetHashesPerSec implements the gethashespersec command.
func handleGetHashesPerSec(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return int64(s.Cfg.CPUMiner.HashesPerSecond()), nil
}

// handleGetHeaders implements the getheaders command. NOTE: This is a btcsuite extension originally ported from github.com/decred/dcrd.
func handleGetHeaders(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*json.GetHeadersCmd)
	// Fetch the requested headers from chain while respecting the provided block locators and stop hash.
	blockLocators := make([]*chainhash.Hash, len(c.BlockLocators))
	for i := range c.BlockLocators {
		blockLocator, err := chainhash.NewHashFromStr(c.BlockLocators[i])
		if err != nil {
			return nil, rpcDecodeHexError(c.BlockLocators[i])
		}
		blockLocators[i] = blockLocator
	}
	var hashStop chainhash.Hash
	if c.HashStop != "" {
		err := chainhash.Decode(&hashStop, c.HashStop)
		if err != nil {
			return nil, rpcDecodeHexError(c.HashStop)
		}
	}
	headers := s.Cfg.SyncMgr.LocateHeaders(blockLocators, &hashStop)
	// Return the serialized block headers as hex-encoded strings.
	hexBlockHeaders := make([]string, len(headers))
	var buf bytes.Buffer
	for i, h := range headers {
		err := h.Serialize(&buf)
		if err != nil {
			return nil, internalRPCError(err.Error(), "Failed to serialize block header")
		}
		hexBlockHeaders[i] = hex.EncodeToString(buf.Bytes())
		buf.Reset()
	}
	return hexBlockHeaders, nil
}

// handleGetInfo implements the getinfo command. We only return the fields that are not related to wallet functionality.
func handleGetInfo(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (ret interface{}, err error) {
	var Difficulty, dBlake2b, dBlake14lr, dBlake2s, dKeccak, dScrypt, dSHA256D, dSkein, dStribog, dX11 float64
	var lastbitsBlake2b, lastbitsBlake14lr, lastbitsBlake2s, lastbitsKeccak, lastbitsScrypt, lastbitsSHA256D, lastbitsSkein, lastbitsStribog, lastbitsX11 uint32
	best := s.Cfg.Chain.BestSnapshot()
	v := s.Cfg.Chain.Index.LookupNode(&best.Hash)
	foundcount, height := 0, best.Height
	switch fork.GetCurrent(height) {
	case 0:
		for foundcount < 9 && height > 0 {
			switch fork.GetAlgoName(v.Header().Version, height) {
			case "sha256d":
				if lastbitsSHA256D == 0 {
					foundcount++
					lastbitsSHA256D = v.Header().Bits
					dSHA256D = getDifficultyRatio(lastbitsSHA256D, s.Cfg.ChainParams, v.Header().Version)
				}
			case "scrypt":
				if lastbitsScrypt == 0 {
					foundcount++
					lastbitsScrypt = v.Header().Bits
					dScrypt = getDifficultyRatio(lastbitsScrypt, s.Cfg.ChainParams, v.Header().Version)
				}
			default:
			}
			v = v.RelativeAncestor(1)
			height--
		}
		switch s.Cfg.Algo {
		case "sha256d":
			Difficulty = dSHA256D
		case "scrypt":
			Difficulty = dScrypt
		default:
		}
		proxy := ""
		if Cfg.Proxy != nil {
			proxy = *Cfg.Proxy
		}
		ret = &json.InfoChainResult0{
			Version:           int32(1000000*appMajor + 10000*appMinor + 100*appPatch),
			ProtocolVersion:   int32(maxProtocolVersion),
			Blocks:            best.Height,
			TimeOffset:        int64(s.Cfg.TimeSource.Offset().Seconds()),
			Connections:       s.Cfg.ConnMgr.ConnectedCount(),
			Proxy:             proxy,
			PowAlgoID:         fork.GetAlgoID(s.Cfg.Algo, height),
			PowAlgo:           s.Cfg.Algo,
			Difficulty:        Difficulty,
			DifficultySHA256D: dSHA256D,
			DifficultyScrypt:  dScrypt,
			TestNet:           *Cfg.TestNet3,
			RelayFee:          StateCfg.ActiveMinRelayTxFee.ToDUO(),
		}
	case 1:
		foundcount, height := 0, best.Height
		for foundcount < 9 &&
			height > fork.List[fork.GetCurrent(height)].ActivationHeight-512 {
			switch fork.GetAlgoName(v.Header().Version, height) {
			case "blake2b":
				if lastbitsBlake2b == 0 {
					foundcount++
					lastbitsBlake2b = v.Header().Bits
					dBlake2b = getDifficultyRatio(lastbitsBlake2b, s.Cfg.ChainParams, v.Header().Version)
				}
			case "blake14lr":
				if lastbitsBlake14lr == 0 {
					foundcount++
					lastbitsBlake14lr = v.Header().Bits
					dBlake14lr = getDifficultyRatio(lastbitsBlake14lr, s.Cfg.ChainParams, v.Header().Version)
				}
			case "blake2s":
				if lastbitsBlake2s == 0 {
					foundcount++
					lastbitsBlake2s = v.Header().Bits
					dBlake2s = getDifficultyRatio(lastbitsBlake2s, s.Cfg.ChainParams, v.Header().Version)
				}
			case "keccak":
				if lastbitsKeccak == 0 {
					foundcount++
					lastbitsKeccak = v.Header().Bits
					dKeccak = getDifficultyRatio(lastbitsKeccak, s.Cfg.ChainParams, v.Header().Version)
				}
			case "scrypt":
				if lastbitsScrypt == 0 {
					foundcount++
					lastbitsScrypt = v.Header().Bits
					dScrypt = getDifficultyRatio(lastbitsScrypt, s.Cfg.ChainParams, v.Header().Version)
				}
			case "sha256d":
				if lastbitsSHA256D == 0 {
					foundcount++
					lastbitsSHA256D = v.Header().Bits
					dSHA256D = getDifficultyRatio(lastbitsSHA256D, s.Cfg.ChainParams, v.Header().Version)
				}
			case "skein":
				if lastbitsSkein == 0 {
					foundcount++
					lastbitsSkein = v.Header().Bits
					dSkein = getDifficultyRatio(lastbitsSkein, s.Cfg.ChainParams, v.Header().Version)
				}
			case "stribog":
				if lastbitsStribog == 0 {
					foundcount++
					lastbitsStribog = v.Header().Bits
					dStribog = getDifficultyRatio(lastbitsStribog, s.Cfg.ChainParams, v.Header().Version)
				}
			case "x11":
				if lastbitsX11 == 0 {
					foundcount++
					lastbitsX11 = v.Header().Bits
					dX11 = getDifficultyRatio(lastbitsX11, s.Cfg.ChainParams, v.Header().Version)
				}
			default:
			}
			v = v.RelativeAncestor(1)
			height--
		}
		switch s.Cfg.Algo {
		case "blake2b":
			Difficulty = dBlake2b
		case "blake14lr":
			Difficulty = dBlake14lr
		case "blake2s":
			Difficulty = dBlake2s
		case "keccak":
			Difficulty = dKeccak
		case "scrypt":
			Difficulty = dScrypt
		case "sha256d":
			Difficulty = dSHA256D
		case "skein":
			Difficulty = dSkein
		case "stribog":
			Difficulty = dStribog
		case "x11":
			Difficulty = dX11
		default:
		}
		ret = &json.InfoChainResult{
			Version:             int32(1000000*appMajor + 10000*appMinor + 100*appPatch),
			ProtocolVersion:     int32(maxProtocolVersion),
			Blocks:              best.Height,
			TimeOffset:          int64(s.Cfg.TimeSource.Offset().Seconds()),
			Connections:         s.Cfg.ConnMgr.ConnectedCount(),
			Proxy:               *Cfg.Proxy,
			PowAlgoID:           fork.GetAlgoID(s.Cfg.Algo, height),
			PowAlgo:             s.Cfg.Algo,
			Difficulty:          Difficulty,
			DifficultyBlake2b:   dBlake2b,
			DifficultyBlake14lr: dBlake14lr,
			DifficultyBlake2s:   dBlake2s,
			DifficultyKeccak:    dKeccak,
			DifficultyScrypt:    dScrypt,
			DifficultySHA256D:   dSHA256D,
			DifficultySkein:     dSkein,
			DifficultyStribog:   dStribog,
			DifficultyX11:       dX11,
			TestNet:             *Cfg.TestNet3,
			RelayFee:            StateCfg.ActiveMinRelayTxFee.ToDUO(),
		}
	}
	return ret, nil
}

// handleGetMempoolInfo implements the getmempoolinfo command.
func handleGetMempoolInfo(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	mempoolTxns := s.Cfg.TxMemPool.TxDescs()
	var numBytes int64
	for _, txD := range mempoolTxns {
		numBytes += int64(txD.Tx.MsgTx().SerializeSize())
	}
	ret := &json.GetMempoolInfoResult{
		Size:  int64(len(mempoolTxns)),
		Bytes: numBytes,
	}
	return ret, nil
}

// handleGetMiningInfo implements the getmininginfo command. We only return the fields that are not related to wallet functionality. This function returns more information than parallelcoind.
func handleGetMiningInfo(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (ret interface{}, err error) {
	// Create a default getnetworkhashps command to use defaults and make use of the existing getnetworkhashps handler.
	gnhpsCmd := json.NewGetNetworkHashPSCmd(nil, nil)
	networkHashesPerSecIface, err := handleGetNetworkHashPS(s, gnhpsCmd, closeChan)
	if err != nil {
		return nil, err
	}
	networkHashesPerSec, ok := networkHashesPerSecIface.(int64)
	if !ok {
		return nil, &json.RPCError{
			Code:    json.ErrRPCInternal.Code,
			Message: "networkHashesPerSec is not an int64",
		}
	}
	var Difficulty, dBlake2b, dBlake14lr, dBlake2s, dKeccak, dScrypt, dSHA256D, dSkein, dStribog, dX11 float64
	var lastbitsBlake2b, lastbitsBlake14lr, lastbitsBlake2s, lastbitsKeccak, lastbitsScrypt, lastbitsSHA256D, lastbitsSkein, lastbitsStribog, lastbitsX11 uint32
	best := s.Cfg.Chain.BestSnapshot()
	v := s.Cfg.Chain.Index.LookupNode(&best.Hash)
	foundcount, height := 0, best.Height
	switch fork.GetCurrent(height) {
	case 0:
		for foundcount < 2 && height > 0 {
			switch fork.GetAlgoName(v.Header().Version, height) {
			case "sha256d":
				if lastbitsSHA256D == 0 {
					foundcount++
					lastbitsSHA256D = v.Header().Bits
					dSHA256D = getDifficultyRatio(lastbitsSHA256D, s.Cfg.ChainParams, v.Header().Version)
				}
			case "scrypt":
				if lastbitsScrypt == 0 {
					foundcount++
					lastbitsScrypt = v.Header().Bits
					dScrypt = getDifficultyRatio(lastbitsScrypt, s.Cfg.ChainParams, v.Header().Version)
				}
			default:
			}
			v = v.RelativeAncestor(1)
			height--
		}
		switch s.Cfg.Algo {
		case "sha256d":
			Difficulty = dSHA256D
		case "scrypt":
			Difficulty = dScrypt
		default:
		}
		ret = &json.GetMiningInfoResult0{
			Blocks:             int64(best.Height),
			CurrentBlockSize:   best.BlockSize,
			CurrentBlockWeight: best.BlockWeight,
			CurrentBlockTx:     best.NumTxns,
			PowAlgoID:          fork.GetAlgoID(s.Cfg.Algo, height),
			PowAlgo:            s.Cfg.Algo,
			Difficulty:         Difficulty,
			DifficultySHA256D:  dSHA256D,
			DifficultyScrypt:   dScrypt,
			Generate:           s.Cfg.CPUMiner.IsMining(),
			GenProcLimit:       s.Cfg.CPUMiner.NumWorkers(),
			HashesPerSec:       int64(s.Cfg.CPUMiner.HashesPerSecond()),
			NetworkHashPS:      networkHashesPerSec,
			PooledTx:           uint64(s.Cfg.TxMemPool.Count()),
			TestNet:            *Cfg.TestNet3,
		}
	case 1:
		foundcount, height := 0, best.Height
		for foundcount < 9 && height > fork.List[fork.GetCurrent(height)].ActivationHeight-512 {
			switch fork.GetAlgoName(v.Header().Version, height) {
			case "blake2b":
				if lastbitsBlake2b == 0 {
					foundcount++
					lastbitsBlake2b = v.Header().Bits
					dBlake2b = getDifficultyRatio(lastbitsBlake2b, s.Cfg.ChainParams, v.Header().Version)
				}
			case "blake14lr":
				if lastbitsBlake14lr == 0 {
					foundcount++
					lastbitsBlake14lr = v.Header().Bits
					dBlake14lr = getDifficultyRatio(lastbitsBlake14lr, s.Cfg.ChainParams, v.Header().Version)
				}
			case "blake2s":
				if lastbitsBlake2s == 0 {
					foundcount++
					lastbitsBlake2s = v.Header().Bits
					dBlake2s = getDifficultyRatio(lastbitsBlake2s, s.Cfg.ChainParams, v.Header().Version)
				}
			case "keccak":
				if lastbitsKeccak == 0 {
					foundcount++
					lastbitsKeccak = v.Header().Bits
					dKeccak = getDifficultyRatio(lastbitsKeccak, s.Cfg.ChainParams, v.Header().Version)
				}
			case "scrypt":
				if lastbitsScrypt == 0 {
					foundcount++
					lastbitsScrypt = v.Header().Bits
					dScrypt = getDifficultyRatio(lastbitsScrypt, s.Cfg.ChainParams, v.Header().Version)
				}
			case "sha256d":
				if lastbitsSHA256D == 0 {
					foundcount++
					lastbitsSHA256D = v.Header().Bits
					dSHA256D = getDifficultyRatio(lastbitsSHA256D, s.Cfg.ChainParams, v.Header().Version)
				}
			case "skein":
				if lastbitsSkein == 0 {
					foundcount++
					lastbitsSkein = v.Header().Bits
					dSkein = getDifficultyRatio(lastbitsSkein, s.Cfg.ChainParams, v.Header().Version)
				}
			case "stribog":
				if lastbitsStribog == 0 {
					foundcount++
					lastbitsStribog = v.Header().Bits
					dStribog = getDifficultyRatio(lastbitsStribog, s.Cfg.ChainParams, v.Header().Version)
				}
			case "x11":
				if lastbitsX11 == 0 {
					foundcount++
					lastbitsX11 = v.Header().Bits
					dX11 = getDifficultyRatio(lastbitsX11, s.Cfg.ChainParams, v.Header().Version)
				}
			default:
			}
			v = v.RelativeAncestor(1)
			height--
		}
		switch s.Cfg.Algo {
		case "blake2b":
			Difficulty = dBlake2b
		case "blake14lr":
			Difficulty = dBlake14lr
		case "blake2s":
			Difficulty = dBlake2s
		case "keccak":
			Difficulty = dKeccak
		case "scrypt":
			Difficulty = dScrypt
		case "sha256d":
			Difficulty = dSHA256D
		case "skein":
			Difficulty = dSkein
		case "stribog":
			Difficulty = dStribog
		case "x11":
			Difficulty = dX11
		default:
		}
		ret = &json.GetMiningInfoResult{
			Blocks:              int64(best.Height),
			CurrentBlockSize:    best.BlockSize,
			CurrentBlockWeight:  best.BlockWeight,
			CurrentBlockTx:      best.NumTxns,
			PowAlgoID:           fork.GetAlgoID(s.Cfg.Algo, height),
			PowAlgo:             s.Cfg.Algo,
			Difficulty:          Difficulty,
			DifficultyBlake2b:   dBlake2b,
			DifficultyBlake14lr: dBlake14lr,
			DifficultyBlake2s:   dBlake2s,
			DifficultyKeccak:    dKeccak,
			DifficultyScrypt:    dScrypt,
			DifficultySHA256D:   dSHA256D,
			DifficultySkein:     dSkein,
			DifficultyStribog:   dStribog,
			DifficultyX11:       dX11,
			Generate:            s.Cfg.CPUMiner.IsMining(),
			GenAlgo:             s.Cfg.CPUMiner.GetAlgo(),
			GenProcLimit:        s.Cfg.CPUMiner.NumWorkers(),
			HashesPerSec:        int64(s.Cfg.CPUMiner.HashesPerSecond()),
			NetworkHashPS:       networkHashesPerSec,
			PooledTx:            uint64(s.Cfg.TxMemPool.Count()),
			TestNet:             *Cfg.TestNet3,
		}
	}
	return ret, nil
}

// handleGetNetTotals implements the getnettotals command.
func handleGetNetTotals(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	totalBytesRecv, totalBytesSent := s.Cfg.ConnMgr.NetTotals()
	reply := &json.GetNetTotalsResult{
		TotalBytesRecv: totalBytesRecv,
		TotalBytesSent: totalBytesSent,
		TimeMillis:     time.Now().UTC().UnixNano() / int64(time.Millisecond),
	}
	return reply, nil
}

// handleGetNetworkHashPS implements the getnetworkhashps command. This command does not default to the same end block as the parallelcoind. TODO: Really this needs to be expanded to show per-algorithm hashrates
func handleGetNetworkHashPS(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// Note: All valid error return paths should return an int64. Literal zeros are inferred as int, and won't coerce to int64 because the return value is an interface{}.
	c := cmd.(*json.GetNetworkHashPSCmd)
	// When the passed height is too high or zero, just return 0 now since we can't reasonably calculate the number of network hashes per second from invalid values.  When it's negative, use the current best block height.
	best := s.Cfg.Chain.BestSnapshot()
	endHeight := int32(-1)
	if c.Height != nil {
		endHeight = int32(*c.Height)
	}
	if endHeight > best.Height || endHeight == 0 {
		return int64(0), nil
	}
	if endHeight < 0 {
		endHeight = best.Height
	}
	// Calculate the number of blocks per retarget interval based on the chain parameters.
	blocksPerRetarget := int32(s.Cfg.ChainParams.TargetTimespan / s.Cfg.ChainParams.TargetTimePerBlock)
	// Calculate the starting block height based on the passed number of blocks.  When the passed value is negative, use the last block the difficulty changed as the starting height.  Also make sure the starting height is not before the beginning of the chain.
	numBlocks := int32(120)
	if c.Blocks != nil {
		numBlocks = int32(*c.Blocks)
	}
	var startHeight int32
	if numBlocks <= 0 {
		startHeight = endHeight - ((endHeight % blocksPerRetarget) + 1)
	} else {
		startHeight = endHeight - numBlocks
	}
	if startHeight < 0 {
		startHeight = 0
	}
	log <- cl.Debugf{
		"calculating network hashes per second from %d to %d",
		startHeight,
		endHeight,
	}
	// Find the min and max block timestamps as well as calculate the total amount of work that happened between the start and end blocks.
	var minTimestamp, maxTimestamp time.Time
	totalWork := big.NewInt(0)
	for curHeight := startHeight; curHeight <= endHeight; curHeight++ {
		hash, err := s.Cfg.Chain.BlockHashByHeight(curHeight)
		if err != nil {
			context := "Failed to fetch block hash"
			return nil, internalRPCError(err.Error(), context)
		}
		// Fetch the header from chain.
		header, err := s.Cfg.Chain.HeaderByHash(hash)
		if err != nil {
			context := "Failed to fetch block header"
			return nil, internalRPCError(err.Error(), context)
		}
		if curHeight == startHeight {
			minTimestamp = header.Timestamp
			maxTimestamp = minTimestamp
		} else {
			totalWork.Add(totalWork, blockchain.CalcWork(header.Bits, best.Height+1, header.Version))
			if minTimestamp.After(header.Timestamp) {
				minTimestamp = header.Timestamp
			}
			if maxTimestamp.Before(header.Timestamp) {
				maxTimestamp = header.Timestamp
			}
		}
	}
	// Calculate the difference in seconds between the min and max block timestamps and avoid division by zero in the case where there is no time difference.
	timeDiff := int64(maxTimestamp.Sub(minTimestamp) / time.Second)
	if timeDiff == 0 {
		return int64(0), nil
	}
	hashesPerSec := new(big.Int).Div(totalWork, big.NewInt(timeDiff))
	return hashesPerSec.Int64(), nil
}

// handleGetPeerInfo implements the getpeerinfo command.
func handleGetPeerInfo(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	peers := s.Cfg.ConnMgr.ConnectedPeers()
	syncPeerID := s.Cfg.SyncMgr.SyncPeerID()
	infos := make([]*json.GetPeerInfoResult, 0, len(peers))
	for _, p := range peers {
		statsSnap := p.ToPeer().StatsSnapshot()
		info := &json.GetPeerInfoResult{
			ID:             statsSnap.ID,
			Addr:           statsSnap.Addr,
			AddrLocal:      p.ToPeer().LocalAddr().String(),
			Services:       fmt.Sprintf("%08d", uint64(statsSnap.Services)),
			RelayTxes:      !p.IsTxRelayDisabled(),
			LastSend:       statsSnap.LastSend.Unix(),
			LastRecv:       statsSnap.LastRecv.Unix(),
			BytesSent:      statsSnap.BytesSent,
			BytesRecv:      statsSnap.BytesRecv,
			ConnTime:       statsSnap.ConnTime.Unix(),
			PingTime:       float64(statsSnap.LastPingMicros),
			TimeOffset:     statsSnap.TimeOffset,
			Version:        statsSnap.Version,
			SubVer:         statsSnap.UserAgent,
			Inbound:        statsSnap.Inbound,
			StartingHeight: statsSnap.StartingHeight,
			CurrentHeight:  statsSnap.LastBlock,
			BanScore:       int32(p.BanScore()),
			FeeFilter:      p.FeeFilter(),
			SyncNode:       statsSnap.ID == syncPeerID,
		}
		if p.ToPeer().LastPingNonce() != 0 {
			wait := float64(time.Since(statsSnap.LastPingTime).Nanoseconds())
			// We actually want microseconds.
			info.PingWait = wait / 1000
		}
		infos = append(infos, info)
	}
	return infos, nil
}

// handleGetRawMempool implements the getrawmempool command.
func handleGetRawMempool(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*json.GetRawMempoolCmd)
	mp := s.Cfg.TxMemPool
	if c.Verbose != nil && *c.Verbose {
		return mp.RawMempoolVerbose(), nil
	}
	// The response is simply an array of the transaction hashes if the verbose flag is not set.
	descs := mp.TxDescs()
	hashStrings := make([]string, len(descs))
	for i := range hashStrings {
		hashStrings[i] = descs[i].Tx.Hash().String()
	}
	return hashStrings, nil
}

// handleGetRawTransaction implements the getrawtransaction command.
func handleGetRawTransaction(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*json.GetRawTransactionCmd)
	// Convert the provided transaction hash hex to a Hash.
	txHash, err := chainhash.NewHashFromStr(c.Txid)
	if err != nil {
		return nil, rpcDecodeHexError(c.Txid)
	}
	verbose := false
	if c.Verbose != nil {
		verbose = *c.Verbose != 0
	}
	// Try to fetch the transaction from the memory pool and if that fails, try the block database.
	var mtx *wire.MsgTx
	var blkHash *chainhash.Hash
	var blkHeight int32
	tx, err := s.Cfg.TxMemPool.FetchTransaction(txHash)
	if err != nil {
		if s.Cfg.TxIndex == nil {
			return nil, &json.RPCError{
				Code: json.ErrRPCNoTxInfo,
				Message: "The transaction index must be " +
					"enabled to query the blockchain " +
					"(specify --txindex)",
			}
		}
		// Look up the location of the transaction.
		blockRegion, err := s.Cfg.TxIndex.TxBlockRegion(txHash)
		if err != nil {
			context := "Failed to retrieve transaction location"
			return nil, internalRPCError(err.Error(), context)
		}
		if blockRegion == nil {
			return nil, rpcNoTxInfoError(txHash)
		}
		// Load the raw transaction bytes from the database.
		var txBytes []byte
		err = s.Cfg.DB.View(func(dbTx database.Tx) error {
			var err error
			txBytes, err = dbTx.FetchBlockRegion(blockRegion)
			return err
		})
		if err != nil {
			return nil, rpcNoTxInfoError(txHash)
		}
		// When the verbose flag isn't set, simply return the serialized transaction as a hex-encoded string.  This is done here to avoid deserializing it only to reserialize it again later.
		if !verbose {
			return hex.EncodeToString(txBytes), nil
		}
		// Grab the block height.
		blkHash = blockRegion.Hash
		blkHeight, err = s.Cfg.Chain.BlockHeightByHash(blkHash)
		if err != nil {
			context := "Failed to retrieve block height"
			return nil, internalRPCError(err.Error(), context)
		}
		// Deserialize the transaction
		var msgTx wire.MsgTx
		err = msgTx.Deserialize(bytes.NewReader(txBytes))
		if err != nil {
			context := "Failed to deserialize transaction"
			return nil, internalRPCError(err.Error(), context)
		}
		mtx = &msgTx
	} else {
		// When the verbose flag isn't set, simply return the network-serialized transaction as a hex-encoded string.
		if !verbose {
			// Note that this is intentionally not directly returning because the first return value is a string and it would result in returning an empty string to the client instead of nothing (nil) in the case of an error.
			mtxHex, err := messageToHex(tx.MsgTx())
			if err != nil {
				return nil, err
			}
			return mtxHex, nil
		}
		mtx = tx.MsgTx()
	}
	// The verbose flag is set, so generate the JSON object and return it.
	var blkHeader *wire.BlockHeader
	var blkHashStr string
	var chainHeight int32
	if blkHash != nil {
		// Fetch the header from chain.
		header, err := s.Cfg.Chain.HeaderByHash(blkHash)
		if err != nil {
			context := "Failed to fetch block header"
			return nil, internalRPCError(err.Error(), context)
		}
		blkHeader = &header
		blkHashStr = blkHash.String()
		chainHeight = s.Cfg.Chain.BestSnapshot().Height
	}
	rawTxn, err := createTxRawResult(s.Cfg.ChainParams, mtx, txHash.String(),
		blkHeader, blkHashStr, blkHeight, chainHeight)
	if err != nil {
		return nil, err
	}
	return *rawTxn, nil
}

// handleGetTxOut handles gettxout commands.
func handleGetTxOut(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*json.GetTxOutCmd)
	// Convert the provided transaction hash hex to a Hash.
	txHash, err := chainhash.NewHashFromStr(c.Txid)
	if err != nil {
		return nil, rpcDecodeHexError(c.Txid)
	}
	// If requested and the tx is available in the mempool try to fetch it from there, otherwise attempt to fetch from the block database.
	var bestBlockHash string
	var confirmations int32
	var value int64
	var pkScript []byte
	var isCoinbase bool
	includeMempool := true
	if c.IncludeMempool != nil {
		includeMempool = *c.IncludeMempool
	}
	// TODO: This is racy.  It should attempt to fetch it directly and check the error.
	if includeMempool && s.Cfg.TxMemPool.HaveTransaction(txHash) {
		tx, err := s.Cfg.TxMemPool.FetchTransaction(txHash)
		if err != nil {
			return nil, rpcNoTxInfoError(txHash)
		}
		mtx := tx.MsgTx()
		if c.Vout > uint32(len(mtx.TxOut)-1) {
			return nil, &json.RPCError{
				Code: json.ErrRPCInvalidTxVout,
				Message: "Output index number (vout) does not " +
					"exist for transaction.",
			}
		}
		txOut := mtx.TxOut[c.Vout]
		if txOut == nil {
			errStr := fmt.Sprintf("Output index: %d for txid: %s "+
				"does not exist", c.Vout, txHash)
			return nil, internalRPCError(errStr, "")
		}
		best := s.Cfg.Chain.BestSnapshot()
		bestBlockHash = best.Hash.String()
		confirmations = 0
		value = txOut.Value
		pkScript = txOut.PkScript
		isCoinbase = blockchain.IsCoinBaseTx(mtx)
	} else {
		out := wire.OutPoint{Hash: *txHash, Index: c.Vout}
		entry, err := s.Cfg.Chain.FetchUtxoEntry(out)
		if err != nil {
			return nil, rpcNoTxInfoError(txHash)
		}
		// To match the behavior of the reference client, return nil (JSON null) if the transaction output is spent by another transaction already in the main chain.  Mined transactions that are spent by a mempool transaction are not affected by this.
		if entry == nil || entry.IsSpent() {
			return nil, nil
		}
		best := s.Cfg.Chain.BestSnapshot()
		bestBlockHash = best.Hash.String()
		confirmations = 1 + best.Height - entry.BlockHeight()
		value = entry.Amount()
		pkScript = entry.PkScript()
		isCoinbase = entry.IsCoinBase()
	}
	// Disassemble script into single line printable format. The disassembled string will contain [error] inline if the script doesn't fully parse, so ignore the error here.
	disbuf, _ := txscript.DisasmString(pkScript)
	// Get further info about the script. Ignore the error here since an error means the script couldn't parse and there is no additional information about it anyways.
	scriptClass, addrs, reqSigs, _ := txscript.ExtractPkScriptAddrs(pkScript, s.Cfg.ChainParams)
	addresses := make([]string, len(addrs))
	for i, addr := range addrs {
		addresses[i] = addr.EncodeAddress()
	}
	txOutReply := &json.GetTxOutResult{
		BestBlock:     bestBlockHash,
		Confirmations: int64(confirmations),
		Value:         util.Amount(value).ToDUO(),
		ScriptPubKey: json.ScriptPubKeyResult{
			Asm:       disbuf,
			Hex:       hex.EncodeToString(pkScript),
			ReqSigs:   int32(reqSigs),
			Type:      scriptClass.String(),
			Addresses: addresses,
		},
		Coinbase: isCoinbase,
	}
	return txOutReply, nil
}

// handleHelp implements the help command.
func handleHelp(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*json.HelpCmd)
	// Provide a usage overview of all commands when no specific command was specified.
	var command string
	if c.Command != nil {
		command = *c.Command
	}
	if command == "" {
		usage, err := s.helpCacher.rpcUsage(false)
		if err != nil {
			context := "Failed to generate RPC usage"
			return nil, internalRPCError(err.Error(), context)
		}
		return usage, nil
	}
	// Check that the command asked for is supported and implemented.  Only search the main list of handlers since help should not be provided for commands that are unimplemented or related to wallet functionality.
	if _, ok := rpcHandlers[command]; !ok {
		return nil, &json.RPCError{
			Code:    json.ErrRPCInvalidParameter,
			Message: "Unknown command: " + command,
		}
	}
	// Get the help for the command.
	help, err := s.helpCacher.rpcMethodHelp(command)
	if err != nil {
		context := "Failed to generate help"
		return nil, internalRPCError(err.Error(), context)
	}
	return help, nil
}

// handleNode handles node commands.
func handleNode(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*json.NodeCmd)
	var addr string
	var nodeID uint64
	var errN, err error
	params := s.Cfg.ChainParams
	switch c.SubCmd {
	case "disconnect":
		// If we have a valid uint disconnect by node id. Otherwise, attempt to disconnect by address, returning an error if a valid IP address is not supplied.
		if nodeID, errN = strconv.ParseUint(c.Target, 10, 32); errN == nil {
			err = s.Cfg.ConnMgr.DisconnectByID(int32(nodeID))
		} else {
			if _, _, errP := net.SplitHostPort(c.Target); errP == nil || net.ParseIP(c.Target) != nil {
				addr = NormalizeAddress(c.Target, params.DefaultPort)
				err = s.Cfg.ConnMgr.DisconnectByAddr(addr)
			} else {
				return nil, &json.RPCError{
					Code:    json.ErrRPCInvalidParameter,
					Message: "invalid address or node ID",
				}
			}
		}
		if err != nil && peerExists(s.Cfg.ConnMgr, addr, int32(nodeID)) {
			return nil, &json.RPCError{
				Code:    json.ErrRPCMisc,
				Message: "can't disconnect a permanent peer, use remove",
			}
		}
	case "remove":
		// If we have a valid uint disconnect by node id. Otherwise, attempt to disconnect by address, returning an error if a valid IP address is not supplied.
		if nodeID, errN = strconv.ParseUint(c.Target, 10, 32); errN == nil {
			err = s.Cfg.ConnMgr.RemoveByID(int32(nodeID))
		} else {
			if _, _, errP := net.SplitHostPort(c.Target); errP == nil || net.ParseIP(c.Target) != nil {
				addr = NormalizeAddress(c.Target, params.DefaultPort)
				err = s.Cfg.ConnMgr.RemoveByAddr(addr)
			} else {
				return nil, &json.RPCError{
					Code:    json.ErrRPCInvalidParameter,
					Message: "invalid address or node ID",
				}
			}
		}
		if err != nil && peerExists(s.Cfg.ConnMgr, addr, int32(nodeID)) {
			return nil, &json.RPCError{
				Code:    json.ErrRPCMisc,
				Message: "can't remove a temporary peer, use disconnect",
			}
		}
	case "connect":
		addr = NormalizeAddress(c.Target, params.DefaultPort)
		// Default to temporary connections.
		subCmd := "temp"
		if c.ConnectSubCmd != nil {
			subCmd = *c.ConnectSubCmd
		}
		switch subCmd {
		case "perm", "temp":
			err = s.Cfg.ConnMgr.Connect(addr, subCmd == "perm")
		default:
			return nil, &json.RPCError{
				Code:    json.ErrRPCInvalidParameter,
				Message: "invalid subcommand for node connect",
			}
		}
	default:
		return nil, &json.RPCError{
			Code:    json.ErrRPCInvalidParameter,
			Message: "invalid subcommand for node",
		}
	}
	if err != nil {
		return nil, &json.RPCError{
			Code:    json.ErrRPCInvalidParameter,
			Message: err.Error(),
		}
	}
	// no data returned unless an error.
	return nil, nil
}

// handlePing implements the ping command.
func handlePing(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// Ask server to ping \o_
	nonce, err := wire.RandomUint64()
	if err != nil {
		return nil, internalRPCError("Not sending ping - failed to generate nonce: "+err.Error(), "")
	}
	s.Cfg.ConnMgr.BroadcastMessage(wire.NewMsgPing(nonce))
	return nil, nil
}

// handleSearchRawTransactions implements the searchrawtransactions command.
func handleSearchRawTransactions(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	// Respond with an error if the address index is not enabled.
	addrIndex := s.Cfg.AddrIndex
	if addrIndex == nil {
		return nil, &json.RPCError{
			Code:    json.ErrRPCMisc,
			Message: "Address index must be enabled (--addrindex)",
		}
	}
	// Override the flag for including extra previous output information in each input if needed.
	c := cmd.(*json.SearchRawTransactionsCmd)
	vinExtra := false
	if c.VinExtra != nil {
		vinExtra = *c.VinExtra != 0
	}
	// Including the extra previous output information requires the transaction index.  Currently the address index relies on the transaction index, so this check is redundant, but it's better to be safe in case the address index is ever changed to not rely on it.
	if vinExtra && s.Cfg.TxIndex == nil {
		return nil, &json.RPCError{
			Code:    json.ErrRPCMisc,
			Message: "Transaction index must be enabled (--txindex)",
		}
	}
	// Attempt to decode the supplied address.
	params := s.Cfg.ChainParams
	addr, err := util.DecodeAddress(c.Address, params)
	if err != nil {
		return nil, &json.RPCError{
			Code:    json.ErrRPCInvalidAddressOrKey,
			Message: "Invalid address or key: " + err.Error(),
		}
	}
	// Override the default number of requested entries if needed.  Also, just return now if the number of requested entries is zero to avoid extra work.
	numRequested := 100
	if c.Count != nil {
		numRequested = *c.Count
		if numRequested < 0 {
			numRequested = 1
		}
	}
	if numRequested == 0 {
		return nil, nil
	}
	// Override the default number of entries to skip if needed.
	var numToSkip int
	if c.Skip != nil {
		numToSkip = *c.Skip
		if numToSkip < 0 {
			numToSkip = 0
		}
	}
	// Override the reverse flag if needed.
	var reverse bool
	if c.Reverse != nil {
		reverse = *c.Reverse
	}
	// Add transactions from mempool first if client asked for reverse order.  Otherwise, they will be added last (as needed depending on the requested counts). NOTE: This code doesn't sort by dependency.  This might be something to do in the future for the client's convenience, or leave it to the client.
	numSkipped := uint32(0)
	addressTxns := make([]retrievedTx, 0, numRequested)
	if reverse {
		// Transactions in the mempool are not in a block header yet, so the block header field in the retieved transaction struct is left nil.
		mpTxns, mpSkipped := fetchMempoolTxnsForAddress(s, addr, uint32(numToSkip), uint32(numRequested))
		numSkipped += mpSkipped
		for _, tx := range mpTxns {
			addressTxns = append(addressTxns, retrievedTx{tx: tx})
		}
	}
	// Fetch transactions from the database in the desired order if more are needed.
	if len(addressTxns) < numRequested {
		err = s.Cfg.DB.View(func(dbTx database.Tx) error {
			regions, dbSkipped, err := addrIndex.TxRegionsForAddress(dbTx, addr, uint32(numToSkip)-numSkipped, uint32(numRequested-len(addressTxns)), reverse)
			if err != nil {
				return err
			}
			// Load the raw transaction bytes from the database.
			serializedTxns, err := dbTx.FetchBlockRegions(regions)
			if err != nil {
				return err
			}
			// Add the transaction and the hash of the block it is contained in to the list.  Note that the transaction is left serialized here since the caller might have requested non-verbose output and hence there would be/ no point in deserializing it just to reserialize it later.
			for i, serializedTx := range serializedTxns {
				addressTxns = append(addressTxns, retrievedTx{
					txBytes: serializedTx,
					blkHash: regions[i].Hash,
				})
			}
			numSkipped += dbSkipped
			return nil
		})
		if err != nil {
			context := "Failed to load address index entries"
			return nil, internalRPCError(err.Error(), context)
		}
	}
	// Add transactions from mempool last if client did not request reverse order and the number of results is still under the number requested.
	if !reverse && len(addressTxns) < numRequested {
		// Transactions in the mempool are not in a block header yet, so the block header field in the retieved transaction struct is left nil.
		mpTxns, mpSkipped := fetchMempoolTxnsForAddress(s, addr, uint32(numToSkip)-numSkipped, uint32(numRequested-len(addressTxns)))
		numSkipped += mpSkipped
		for _, tx := range mpTxns {
			addressTxns = append(addressTxns, retrievedTx{tx: tx})
		}
	}
	// Address has never been used if neither source yielded any results.
	if len(addressTxns) == 0 {
		return nil, &json.RPCError{
			Code:    json.ErrRPCNoTxInfo,
			Message: "No information available about address",
		}
	}
	// Serialize all of the transactions to hex.
	hexTxns := make([]string, len(addressTxns))
	for i := range addressTxns {
		// Simply encode the raw bytes to hex when the retrieved transaction is already in serialized form.
		rtx := &addressTxns[i]
		if rtx.txBytes != nil {
			hexTxns[i] = hex.EncodeToString(rtx.txBytes)
			continue
		}
		// Serialize the transaction first and convert to hex when the retrieved transaction is the deserialized structure.
		hexTxns[i], err = messageToHex(rtx.tx.MsgTx())
		if err != nil {
			return nil, err
		}
	}
	// When not in verbose mode, simply return a list of serialized txns.
	if c.Verbose != nil && *c.Verbose == 0 {
		return hexTxns, nil
	}
	// Normalize the provided filter addresses (if any) to ensure there are no duplicates.
	filterAddrMap := make(map[string]struct{})
	if c.FilterAddrs != nil && len(*c.FilterAddrs) > 0 {
		for _, addr := range *c.FilterAddrs {
			filterAddrMap[addr] = struct{}{}
		}
	}
	// The verbose flag is set, so generate the JSON object and return it.
	best := s.Cfg.Chain.BestSnapshot()
	srtList := make([]json.SearchRawTransactionsResult, len(addressTxns))
	for i := range addressTxns {
		// The deserialized transaction is needed, so deserialize the retrieved transaction if it's in serialized form (which will be the case when it was lookup up from the database). Otherwise, use the existing deserialized transaction.
		rtx := &addressTxns[i]
		var mtx *wire.MsgTx
		if rtx.tx == nil {
			// Deserialize the transaction.
			mtx = new(wire.MsgTx)
			err := mtx.Deserialize(bytes.NewReader(rtx.txBytes))
			if err != nil {
				context := "Failed to deserialize transaction"
				return nil, internalRPCError(err.Error(), context)
			}
		} else {
			mtx = rtx.tx.MsgTx()
		}
		result := &srtList[i]
		result.Hex = hexTxns[i]
		result.Txid = mtx.TxHash().String()
		result.Vin, err = createVinListPrevOut(s, mtx, params, vinExtra, filterAddrMap)
		if err != nil {
			return nil, err
		}
		result.Vout = createVoutList(mtx, params, filterAddrMap)
		result.Version = mtx.Version
		result.LockTime = mtx.LockTime
		// Transactions grabbed from the mempool aren't yet in a block, so conditionally fetch block details here.  This will be reflected in the final JSON output (mempool won't have confirmations or block information).
		var blkHeader *wire.BlockHeader
		var blkHashStr string
		var blkHeight int32
		if blkHash := rtx.blkHash; blkHash != nil {
			// Fetch the header from chain.
			header, err := s.Cfg.Chain.HeaderByHash(blkHash)
			if err != nil {
				return nil, &json.RPCError{
					Code:    json.ErrRPCBlockNotFound,
					Message: "Block not found",
				}
			}
			// Get the block height from chain.
			height, err := s.Cfg.Chain.BlockHeightByHash(blkHash)
			if err != nil {
				context := "Failed to obtain block height"
				return nil, internalRPCError(err.Error(), context)
			}
			blkHeader = &header
			blkHashStr = blkHash.String()
			blkHeight = height
		}
		// Add the block information to the result if there is any.
		if blkHeader != nil {
			// This is not a typo, they are identical in Bitcoin Core as well.
			result.Time = blkHeader.Timestamp.Unix()
			result.Blocktime = blkHeader.Timestamp.Unix()
			result.BlockHash = blkHashStr
			result.Confirmations = uint64(1 + best.Height - blkHeight)
		}
	}
	return srtList, nil
}

// handleSendRawTransaction implements the sendrawtransaction command.
func handleSendRawTransaction(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*json.SendRawTransactionCmd)
	// Deserialize and send off to tx relay
	hexStr := c.HexTx
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}
	serializedTx, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, rpcDecodeHexError(hexStr)
	}
	var msgTx wire.MsgTx
	err = msgTx.Deserialize(bytes.NewReader(serializedTx))
	if err != nil {
		return nil, &json.RPCError{
			Code:    json.ErrRPCDeserialization,
			Message: "TX decode failed: " + err.Error(),
		}
	}
	// Use 0 for the tag to represent local node.
	tx := util.NewTx(&msgTx)
	acceptedTxs, err := s.Cfg.TxMemPool.ProcessTransaction(tx, false, false, 0)
	if err != nil {
		// When the error is a rule error, it means the transaction was simply rejected as opposed to something actually going wrong, so log it as such.  Otherwise, something really did go wrong, so log it as an actual error.  In both cases, a JSON-RPC error is returned to the client with the deserialization error code (to match bitcoind behavior).
		if _, ok := err.(mempool.RuleError); ok {
			log <- cl.Debugf{
				"rejected transaction %v: %v", tx.Hash(), err,
			}
		} else {
			log <- cl.Errorf{
				"failed to process transaction %v: %v", tx.Hash(), err,
			}
		}
		return nil, &json.RPCError{
			Code:    json.ErrRPCDeserialization,
			Message: "TX rejected: " + err.Error(),
		}
	}
	// When the transaction was accepted it should be the first item in the returned array of accepted transactions.  The only way this will not be true is if the API for ProcessTransaction changes and this code is not properly updated, but ensure the condition holds as a safeguard. Also, since an error is being returned to the caller, ensure the transaction is removed from the memory pool.
	if len(acceptedTxs) == 0 || !acceptedTxs[0].Tx.Hash().IsEqual(tx.Hash()) {
		s.Cfg.TxMemPool.RemoveTransaction(tx, true)
		errStr := fmt.Sprintf("transaction %v is not in accepted list", tx.Hash())
		return nil, internalRPCError(errStr, "")
	}
	// Generate and relay inventory vectors for all newly accepted transactions into the memory pool due to the original being accepted.
	s.Cfg.ConnMgr.RelayTransactions(acceptedTxs)
	// Notify both websocket and getblocktemplate long poll clients of all newly accepted transactions.
	s.NotifyNewTransactions(acceptedTxs)
	// Keep track of all the sendrawtransaction request txns so that they can be rebroadcast if they don't make their way into a block.
	txD := acceptedTxs[0]
	iv := wire.NewInvVect(wire.InvTypeTx, txD.Tx.Hash())
	s.Cfg.ConnMgr.AddRebroadcastInventory(iv, txD)
	return tx.Hash().String(), nil
}

// handleSetGenerate implements the setgenerate command.
func handleSetGenerate(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*json.SetGenerateCmd)
	// Disable generation regardless of the provided generate flag if the maximum number of threads (goroutines for our purposes) is 0. Otherwise enable or disable it depending on the provided flag.
	fmt.Println(*c.GenProcLimit, c.Generate)
	generate := c.Generate
	genProcLimit := -1
	if c.GenProcLimit != nil {
		genProcLimit = *c.GenProcLimit
	}
	if genProcLimit == 0 {
		generate = false
	}
	if s.Cfg.CPUMiner.IsMining() {
		// if s.Cfg.CPUMiner.GetAlgo() != s.Cfg.Algo {
		s.Cfg.CPUMiner.Stop()
		generate = true
		// }
	}
	if !generate {
		s.Cfg.CPUMiner.Stop()
	} else {
		// Respond with an error if there are no addresses to pay the created blocks to.
		if len(StateCfg.ActiveMiningAddrs) == 0 {
			return nil, &json.RPCError{
				Code:    json.ErrRPCInternal.Code,
				Message: "no payment addresses specified via --miningaddr",
			}
		}
		// fmt.Println("generating with algo", s.Cfg.Algo)
		// s.Cfg.CPUMiner.SetAlgo(s.Cfg.Algo)
		// It's safe to call start even if it's already started.
		s.Cfg.CPUMiner.SetNumWorkers(int32(genProcLimit))
		s.Cfg.CPUMiner.Start()
	}
	return nil, nil
}

// handleStop implements the stop command.
func handleStop(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	select {
	case s.requestProcessShutdown <- struct{}{}:
		// fmt.Println("chan:s.requestProcessShutdown <- struct{}{}")
	default:
	}
	return "node stopping", nil
}

// handleSubmitBlock implements the submitblock command.
func handleSubmitBlock(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*json.SubmitBlockCmd)
	// Deserialize the submitted block.
	hexStr := c.HexBlock
	if len(hexStr)%2 != 0 {
		hexStr = "0" + c.HexBlock
	}
	serializedBlock, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, rpcDecodeHexError(hexStr)
	}
	block, err := util.NewBlockFromBytes(serializedBlock)
	if err != nil {
		return nil, &json.RPCError{
			Code:    json.ErrRPCDeserialization,
			Message: "Block decode failed: " + err.Error(),
		}
	}
	// Process this block using the same rules as blocks coming from other nodes.  This will in turn relay it to the network like normal.
	_, err = s.Cfg.SyncMgr.SubmitBlock(block, blockchain.BFNone)
	if err != nil {
		return fmt.Sprintf("rejected: %s", err.Error()), nil
	}
	log <- cl.Infof{
		"accepted block %s via submitblock", block.Hash(),
	}
	return nil, nil
}

// handleUnimplemented is the handler for commands that should ultimately be supported but are not yet implemented.
func handleUnimplemented(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return nil, ErrRPCUnimplemented
}

// handleUptime implements the uptime command.
func handleUptime(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return time.Now().Unix() - s.Cfg.StartupTime, nil
}

// handleValidateAddress implements the validateaddress command.
func handleValidateAddress(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*json.ValidateAddressCmd)
	result := json.ValidateAddressChainResult{}
	addr, err := util.DecodeAddress(c.Address, s.Cfg.ChainParams)
	if err != nil {
		// Return the default value (false) for IsValid.
		return result, nil
	}
	result.Address = addr.EncodeAddress()
	result.IsValid = true
	return result, nil
}

// handleVerifyChain implements the verifychain command.
func handleVerifyChain(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*json.VerifyChainCmd)
	var checkLevel, checkDepth int32
	if c.CheckLevel != nil {
		checkLevel = *c.CheckLevel
	}
	if c.CheckDepth != nil {
		checkDepth = *c.CheckDepth
	}
	err := verifyChain(s, checkLevel, checkDepth)
	return err == nil, nil
}

// handleVerifyMessage implements the verifymessage command.
func handleVerifyMessage(
	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	c := cmd.(*json.VerifyMessageCmd)
	// Decode the provided address.
	params := s.Cfg.ChainParams
	addr, err := util.DecodeAddress(c.Address, params)
	if err != nil {
		return nil, &json.RPCError{
			Code:    json.ErrRPCInvalidAddressOrKey,
			Message: "Invalid address or key: " + err.Error(),
		}
	}
	// Only P2PKH addresses are valid for signing.
	if _, ok := addr.(*util.AddressPubKeyHash); !ok {
		return nil, &json.RPCError{
			Code:    json.ErrRPCType,
			Message: "Address is not a pay-to-pubkey-hash address",
		}
	}
	// Decode base64 signature.
	sig, err := base64.StdEncoding.DecodeString(c.Signature)
	if err != nil {
		return nil, &json.RPCError{
			Code:    json.ErrRPCParse.Code,
			Message: "Malformed base64 encoding: " + err.Error(),
		}
	}
	// Validate the signature - this just shows that it was valid at all. we will compare it with the key next.
	var buf bytes.Buffer
	wire.WriteVarString(&buf, 0, "Bitcoin Signed Message:\n")
	wire.WriteVarString(&buf, 0, c.Message)
	expectedMessageHash := chainhash.DoubleHashB(buf.Bytes())
	pk, wasCompressed, err := ec.RecoverCompact(ec.S256(), sig,
		expectedMessageHash)
	if err != nil {
		// Mirror Bitcoin Core behavior, which treats error in RecoverCompact as invalid signature.
		return false, nil
	}
	// Reconstruct the pubkey hash.
	var serializedPK []byte
	if wasCompressed {
		serializedPK = pk.SerializeCompressed()
	} else {
		serializedPK = pk.SerializeUncompressed()
	}
	address, err := util.NewAddressPubKey(serializedPK, params)
	if err != nil {
		// Again mirror Bitcoin Core behavior, which treats error in public key reconstruction as invalid signature.
		return false, nil
	}
	// Return boolean if addresses match.
	return address.EncodeAddress() == c.Address, nil
}

// handleVersion implements the version command. NOTE: This is a btcsuite extension ported from github.com/decred/dcrd.
func handleVersion(
	s *rpcServer,
	cmd interface{},
	closeChan <-chan struct{},
) (
	interface{},
	error,
) {
	result := map[string]json.VersionResult{
		"podjsonrpcapi": {
			VersionString: jsonrpcSemverString,
			Major:         jsonrpcSemverMajor,
			Minor:         jsonrpcSemverMinor,
			Patch:         jsonrpcSemverPatch,
		},
	}
	return result, nil
}
func init() {
	rpcHandlers = rpcHandlersBeforeInit
	rand.Seed(time.Now().UnixNano())
}

// internalRPCError is a convenience function to convert an internal error to an RPC error with the appropriate code set.  It also logs the error to the RPC server subsystem since internal errors really should not occur.  The context parameter is only used in the log message and may be empty if it's not needed.
func internalRPCError(
	errStr, context string) *json.RPCError {
	logStr := errStr
	if context != "" {
		logStr = context + ": " + errStr
	}
	log <- cl.Err(logStr)
	return json.NewRPCError(json.ErrRPCInternal.Code, errStr)
}

// jsonAuthFail sends a message back to the client if the http auth is rejected.
func jsonAuthFail(
	w http.ResponseWriter,
) {
	w.Header().Add("WWW-Authenticate", `Basic realm="pod RPC"`)
	http.Error(w, "401 Unauthorized.", http.StatusUnauthorized)
}

// messageToHex serializes a message to the wire protocol encoding using the latest protocol version and returns a hex-encoded string of the result.
func messageToHex(
	msg wire.Message,
) (
	string,
	error,
) {
	var buf bytes.Buffer
	if err := msg.BtcEncode(&buf, maxProtocolVersion, wire.WitnessEncoding); err != nil {
		context := fmt.Sprintf("Failed to encode msg of type %T", msg)
		return "", internalRPCError(err.Error(), context)
	}
	return hex.EncodeToString(buf.Bytes()), nil
}

// newGbtWorkState returns a new instance of a gbtWorkState with all internal fields initialized and ready to use.
func newGbtWorkState(
	timeSource blockchain.MedianTimeSource,
	algoname string,
) *gbtWorkState {
	return &gbtWorkState{
		notifyMap:  make(map[chainhash.Hash]map[int64]chan struct{}),
		timeSource: timeSource,
		algo:       algoname,
	}
}

// newRPCServer returns a new instance of the rpcServer struct.
func newRPCServer(
	config *rpcserverConfig,
) (
	*rpcServer,
	error,
) {
	rpc := rpcServer{
		Cfg:                    *config,
		statusLines:            make(map[int]string),
		gbtWorkState:           newGbtWorkState(config.TimeSource, config.Algo),
		helpCacher:             newHelpCacher(),
		requestProcessShutdown: make(chan struct{}),
		quit:                   make(chan int),
	}
	if *Cfg.Username != "" && *Cfg.Password != "" {
		login := *Cfg.Username + ":" + *Cfg.Password
		auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(login))
		rpc.authsha = sha256.Sum256([]byte(auth))
	}
	if *Cfg.LimitUser != "" && *Cfg.LimitPass != "" {
		login := *Cfg.LimitUser + ":" + *Cfg.LimitPass
		auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(login))
		rpc.limitauthsha = sha256.Sum256([]byte(auth))
	}
	rpc.ntfnMgr = newWsNotificationManager(&rpc)
	rpc.Cfg.Chain.Subscribe(rpc.handleBlockchainNotification)
	return &rpc, nil
}

// parseCmd parses a JSON-RPC request object into known concrete command.  The err field of the returned parsedRPCCmd struct will contain an RPC error that is suitable for use in replies if the command is invalid in some way such as an unregistered command or invalid parameters.
func parseCmd(
	request *json.Request,
) *parsedRPCCmd {
	var parsedCmd parsedRPCCmd
	parsedCmd.id = request.ID
	parsedCmd.method = request.Method
	cmd, err := json.UnmarshalCmd(request)
	if err != nil {
		// When the error is because the method is not registered, produce a method not found RPC error.
		if jerr, ok := err.(json.Error); ok &&
			jerr.ErrorCode == json.ErrUnregisteredMethod {
			parsedCmd.err = json.ErrRPCMethodNotFound
			return &parsedCmd
		}
		// Otherwise, some type of invalid parameters is the cause, so produce the equivalent RPC error.
		parsedCmd.err = json.NewRPCError(
			json.ErrRPCInvalidParams.Code, err.Error())
		return &parsedCmd
	}
	parsedCmd.cmd = cmd
	return &parsedCmd
}

// peerExists determines if a certain peer is currently connected given information about all currently connected peers. Peer existence is determined using either a target address or node id.
func peerExists(
	connMgr rpcserverConnManager,
	addr string,
	nodeID int32,
) bool {
	for _, p := range connMgr.ConnectedPeers() {
		if p.ToPeer().ID() == nodeID || p.ToPeer().Addr() == addr {
			return true
		}
	}
	return false
}

// rpcDecodeHexError is a convenience function for returning a nicely formatted RPC error which indicates the provided hex string failed to decode.
func rpcDecodeHexError(
	gotHex string,
) *json.RPCError {
	return json.NewRPCError(json.ErrRPCDecodeHexString,
		fmt.Sprintf("Argument must be hexadecimal string (not %q)",
			gotHex))
}

// rpcNoTxInfoError is a convenience function for returning a nicely formatted RPC error which indicates there is no information available for the provided transaction hash.
func rpcNoTxInfoError(
	txHash *chainhash.Hash,
) *json.RPCError {
	return json.NewRPCError(json.ErrRPCNoTxInfo,
		fmt.Sprintf("No information available about transaction %v",
			txHash))
}

// softForkStatus converts a ThresholdState state into a human readable string corresponding to the particular state.
func softForkStatus(
	state blockchain.ThresholdState,
) (
	string,
	error,
) {
	switch state {
	case blockchain.ThresholdDefined:
		return "defined", nil
	case blockchain.ThresholdStarted:
		return "started", nil
	case blockchain.ThresholdLockedIn:
		return "lockedin", nil
	case blockchain.ThresholdActive:
		return "active", nil
	case blockchain.ThresholdFailed:
		return "failed", nil
	default:
		return "", fmt.Errorf("unknown deployment state: %v", state)
	}
}
func verifyChain(
	s *rpcServer,
	level,
	depth int32,
) error {
	best := s.Cfg.Chain.BestSnapshot()
	finishHeight := best.Height - depth
	if finishHeight < 0 {
		finishHeight = 0
	}
	log <- cl.Infof{
		"verifying chain for %d blocks at level %d",
		best.Height - finishHeight,
		level,
	}
	for height := best.Height; height > finishHeight; height-- {
		// Level 0 just looks up the block.
		block, err := s.Cfg.Chain.BlockByHeight(height)
		if err != nil {
			log <- cl.Errorf{
				"verify is unable to fetch block at height %d: %v",
				height,
				err,
			}
			return err
		}
		powLimit := fork.GetMinDiff(fork.GetAlgoName(block.MsgBlock().Header.Version, height), height)
		// Level 1 does basic chain sanity checks.
		if level > 0 {
			err := blockchain.CheckBlockSanity(block, powLimit, s.Cfg.TimeSource, true, block.Height(), s.Cfg.ChainParams.Name == "testnet")
			if err != nil {
				log <- cl.Errorf{
					"verify is unable to validate block at hash %v height %d: %v",
					block.Hash(), height, err}
				return err
			}
		}
	}
	log <- cl.Inf("chain verify completed successfully")
	return nil
}

/*
// handleDebugLevel handles debuglevel commands.
func handleDebugLevel(	s *rpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {


	c := cmd.(*json.DebugLevelCmd)
	// Special show command to list supported subsystems.


	if c.LevelSpec == "show" {

		return fmt.Sprintf("Supported subsystems %v",
			supportedSubsystems()), nil
	}
	err := parseAndSetDebugLevels(c.LevelSpec)


	if err != nil {


		return nil, &json.RPCError{
			Code:    json.ErrRPCInvalidParams.Code,
			Message: err.Error(),
		}
	}
	return "Done.", nil
}
*/
// witnessToHex formats the passed witness stack as a slice of hex-encoded strings to be used in a JSON response.
func witnessToHex(
	witness wire.TxWitness) []string {
	// Ensure nil is returned when there are no entries versus an empty slice so it can properly be omitted as necessary.
	if len(witness) == 0 {
		return nil
	}
	result := make([]string, 0, len(witness))
	for _, wit := range witness {
		result = append(result, hex.EncodeToString(wit))
	}
	return result
}
