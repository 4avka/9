package node
import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"git.parallelcoin.io/dev/9/cmd/node/mempool"
	blockchain "git.parallelcoin.io/dev/9/pkg/chain"
	chaincfg "git.parallelcoin.io/dev/9/pkg/chain/config"
	chainhash "git.parallelcoin.io/dev/9/pkg/chain/hash"
	indexers "git.parallelcoin.io/dev/9/pkg/chain/index"
	"git.parallelcoin.io/dev/9/pkg/chain/mining"
	cpuminer "git.parallelcoin.io/dev/9/pkg/chain/mining/cpu"
	netsync "git.parallelcoin.io/dev/9/pkg/chain/sync"
	txscript "git.parallelcoin.io/dev/9/pkg/chain/tx/script"
	"git.parallelcoin.io/dev/9/pkg/chain/wire"
	database "git.parallelcoin.io/dev/9/pkg/db"
	"git.parallelcoin.io/dev/9/pkg/peer"
	"git.parallelcoin.io/dev/9/pkg/peer/addrmgr"
	"git.parallelcoin.io/dev/9/pkg/peer/connmgr"
	"git.parallelcoin.io/dev/9/pkg/util"
	"git.parallelcoin.io/dev/9/pkg/util/bloom"
	cl "git.parallelcoin.io/dev/9/pkg/util/cl"
	"git.parallelcoin.io/dev/9/pkg/util/interrupt"
)
// broadcastInventoryAdd is a type used to declare that the InvVect it contains needs to be added to the rebroadcast map
type broadcastInventoryAdd relayMsg
// broadcastInventoryDel is a type used to declare that the InvVect it contains needs to be removed from the rebroadcast map
type broadcastInventoryDel *wire.InvVect
// broadcastMsg provides the ability to house a bitcoin message to be broadcast to all connected peers except specified excluded peers.
type broadcastMsg struct {
	message      wire.Message
	excludePeers []*serverPeer
}
// cfHeaderKV is a tuple of a filter header and its associated block hash. The struct is used to cache cfcheckpt responses.
type cfHeaderKV struct {
	blockHash    chainhash.Hash
	filterHeader chainhash.Hash
}
// checkpointSorter implements sort.Interface to allow a slice of checkpoints to be sorted.
type checkpointSorter []chaincfg.Checkpoint
type connectNodeMsg struct {
	addr      string
	permanent bool
	reply     chan error
}
type disconnectNodeMsg struct {
	cmp   func(*serverPeer) bool
	reply chan error
}
type getAddedNodesMsg struct {
	reply chan []*serverPeer
}
type getConnCountMsg struct {
	reply chan int32
}
type getOutboundGroup struct {
	key   string
	reply chan int
}
type getPeersMsg struct {
	reply chan []*serverPeer
}
// onionAddr implements the net.Addr interface and represents a tor address.
type onionAddr struct {
	addr string
}
// peerState maintains state of inbound, persistent, outbound peers as well as banned peers and outbound groups.
type peerState struct {
	inboundPeers    map[int32]*serverPeer
	outboundPeers   map[int32]*serverPeer
	persistentPeers map[int32]*serverPeer
	banned          map[string]time.Time
	outboundGroups  map[string]int
}
// relayMsg packages an inventory vector along with the newly discovered inventory so the relay has access to that information.
type relayMsg struct {
	invVect *wire.InvVect
	data    interface{}
}
type removeNodeMsg struct {
	cmp   func(*serverPeer) bool
	reply chan error
}
// server provides a bitcoin server for handling communications to and from bitcoin peers.
type server struct {
	// The following variables must only be used atomically. Putting the uint64s first makes them 64-bit aligned for 32-bit systems.
	bytesReceived uint64 // Total bytes received from all peers since start.
	bytesSent     uint64 // Total bytes sent by all peers since start.
	started       int32
	shutdown      int32
	shutdownSched int32
	startupTime   int64
	chainParams   *chaincfg.Params
	addrManager   *addrmgr.AddrManager
	connManager   *connmgr.ConnManager
	sigCache      *txscript.SigCache
	hashCache     *txscript.HashCache
	rpcServers    []*rpcServer
	syncManager   *netsync.SyncManager
	chain         *blockchain.BlockChain
	txMemPool     *mempool.TxPool
	cpuMiner      *cpuminer.CPUMiner
	// minerController      *controller.Controller
	modifyRebroadcastInv chan interface{}
	newPeers             chan *serverPeer
	donePeers            chan *serverPeer
	banPeers             chan *serverPeer
	query                chan interface{}
	relayInv             chan relayMsg
	broadcast            chan broadcastMsg
	peerHeightsUpdate    chan updatePeerHeightsMsg
	wg                   sync.WaitGroup
	quit                 chan struct{}
	nat                  NAT
	db                   database.DB
	timeSource           blockchain.MedianTimeSource
	services             wire.ServiceFlag
	// The following fields are used for optional indexes.  They will be nil if the associated index is not enabled.  These fields are set during initial creation of the server and never changed afterwards, so they do not need to be protected for concurrent access.
	txIndex   *indexers.TxIndex
	addrIndex *indexers.AddrIndex
	cfIndex   *indexers.CfIndex
	// The fee estimator keeps track of how long transactions are left in the mempool before they are mined into blocks.
	feeEstimator *mempool.FeeEstimator
	// cfCheckptCaches stores a cached slice of filter headers for cfcheckpt messages for each filter type.
	cfCheckptCaches    map[wire.FilterType][]cfHeaderKV
	cfCheckptCachesMtx sync.RWMutex
	algo               string
	numthreads         uint32
}
// serverPeer extends the peer to maintain state shared by the server and the blockmanager.
type serverPeer struct {
	// The following variables must only be used atomically
	feeFilter int64
	*peer.Peer
	connReq        *connmgr.ConnReq
	server         *server
	persistent     bool
	continueHash   *chainhash.Hash
	relayMtx       sync.Mutex
	disableRelayTx bool
	sentAddrs      bool
	isWhitelisted  bool
	filter         *bloom.Filter
	knownAddresses map[string]struct{}
	banScore       connmgr.DynamicBanScore
	quit           chan struct{}
	// The following chans are used to sync blockmanager and server.
	txProcessed    chan struct{}
	blockProcessed chan struct{}
}
// simpleAddr implements the net.Addr interface with two struct fields
type simpleAddr struct {
	net, addr string
}
// updatePeerHeightsMsg is a message sent from the blockmanager to the server after a new block has been accepted. The purpose of the message is to update the heights of peers that were known to announce the block before we connected it to the main chain or recognized it as an orphan. With these updates, peer heights will be kept up to date, allowing for fresh data when selecting sync peer candidacy.
type updatePeerHeightsMsg struct {
	newHash    *chainhash.Hash
	newHeight  int32
	originPeer *peer.Peer
}
// defaultServices describes the default services that are supported by the server.
const defaultServices = wire.SFNodeNetwork | wire.SFNodeBloom |
	wire.SFNodeWitness | wire.SFNodeCF
