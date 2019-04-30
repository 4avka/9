package bola

import (
	"git.parallelcoin.io/dev/9/pkg/fek"
	"net"
	"time"
)

var (
	UNet           = "udp4"
	MaxMessageSize = 3072
	// default channel buffer sizes for Base
	BaseChanBufs = 128
	// latency maximum
	LatencyMax = time.Millisecond * 250
)

// BaseCfg is the configuration for a Base
type BaseCfg struct {
	Handler    func(message Message)
	Listener   string
	FEC        fek.RS
	BufferSize int
}

// Base is the common structure between a worker and a node
type Base struct {
	Cfg       BaseCfg
	Listener  *net.UDPConn
	Packets   chan Packet
	Incoming  chan Bundle
	Returning chan Bundle
	Trash     chan Bundle
	DoneRet   chan bool
	Message   chan Message
	Quit      chan bool
}

// Packet is the structure of individual encoded packets of the message. These
// are made from a 9/3 Reed Solomon code and 9 are sent in distinct packets and
// only 3 are required to guarantee retransmit-free delivery.
type Packet struct {
	UUID     int32
	Sender   string
	Received time.Time
	Data     []byte
}

// A Bundle is a collection of the received packets received from the same
// sender with up to 9 pieces.
type Bundle struct {
	Sender   string
	UUID     int32
	Received time.Time
	Packets  [][]byte
}

// Message is the data reconstructed from a complete Bundle, containing data in
// messagepack format
type Message struct {
	UUID     int32
	Sender   string
	Received time.Time
	Data     []byte
}
