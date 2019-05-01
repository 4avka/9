package bola

import (
	"net"
	"time"

	"git.parallelcoin.io/dev/9/pkg/fek"
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
	FEC        *fek.RS
	BufferSize int
}

// Base is the common structure between a worker and a node
type Base struct {
	Cfg       BaseCfg
	Address   Addr
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
	Sender   Addr
	UUID     uint32
	Size     uint16
	Data     []byte
	Received time.Time
}

// Packets lets us attach a method to unwrap to [][]byte
type Packets []Packet

func (p *Packets) GetShards() (out [][]byte) {
	for _, x := range *p {
		out = append(out, x.Data)
	}
	return
}

// A Bundle is a collection of the received packets received from the same
// sender with up to 9 pieces.
type Bundle struct {
	Sender  Addr
	UUID    uint32
	Started time.Time
	Packets Packets
	Spent   bool
}

// Message is the data reconstructed from a complete Bundle, containing data
type Message struct {
	Sender   Addr
	UUID     uint32
	Received time.Time
	Data     []byte
}