// defaultRequiredServices describes the default services that are required to be supported by outbound peers.
const defaultRequiredServices = wire.SFNodeNetwork
// defaultTargetOutbound is the default number of outbound peers to target.
const defaultTargetOutbound = 9
// connectionRetryInterval is the base amount of time to wait in between retries when connecting to persistent peers.  It is adjusted by the number of retries such that there is a retry backoff.
const connectionRetryInterval = time.Second
// Ensure simpleAddr implements the net.Addr interface.
var _ net.Addr = simpleAddr{}
// userAgentName is the user agent name and is used to help identify ourselves to peers.
var userAgentName = "pod"
// userAgentVersion is the user agent version and is used to help identify ourselves to peers.
var userAgentVersion = fmt.Sprintf("%d.%d.%d", appMajor, appMinor, appPatch)
// zeroHash is the zero value hash (all zeros).  It is defined as a convenience.
var zeroHash chainhash.Hash
// Network returns "onion". This is part of the net.Addr interface.
func (
	oa *onionAddr,
) Network() string {
	return "onion"
}
// String returns the onion address. This is part of the net.Addr interface.
func (
	oa *onionAddr,
) String() string {
	return oa.addr
}
// Count returns the count of all known peers.
func (
	ps *peerState,
) Count() int {
	return len(ps.inboundPeers) + len(ps.outboundPeers) +
		len(ps.persistentPeers)
}
// forAllOutboundPeers is a helper function that runs closure on all outbound peers known to peerState.
func (
	ps *peerState,
) forAllOutboundPeers(
	closure func(sp *serverPeer)) {
	for _, e := range ps.outboundPeers {
		closure(e)
	}
	for _, e := range ps.persistentPeers {
		closure(e)
	}
}
// forAllPeers is a helper function that runs closure on all peers known to peerState.
func (
	ps *peerState,
) forAllPeers(
	closure func(sp *serverPeer)) {
	for _, e := range ps.inboundPeers {
		closure(e)
	}
	ps.forAllOutboundPeers(closure)
}
// AddBytesReceived adds the passed number of bytes to the total bytes received counter for the server.  It is safe for concurrent access.
func (
	s *server,
) AddBytesReceived(
	bytesReceived uint64) {
	atomic.AddUint64(&s.bytesReceived, bytesReceived)
}
// AddBytesSent adds the passed number of bytes to the total bytes sent counter for the server.  It is safe for concurrent access.
func (
	s *server,
) AddBytesSent(
	bytesSent uint64) {
	atomic.AddUint64(&s.bytesSent, bytesSent)
}
// AddPeer adds a new peer that has already been connected to the server.
func (
	s *server,
) AddPeer(
	sp *serverPeer) {
	s.newPeers <- sp
}
// AddRebroadcastInventory adds 'iv' to the list of inventories to be rebroadcasted at random intervals until they show up in a block.
func (
	s *server,
) AddRebroadcastInventory(
	iv *wire.InvVect, data interface{}) {
	// Ignore if shutting down.
	if atomic.LoadInt32(&s.shutdown) != 0 {
		return
	}
	s.modifyRebroadcastInv <- broadcastInventoryAdd{invVect: iv, data: data}
}
// AnnounceNewTransactions generates and relays inventory vectors and notifies both websocket and getblocktemplate long poll clients of the passed transactions.  This function should be called whenever new transactions are added to the mempool.
func (
	s *server,
) AnnounceNewTransactions(
	txns []*mempool.TxDesc) {
	// Generate and relay inventory vectors for all newly accepted transactions.
	s.relayTransactions(txns)
	// Notify both websocket and getblocktemplate long poll clients of all newly accepted transactions.
	for i := range s.rpcServers {
		if s.rpcServers[i] != nil {
			s.rpcServers[i].NotifyNewTransactions(txns)
		}
	}
}
// BanPeer bans a peer that has already been connected to the server by ip.
func (
	s *server,
) BanPeer(
	sp *serverPeer) {
	s.banPeers <- sp
}
// BroadcastMessage sends msg to all peers currently connected to the server except those in the passed peers to exclude.
func (
	s *server,
) BroadcastMessage(
	msg wire.Message, exclPeers ...*serverPeer) {
	// XXX: Need to determine if this is an alert that has already been broadcast and refrain from broadcasting again.
	bmsg := broadcastMsg{message: msg, excludePeers: exclPeers}
	s.broadcast <- bmsg
}
// ConnectedCount returns the number of currently connected peers.
func (
	s *server,
) ConnectedCount() int32 {
	replyChan := make(chan int32)
	s.query <- getConnCountMsg{reply: replyChan}
	return <-replyChan
}
// NetTotals returns the sum of all bytes received and sent across the network for all peers.  It is safe for concurrent access.
func (
	s *server,
) NetTotals() (uint64, uint64) {
	return atomic.LoadUint64(&s.bytesReceived),
		atomic.LoadUint64(&s.bytesSent)
}
// OutboundGroupCount returns the number of peers connected to the given outbound group key.
func (
	s *server,
) OutboundGroupCount(
	key string) int {
	replyChan := make(chan int)
	s.query <- getOutboundGroup{key: key, reply: replyChan}
	return <-replyChan
}
// RelayInventory relays the passed inventory vector to all connected peers that are not already known to have it.
func (
	s *server,
) RelayInventory(
	invVect *wire.InvVect, data interface{}) {
	s.relayInv <- relayMsg{invVect: invVect, data: data}
}
// RemoveRebroadcastInventory removes 'iv' from the list of items to be rebroadcasted if present.
func (
	s *server,
) RemoveRebroadcastInventory(
	iv *wire.InvVect) {
	// Log<-cl.Debug{emoveBroadcastInventory"
	// Ignore if shutting down.
	if atomic.LoadInt32(&s.shutdown) != 0 {
		// Log<-cl.Debug{gnoring due to shutdown"
		return
	}
	s.modifyRebroadcastInv <- broadcastInventoryDel(iv)
}
// ScheduleShutdown schedules a server shutdown after the specified duration. It also dynamically adjusts how often to warn the server is going down based on remaining duration.
func (
	s *server,
) ScheduleShutdown(
	duration time.Duration) {
	// Don't schedule shutdown more than once.
	if atomic.AddInt32(&s.shutdownSched, 1) != 1 {
		return
	}
	log <- cl.Warnf{"server shutdown in %v", duration}
	go func() {
		remaining := duration
		tickDuration := dynamicTickDuration(remaining)
		done := time.After(remaining)
		ticker := time.NewTicker(tickDuration)
	out:
		for {
			select {
			case <-done:
				// fmt.Println("chan:<-done")
				ticker.Stop()
				s.Stop()
				break out
			case <-ticker.C:
				// fmt.Println("chan:<-ticker.C")
				remaining -= -tickDuration
				if remaining < time.Second {
					continue
				}
				// Change tick duration dynamically based on remaining time.
				newDuration := dynamicTickDuration(remaining)
				if tickDuration != newDuration {
					tickDuration = newDuration
					ticker.Stop()
					ticker = time.NewTicker(tickDuration)
				}
				log <- cl.Warnf{"server shutdown in %v", remaining}
			}
		}
	}()
}
// Start begins accepting connections from peers.
func (s *server) Start() {
	// Already started?
	if atomic.AddInt32(&s.started, 1) != 1 {
		return
	}
	log <- cl.Trace{"starting server"}
	// Server startup time. Used for the uptime command for uptime calculation.
	s.startupTime = time.Now().Unix()
	// Start the peer handler which in turn starts the address and block managers.
	s.wg.Add(1)
	go s.peerHandler()
	if s.nat != nil {
		s.wg.Add(1)
		go s.upnpUpdateThread()
	}
	if !*Cfg.DisableRPC {
		s.wg.Add(1)
		// Log<-cl.Debug{tarting rebroadcast handler"
		// Start the rebroadcastHandler, which ensures user tx received by the RPC server are rebroadcast until being included in a block.
		go s.rebroadcastHandler()
		for i := range s.rpcServers {
			s.rpcServers[i].Start()
		}
	} else {
		panic("cannot run without RPC")
	}
	// Start the CPU miner if generation is enabled.
	if *Cfg.Generate {
		s.cpuMiner.Start()
	}
	// if *Cfg.MinerListener != "" {
	// 	s.minerController.Start()
	// }
}
// Stop gracefully shuts down the server by stopping and disconnecting all peers and the main listener.
func (
	s *server,
) Stop() error {
	// Make sure this only happens once.
	if atomic.AddInt32(&s.shutdown, 1) != 1 {
		log <- cl.Infof{"server is already in the process of shutting down"}
		return nil
	}
	log <- cl.Wrn("server shutting down")
	// Stop the CPU miner if needed
	s.cpuMiner.Stop()
	// Stop miner controller if needed
	// s.minerController.Stop()
	// Shutdown the RPC server if it's not disabled.
	if !*Cfg.DisableRPC {
		for i := range s.rpcServers {
			s.rpcServers[i].Stop()
		}
	}
	// Save fee estimator state in the database.
	s.db.Update(func(tx database.Tx) error {
		metadata := tx.Metadata()
		metadata.Put(mempool.EstimateFeeDatabaseKey, s.feeEstimator.Save())
		return nil
	})
	// Signal the remaining goroutines to quit.
	close(s.quit)
	return nil
}
// Transaction has one confirmation on the main chain. Now we can mark it as no longer needing rebroadcasting.
func (
	s *server,
) TransactionConfirmed(
	tx *util.Tx) {
	// Log<-cl.Debug{ransactionConfirmed"
	// Rebroadcasting is only necessary when the RPC server is active.
	for i := range s.rpcServers {
		// Log<-cl.Debug{ending to RPC servers"
		if s.rpcServers[i] == nil {
			return
		}
	}
	// Log<-cl.Debug{etting new inventory vector"
	iv := wire.NewInvVect(wire.InvTypeTx, tx.Hash())
	// Log<-cl.Debug{emoving broadcast inventory"
	s.RemoveRebroadcastInventory(iv)
	// Log<-cl.Debug{one TransactionConfirmed"
}
// UpdatePeerHeights updates the heights of all peers who have have announced the latest connected main chain block, or a recognized orphan. These height updates allow us to dynamically refresh peer heights, ensuring sync peer selection has access to the latest block heights for each peer.
func (
	s *server,
) UpdatePeerHeights(
	latestBlkHash *chainhash.Hash, latestHeight int32, updateSource *peer.Peer) {
	s.peerHeightsUpdate <- updatePeerHeightsMsg{
		newHash:    latestBlkHash,
		newHeight:  latestHeight,
		originPeer: updateSource,
	}
}
// WaitForShutdown blocks until the main listener and peer handlers are stopped.
func (
	s *server,
) WaitForShutdown() {
	s.wg.Wait()
}
// handleAddPeerMsg deals with adding new peers.  It is invoked from the peerHandler goroutine.
func (
	s *server,
) handleAddPeerMsg(
	state *peerState, sp *serverPeer) bool {
	if sp == nil {
		return false
	}
	// Ignore new peers if we're shutting down.
	if atomic.LoadInt32(&s.shutdown) != 0 {
		log <- cl.Infof{
			"new peer %s ignored - server is shutting down", sp,
		}
		sp.Disconnect()
		return false
	}
	// Disconnect banned peers.
	host, _, err := net.SplitHostPort(sp.Addr())
	if err != nil {
		log <- cl.Debug{"can't split hostport", err}
		sp.Disconnect()
		return false
	}
	if banEnd, ok := state.banned[host]; ok {
		if time.Now().Before(banEnd) {
			log <- cl.Debugf{
				"peer %s is banned for another %v - disconnecting",
				host, time.Until(banEnd),
			}
			sp.Disconnect()
			return false
		}
		log <- cl.Infof{"peer %s is no longer banned", host}
		delete(state.banned, host)
	}
	// TODO: Check for max peers from a single IP. Limit max number of total peers.
	if state.Count() >= *Cfg.MaxPeers {
		log <- cl.Infof{
			"max peers reached [%d] - disconnecting peer %s",
			Cfg.MaxPeers, sp,
		}
		sp.Disconnect()
		// TODO: how to handle permanent peers here? they should be rescheduled.
		return false
	}
	// Add the new peer and start it.
	log <- cl.Debug{"new peer", sp}
	if sp.Inbound() {
		state.inboundPeers[sp.ID()] = sp
	} else {
		state.outboundGroups[addrmgr.GroupKey(sp.NA())]++
		if sp.persistent {
			state.persistentPeers[sp.ID()] = sp
		} else {
			state.outboundPeers[sp.ID()] = sp
		}
	}
	return true
}
// handleBanPeerMsg deals with banning peers.  It is invoked from the peerHandler goroutine.
func (
	s *server,
) handleBanPeerMsg(
	state *peerState, sp *serverPeer) {
	host, _, err := net.SplitHostPort(sp.Addr())
	if err != nil {
		log <- cl.Debugf{"can't split ban peer %s %v", sp.Addr(), err}
		return
	}
	direction := directionString(sp.Inbound())
	log <- cl.Infof{
		"banned peer %s (%s) for %v", host, direction, *Cfg.BanDuration,
	}
	state.banned[host] = time.Now().Add(*Cfg.BanDuration)
}
// handleBroadcastMsg deals with broadcasting messages to peers.  It is invoked from the peerHandler goroutine.
func (
	s *server,
) handleBroadcastMsg(
	state *peerState, bmsg *broadcastMsg) {
	state.forAllPeers(func(sp *serverPeer) {
		if !sp.Connected() {
			return
		}
		for _, ep := range bmsg.excludePeers {
			if sp == ep {
				return
			}
		}
		sp.QueueMessage(bmsg.message, nil)
	})
}
// handleDonePeerMsg deals with peers that have signalled they are done.  It is invoked from the peerHandler goroutine.
func (
	s *server,
) handleDonePeerMsg(
	state *peerState, sp *serverPeer) {
	var list map[int32]*serverPeer
	if sp.persistent {
		list = state.persistentPeers
	} else if sp.Inbound() {
		list = state.inboundPeers
	} else {
		list = state.outboundPeers
	}
	if _, ok := list[sp.ID()]; ok {
		if !sp.Inbound() && sp.VersionKnown() {
			state.outboundGroups[addrmgr.GroupKey(sp.NA())]--
		}
		if !sp.Inbound() && sp.connReq != nil {
			s.connManager.Disconnect(sp.connReq.ID())
		}
		delete(list, sp.ID())
		log <- cl.Debug{"removed peer", sp}
		return
	}
	if sp.connReq != nil {
		s.connManager.Disconnect(sp.connReq.ID())
	}
	// Update the address' last seen time if the peer has acknowledged our version and has sent us its version as well.
	if sp.VerAckReceived() && sp.VersionKnown() && sp.NA() != nil {
		s.addrManager.Connected(sp.NA())
	}
	// If we get here it means that either we didn't know about the peer or we purposefully deleted it.
}
// handleQuery is the central handler for all queries and commands from other goroutines related to peer state.
func (s *server) handleQuery(state *peerState, querymsg interface{}) {
	switch msg := querymsg.(type) {
	case getConnCountMsg:
		nconnected := int32(0)
		state.forAllPeers(func(sp *serverPeer) {
			if sp.Connected() {
				nconnected++
			}
		})
		msg.reply <- nconnected
	case getPeersMsg:
		peers := make([]*serverPeer, 0, state.Count())
		state.forAllPeers(func(sp *serverPeer) {
			if !sp.Connected() {
				return
			}
			peers = append(peers, sp)
		})
		msg.reply <- peers
	case connectNodeMsg:
		// TODO: duplicate oneshots? Limit max number of total peers.
		if state.Count() >= *Cfg.MaxPeers {
			msg.reply <- errors.New("max peers reached")
			return
		}
		for _, peer := range state.persistentPeers {
			if peer.Addr() == msg.addr {
				if msg.permanent {
					msg.reply <- errors.New("peer already connected")
				} else {
					msg.reply <- errors.New("peer exists as a permanent peer")
				}
				return
			}
		}
		netAddr, err := addrStringToNetAddr(msg.addr)
		if err != nil {
			msg.reply <- err
			return
		}
		// TODO: if too many, nuke a non-perm peer.
		go s.connManager.Connect(&connmgr.ConnReq{
			Addr:      netAddr,
			Permanent: msg.permanent,
		})
		msg.reply <- nil
	case removeNodeMsg:
		found := disconnectPeer(state.persistentPeers, msg.cmp, func(sp *serverPeer) {
			// Keep group counts ok since we remove from the list now.
			state.outboundGroups[addrmgr.GroupKey(sp.NA())]--
		})
		if found {
			msg.reply <- nil
		} else {
			msg.reply <- errors.New("peer not found")
		}
	case getOutboundGroup:
		count, ok := state.outboundGroups[msg.key]
		if ok {
			msg.reply <- count
		} else {
			msg.reply <- 0
		}
	// Request a list of the persistent (added) peers.
	case getAddedNodesMsg:
		// Respond with a slice of the relevant peers.
		peers := make([]*serverPeer, 0, len(state.persistentPeers))
		for _, sp := range state.persistentPeers {
			peers = append(peers, sp)
		}
		msg.reply <- peers
	case disconnectNodeMsg:
		// Check inbound peers. We pass a nil callback since we don't require any additional actions on disconnect for inbound peers.
		found := disconnectPeer(state.inboundPeers, msg.cmp, nil)
		if found {
			msg.reply <- nil
			return
		}
		// Check outbound peers.
		found = disconnectPeer(state.outboundPeers, msg.cmp, func(sp *serverPeer) {
			// Keep group counts ok since we remove from the list now.
			state.outboundGroups[addrmgr.GroupKey(sp.NA())]--
		})
		if found {
			// If there are multiple outbound connections to the same ip:port, continue disconnecting them all until no such peers are found.
			for found {
				found = disconnectPeer(state.outboundPeers, msg.cmp, func(sp *serverPeer) {
					state.outboundGroups[addrmgr.GroupKey(sp.NA())]--
				})
			}
			msg.reply <- nil
			return
		}
		msg.reply <- errors.New("peer not found")
	}
}
// handleRelayInvMsg deals with relaying inventory to peers that are not already known to have it.  It is invoked from the peerHandler goroutine.
func (
	s *server,
) handleRelayInvMsg(
	state *peerState, msg relayMsg) {
	state.forAllPeers(func(sp *serverPeer) {
		if !sp.Connected() {
			return
		}
		// If the inventory is a block and the peer prefers headers, generate and send a headers message instead of an inventory message.
		if msg.invVect.Type == wire.InvTypeBlock && sp.WantsHeaders() {
			blockHeader, ok := msg.data.(wire.BlockHeader)
			if !ok {
				log <- cl.Wrn("underlying data for headers is not a block header")
				return
			}
			msgHeaders := wire.NewMsgHeaders()
			if err := msgHeaders.AddBlockHeader(&blockHeader); err != nil {
				log <- cl.Error{"failed to add block header:", err}
				return
			}
			sp.QueueMessage(msgHeaders, nil)
			return
		}
		if msg.invVect.Type == wire.InvTypeTx {
			// Don't relay the transaction to the peer when it has transaction relaying disabled.
			if sp.relayTxDisabled() {
				return
			}
			txD, ok := msg.data.(*mempool.TxDesc)
			if !ok {
				log <- cl.Warnf{
					"underlying data for tx inv relay is not a *mempool.TxDesc: %T",
					msg.data,
				}
				return
			}
			// Don't relay the transaction if the transaction fee-per-kb is less than the peer's feefilter.
			feeFilter := atomic.LoadInt64(&sp.feeFilter)
			if feeFilter > 0 && txD.FeePerKB < feeFilter {
				return
			}
			// Don't relay the transaction if there is a bloom filter loaded and the transaction doesn't match it.
			if sp.filter.IsLoaded() {
				if !sp.filter.MatchTxAndUpdate(txD.Tx) {
					return
				}
			}
		}
		// Queue the inventory to be relayed with the next batch. It will be ignored if the peer is already known to have the inventory.
		sp.QueueInventory(msg.invVect)
	})
}
// handleUpdatePeerHeight updates the heights of all peers who were known to announce a block we recently accepted.
func (
	s *server,
) handleUpdatePeerHeights(
	state *peerState, umsg updatePeerHeightsMsg) {
	state.forAllPeers(func(sp *serverPeer) {
		// The origin peer should already have the updated height.
		if sp.Peer == umsg.originPeer {
			return
		}
		// This is a pointer to the underlying memory which doesn't change.
		latestBlkHash := sp.LastAnnouncedBlock()
		// Skip this peer if it hasn't recently announced any new blocks.
		if latestBlkHash == nil {
			return
		}
		// If the peer has recently announced a block, and this block matches our newly accepted block, then update their block height.
		if *latestBlkHash == *umsg.newHash {
			sp.UpdateLastBlockHeight(umsg.newHeight)
			sp.UpdateLastAnnouncedBlock(nil)
		}
	})
}
// inboundPeerConnected is invoked by the connection manager when a new inbound connection is established.  It initializes a new inbound server peer instance, associates it with the connection, and starts a goroutine to wait for disconnection.
func (
	s *server,
) inboundPeerConnected(
	conn net.Conn) {
	sp := newServerPeer(s, false)
	sp.isWhitelisted = isWhitelisted(conn.RemoteAddr())
	sp.Peer = peer.NewInboundPeer(newPeerConfig(sp))
	sp.AssociateConnection(conn)
	go s.peerDoneHandler(sp)
}
// outboundPeerConnected is invoked by the connection manager when a new outbound connection is established.  It initializes a new outbound server peer instance, associates it with the relevant state such as the connection request instance and the connection itself, and finally notifies the address manager of the attempt.
func (
	s *server,
) outboundPeerConnected(
	c *connmgr.ConnReq, conn net.Conn) {
	sp := newServerPeer(s, c.Permanent)
	p, err := peer.NewOutboundPeer(newPeerConfig(sp), c.Addr.String())
	if err != nil {
		log <- cl.Debugf{"cannot create outbound peer %s: %v", c.Addr, err}
		s.connManager.Disconnect(c.ID())
	}
	sp.Peer = p
	sp.connReq = c
	sp.isWhitelisted = isWhitelisted(conn.RemoteAddr())
	sp.AssociateConnection(conn)
	go s.peerDoneHandler(sp)
	s.addrManager.Attempt(sp.NA())
}
// peerDoneHandler handles peer disconnects by notifiying the server that it's done along with other performing other desirable cleanup.
func (
	s *server,
) peerDoneHandler(
	sp *serverPeer) {
	sp.WaitForDisconnect()
	s.donePeers <- sp
	// Only tell sync manager we are gone if we ever told it we existed.
	if sp.VersionKnown() {
		s.syncManager.DonePeer(sp.Peer)
		// Evict any remaining orphans that were sent by the peer.
		numEvicted := s.txMemPool.RemoveOrphansByTag(mempool.Tag(sp.ID()))
		if numEvicted > 0 {
			log <- cl.Debugf{
				"Evicted %d %s from peer %v (id %d)",
				numEvicted, pickNoun(numEvicted, "orphan", "orphans"),
				sp, sp.ID(),
			}
		}
	}
	close(sp.quit)
}
// peerHandler is used to handle peer operations such as adding and removing peers to and from the server, banning peers, and broadcasting messages to peers.  It must be run in a goroutine.
func (
	s *server,
) peerHandler() {
	// Start the address manager and sync manager, both of which are needed by peers.  This is done here since their lifecycle is closely tied to this handler and rather than adding more channels to sychronize things, it's easier and slightly faster to simply start and stop them in this handler.
	s.addrManager.Start()
	s.syncManager.Start()
	log <- cl.Trc("starting peer handler")
	state := &peerState{
		inboundPeers:    make(map[int32]*serverPeer),
		persistentPeers: make(map[int32]*serverPeer),
		outboundPeers:   make(map[int32]*serverPeer),
		banned:          make(map[string]time.Time),
		outboundGroups:  make(map[string]int),
	}
	if !*Cfg.DisableDNSSeed {
		log <- cl.Trc("seeding from DNS")
		// Add peers discovered through DNS to the address manager.
		connmgr.SeedFromDNS(ActiveNetParams.Params, defaultRequiredServices,
			podLookup, func(addrs []*wire.NetAddress) {
				// Bitcoind uses a lookup of the dns seeder here. This is rather strange since the values looked up by the DNS seed lookups will vary quite a lot. to replicate this behaviour we put all addresses as having come from the first one.
				s.addrManager.AddAddresses(addrs, addrs[0])
			})
	}
	log <- cl.Trc("starting connmgr")
	go s.connManager.Start()
out:
	for {
		select {
		// New peers connected to the server.
		case p := <-s.newPeers:
			// fmt.Println("chan:p := <-s.newPeers")
			s.handleAddPeerMsg(state, p)
		// Disconnected peers.
		case p := <-s.donePeers:
			// fmt.Println("chan:p := <-s.donePeers")
			s.handleDonePeerMsg(state, p)
		// Block accepted in mainchain or orphan, update peer height.
		case umsg := <-s.peerHeightsUpdate:
			// fmt.Println("chan:umsg := <-s.peerHeightsUpdate")
			s.handleUpdatePeerHeights(state, umsg)
		// Peer to ban.
		case p := <-s.banPeers:
			// fmt.Println("chan:p := <-s.banPeers")
			s.handleBanPeerMsg(state, p)
		// New inventory to potentially be relayed to other peers.
		case invMsg := <-s.relayInv:
			// fmt.Println("chan:invMsg := <-s.relayInv")
			s.handleRelayInvMsg(state, invMsg)
		// Message to broadcast to all connected peers except those which are excluded by the message.
		case bmsg := <-s.broadcast:
			// fmt.Println("chan:bmsg := <-s.broadcast")
			s.handleBroadcastMsg(state, &bmsg)
		case qmsg := <-s.query:
			// fmt.Println("chan:qmsg := <-s.query")
			s.handleQuery(state, qmsg)
		case <-s.quit:
			// fmt.Println("chan:<-s.quit")
			// Disconnect all peers on server shutdown.
			state.forAllPeers(func(sp *serverPeer) {
				log <- cl.Tracef{"shutdown peer %s", sp}
				sp.Disconnect()
			})
			break out
		}
	}
	s.connManager.Stop()
	s.syncManager.Stop()
	s.addrManager.Stop()
	// Drain channels before exiting so nothing is left waiting around to send.
cleanup:
	for {
		select {
		case <-s.newPeers:
		case <-s.donePeers:
		case <-s.peerHeightsUpdate:
		case <-s.relayInv:
		case <-s.broadcast:
		case <-s.query:
		default:
			break cleanup
		}
	}
	s.wg.Done()
	log <- cl.Tracef{"peer handler done"}
}
// pushBlockMsg sends a block message for the provided block hash to the connected peer.  An error is returned if the block hash is not known.
func (
	s *server,
) pushBlockMsg(
	sp *serverPeer, hash *chainhash.Hash, doneChan chan<- struct {
	},
	waitChan <-chan struct{}, encoding wire.MessageEncoding) error {
	// Fetch the raw block bytes from the database.
	var blockBytes []byte
	err := sp.server.db.View(func(dbTx database.Tx) error {
		var err error
		blockBytes, err = dbTx.FetchBlock(hash)
		return err
	})
	if err != nil {
		log <- cl.Tracef{
			"unable to fetch requested block hash %v: %v",
			hash, err,
		}
		if doneChan != nil {
			doneChan <- struct{}{}
		}
		return err
	}
	// Deserialize the block.
	var msgBlock wire.MsgBlock
	err = msgBlock.Deserialize(bytes.NewReader(blockBytes))
	if err != nil {
		log <- cl.Tracef{
			"unable to deserialize requested block hash %v: %v",
			hash, err,
		}
		if doneChan != nil {
			doneChan <- struct{}{}
		}
		return err
	}
	// Once we have fetched data wait for any previous operation to finish.
	if waitChan != nil {
		<-waitChan
	}
	// We only send the channel for this message if we aren't sending an inv straight after.
	var dc chan<- struct{}
	continueHash := sp.continueHash
	sendInv := continueHash != nil && continueHash.IsEqual(hash)
	if !sendInv {
		dc = doneChan
	}
	sp.QueueMessageWithEncoding(&msgBlock, dc, encoding)
	// When the peer requests the final block that was advertised in response to a getblocks message which requested more blocks than would fit into a single message, send it a new inventory message to trigger it to issue another getblocks message for the next batch of inventory.
	if sendInv {
		best := sp.server.chain.BestSnapshot()
		invMsg := wire.NewMsgInvSizeHint(1)
		iv := wire.NewInvVect(wire.InvTypeBlock, &best.Hash)
		invMsg.AddInvVect(iv)
		sp.QueueMessage(invMsg, doneChan)
		sp.continueHash = nil
	}
	return nil
}
// pushMerkleBlockMsg sends a merkleblock message for the provided block hash to the connected peer.  Since a merkle block requires the peer to have a filter loaded, this call will simply be ignored if there is no filter loaded.  An error is returned if the block hash is not known.
func (
	s *server,
) pushMerkleBlockMsg(
	sp *serverPeer, hash *chainhash.Hash,
	doneChan chan<- struct{}, waitChan <-chan struct{}, encoding wire.MessageEncoding) error {
	// Do not send a response if the peer doesn't have a filter loaded.
	if !sp.filter.IsLoaded() {
		if doneChan != nil {
			doneChan <- struct{}{}
		}
		return nil
	}
	// Fetch the raw block bytes from the database.
	blk, err := sp.server.chain.BlockByHash(hash)
	if err != nil {
		log <- cl.Tracef{
			"unable to fetch requested block hash %v: %v",
			hash, err,
		}
		if doneChan != nil {
			doneChan <- struct{}{}
		}
		return err
	}
	// Generate a merkle block by filtering the requested block according to the filter for the peer.
	merkle, matchedTxIndices := bloom.NewMerkleBlock(blk, sp.filter)
	// Once we have fetched data wait for any previous operation to finish.
	if waitChan != nil {
		<-waitChan
	}
	// Send the merkleblock.  Only send the done channel with this message if no transactions will be sent afterwards.
	var dc chan<- struct{}
	if len(matchedTxIndices) == 0 {
		dc = doneChan
	}
	sp.QueueMessage(merkle, dc)
	// Finally, send any matched transactions.
	blkTransactions := blk.MsgBlock().Transactions
	for i, txIndex := range matchedTxIndices {
		// Only send the done channel on the final transaction.
		var dc chan<- struct{}
		if i == len(matchedTxIndices)-1 {
			dc = doneChan
		}
		if txIndex < uint32(len(blkTransactions)) {
			sp.QueueMessageWithEncoding(blkTransactions[txIndex], dc,
				encoding)
		}
	}
	return nil
}
// pushTxMsg sends a tx message for the provided transaction hash to the connected peer.  An error is returned if the transaction hash is not known.
func (
	s *server,
) pushTxMsg(
	sp *serverPeer, hash *chainhash.Hash, doneChan chan<- struct {
	},
	waitChan <-chan struct{}, encoding wire.MessageEncoding) error {
	// Attempt to fetch the requested transaction from the pool.  A call could be made to check for existence first, but simply trying to fetch a missing transaction results in the same behavior.
	tx, err := s.txMemPool.FetchTransaction(hash)
	if err != nil {
		log <- cl.Tracef{
			"unable to fetch tx %v from transaction pool: %v", hash, err,
		}
		if doneChan != nil {
			doneChan <- struct{}{}
		}
		return err
	}
	// Once we have fetched data wait for any previous operation to finish.
	if waitChan != nil {
		<-waitChan
	}
	sp.QueueMessageWithEncoding(tx.MsgTx(), doneChan, encoding)
	return nil
}
// rebroadcastHandler keeps track of user submitted inventories that we have sent out but have not yet made it into a block. We periodically rebroadcast them in case our peers restarted or otherwise lost track of them.
func (
	s *server,
) rebroadcastHandler() {
	// Log<-cl.Debug{tarting rebroadcastHandler"
	// Wait 5 min before first tx rebroadcast.
	timer := time.NewTimer(5 * time.Minute)
	pendingInvs := make(map[wire.InvVect]interface{})
out:
	for {
		select {
		case riv := <-s.modifyRebroadcastInv:
			// fmt.Println("chan:riv := <-s.modifyRebroadcastInv")
			// Log<-cl.Debug{eceived modify rebroadcast inventory"
			switch msg := riv.(type) {
			// Incoming InvVects are added to our map of RPC txs.
			case broadcastInventoryAdd:
				// Log<-cl.Debug{roadcast inventory add"
				pendingInvs[*msg.invVect] = msg.data
			// When an InvVect has been added to a block, we can now remove it, if it was present.
			case broadcastInventoryDel:
				// Log<-cl.Debug{roadcast inventory delete"
				if _, ok := pendingInvs[*msg]; ok {
					delete(pendingInvs, *msg)
				}
			}
		case <-timer.C:
			// fmt.Println("chan:<-timer.C")
			// Any inventory we have has not made it into a block yet. We periodically resubmit them until they have.
			for iv, data := range pendingInvs {
				ivCopy := iv
				s.RelayInventory(&ivCopy, data)
			}
			// Process at a random time up to 30mins (in seconds) in the future.
			timer.Reset(time.Second *
				time.Duration(randomUint16Number(1800)))
		case <-s.quit:
			// fmt.Println("chan:<-s.quit")
			break out
			// default:
		}
	}
	timer.Stop()
	// Drain channels before exiting so nothing is left waiting around to send.
cleanup:
	for {
		select {
		case <-s.modifyRebroadcastInv:
		default:
			break cleanup
		}
	}
	s.wg.Done()
}
// relayTransactions generates and relays inventory vectors for all of the passed transactions to all connected peers.
func (
	s *server,
) relayTransactions(
	txns []*mempool.TxDesc) {
	for _, txD := range txns {
		iv := wire.NewInvVect(wire.InvTypeTx, txD.Tx.Hash())
		s.RelayInventory(iv, txD)
	}
}
func (
	s *server,
) upnpUpdateThread() {
	// Go off immediately to prevent code duplication, thereafter we renew lease every 15 minutes.
	timer := time.NewTimer(0 * time.Second)
	lport, _ := strconv.ParseInt(ActiveNetParams.DefaultPort, 10, 16)
	first := true
out:
	for {
		select {
		case <-timer.C:
			// TODO: pick external port  more cleverly
			// TODO: know which ports we are listening to on an external net.
			// TODO: if specific listen port doesn't work then ask for wildcard
			// listen port?
			// XXX this assumes timeout is in seconds.
			listenPort, err := s.nat.AddPortMapping("tcp", int(lport), int(lport),
				"pod listen port", 20*60)
			if err != nil {
				log <- cl.Warnf{"can't add UPnP port mapping: %v", err}
			}
			if first && err == nil {
				// TODO: look this up periodically to see if upnp domain changed and so did ip.
				externalip, err := s.nat.GetExternalAddress()
				if err != nil {
					log <- cl.Warnf{"UPnP can't get external address: %v", err}
					continue out
				}
				na := wire.NewNetAddressIPPort(externalip, uint16(listenPort),
					s.services)
				err = s.addrManager.AddLocalAddress(na, addrmgr.UpnpPrio)
				if err != nil {
					// XXX DeletePortMapping?
				}
				log <- cl.Warnf{"successfully bound via UPnP to %s", addrmgr.NetAddressKey(na)}
				first = false
			}
			timer.Reset(time.Minute * 15)
		case <-s.quit:
			fmt.Println("<-s.quit")
			break out
		}
	}
	timer.Stop()
	if err := s.nat.DeletePortMapping("tcp", int(lport), int(lport)); err != nil {
		log <- cl.Warnf{"unable to remove UPnP port mapping: %v", err}
	} else {
		log <- cl.Debugf{"successfully disestablished UPnP port mapping"}
	}
	s.wg.Done()
}
// OnAddr is invoked when a peer receives an addr bitcoin message and is used to notify the server about advertised addresses.
func (
	sp *serverPeer,
) OnAddr(
	_ *peer.Peer, msg *wire.MsgAddr) {
	// Ignore addresses when running on the simulation test network.  This helps prevent the network from becoming another public test network since it will not be able to learn about other peers that have not specifically been provided.
	if *Cfg.SimNet {
		return
	}
	// Ignore old style addresses which don't include a timestamp.
	if sp.ProtocolVersion() < wire.NetAddressTimeVersion {
		return
	}
	// A message that has no addresses is invalid.
	if len(msg.AddrList) == 0 {
		log <- cl.Errorf{
			"command [%s] from %s does not contain any addresses",
			msg.Command(), sp.Peer,
		}
		sp.Disconnect()
		return
	}
	for _, na := range msg.AddrList {
		// Don't add more address if we're disconnecting.
		if !sp.Connected() {
			return
		}
		// Set the timestamp to 5 days ago if it's more than 24 hours in the future so this address is one of the first to be removed when space is needed.
		now := time.Now()
		if na.Timestamp.After(now.Add(time.Minute * 10)) {
			na.Timestamp = now.Add(-1 * time.Hour * 24 * 5)
		}
		// Add address to known addresses for this peer.
		sp.addKnownAddresses([]*wire.NetAddress{na})
	}
	// Add addresses to server address manager.  The address manager handles the details of things such as preventing duplicate addresses, max addresses, and last seen updates. XXX bitcoind gives a 2 hour time penalty here, do we want to do the same?
	sp.server.addrManager.AddAddresses(msg.AddrList, sp.NA())
}
// OnBlock is invoked when a peer receives a block bitcoin message.  It blocks until the bitcoin block has been fully processed.
func (
	sp *serverPeer,
) OnBlock(
	_ *peer.Peer, msg *wire.MsgBlock, buf []byte) {
	// Convert the raw MsgBlock to a util.Block which provides some convenience methods and things such as hash caching.
	block := util.NewBlockFromBlockAndBytes(msg, buf)
	// Add the block to the known inventory for the peer.
	iv := wire.NewInvVect(wire.InvTypeBlock, block.Hash())
	sp.AddKnownInventory(iv)
	// Queue the block up to be handled by the block manager and intentionally block further receives until the bitcoin block is fully processed and known good or bad.  This helps prevent a malicious peer from queuing up a bunch of bad blocks before disconnecting (or being disconnected) and wasting memory.  Additionally, this behavior is depended on by at least the block acceptance test tool as the reference implementation processes blocks in the same thread and therefore blocks further messages until the bitcoin block has been fully processed.
	sp.server.syncManager.QueueBlock(block, sp.Peer, sp.blockProcessed)
	<-sp.blockProcessed
}
// OnFeeFilter is invoked when a peer receives a feefilter bitcoin message and is used by remote peers to request that no transactions which have a fee rate lower than provided value are inventoried to them.  The peer will be disconnected if an invalid fee filter value is provided.
func (
	sp *serverPeer,
) OnFeeFilter(
	_ *peer.Peer, msg *wire.MsgFeeFilter) {
	// Check that the passed minimum fee is a valid amount.
	if msg.MinFee < 0 || msg.MinFee > util.MaxSatoshi {
		log <- cl.Debugf{
			"peer %v sent an invalid feefilter '%v' -- disconnecting",
			sp, util.Amount(msg.MinFee)}
		sp.Disconnect()
		return
	}
	atomic.StoreInt64(&sp.feeFilter, msg.MinFee)
}
// OnFilterAdd is invoked when a peer receives a filteradd bitcoin message and is used by remote peers to add data to an already loaded bloom filter.  The peer will be disconnected if a filter is not loaded when this message is received or the server is not configured to allow bloom filters.
func (
	sp *serverPeer,
) OnFilterAdd(
	_ *peer.Peer, msg *wire.MsgFilterAdd) {
	// Disconnect and/or ban depending on the node bloom services flag and negotiated protocol version.
	if !sp.enforceNodeBloomFlag(msg.Command()) {
		return
	}
	if !sp.filter.IsLoaded() {
		log <- cl.Debugf{
			"%s sent a filteradd request with no filter loaded -- disconnecting", sp,
		}
		sp.Disconnect()
		return
	}
	sp.filter.Add(msg.Data)
}
// OnFilterClear is invoked when a peer receives a filterclear bitcoin message and is used by remote peers to clear an already loaded bloom filter. The peer will be disconnected if a filter is not loaded when this message is received  or the server is not configured to allow bloom filters.
func (
	sp *serverPeer,
) OnFilterClear(
	_ *peer.Peer, msg *wire.MsgFilterClear) {
	// Disconnect and/or ban depending on the node bloom services flag and negotiated protocol version.
	if !sp.enforceNodeBloomFlag(msg.Command()) {
		return
	}
	if !sp.filter.IsLoaded() {
		log <- cl.Debugf{
			"%s sent a filterclear request with no filter loaded -- disconnecting", sp,
		}
		sp.Disconnect()
		return
	}
	sp.filter.Unload()
}
// OnFilterLoad is invoked when a peer receives a filterload bitcoin message and it used to load a bloom filter that should be used for delivering merkle blocks and associated transactions that match the filter. The peer will be disconnected if the server is not configured to allow bloom filters.
func (
	sp *serverPeer,
) OnFilterLoad(
	_ *peer.Peer, msg *wire.MsgFilterLoad) {
	// Disconnect and/or ban depending on the node bloom services flag and negotiated protocol version.
	if !sp.enforceNodeBloomFlag(msg.Command()) {
		return
	}
	sp.setDisableRelayTx(false)
	sp.filter.Reload(msg)
}
// OnGetAddr is invoked when a peer receives a getaddr bitcoin message and is used to provide the peer with known addresses from the address manager.
func (
	sp *serverPeer,
) OnGetAddr(
	_ *peer.Peer, msg *wire.MsgGetAddr) {
	// Don't return any addresses when running on the simulation test network.  This helps prevent the network from becoming another public test network since it will not be able to learn about other peers that have not specifically been provided.
	if *Cfg.SimNet {
		return
	}
	// Do not accept getaddr requests from outbound peers.  This reduces fingerprinting attacks.
	if !sp.Inbound() {
		log <- cl.Debug{"ignoring getaddr request from outbound peer", sp}
		return
	}
	// Only allow one getaddr request per connection to discourage address stamping of inv announcements.
	if sp.sentAddrs {
		log <- cl.Debugf{"ignoring repeated getaddr request from peer", sp}
		return
	}
	sp.sentAddrs = true
	// Get the current known addresses from the address manager.
	addrCache := sp.server.addrManager.AddressCache()
	// Push the addresses.
	sp.pushAddrMsg(addrCache)
}
// OnGetBlocks is invoked when a peer receives a getblocks bitcoin message.
func (
	sp *serverPeer,
) OnGetBlocks(
	_ *peer.Peer, msg *wire.MsgGetBlocks) {
	// Find the most recent known block in the best chain based on the block locator and fetch all of the block hashes after it until either wire.MaxBlocksPerMsg have been fetched or the provided stop hash is encountered. Use the block after the genesis block if no other blocks in the provided locator are known.  This does mean the client will start over with the genesis block if unknown block locators are provided. This mirrors the behavior in the reference implementation.
	chain := sp.server.chain
	hashList := chain.LocateBlocks(msg.BlockLocatorHashes, &msg.HashStop,
		wire.MaxBlocksPerMsg)
	// Generate inventory message.
	invMsg := wire.NewMsgInv()
	for i := range hashList {
		iv := wire.NewInvVect(wire.InvTypeBlock, &hashList[i])
		invMsg.AddInvVect(iv)
	}
	// Send the inventory message if there is anything to send.
	if len(invMsg.InvList) > 0 {
		invListLen := len(invMsg.InvList)
		if invListLen == wire.MaxBlocksPerMsg {
			// Intentionally use a copy of the final hash so there is not a reference into the inventory slice which would prevent the entire slice from being eligible for GC as soon as it's sent.
			continueHash := invMsg.InvList[invListLen-1].Hash
			sp.continueHash = &continueHash
		}
		sp.QueueMessage(invMsg, nil)
	}
}
// OnGetCFCheckpt is invoked when a peer receives a getcfcheckpt bitcoin message.
func (
	sp *serverPeer,
) OnGetCFCheckpt(
	_ *peer.Peer, msg *wire.MsgGetCFCheckpt) {
	// Ignore getcfcheckpt requests if not in sync.
	if !sp.server.syncManager.IsCurrent() {
		return
	}
	// We'll also ensure that the remote party is requesting a set of checkpoints for filters that we actually currently maintain.
	switch msg.FilterType {
	case wire.GCSFilterRegular:
		break
	default:
		log <- cl.Debugf{
			"filter request for unknown checkpoints for filter:", msg.FilterType,
		}
		return
	}
	// Now that we know the client is fetching a filter that we know of, we'll fetch the block hashes et each check point interval so we can compare against our cache, and create new check points if necessary.
	blockHashes, err := sp.server.chain.IntervalBlockHashes(
		&msg.StopHash, wire.CFCheckptInterval,
	)
	if err != nil {
		log <- cl.Debug{"invalid getcfilters request:", err}
		return
	}
	checkptMsg := wire.NewMsgCFCheckpt(
		msg.FilterType, &msg.StopHash, len(blockHashes),
	)
	// Fetch the current existing cache so we can decide if we need to extend it or if its adequate as is.
	sp.server.cfCheckptCachesMtx.RLock()
	checkptCache := sp.server.cfCheckptCaches[msg.FilterType]
	// If the set of block hashes is beyond the current size of the cache, then we'll expand the size of the cache and also retain the write lock.
	var updateCache bool
	if len(blockHashes) > len(checkptCache) {
		// Now that we know we'll need to modify the size of the cache, we'll release the read lock and grab the write lock to possibly expand the cache size.
		sp.server.cfCheckptCachesMtx.RUnlock()
		sp.server.cfCheckptCachesMtx.Lock()
		defer sp.server.cfCheckptCachesMtx.Unlock()
		// Now that we have the write lock, we'll check again as it's possible that the cache has already been expanded.
		checkptCache = sp.server.cfCheckptCaches[msg.FilterType]
		// If we still need to expand the cache, then We'll mark that we need to update the cache for below and also expand the size of the cache in place.
		if len(blockHashes) > len(checkptCache) {
			updateCache = true
			additionalLength := len(blockHashes) - len(checkptCache)
			newEntries := make([]cfHeaderKV, additionalLength)
			log <- cl.Infof{
				"growing size of checkpoint cache from %v to %v block hashes",
				len(checkptCache), len(blockHashes),
			}
			checkptCache = append(
				sp.server.cfCheckptCaches[msg.FilterType],
				newEntries...,
			)
		}
	} else {
		// Otherwise, we'll hold onto the read lock for the remainder of this method.
		defer sp.server.cfCheckptCachesMtx.RUnlock()
		log <- cl.Tracef{
			"serving stale cache of size %v", len(checkptCache),
		}
	}
	// Now that we know the cache is of an appropriate size, we'll iterate backwards until the find the block hash. We do this as it's possible a re-org has occurred so items in the db are now in the main china while the cache has been partially invalidated.
	var forkIdx int
	for forkIdx = len(blockHashes); forkIdx > 0; forkIdx-- {
		if checkptCache[forkIdx-1].blockHash == blockHashes[forkIdx-1] {
			break
		}
	}
	// Now that we know the how much of the cache is relevant for this query, we'll populate our check point message with the cache as is. Shortly below, we'll populate the new elements of the cache.
	for i := 0; i < forkIdx; i++ {
		checkptMsg.AddCFHeader(&checkptCache[i].filterHeader)
	}
	// We'll now collect the set of hashes that are beyond our cache so we can look up the filter headers to populate the final cache.
	blockHashPtrs := make([]*chainhash.Hash, 0, len(blockHashes)-forkIdx)
	for i := forkIdx; i < len(blockHashes); i++ {
		blockHashPtrs = append(blockHashPtrs, &blockHashes[i])
	}
	filterHeaders, err := sp.server.cfIndex.FilterHeadersByBlockHashes(
		blockHashPtrs, msg.FilterType,
	)
	if err != nil {
		log <- cl.Error{"error retrieving cfilter headers:", err}
		return
	}
	// Now that we have the full set of filter headers, we'll add them to the checkpoint message, and also update our cache in line.
	for i, filterHeaderBytes := range filterHeaders {
		if len(filterHeaderBytes) == 0 {
			log <- cl.Warn{
				"could not obtain CF header for", blockHashPtrs[i],
			}
			return
		}
		filterHeader, err := chainhash.NewHash(filterHeaderBytes)
		if err != nil {
			log <- cl.Warn{
				"committed filter header deserialize failed:", err,
			}
			return
		}
		checkptMsg.AddCFHeader(filterHeader)
		// If the new main chain is longer than what's in the cache, then we'll override it beyond the fork point.
		if updateCache {
			checkptCache[forkIdx+i] = cfHeaderKV{
				blockHash:    blockHashes[forkIdx+i],
				filterHeader: *filterHeader,
			}
		}
	}
	// Finally, we'll update the cache if we need to, and send the final message back to the requesting peer.
	if updateCache {
		sp.server.cfCheckptCaches[msg.FilterType] = checkptCache
	}
	sp.QueueMessage(checkptMsg, nil)
}
// OnGetCFHeaders is invoked when a peer receives a getcfheader bitcoin message.
func (
	sp *serverPeer,
) OnGetCFHeaders(
	_ *peer.Peer, msg *wire.MsgGetCFHeaders) {
	// Ignore getcfilterheader requests if not in sync.
	if !sp.server.syncManager.IsCurrent() {
		return
	}
	// We'll also ensure that the remote party is requesting a set of headers for filters that we actually currently maintain.
	switch msg.FilterType {
	case wire.GCSFilterRegular:
		break
	default:
		log <- cl.Debug{
			"filter request for unknown headers for filter:", msg.FilterType,
		}
		return
	}
	startHeight := int32(msg.StartHeight)
	maxResults := wire.MaxCFHeadersPerMsg
	// If StartHeight is positive, fetch the predecessor block hash so we can populate the PrevFilterHeader field.
	if msg.StartHeight > 0 {
		startHeight--
		maxResults++
	}
	// Fetch the hashes from the block index.
	hashList, err := sp.server.chain.HeightToHashRange(
		startHeight, &msg.StopHash, maxResults,
	)
	if err != nil {
		log <- cl.Debug{
			"invalid getcfheaders request:", err,
		}
	}
	// This is possible if StartHeight is one greater that the height of StopHash, and we pull a valid range of hashes including the previous filter header.
	if len(hashList) == 0 || (msg.StartHeight > 0 && len(hashList) == 1) {
		log <- cl.Dbg("no results for getcfheaders request")
		return
	}
	// Create []*chainhash.Hash from []chainhash.Hash to pass to FilterHeadersByBlockHashes.
	hashPtrs := make([]*chainhash.Hash, len(hashList))
	for i := range hashList {
		hashPtrs[i] = &hashList[i]
	}
	// Fetch the raw filter hash bytes from the database for all blocks.
	filterHashes, err := sp.server.cfIndex.FilterHashesByBlockHashes(
		hashPtrs, msg.FilterType,
	)
	if err != nil {
		log <- cl.Error{"error retrieving cfilter hashes:", err}
		return
	}
	// Generate cfheaders message and send it.
	headersMsg := wire.NewMsgCFHeaders()
	// Populate the PrevFilterHeader field.
	if msg.StartHeight > 0 {
		prevBlockHash := &hashList[0]
		// Fetch the raw committed filter header bytes from the database.
		headerBytes, err := sp.server.cfIndex.FilterHeaderByBlockHash(
			prevBlockHash, msg.FilterType)
		if err != nil {
			log <- cl.Error{"error retrieving CF header:", err}
			return
		}
		if len(headerBytes) == 0 {
			log <- cl.Warn{"could not obtain CF header for", prevBlockHash}
			return
		}
		// Deserialize the hash into PrevFilterHeader.
		err = headersMsg.PrevFilterHeader.SetBytes(headerBytes)
		if err != nil {
			log <- cl.Warn{
				"committed filter header deserialize failed:", err,
			}
			return
		}
		hashList = hashList[1:]
		filterHashes = filterHashes[1:]
	}
	// Populate HeaderHashes.
	for i, hashBytes := range filterHashes {
		if len(hashBytes) == 0 {
			log <- cl.Warn{
				"could not obtain CF hash for", hashList[i],
			}
			return
		}
		// Deserialize the hash.
		filterHash, err := chainhash.NewHash(hashBytes)
		if err != nil {
			log <- cl.Warn{
				"committed filter hash deserialize failed:", err,
			}
			return
		}
		headersMsg.AddCFHash(filterHash)
	}
	headersMsg.FilterType = msg.FilterType
	headersMsg.StopHash = msg.StopHash
	sp.QueueMessage(headersMsg, nil)
}
// OnGetCFilters is invoked when a peer receives a getcfilters bitcoin message.
func (
	sp *serverPeer,
) OnGetCFilters(
	_ *peer.Peer, msg *wire.MsgGetCFilters) {
	// Ignore getcfilters requests if not in sync.
	if !sp.server.syncManager.IsCurrent() {
		return
	}
	// We'll also ensure that the remote party is requesting a set of filters that we actually currently maintain.
	switch msg.FilterType {
	case wire.GCSFilterRegular:
		break
	default:
		log <- cl.Debug{"filter request for unknown filter:", msg.FilterType}
		return
	}
	hashes, err := sp.server.chain.HeightToHashRange(
		int32(msg.StartHeight), &msg.StopHash, wire.MaxGetCFiltersReqRange,
	)
	if err != nil {
		log <- cl.Debug{"invalid getcfilters request:", err}
		return
	}
	// Create []*chainhash.Hash from []chainhash.Hash to pass to FiltersByBlockHashes.
	hashPtrs := make([]*chainhash.Hash, len(hashes))
	for i := range hashes {
		hashPtrs[i] = &hashes[i]
	}
	filters, err := sp.server.cfIndex.FiltersByBlockHashes(
		hashPtrs, msg.FilterType,
	)
	if err != nil {
		log <- cl.Error{"error retrieving cfilters:", err}
		return
	}
	for i, filterBytes := range filters {
		if len(filterBytes) == 0 {
			log <- cl.Warn{"could not obtain cfilter for", hashes[i]}
			return
		}
		filterMsg := wire.NewMsgCFilter(
			msg.FilterType, &hashes[i], filterBytes,
		)
		sp.QueueMessage(filterMsg, nil)
	}
}
// handleGetData is invoked when a peer receives a getdata bitcoin message and is used to deliver block and transaction information.
func (
	sp *serverPeer,
) OnGetData(
	_ *peer.Peer, msg *wire.MsgGetData) {
	numAdded := 0
	notFound := wire.NewMsgNotFound()
	length := len(msg.InvList)
	// A decaying ban score increase is applied to prevent exhausting resources with unusually large inventory queries. Requesting more than the maximum inventory vector length within a short period of time yields a score above the default ban threshold. Sustained bursts of small requests are not penalized as that would potentially ban peers performing IBD. This incremental score decays each minute to half of its value.
	sp.addBanScore(0, uint32(length)*99/wire.MaxInvPerMsg, "getdata")
	// We wait on this wait channel periodically to prevent queuing far more data than we can send in a reasonable time, wasting memory. The waiting occurs after the database fetch for the next one to provide a little pipelining.
	var waitChan chan struct{}
	doneChan := make(chan struct{}, 1)
	for i, iv := range msg.InvList {
		var c chan struct{}
		// If this will be the last message we send.
		if i == length-1 && len(notFound.InvList) == 0 {
			c = doneChan
		} else if (i+1)%3 == 0 {
			// Buffered so as to not make the send goroutine block.
			c = make(chan struct{}, 1)
		}
		var err error
		switch iv.Type {
		case wire.InvTypeWitnessTx:
			err = sp.server.pushTxMsg(sp, &iv.Hash, c, waitChan, wire.WitnessEncoding)
		case wire.InvTypeTx:
			err = sp.server.pushTxMsg(sp, &iv.Hash, c, waitChan, wire.BaseEncoding)
		case wire.InvTypeWitnessBlock:
			err = sp.server.pushBlockMsg(sp, &iv.Hash, c, waitChan, wire.WitnessEncoding)
		case wire.InvTypeBlock:
			err = sp.server.pushBlockMsg(sp, &iv.Hash, c, waitChan, wire.BaseEncoding)
		case wire.InvTypeFilteredWitnessBlock:
			err = sp.server.pushMerkleBlockMsg(sp, &iv.Hash, c, waitChan, wire.WitnessEncoding)
		case wire.InvTypeFilteredBlock:
			err = sp.server.pushMerkleBlockMsg(sp, &iv.Hash, c, waitChan, wire.BaseEncoding)
		default:
			log <- cl.Warn{"unknown type in inventory request", iv.Type}
			continue
		}
		if err != nil {
			notFound.AddInvVect(iv)
			// When there is a failure fetching the final entry and the done channel was sent in due to there being no outstanding not found inventory, consume it here because there is now not found inventory that will use the channel momentarily.
			if i == len(msg.InvList)-1 && c != nil {
				<-c
			}
		}
		numAdded++
		waitChan = c
	}
	if len(notFound.InvList) != 0 {
		sp.QueueMessage(notFound, doneChan)
	}
	// Wait for messages to be sent. We can send quite a lot of data at this point and this will keep the peer busy for a decent amount of time. We don't process anything else by them in this time so that we have an idea of when we should hear back from them - else the idle timeout could fire when we were only half done sending the blocks.
	if numAdded > 0 {
		<-doneChan
	}
}
// OnGetHeaders is invoked when a peer receives a getheaders bitcoin message.
func (
	sp *serverPeer,
) OnGetHeaders(
	_ *peer.Peer, msg *wire.MsgGetHeaders) {
	// Ignore getheaders requests if not in sync.
	if !sp.server.syncManager.IsCurrent() {
		return
	}
	// Find the most recent known block in the best chain based on the block locator and fetch all of the headers after it until either wire.MaxBlockHeadersPerMsg have been fetched or the provided stop hash is encountered. Use the block after the genesis block if no other blocks in the provided locator are known.  This does mean the client will start over with the genesis block if unknown block locators are provided. This mirrors the behavior in the reference implementation.
	chain := sp.server.chain
	headers := chain.LocateHeaders(msg.BlockLocatorHashes, &msg.HashStop)
	// Send found headers to the requesting peer.
	blockHeaders := make([]*wire.BlockHeader, len(headers))
	for i := range headers {
		blockHeaders[i] = &headers[i]
	}
	sp.QueueMessage(&wire.MsgHeaders{Headers: blockHeaders}, nil)
}
// OnHeaders is invoked when a peer receives a headers bitcoin message.  The message is passed down to the sync manager.
func (
	sp *serverPeer,
) OnHeaders(
	_ *peer.Peer, msg *wire.MsgHeaders) {
	sp.server.syncManager.QueueHeaders(msg, sp.Peer)
}
// OnInv is invoked when a peer receives an inv bitcoin message and is used to examine the inventory being advertised by the remote peer and react accordingly.  We pass the message down to blockmanager which will call QueueMessage with any appropriate responses.
func (
	sp *serverPeer,
) OnInv(
	_ *peer.Peer, msg *wire.MsgInv) {
	if !*Cfg.BlocksOnly {
		if len(msg.InvList) > 0 {
			sp.server.syncManager.QueueInv(msg, sp.Peer)
		}
		return
	}
	newInv := wire.NewMsgInvSizeHint(uint(len(msg.InvList)))
	for _, invVect := range msg.InvList {
		if invVect.Type == wire.InvTypeTx {
			log <- cl.Tracef{
				"ignoring tx %v in inv from %v -- blocksonly enabled",
				invVect.Hash, sp,
			}
			if sp.ProtocolVersion() >= wire.BIP0037Version {
				log <- cl.Infof{
					"peer %v is announcing transactions -- disconnecting", sp,
				}
				sp.Disconnect()
				return
			}
			continue
		}
		err := newInv.AddInvVect(invVect)
		if err != nil {
			log <- cl.Error{"failed to add inventory vector:", err}
			break
		}
	}
	if len(newInv.InvList) > 0 {
		sp.server.syncManager.QueueInv(newInv, sp.Peer)
	}
}
// OnMemPool is invoked when a peer receives a mempool bitcoin message. It creates and sends an inventory message with the contents of the memory pool up to the maximum inventory allowed per message.  When the peer has a bloom filter loaded, the contents are filtered accordingly.
func (
	sp *serverPeer,
) OnMemPool(
	_ *peer.Peer, msg *wire.MsgMemPool) {
	// Only allow mempool requests if the server has bloom filtering enabled.
	if sp.server.services&wire.SFNodeBloom != wire.SFNodeBloom {
		log <- cl.Debugf{
			"peer", sp, "sent mempool request with bloom filtering disabled -- disconnecting",
		}
		sp.Disconnect()
		return
	}
	// A decaying ban score increase is applied to prevent flooding. The ban score accumulates and passes the ban threshold if a burst of mempool messages comes from a peer. The score decays each minute to half of its value.
	sp.addBanScore(0, 33, "mempool")
	// Generate inventory message with the available transactions in the transaction memory pool.  Limit it to the max allowed inventory per message.  The NewMsgInvSizeHint function automatically limits the passed hint to the maximum allowed, so it's safe to pass it without double checking it here.
	txMemPool := sp.server.txMemPool
	txDescs := txMemPool.TxDescs()
	invMsg := wire.NewMsgInvSizeHint(uint(len(txDescs)))
	for _, txDesc := range txDescs {
		// Either add all transactions when there is no bloom filter, or only the transactions that match the filter when there is one.
		if !sp.filter.IsLoaded() || sp.filter.MatchTxAndUpdate(txDesc.Tx) {
			iv := wire.NewInvVect(wire.InvTypeTx, txDesc.Tx.Hash())
			invMsg.AddInvVect(iv)
			if len(invMsg.InvList)+1 > wire.MaxInvPerMsg {
				break
			}
		}
	}
	// Send the inventory message if there is anything to send.
	if len(invMsg.InvList) > 0 {
		sp.QueueMessage(invMsg, nil)
	}
}
// OnRead is invoked when a peer receives a message and it is used to update the bytes received by the server.
func (
	sp *serverPeer,
) OnRead(
	_ *peer.Peer, bytesRead int, msg wire.Message, err error) {
	sp.server.AddBytesReceived(uint64(bytesRead))
}
// OnTx is invoked when a peer receives a tx bitcoin message.  It blocks until the bitcoin transaction has been fully processed.  Unlock the block handler this does not serialize all transactions through a single thread transactions don't rely on the previous one in a linear fashion like blocks.
func (
	sp *serverPeer,
) OnTx(
	_ *peer.Peer, msg *wire.MsgTx) {
	if *Cfg.BlocksOnly {
		log <- cl.Tracef{
			"ignoring tx %v from %v - blocksonly enabled",
			msg.TxHash(), sp,
		}
		return
	}
	// Add the transaction to the known inventory for the peer. Convert the raw MsgTx to a util.Tx which provides some convenience methods and things such as hash caching.
	tx := util.NewTx(msg)
	iv := wire.NewInvVect(wire.InvTypeTx, tx.Hash())
	sp.AddKnownInventory(iv)
	// Queue the transaction up to be handled by the sync manager and intentionally block further receives until the transaction is fully processed and known good or bad.  This helps prevent a malicious peer from queuing up a bunch of bad transactions before disconnecting (or being disconnected) and wasting memory.
	sp.server.syncManager.QueueTx(tx, sp.Peer, sp.txProcessed)
	<-sp.txProcessed
}
// OnVersion is invoked when a peer receives a version bitcoin message and is used to negotiate the protocol version details as well as kick start the communications.
func (
	sp *serverPeer,
) OnVersion(
	_ *peer.Peer, msg *wire.MsgVersion) *wire.MsgReject {
	// Update the address manager with the advertised services for outbound connections in case they have changed.  This is not done for inbound connections to help prevent malicious behavior and is skipped when running on the simulation test network since it is only intended to connect to specified peers and actively avoids advertising and connecting to discovered peers. NOTE: This is done before rejecting peers that are too old to ensure it is updated regardless in the case a new minimum protocol version is enforced and the remote node has not upgraded yet.
	isInbound := sp.Inbound()
	remoteAddr := sp.NA()
	addrManager := sp.server.addrManager
	if !*Cfg.SimNet && !isInbound {
		addrManager.SetServices(remoteAddr, msg.Services)
	}
	// Ignore peers that have a protcol version that is too old.  The peer negotiation logic will disconnect it after this callback returns.
	if msg.ProtocolVersion < int32(peer.MinAcceptableProtocolVersion) {
		return nil
	}
	// Reject outbound peers that are not full nodes.
	wantServices := wire.SFNodeNetwork
	if !isInbound && !hasServices(msg.Services, wantServices) {
		missingServices := wantServices & ^msg.Services
		log <- cl.Debugf{
			"rejecting peer %s with services %v due to not providing desired services %v",
			sp.Peer, msg.Services, missingServices,
		}
		reason := fmt.Sprintf("required services %#x not offered",
			uint64(missingServices))
		return wire.NewMsgReject(msg.Command(), wire.RejectNonstandard, reason)
	}
	// Update the address manager and request known addresses from the remote peer for outbound connections.  This is skipped when running on the simulation test network since it is only intended to connect to specified peers and actively avoids advertising and connecting to discovered peers.
	if !*Cfg.SimNet && !isInbound {
		// After soft-fork activation, only make outbound connection to peers if they flag that they're segwit enabled.
		chain := sp.server.chain
		segwitActive, err := chain.IsDeploymentActive(chaincfg.DeploymentSegwit)
		if err != nil {
			log <- cl.Error{
				"unable to query for segwit soft-fork state:", err,
			}
			return nil
		}
		if segwitActive && !sp.IsWitnessEnabled() {
			log <- cl.Info{
				"disconnecting non-segwit peer", sp,
				"as it isn't segwit enabled and we need more segwit enabled peers",
			}
			sp.Disconnect()
			return nil
		}
		// Advertise the local address when the server accepts incoming connections and it believes itself to be close to the best known tip.
		if !*Cfg.DisableListen && sp.server.syncManager.IsCurrent() {
			// Get address that best matches.
			lna := addrManager.GetBestLocalAddress(remoteAddr)
			if addrmgr.IsRoutable(lna) {
				// Filter addresses the peer already knows about.
				addresses := []*wire.NetAddress{lna}
				sp.pushAddrMsg(addresses)
			}
		}
		// Request known addresses if the server address manager needs more and the peer has a protocol version new enough to include a timestamp with addresses.
		hasTimestamp := sp.ProtocolVersion() >= wire.NetAddressTimeVersion
		if addrManager.NeedMoreAddresses() && hasTimestamp {
			sp.QueueMessage(wire.NewMsgGetAddr(), nil)
		}
		// Mark the address as a known good address.
		addrManager.Good(remoteAddr)
	}
	// Add the remote peer time as a sample for creating an offset against the local clock to keep the network time in sync.
	sp.server.timeSource.AddTimeSample(sp.Addr(), msg.Timestamp)
	// Signal the sync manager this peer is a new sync candidate.
	sp.server.syncManager.NewPeer(sp.Peer)
	// Choose whether or not to relay transactions before a filter command is received.
	sp.setDisableRelayTx(msg.DisableRelayTx)
	// Add valid peer to the server.
	sp.server.AddPeer(sp)
	return nil
}
// OnWrite is invoked when a peer sends a message and it is used to update the bytes sent by the server.
func (
	sp *serverPeer,
) OnWrite(
	_ *peer.Peer, bytesWritten int, msg wire.Message, err error) {
	sp.server.AddBytesSent(uint64(bytesWritten))
}
// addBanScore increases the persistent and decaying ban score fields by the values passed as parameters. If the resulting score exceeds half of the ban threshold, a warning is logged including the reason provided. Further, if the score is above the ban threshold, the peer will be banned and disconnected.
func (
	sp *serverPeer,
) addBanScore(
	persistent, transient uint32, reason string) {
	// No warning is logged and no score is calculated if banning is disabled.
	if *Cfg.DisableBanning {
		return
	}
	if sp.isWhitelisted {
		log <- cl.Debugf{
			"misbehaving whitelisted peer %s: %s", sp, reason,
		}
		return
	}
	warnThreshold := *Cfg.BanThreshold >> 1
	if transient == 0 && persistent == 0 {
		// The score is not being increased, but a warning message is still logged if the score is above the warn threshold.
		score := sp.banScore.Int()
		if int(score) > warnThreshold {
			log <- cl.Warnf{
				"misbehaving peer %s: %s -- ban score is %d, it was not increased this time",
				sp, reason, score,
			}
		}
		return
	}
	score := sp.banScore.Increase(persistent, transient)
	if int(score) > warnThreshold {
		log <- cl.Warnf{
			"misbehaving peer %s: %s -- ban score increased to %d",
			sp, reason, score,
		}
		if int(score) > *Cfg.BanThreshold {
			log <- cl.Warnf{
				"misbehaving peer %s -- banning and disconnecting", sp,
			}
			sp.server.BanPeer(sp)
			sp.Disconnect()
		}
	}
}
// addKnownAddresses adds the given addresses to the set of known addresses to the peer to prevent sending duplicate addresses.
func (
	sp *serverPeer,
) addKnownAddresses(
	addresses []*wire.NetAddress) {
	for _, na := range addresses {
		sp.knownAddresses[addrmgr.NetAddressKey(na)] = struct{}{}
	}
}
// addressKnown true if the given address is already known to the peer.
func (
	sp *serverPeer,
) addressKnown(
	na *wire.NetAddress) bool {
	_, exists := sp.knownAddresses[addrmgr.NetAddressKey(na)]
	return exists
}
// enforceNodeBloomFlag disconnects the peer if the server is not configured to allow bloom filters.  Additionally, if the peer has negotiated to a protocol version  that is high enough to observe the bloom filter service support bit, it will be banned since it is intentionally violating the protocol.
func (
	sp *serverPeer,
) enforceNodeBloomFlag(
	cmd string) bool {
	if sp.server.services&wire.SFNodeBloom != wire.SFNodeBloom {
		// Ban the peer if the protocol version is high enough that the peer is knowingly violating the protocol and banning is enabled. NOTE: Even though the addBanScore function already examines whether or not banning is enabled, it is checked here as well to ensure the violation is logged and the peer is disconnected regardless.
		if sp.ProtocolVersion() >= wire.BIP0111Version &&
			!*Cfg.DisableBanning {
			// Disconnect the peer regardless of whether it was banned.
			sp.addBanScore(100, 0, cmd)
			sp.Disconnect()
			return false
		}
		// Disconnect the peer regardless of protocol version or banning state.
		log <- cl.Debugf{
			"%s sent an unsupported %s request -- disconnecting", sp, cmd,
		}
		sp.Disconnect()
		return false
	}
	return true
}
// newestBlock returns the current best block hash and height using the format required by the configuration for the peer package.
func (
	sp *serverPeer,
) newestBlock() (*chainhash.Hash, int32, error) {
	best := sp.server.chain.BestSnapshot()
	return &best.Hash, best.Height, nil
}
// pushAddrMsg sends an addr message to the connected peer using the provided addresses.
func (
	sp *serverPeer,
) pushAddrMsg(
	addresses []*wire.NetAddress) {
	// Filter addresses already known to the peer.
	addrs := make([]*wire.NetAddress, 0, len(addresses))
	for _, addr := range addresses {
		if !sp.addressKnown(addr) {
			addrs = append(addrs, addr)
		}
	}
	known, err := sp.PushAddrMsg(addrs)
	if err != nil {
		log <- cl.Errorf{
			"can't push address message to %s: %v", sp.Peer, err,
		}
		sp.Disconnect()
		return
	}
	sp.addKnownAddresses(known)
}
// relayTxDisabled returns whether or not relaying of transactions for the given peer is disabled. It is safe for concurrent access.
func (
	sp *serverPeer,
) relayTxDisabled() bool {
	sp.relayMtx.Lock()
	isDisabled := sp.disableRelayTx
	sp.relayMtx.Unlock()
	return isDisabled
}
// setDisableRelayTx toggles relaying of transactions for the given peer. It is safe for concurrent access.
func (
	sp *serverPeer,
) setDisableRelayTx(
	disable bool) {
	sp.relayMtx.Lock()
	sp.disableRelayTx = disable
	sp.relayMtx.Unlock()
}
// Len returns the number of checkpoints in the slice.  It is part of the sort.Interface implementation.
func (
	s checkpointSorter,
) Len() int {
	return len(s)
}
/*	Less returns whether the checkpoint with index i should sort before the
	checkpoint with index j.  It is part of the sort.Interface implementation. */
