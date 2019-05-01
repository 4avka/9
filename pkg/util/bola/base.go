package bola

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net"
	"time"

	"git.parallelcoin.io/dev/9/pkg/fek"
	"git.parallelcoin.io/dev/9/pkg/util/cl"
	"github.com/minio/highwayhash"
)

// Implementations of common parts for node and worker

// NewBase creates a new base listener
func NewBase(cfg BaseCfg) (b *Base) {
	log <- cl.Debug{"creating new Base"}
	b = &Base{
		Cfg:       cfg,
		Packets:   make(chan Packet, BaseChanBufs),
		Incoming:  make(chan Bundle, BaseChanBufs),
		Returning: make(chan Bundle, BaseChanBufs),
		DoneRet:   make(chan bool),
		Trash:     make(chan Bundle),
		Message:   make(chan Message),
		Quit:      make(chan bool),
	}
	return
}

// Start attempts to open a listener and commences receiving packets and assembling them into messages
func (b *Base) Start() (err error) {
	var addr *net.UDPAddr
	addr, err = net.ResolveUDPAddr(UNet, b.Cfg.Listener)
	if err != nil {
		log <- cl.Fatal{"sub.Base.Start ResolveUDPAddr", err}
		panic(err)
	}
	b.Listener, err = net.ListenUDP(UNet, addr)
	if err != nil {
		log <- cl.Fatal{"sub.Base.Start ListenUDP", err}
		panic(err)
	}
	// Start up reader to push packets into packet channel
	go b.readFromSocket()
	go b.processPackets()
	go b.processBundles()
	go func() {
		for {
			select {
			case <-b.Quit:
				break
			default:
			}
			select {
			case msg := <-b.Message:
				if b.Cfg.Handler != nil {
					go b.Cfg.Handler(msg)
				}
				// continue
			default:
			}
		}
	}()
	return
}

// Stop shuts down the listener
func (b *Base) Stop() {
	log <- cl.Trace{"shutting down listener"}
	b.Quit <- true
	b.Listener.Close()
}

func (b *Base) readFromSocket() {
	log <- cl.Debug{"reading from socket"}
	for {
		select {
		case <-b.Quit:
			log <- cl.Trace{"quitting readFromSocket"}
			break
		default:
		}
		var data = make([]byte, b.Cfg.BufferSize)
		count, _, err := b.Listener.ReadFromUDP(data)
		if err != nil {
			log <- cl.Info{"sub.Base.readFromSocket.ReadFromUDP", err}
			continue
		}
		received := time.Now()
		if count > 12 {
			bSize :=
				data[10:12]
			iSize := uint16(int(bSize[0]) | int(bSize[1])<<8)
			tSize := iSize + 20
			if count < int(tSize) {
				log <- cl.Debug{"received truncated packet"}
				continue
			}
			packet := data[:tSize]
			pl, ok := fek.VerifyChecksum(packet)
			if !ok {
				continue
			}
			sender := Addr(pl[:6])
			UUID := int(pl[6]) | int(pl[7])<<8 | int(pl[8])<<16 | int(pl[9])<<24
			data := pl[12:]
			b.Packets <- Packet{
				sender, uint32(UUID), iSize, data, received,
			}
		}
	}
}

func (b *Base) processPackets() {
	for {
		select {
		case <-b.Quit:
			break
		default:
		}
		select {
		case p := <-b.Packets:
			go func() {
				for {
					select {
					case <-b.DoneRet:
						log <- cl.Trace{"returning items to incoming"}
						for i := range b.Returning {
							b.Incoming <- i
						}
						break
					case <-b.Returning:
						continue
					case _ = <-b.Trash:
						continue
					}
				}
			}()
			var bundled bool
			for bi := range b.Incoming {
				if bi.UUID == p.UUID {
					log <- cl.Trace{"appending packet to bundle"}
					bi.Packets = append(bi.Packets, p)
					bundled = true
					if len(bi.Packets) > 2 {
						log <- cl.Trace{"if we have 3 or more it should be possible to now assemble the message"}
						b.Incoming <- bi
						break
					}
				}
				if bi.Started.Sub(time.Now()) > LatencyMax {
					log <- cl.Trace{"delete all bundles that fall outside the latency maximum"}
					b.Trash <- bi
					continue
				} else {
					log <- cl.Trace{"returning bundle to channel queue"}
					b.Returning <- bi
				}
			}
			if !bundled {
				log <- cl.Trace{"bundling new packet and placing in the return queue"}
				b.Returning <- Bundle{p.Sender, p.UUID, p.Received, []Packet{p}, false}
			}
			b.DoneRet <- true
		}
	}
}

func (b *Base) processBundles() {
	for {
		select {
		case <-b.Quit:
			break
		default:
		}
		select {
		case bundle := <-b.Incoming:
			data := b.Cfg.FEC.Decode(bundle.Packets.GetShards())
			if data != nil {
				b.Message <- Message{
					bundle.Sender, bundle.UUID, bundle.Started, data,
				}
			}
		}
	}
}

// Send a message of up to maxMessageSize bytes to a given UDP address
func (b *Base) Send(data []byte, addr *net.UDPAddr) (err error) {
	if len(data) > 3072 {
		err = errors.New("maximum message size is " + fmt.Sprint(MaxMessageSize) + " bytes")
	}
	var shards, pkg [][]byte
	if shards = b.Cfg.FEC.Encode(data); shards != nil {
		shardsLen := len(shards[0])
		buuid := make([]byte, 4)
		_, e := rand.Read(buuid)
		if e != nil {
			panic(err)
		}
		pkg = make([][]byte, 9)
		for i := range pkg {
			pkg[i] = make([]byte, shardsLen+20)
		}
		for i := range pkg {
			for j, y := range b.Address {
				pkg[i][j] = y
			}
			pkg[i][6] = buuid[0]
			pkg[i][7] = buuid[1]
			pkg[i][8] = buuid[2]
			pkg[i][9] = buuid[3]
			pkg[i][10] = byte(shardsLen)
			pkg[i][11] = byte(shardsLen >> 8)
			for j, y := range shards[i] {
				pkg[i][j+12] = y
			}
			hh := fek.Uint64ToBytes(highwayhash.Sum64(pkg[i][:12+shardsLen], fek.Zerokey))
			for j, y := range hh {
				pkg[i][j+12+shardsLen] = y
			}
			addr, err = net.ResolveUDPAddr(UNet, addr.String())
			if err != nil {
				log <- cl.Debug{"sub.Base.Send.ResolveUDPAddr", err}
			}
			conn, e := net.DialUDP(UNet, nil, addr)
			if e != nil {
				log <- cl.Debug{"sub.Base.Send.DialUDP", err}
				return
			}
			_, err = conn.Write(pkg[i][:20+shardsLen])
			if err != nil {
				log <- cl.Debug{"sub.Base.Send.Write", err}
				return
			}
		}
	}
	return
}
