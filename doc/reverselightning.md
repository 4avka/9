# Multi-Token Protocols and Mining
### Flip Lightning on its head!

Many cryptocurrencies sport multiple and user-issuable arbitrary tokens, such as Ethereum ERC20 tokens, Bitshares tokens, and others. These chains execute transactions faster, on average, than Lightning clears a cross-chain atomic swap. 

However, their own architecture complicates a complete emulation of the supply parameters and compatibility of the script engine that validates transactions. Centralised and permissioned, especially geographically and network-distance constrained distributed databases could provide the millisecond speed required to sufficiently allow a complete simulation of the timing properties and supply of new tokens in a multi-token network protocol.

If a small chain with scarce resources and only a handful of developers wants to improve thteir chances of making a mark on the marketplace, the most important thing is access to the market.

If another network were able to host a functionally identical (but with instant clearance) network but bit for bit compatible blocks and transactions alongside its own native base token, it would be trivial to implement instant clearing cross-chain transactions, aka Atomic Swaps.

For the small chains, beset on all sides by ASIC makers, pump and dumpers, predatory pool miners and opportunistic large mining farms (such as Nicehash), the combination of all these threats, plus thin connections to the marketplace compound the early stage development process.

However, if a consortium of similarly afflicted small projects were to be able to merge their tokens together in a multi-token protocol with on-chain exchange and sub-second clearance to finality, firstly between themselves they can easily shift tokens, taking advantage of differences in markets or geofenced access (by network or by fiat) - then they would ultimately eventually sell very easily the idea of all cryptocurrencies moving to a single base protocol.

### How to do it???!

Obviously, with current consensus protocols based on PoW, staking, and weakly synchronous, a ledger's latency will only increase with more transactions.

Both approaches trade off things that have a similar end result - the development of a main miner group numbering around 20. Whether it is Delegated Proof of Stake, Proof of Stake (like Pivx) or Hashcash Proof of Work, the numbers we see in the top 95% of blocks mined by only a score of miners.

### Gossip first, check later

One key element to this is the latency of message distribution. Every message cycle has certain structural limitations. If a node must verify every bit of jibber jabber it is sent, this adds the time of verification to the network message propagation rate.

So, firstly, nodes in a network that runs at minimum latency must relay chatter as promptly as possible. Even just to hash it, store the index and not relay it again after some number of relay cycles. The latency additional for this is somewhere around a microsecond. A full check could cost up to 10 microseconds or even more with a 1 megabyte or larger sized block.

It is possible to tweak these parameters in most Nakamoto Consensus based cryptocurrency servers, to defer verification before relay and often not much customisation would allow this to be made more clever.

### Gossip means share your view of the network

The information that nodes will be relaying is intended to allow the entire network, based on a sufficient number of messages of their lists of recent transactions, to determine with full 100% finality the total ordering of transactions.

#### Total order forbids double spends

It's a bit like looking down on Flatland. You can see where everything is all at once, but flatlanders have a horizon and a visibility distance.

However, if the flatlanders all take photos of their view and compare them together, if they can share and compare quickly, they can all come to agree on the state of things both directly in their sphere and to the antipodes, the furthest possible network location away from oneself.

The very Byzantine Generals Problem exactly presupposes a concurrent network with corruptible messengers, some amount of latency of information that can conceal mischief.

One may not be able to necessarily trust any other communication partner in such a scenario, but the odds exponentially collapse with the number of subjective reports, as deceptions require artificial manipulation, not only is it more susceptible to being obviously inconsistent with the rest of the data.

### Sufficient Gossip yields Consensus

If all nodes are frequently reporting to each other what they have been seeing, and economically motivated by the chance of being the signer of the final version of a transaction, they will more than overwhelm even a substantial majority of corrupt nodes.

With enough, even largely manipulated versions of the network traffic, the overlap between truth versus the overlap of lies inherently and naturally tends to favour the truth, both because of the reduced time to response of unmodified subjective data, means that even the truths told by liars will undo their lies.

Unlike democracy, with its bottleneck of calling an election or quorum, a gossip network does not flip between provisional and final atomically, it is as fluid as the number of potential message paths and endpoints on the network (ie, factorial).

This means that even the best laid plans of deception, given absolutely huge resources, will not be able to predict the potential flux of epidemic transmission paths, and position itself to funnel and channel it to separate "payer" and "payee" history for long enough to make off with the booty and disappear back into the woodwork.