func (
	s checkpointSorter,
) Less(
	i, j int) bool {
	return s[i].Height < s[j].Height
}
// Swap swaps the checkpoints at the passed indices.  It is part of the sort.Interface implementation.
func (
	s checkpointSorter,
) Swap(
	i, j int) {
	s[i], s[j] = s[j], s[i]
}
// Network returns the network. This is part of the net.Addr interface.
func (
	a simpleAddr,
) Network() string {
	return a.net
}
// String returns the address. This is part of the net.Addr interface.
func (
	a simpleAddr,
) String() string {
	return a.addr
}
/*	addLocalAddress adds an address that this node is listening on to the
	address manager so that it may be relayed to peers. */
func addLocalAddress(
	addrMgr *addrmgr.AddrManager, addr string, services wire.ServiceFlag) error {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return err
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsUnspecified() {
		// If bound to unspecified address, advertise all local interfaces
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			return err
		}
		for _, addr := range addrs {
			ifaceIP, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}
			/*	If bound to 0.0.0.0, do not add IPv6 interfaces and if bound to
				::, do not add IPv4 interfaces. */
			if (ip.To4() == nil) != (ifaceIP.To4() == nil) {
				continue
			}
			netAddr := wire.NewNetAddressIPPort(ifaceIP, uint16(port), services)
			addrMgr.AddLocalAddress(netAddr, addrmgr.BoundPrio)
		}
	} else {
		netAddr, err := addrMgr.HostToNetAddress(host, uint16(port), services)
		if err != nil {
			return err
		}
		addrMgr.AddLocalAddress(netAddr, addrmgr.BoundPrio)
	}
	return nil
}
/*	addrStringToNetAddr takes an address in the form of 'host:port' and returns
	a net.Addr which maps to the original address with any host names resolved
	to IP addresses.  It also handles tor addresses properly by returning a
	net.Addr that encapsulates the address. */
