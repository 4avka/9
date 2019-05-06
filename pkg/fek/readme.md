# fek - forward error korrection

Reed-Solomon encoding is a very widely used method of adding redundant data to streams of bytes that are transmitted in discrete packets over an unreliable channel, to ensure the signal is received verbatim while avoiding the need for retransmission.

It was originally devised for space probes to send back images and telemetry data, where there basically is no option of retransmit, and is used in many devices and protocols including optical disks, some solid state memory and storage devices, and many realtime and streaming codecs for various kinds of data including video, audio and general data.

The essential principle to understand in how it works, is that it generates chunks of data that map onto the same polynomial curve as the 'shards' of the original data, and then allows the reverse derivation from these extra points to regain the original data.

The shards must be in the specified order as they are created, but out of the total generated, any number of the required number of shards can reassemble the first 'data' shards that reconstitutes the message, even if several of the pieces did not make it across the channel, or came across corrupted, so long as you can identify which pieces were not corrupted and zero out the corrupted pieces, hence the name 'Erasure Coding'.

## Error Detection

In many implementations of Reed Solomon encoding, there is the use of several brute-force algorithms that interpolate shards to determine the valid and erase the invalid and regenerate them. These techniques save on the inflation required by adding checksums, but add a cost in that the regeneration process is longer as it operates on permutations, thus the more pieces, the factorial more amount of processing required.

Instead, since the intent of this library is for the implementation of low latency network transport (or potentially, redundant storage) and we don't mind a little extra byte size since we want to have the packets decoded as quickly as possible, instead, `fek` uses the fastest, collision resistant non-cryptographic hash function to generate 64 bit checksums, HighwayHash, a hash function designed in google's dev shop specifically for the purpose of high throughput data processing.

## Usage

The use of this library is extremely straightforward. You feed in arbitrary data, of up to `required` pieces, maximum shard size of 64kb (thus `required*2<<16` max data payload size, and 11 bytes overhead for checksum each piece), and it returns an array of shards of the specified number, the first `required` number are the original message, plus padding, the remainder up to `total` size, and each shard has its shard number as a 1 byte prefix and at the beginning of shard 0 is the total payload length covered by the checksum at the end.

Thus, this implementation is quite opinionated, and makes the tradeoff in favour of low processing cost over redundant data overhead, as it is intended to provide the minimum cost of latency.