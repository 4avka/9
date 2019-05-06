# bola

#### low latency reliable UDP with authentication and integrity

## Rationale for protocol

`bola` is being written to implement a low latency work dispatch system for the parallelcoin cryptocurrency plan 9 proof of work to facilitate the deployment of network connected clusters of worker nodes.

As such, latency is the utmost priority for this implementation, and sufficient protection against the possibility of spoofing attacks leading to denial of service.

Thus, the messages use no encryption, but they cannot be forged in the amount of time during which the solution is valuable. Mainly it is solutions returned, but without correct blocks to solve, the miner's cluster cannot work correctly either.

The block headers and the nonce/timestamp combination returned by workers do not constitute sufficient data to steal the solution and issue a block before the legitimate winner, as the contents of the transactions are not shared as part of this protocol.

This is to contrast to the 'GetBlockTemplate' method in the Bitcoin JSONRPC API, which was devised to prevent malicious pools. Pools are not considered to be a viable means of aggregating greater hashpower under the post plan9 hard fork regime, as they have 1-2 second lag time on propagating work, and it is intended that private, coordinated 

## Selection of Hash Algorithm

For reasons of performance, the ideal hash function for this exact use is the one with the fastest processing time and the best collision resistance. It need not be 128 bit cryptographic grade. The HighwayHash function has been selected as its 128 bit implementation provides about 53 bits of collision resistance, which is enough to delay an attacker for more than a few minutes.

In addition to to this, as a further protection, the protocol uses a second tier from the pre-shared secret key for the connection (a 256 bit key) rather than the key itself, it has a 64 bit value appended with an increment for each forward and back message half-cycle between the two nodes, providing a further layer of protection against directly deriving the hash in enough time to make use of the data, or to disrupt it before it can be used to produce a solution by the node receiving it.

## Why reliable UDP?

Reliable UDP, in this case only for a uniform maximum single packet message up to 3kb in size, allows the messages to be sent and immediately received and more than 6 out of the 9 packets in a message have to be lost to lose the message. The recipient can then send a resend request, which in this case is not for a repeat send but just to send the current data. This may change anyway.

The miner has a strategy for distributing the pending transactions that can go into blocks so that it reduces the occurrance of all of the nodes in the cluster from having to halt work at the same time. This somewhat reduces the guarantee of clearance time to between 1 and 3 blocks from the posting of the transaction, but this benefits miners and benefits the network's security in that it improves the granularity of the likely occurrance of solutions. If nodes always publish every tx in the block they can, firstly, this guarantees that all of a cluster has to stop work until new work is constructed, whereas if there is always a leftover transaction or two after a solution is found, some nodes can continue to work while others wait on new work.

The messages are sent out in a burst of nine UDP packets per message, and 3 must get through to get a solution. The sooner workers start on new work, the odds of solution being directly proportional to operation cycles, the less time between work being ready and starting on it, the better. This latency advantage is intended to obstruct the use of Pool mining as well as botnets, who both have a lowest-common-denominator latency cost that a UDP connected cluster can always stay ahead by a few message cycles in most cases.

Thus, the only real attack that can be applied against a cluster, if it happens to have any open access into the connections between node and workers, is to either prevent communication between nodes, or to steal solutions somehow.

Stealing solutions is impractical as the node delivers only the block header, and not the transaction payload that generates the merkle root in the header, it is an assumption and contract of this protocol that it is for private, centrally controlled mining work pools, as this protocol facilitates a pool operator to use its users hashpower to attack a cryptocurrency.

Thus, the main way an attack can take place on this protocol is the prevention of correct decoding of packets, by injecting valid but wrong-payload bearing shards, making the reception of packets more unpredictable. Reed Solomon encoding has the limitation that all shards must be in order and if an attacker could construct a valid shard, or two, it can prevent the correct decoding and dropping of the packet, and loss, potentially, of a solution thereby.

Thus, the necessary security is prevention of the ability to construct valid but incorrect packets that break the decoding process. The lifetime of these packets' value is about 1 second, which is way shorter than 52 bits of cryptographic security can be reversed without formidable resources located close to the target.