func addrStringToNetAddr(
	addr string) (net.Addr, error) {
	host, strPort, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(strPort)
	if err != nil {
		return nil, err
	}
	// Skip if host is already an IP address.
	if ip := net.ParseIP(host); ip != nil {
		return &net.TCPAddr{
				IP:   ip,
				Port: port,
			},
			nil
	}
	// Tor addresses cannot be resolved to an IP, so just return an onion address instead.
	if strings.HasSuffix(host, ".onion") {
		if !*Cfg.Onion {
			return nil, errors.New("tor has been disabled")
		}
		return &onionAddr{addr: addr}, nil
	}
	// Attempt to look up an IP address associated with the parsed host.
	ips, err := podLookup(host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("no addresses found for %s", host)
	}
	return &net.TCPAddr{
			IP:   ips[0],
			Port: port,
		},
		nil
}
// disconnectPeer attempts to drop the connection of a targeted peer in the passed peer list. Targets are identified via usage of the passed `compareFunc`, which should return `true` if the passed peer is the target peer. This function returns true on success and false if the peer is unable to be located. If the peer is found, and the passed callback: `whenFound' isn't nil, we call it with the peer as the argument before it is removed from the peerList, and is disconnected from the server.
func disconnectPeer(
	peerList map[int32]*serverPeer, compareFunc func(
		*serverPeer) bool, whenFound func(*serverPeer)) bool {
	for addr, peer := range peerList {
		if compareFunc(peer) {
			if whenFound != nil {
				whenFound(peer)
			}
			// This is ok because we are not continuing to iterate so won't corrupt the loop.
			delete(peerList, addr)
			peer.Disconnect()
			return true
		}
	}
	return false
}
/*	dynamicTickDuration is a convenience function used to dynamically choose a
	tick duration based on remaining time.  It is primarily used during
	server shutdown to make shutdown warnings more frequent as the shutdown time
	approaches. */
