package bola

import "time"

const (
	uNet = "udp4"
	// Maximum of 9 packets per message, so 16kb round is enough
	defaultBufferSize = 16384
	// FEC expands message by 150%, we don't split message chunks over more than one packet
	maxMessageSize = 3072
	// default channel buffer sizes for Base
	baseChanBufs = 128
	// timeout for bundles
	timeout = time.Millisecond * 250
)

type Packet struct {
	UUID        uint32
	ShardNumber byte
	Payload     []byte
	Checksum    uint64
}