func dynamicTickDuration(
	remaining time.Duration) time.Duration {
	switch {
	case remaining <= time.Second*5:
		return time.Second
	case remaining <= time.Second*15:
		return time.Second * 5
	case remaining <= time.Minute:
		return time.Second * 15
	case remaining <= time.Minute*5:
		return time.Minute
	case remaining <= time.Minute*15:
		return time.Minute * 5
	case remaining <= time.Hour:
		return time.Minute * 15
	}
	return time.Hour
}
// hasServices returns whether or not the provided advertised service flags have all of the provided desired service flags set.
func hasServices(
	advertised, desired wire.ServiceFlag) bool {
	return advertised&desired == desired
}
/*	initListeners initializes the configured net listeners and adds any bound
	addresses to the address manager. Returns the listeners and a NAT interface,
	which is non-nil if UPnP is in use. */
func initListeners(
	amgr *addrmgr.AddrManager, listenAddrs []string, services wire.ServiceFlag) ([]net.Listener, NAT, error) {
	// Listen for TCP connections at the configured addresses
	netAddrs, err := parseListeners(listenAddrs)
	if err != nil {
		return nil, nil, err
	}
	listeners := make([]net.Listener, 0, len(netAddrs))
	for _, addr := range netAddrs {
		listener, err := net.Listen(addr.Network(), addr.String())
		if err != nil {
			log <- cl.Warnf{"can't listen on %s: %v", addr, err}
			continue
		}
		listeners = append(listeners, listener)
	}
	var nat NAT
	if Cfg.ExternalIPs != nil {
		defaultPort, err := strconv.ParseUint(ActiveNetParams.DefaultPort, 10, 16)
		if err != nil {
			log <- cl.Errorf{"can not parse default port %s for active chain: %v",
				ActiveNetParams.DefaultPort, err}
			return nil, nil, err
		}
		for _, sip := range *Cfg.ExternalIPs {
			eport := uint16(defaultPort)
			host, portstr, err := net.SplitHostPort(sip)
			if err != nil {
				// no port, use default.
				host = sip
			} else {
				port, err := strconv.ParseUint(portstr, 10, 16)
				if err != nil {
					log <- cl.Warnf{"can not parse port from %s for " +
						"externalip: %v", sip, err}
					continue
				}
				eport = uint16(port)
			}
			na, err := amgr.HostToNetAddress(host, eport, services)
			if err != nil {
				log <- cl.Warnf{"not adding %s as externalip: %v", sip, err}
				continue
			}
			err = amgr.AddLocalAddress(na, addrmgr.ManualPrio)
			if err != nil {
				log <- cl.Warnf{"skipping specified external IP: %v", err}
			}
		}
	} else {
		if *Cfg.Upnp {
			var err error
			nat, err = Discover()
			if err != nil {
				log <- cl.Warnf{"can't discover upnp: %v", err}
			}
			// nil nat here is fine, just means no upnp on network.
		}
		// Add bound addresses to address manager to be advertised to peers.
		for _, listener := range listeners {
			addr := listener.Addr().String()
			err := addLocalAddress(amgr, addr, services)
			if err != nil {
				log <- cl.Warnf{"skipping bound address %s: %v", addr, err}
			}
		}
	}
	return listeners, nat, nil
}
/*	isWhitelisted returns whether the IP address is included in the whitelisted
	networks and IPs. */
func isWhitelisted(
	addr net.Addr) bool {
	if len(StateCfg.ActiveWhitelists) == 0 {
		return false
	}
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		log <- cl.Warnf{"unable to SplitHostPort on '%s': %v", addr, err}
		return false
	}
	ip := net.ParseIP(host)
	if ip == nil {
		log <- cl.Warnf{"unable to parse IP '%s'", addr}
		return false
	}
	for _, ipnet := range StateCfg.ActiveWhitelists {
		if ipnet.Contains(ip) {
			return true
		}
	}
	return false
}
/*	mergeCheckpoints returns two slices of checkpoints merged into one slice
	such that the checkpoints are sorted by height.  In the case the additional
	checkpoints contain a checkpoint with the same height as a checkpoint in the
	default checkpoints, the additional checkpoint will take precedence and
	overwrite the default one. */
func mergeCheckpoints(
	defaultCheckpoints, additional []chaincfg.Checkpoint) []chaincfg.Checkpoint {
	/*	Create a map of the additional checkpoints to remove duplicates while
		leaving the most recently-specified checkpoint. */
	extra := make(map[int32]chaincfg.Checkpoint)
	for _, checkpoint := range additional {
		extra[checkpoint.Height] = checkpoint
	}
	// Add all default checkpoints that do not have an override in the additional checkpoints.
	numDefault := len(defaultCheckpoints)
	checkpoints := make([]chaincfg.Checkpoint, 0, numDefault+len(extra))
	for _, checkpoint := range defaultCheckpoints {
		if _, exists := extra[checkpoint.Height]; !exists {
			checkpoints = append(checkpoints, checkpoint)
		}
	}
	// Append the additional checkpoints and return the sorted results.
	for _, checkpoint := range extra {
		checkpoints = append(checkpoints, checkpoint)
	}
	sort.Sort(checkpointSorter(checkpoints))
	return checkpoints
}
// newPeerConfig returns the configuration for the given serverPeer.
func newPeerConfig(
	sp *serverPeer) *peer.Config {
	proxyaddress := new(string)
	if Cfg.Proxy != nil {
		*proxyaddress = *Cfg.Proxy
	}
	useragentcomments := new([]string)
	if Cfg.UserAgentComments != nil {
		*useragentcomments = *Cfg.UserAgentComments
	}
	return &peer.Config{
		Listeners: peer.MessageListeners{
			OnVersion:      sp.OnVersion,
			OnMemPool:      sp.OnMemPool,
			OnTx:           sp.OnTx,
			OnBlock:        sp.OnBlock,
			OnInv:          sp.OnInv,
			OnHeaders:      sp.OnHeaders,
			OnGetData:      sp.OnGetData,
			OnGetBlocks:    sp.OnGetBlocks,
			OnGetHeaders:   sp.OnGetHeaders,
			OnGetCFilters:  sp.OnGetCFilters,
			OnGetCFHeaders: sp.OnGetCFHeaders,
			OnGetCFCheckpt: sp.OnGetCFCheckpt,
			OnFeeFilter:    sp.OnFeeFilter,
			OnFilterAdd:    sp.OnFilterAdd,
			OnFilterClear:  sp.OnFilterClear,
			OnFilterLoad:   sp.OnFilterLoad,
			OnGetAddr:      sp.OnGetAddr,
			OnAddr:         sp.OnAddr,
			OnRead:         sp.OnRead,
			OnWrite:        sp.OnWrite,
			// Note: The reference client currently bans peers that send alerts not signed with its key.  We could verify against their key, but since the reference client is currently unwilling to support other implementations' alert messages, we will not relay theirs.
			OnAlert: nil,
		},
		NewestBlock:       sp.newestBlock,
		HostToNetAddress:  sp.server.addrManager.HostToNetAddress,
		Proxy:             *proxyaddress,
		UserAgentName:     userAgentName,
		UserAgentVersion:  userAgentVersion,
		UserAgentComments: *useragentcomments,
		ChainParams:       sp.server.chainParams,
		Services:          sp.server.services,
		DisableRelayTx:    *Cfg.BlocksOnly,
		ProtocolVersion:   peer.MaxProtocolVersion,
		TrickleInterval:   *Cfg.TrickleInterval,
	}
}
// newServer returns a new pod server configured to listen on addr for the bitcoin network type specified by chainParams.  Use start to begin accepting connections from peers.
func newServer(
	listenAddrs []string, db database.DB, chainParams *chaincfg.Params, interruptChan <-chan struct{}, algo string,
) (*server, error) {
	services := defaultServices
	if *Cfg.NoPeerBloomFilters {
		services &^= wire.SFNodeBloom
	}
	if *Cfg.NoCFilters {
		services &^= wire.SFNodeCF
	}
	amgr := addrmgr.New(filepath.Join(
		*Cfg.AppDataDir, NetName(ActiveNetParams)), podLookup)
	var listeners []net.Listener
	var nat NAT
	if !*Cfg.DisableListen {
		var err error
		listeners, nat, err = initListeners(amgr, listenAddrs, services)
		if err != nil {
			return nil, err
		}
		if len(listeners) == 0 {
			return nil, errors.New("no valid listen address")
		}
	}
	nthr := uint32(runtime.NumCPU())
	var thr uint32
	if *Cfg.GenThreads == -1 || thr > nthr {
		thr = uint32(nthr)
	} else {
		thr = uint32(*Cfg.GenThreads)
	}
	s := server{
		chainParams:          chainParams,
		addrManager:          amgr,
		newPeers:             make(chan *serverPeer, *Cfg.MaxPeers),
		donePeers:            make(chan *serverPeer, *Cfg.MaxPeers),
		banPeers:             make(chan *serverPeer, *Cfg.MaxPeers),
		query:                make(chan interface{}),
		relayInv:             make(chan relayMsg, *Cfg.MaxPeers),
		broadcast:            make(chan broadcastMsg, *Cfg.MaxPeers),
		quit:                 make(chan struct{}),
		modifyRebroadcastInv: make(chan interface{}),
		peerHeightsUpdate:    make(chan updatePeerHeightsMsg),
		nat:                  nat,
		db:                   db,
		timeSource:           blockchain.NewMedianTime(),
		services:             services,
		sigCache:             txscript.NewSigCache(uint(*Cfg.SigCacheMaxSize)),
		hashCache:            txscript.NewHashCache(uint(*Cfg.SigCacheMaxSize)),
		cfCheckptCaches:      make(map[wire.FilterType][]cfHeaderKV),
		numthreads:           thr,
		algo:                 algo,
	}
	// Create the transaction and address indexes if needed.
	// CAUTION: the txindex needs to be first in the indexes array because the addrindex uses data from the txindex during catchup.  If the addrindex is run first, it may not have the transactions from the current block indexed.
	var indexes []indexers.Indexer
	if *Cfg.TxIndex || *Cfg.AddrIndex {
		// Enable transaction index if address index is enabled since it requires it.
		if !*Cfg.TxIndex {
			log <- cl.Infof{
				"transaction index enabled because it is required by the address index"}
			*Cfg.TxIndex = true
		} else {
			log <- cl.Info{"transaction index is enabled"}
		}
		s.txIndex = indexers.NewTxIndex(db)
		indexes = append(indexes, s.txIndex)
	}
	if *Cfg.AddrIndex {
		log <- cl.Info{"address index is enabled"}
		s.addrIndex = indexers.NewAddrIndex(db, chainParams)
		indexes = append(indexes, s.addrIndex)
	}
	if !*Cfg.NoCFilters {
		log <- cl.Info{"committed filter index is enabled"}
		s.cfIndex = indexers.NewCfIndex(db, chainParams)
		indexes = append(indexes, s.cfIndex)
	}
	// Create an index manager if any of the optional indexes are enabled.
	var indexManager blockchain.IndexManager
	if len(indexes) > 0 {
		indexManager = indexers.NewManager(db, indexes)
	}
	// Merge given checkpoints with the default ones unless they are disabled.
	var checkpoints []chaincfg.Checkpoint
	if !*Cfg.DisableCheckpoints {
		checkpoints = mergeCheckpoints(
			s.chainParams.Checkpoints,
			StateCfg.AddedCheckpoints)
	}
	// Create a new block chain instance with the appropriate configuration.
	var err error
	s.chain, err = blockchain.New(
		&blockchain.Config{
			DB:           s.db,
			Interrupt:    interruptChan,
			ChainParams:  s.chainParams,
			Checkpoints:  checkpoints,
			TimeSource:   s.timeSource,
			SigCache:     s.sigCache,
			IndexManager: indexManager,
			HashCache:    s.hashCache,
		},
	)
	if err != nil {
		return nil, err
	}
	s.chain.DifficultyAdjustments = make(map[string]float64)
	// Search for a FeeEstimator state in the database. If none can be found or if it cannot be loaded, create a new one.
	e := db.Update(func(tx database.Tx) error {
		metadata := tx.Metadata()
		feeEstimationData := metadata.Get(mempool.EstimateFeeDatabaseKey)
		if feeEstimationData != nil {
			// delete it from the database so that we don't try to restore the same thing again somehow.
			e := metadata.Delete(mempool.EstimateFeeDatabaseKey)
			if e != nil {
				return e
			}
			// If there is an error, log it and make a new fee estimator.
			var err error
			s.feeEstimator, err = mempool.RestoreFeeEstimator(feeEstimationData)
			if err != nil {
				return fmt.Errorf("Failed to restore fee estimator %v", err)
			}
		}
		return nil
	})
	if e != nil {
		log <- cl.Error{e}
	}
	// If no feeEstimator has been found, or if the one that has been found is behind somehow, create a new one and start over.
	if s.feeEstimator == nil || s.feeEstimator.LastKnownHeight() != s.chain.BestSnapshot().Height {
		s.feeEstimator = mempool.NewFeeEstimator(
			mempool.DefaultEstimateFeeMaxRollback,
			mempool.DefaultEstimateFeeMinRegisteredBlocks,
		)
	}
	txC := mempool.Config{
		Policy: mempool.Policy{
			DisableRelayPriority: *Cfg.NoRelayPriority,
			AcceptNonStd:         *Cfg.RelayNonStd,
			FreeTxRelayLimit:     *Cfg.FreeTxRelayLimit,
			MaxOrphanTxs:         *Cfg.MaxOrphanTxs,
			MaxOrphanTxSize:      DefaultMaxOrphanTxSize,
			MaxSigOpCostPerTx:    blockchain.MaxBlockSigOpsCost / 4,
			MinRelayTxFee:        StateCfg.ActiveMinRelayTxFee,
			MaxTxVersion:         2,
		},
		ChainParams:   chainParams,
		FetchUtxoView: s.chain.FetchUtxoView,
		BestHeight: func() int32 {
			return s.chain.BestSnapshot().Height
		},
		MedianTimePast: func() time.Time {
			return s.chain.BestSnapshot().MedianTime
		},
		CalcSequenceLock: func(tx *util.Tx, view *blockchain.UtxoViewpoint) (*blockchain.SequenceLock, error) {
			return s.chain.CalcSequenceLock(tx, view, true)
		},
		IsDeploymentActive: s.chain.IsDeploymentActive,
		SigCache:           s.sigCache,
		HashCache:          s.hashCache,
		AddrIndex:          s.addrIndex,
		FeeEstimator:       s.feeEstimator,
	}
	s.txMemPool = mempool.New(&txC)
	s.syncManager, err =
		netsync.New(
			&netsync.Config{
				PeerNotifier:       &s,
				Chain:              s.chain,
				TxMemPool:          s.txMemPool,
				ChainParams:        s.chainParams,
				DisableCheckpoints: *Cfg.DisableCheckpoints,
				MaxPeers:           *Cfg.MaxPeers,
				FeeEstimator:       s.feeEstimator,
			},
		)
	if err != nil {
		return nil, err
	}
	// Create the mining policy and block template generator based on the configuration options.
	// NOTE: The CPU miner relies on the mempool, so the mempool has to be created before calling the function to create the CPU miner.
	policy := mining.Policy{
		BlockMinWeight:    uint32(*Cfg.BlockMinWeight),
		BlockMaxWeight:    uint32(*Cfg.BlockMaxWeight),
		BlockMinSize:      uint32(*Cfg.BlockMinSize),
		BlockMaxSize:      uint32(*Cfg.BlockMaxSize),
		BlockPrioritySize: uint32(*Cfg.BlockPrioritySize),
		TxMinFreeFee:      StateCfg.ActiveMinRelayTxFee,
	}
	blockTemplateGenerator := mining.NewBlkTmplGenerator(&policy,
		s.chainParams, s.txMemPool, s.chain, s.timeSource,
		s.sigCache, s.hashCache, s.algo)
	s.cpuMiner = cpuminer.New(&cpuminer.Config{
		Blockchain:             s.chain,
		ChainParams:            chainParams,
		BlockTemplateGenerator: blockTemplateGenerator,
		MiningAddrs:            StateCfg.ActiveMiningAddrs,
		ProcessBlock:           s.syncManager.ProcessBlock,
		ConnectedCount:         s.ConnectedCount,
		IsCurrent:              s.syncManager.IsCurrent,
		NumThreads:             s.numthreads,
		Algo:                   s.algo,
	})
	// s.minerController = controller.New(&controller.Config{
	// 	Blockchain:             s.chain,
	// 	ChainParams:            chainParams,
	// 	BlockTemplateGenerator: blockTemplateGenerator,
	// 	MiningAddrs:            StateCfg.ActiveMiningAddrs,
	// 	ProcessBlock:           s.syncManager.ProcessBlock,
	// 	MinerListener:          *Cfg.MinerListener,
	// 	MinerKey:               StateCfg.ActiveMinerKey,
	// 	ConnectedCount:         s.ConnectedCount,
	// 	IsCurrent:              s.syncManager.IsCurrent,
	// })
	/*	Only setup a function to return new addresses to connect to when
		not running in connect-only mode.  The simulation network is always
		in connect-only mode since it is only intended to connect to
		specified peers and actively avoid advertising and connecting to
		discovered peers in order to prevent it from becoming a public test
		network. */
	var newAddressFunc func() (net.Addr, error)
	if !*Cfg.SimNet && Cfg.ConnectPeers == nil {
		newAddressFunc = func() (net.Addr, error) {
			for tries := 0; tries < 100; tries++ {
				addr := s.addrManager.GetAddress()
				if addr == nil {
					break
				}
				/*	Address will not be invalid, local or unroutable
					because addrmanager rejects those on addition.
					Just check that we don't already have an address
					in the same group so that we are not connecting
					to the same network segment at the expense of
					others. */
				key := addrmgr.GroupKey(addr.NetAddress())
				if s.OutboundGroupCount(key) != 0 {
					continue
				}
				// only allow recent nodes (10mins) after we failed 30 times
				if tries < 30 && time.Since(addr.LastAttempt()) < 10*time.Minute {
					continue
				}
				// allow nondefault ports after 50 failed tries.
				if tries < 50 && fmt.Sprintf("%d", addr.NetAddress().Port) !=
					ActiveNetParams.DefaultPort {
					continue
				}
				addrString := addrmgr.NetAddressKey(addr.NetAddress())
				return addrStringToNetAddr(addrString)
			}
			return nil, errors.New("no valid connect address")
		}
	}
	// Create a connection manager.
	targetOutbound := defaultTargetOutbound
	if *Cfg.MaxPeers < targetOutbound {
		targetOutbound = *Cfg.MaxPeers
	}
	cmgr, err :=
		connmgr.New(
			&connmgr.Config{
				Listeners:      listeners,
				OnAccept:       s.inboundPeerConnected,
				RetryDuration:  connectionRetryInterval,
				TargetOutbound: uint32(targetOutbound),
				Dial:           podDial,
				OnConnection:   s.outboundPeerConnected,
				GetNewAddress:  newAddressFunc,
			},
		)
	if err != nil {
		return nil, err
	}
	s.connManager = cmgr
	// Start up persistent peers.
	if Cfg.ConnectPeers != nil {
		permanentPeers := *Cfg.ConnectPeers
		if len(permanentPeers) == 0 {
			permanentPeers = *Cfg.AddPeers
		}
		for _, addr := range permanentPeers {
			netAddr, err := addrStringToNetAddr(addr)
			if err != nil {
				return nil, err
			}
			go s.connManager.Connect(
				&connmgr.ConnReq{
					Addr:      netAddr,
					Permanent: true,
				},
			)
		}
	}
	if !*Cfg.DisableRPC {
		/*	Setup listeners for the configured RPC listen addresses and
			TLS settings. */
		listeners := map[string][]string{
			"sha256d": *Cfg.RPCListeners,
		}
		for l := range listeners {
			rpcListeners, err := setupRPCListeners(listeners[l])
			if err != nil {
				return nil, err
			}
			if len(rpcListeners) == 0 {
				return nil, errors.New("RPCS: No valid listen address")
			}
			rp, err := newRPCServer(&rpcserverConfig{
				Listeners:    rpcListeners,
				StartupTime:  s.startupTime,
				ConnMgr:      &rpcConnManager{&s},
				SyncMgr:      &rpcSyncMgr{&s, s.syncManager},
				TimeSource:   s.timeSource,
				Chain:        s.chain,
				ChainParams:  chainParams,
				DB:           db,
				TxMemPool:    s.txMemPool,
				Generator:    blockTemplateGenerator,
				CPUMiner:     s.cpuMiner,
				TxIndex:      s.txIndex,
				AddrIndex:    s.addrIndex,
				CfIndex:      s.cfIndex,
				FeeEstimator: s.feeEstimator,
				Algo:         l,
			})
			if err != nil {
				return nil, err
			}
			s.rpcServers = append(s.rpcServers, rp)
		}
		// Signal process shutdown when the RPC server requests it.
		go func() {
			for i := range s.rpcServers {
				<-s.rpcServers[i].RequestedProcessShutdown()
			}
			interrupt.Request()
		}()
	}
	return &s, nil
}
// newServerPeer returns a new serverPeer instance. The peer needs to be set by the caller.
func newServerPeer(s *server, isPersistent bool) *serverPeer {
	return &serverPeer{
		server:         s,
		persistent:     isPersistent,
		filter:         bloom.LoadFilter(nil),
		knownAddresses: make(map[string]struct{}),
		quit:           make(chan struct{}),
		txProcessed:    make(chan struct{}, 1),
		blockProcessed: make(chan struct{}, 1),
	}
}
// parseListeners determines whether each listen address is IPv4 and IPv6 and returns a slice of appropriate net.Addrs to listen on with TCP. It also properly detects addresses which apply to "all interfaces" and adds the address as both IPv4 and IPv6.
func parseListeners(
	addrs []string) ([]net.Addr, error) {
	netAddrs := make([]net.Addr, 0, len(addrs)*2)
	for _, addr := range addrs {
		log <- cl.Debug{"addr", addr}
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			// Shouldn't happen due to already being normalized.
			return nil, err
		}
		// Empty host or host of * on plan9 is both IPv4 and IPv6.
		if host == "" || (host == "*" && runtime.GOOS == "plan9") {
			netAddrs = append(netAddrs, simpleAddr{net: "tcp4", addr: addr})
			netAddrs = append(netAddrs, simpleAddr{net: "tcp6", addr: addr})
			continue
		}
		// Strip IPv6 zone id if present since net.ParseIP does not handle it.
		zoneIndex := strings.LastIndex(host, "%")
		if zoneIndex > 0 {
			host = host[:zoneIndex]
		}
		// Parse the IP.
		ip := net.ParseIP(host)
		if ip == nil {
			return nil, fmt.Errorf("'%s' is not a valid IP address", host)
		}
		// To4 returns nil when the IP is not an IPv4 address, so use this determine the address type.
		if ip.To4() == nil {
			netAddrs = append(netAddrs, simpleAddr{net: "tcp6", addr: addr})
		} else {
			netAddrs = append(netAddrs, simpleAddr{net: "tcp4", addr: addr})
		}
	}
	return netAddrs, nil
}
// randomUint16Number returns a random uint16 in a specified input range.  Note that the range is in zeroth ordering; if you pass it 1800, you will get values from 0 to 1800.
func randomUint16Number(
	max uint16) uint16 {
	// In order to avoid modulo bias and ensure every possible outcome in [0, max) has equal probability, the random number must be sampled from a random source that has a range limited to a multiple of the modulus.
	var randomNumber uint16
	var limitRange = (math.MaxUint16 / max) * max
	for {
		err := binary.Read(rand.Reader, binary.LittleEndian, &randomNumber)
		if err != nil {
			panic(err)
		}
		if randomNumber < limitRange {
			return (randomNumber % max)
		}
	}
}
// setupRPCListeners returns a slice of listeners that are configured for use with the RPC server depending on the configuration settings for listen addresses and TLS.
func setupRPCListeners(
	urls []string) ([]net.Listener, error) {
	// Setup TLS if not disabled.
	listenFunc := net.Listen
	if !*Cfg.NoTLS {
		// Generate the TLS cert and key file if both don't already exist.
		if !FileExists(*Cfg.RPCKey) && !FileExists(*Cfg.RPCCert) {
			err := genCertPair(*Cfg.RPCCert, *Cfg.RPCKey)
			if err != nil {
				return nil, err
			}
		}
		keypair, err := tls.LoadX509KeyPair(*Cfg.RPCCert, *Cfg.RPCKey)
		if err != nil {
			return nil, err
		}
		tlsConfig := tls.Config{
			Certificates: []tls.Certificate{keypair},
			MinVersion:   tls.VersionTLS12,
		}
		// Change the standard net.Listen function to the tls one.
		listenFunc = func(net string, laddr string) (net.Listener, error) {
			return tls.Listen(net, laddr, &tlsConfig)
		}
	}
	netAddrs, err := parseListeners(urls)
	if err != nil {
		return nil, err
	}
	listeners := make([]net.Listener, 0, len(netAddrs))
	for _, addr := range netAddrs {
		listener, err := listenFunc(addr.Network(), addr.String())
		if err != nil {
			log <- cl.Warnf{"can't listen on %s: %v", addr, err}
			continue
		}
		listeners = append(listeners, listener)
	}
	return listeners, nil
}
