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
	// log <- cl.Debug{cl.Ine(), cfg.Listener, "creating new Base"}
	b = &Base{
		Cfg:       cfg,
		Address:   NewAddr().Put(cfg.Listener),
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
	// log <- cl.Debug{cl.Ine(), "resolving listener address"}
	addr, err = net.ResolveUDPAddr(UNet, b.Cfg.Listener)
	if err != nil {
		log <- cl.Fatal{cl.Ine(), err}
		panic(err)
	}
	// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "creating listener"}
	b.Listener, err = net.ListenUDP(UNet, addr)
	if err != nil {
		log <- cl.Fatal{cl.Ine(), err}
		panic(err)
	}
	// Start up reader to push packets into packet channel
	go b.processPackets()
	go b.readFromSocket()
	go b.processBundles()
	go func() {
		done := false
	startLoop:
		for !done {
			select {
			case qm := <-b.Quit:
				// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "quitting main handler"}
				b.Quit <- qm
				done = true
				goto startLoop
			default:
			}
			select {
			case msg := <-b.Message:
				if b.Cfg.Handler != nil {
					go b.Cfg.Handler(msg)
				}
				continue
			default:
			}
		}
	}()
	return
}

// Stop shuts down the listener
func (b *Base) Stop() {
	// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "shutting down listener"}
	b.Quit <- true
	b.Listener.Close()
}

func (b *Base) readFromSocket() {
	// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "starting socket listener"}
	var data = make([]byte, b.Cfg.BufferSize)
	for {
		var received time.Time
		var count int
		// var remoteAddr *net.UDPAddr
		var err error
		select {
		case <-b.Quit:
			// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "quitting readFromSocket"}
			b.Quit <- true
			goto readQuit
		default:
		}
		// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "reading from socket"}
		count, _, err = b.Listener.ReadFromUDP(data)
		if err != nil {
			// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "sub.Base.readFromSocket.ReadFromUDP", err}
			continue
		}
		// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "received", count, "bytes from", remoteAddr}
		received = time.Now()
		if count > 12 {
			// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "slicing up message"}
			bSize :=
				data[10:12]
			iSize := uint16(int(bSize[0]) | int(bSize[1])<<8)
			tSize := iSize + 20
			// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "message should be", tSize, "bytes"}
			if count < int(tSize) {
				// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "received truncated packet"}
				continue
			}
			packet := data[:tSize]
			// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "verifying checksum"}
			pl, ok := fek.VerifyChecksum(packet)
			if !ok {
				// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "failed to verify checksum"}
				continue
			}
			sender := Addr(pl[:6])
			UUID := int(pl[6]) | int(pl[7])<<8 | int(pl[8])<<16 | int(pl[9])<<24
			data := pl[12:]
			b.Packets <- Packet{
				sender, uint32(UUID), iSize, data, received,
			}
			// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "pushed packet to packets channel"}
		}
		// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "waiting for next packet"}
		continue
	readQuit:
		break
	}
}

func (b *Base) processPackets() {
	go func() {
		log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "starting return channel"}
		for {
			select {
			case <-b.DoneRet:
				log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "returning items to incoming"}
				// go func() {
				log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "starting returner"}
				for i := range b.Returning {
					log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "bundle", i, "returned"}
					b.Incoming <- i
				}
				log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "finishing returner"}
				// }()
			// case rb := <-b.Returning:
			// 	log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "returning bundle"}
			// 	b.Incoming <- rb
			// 	continue
			case _ = <-b.Trash:
				log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "discarding bundle"}
				continue
			default:
			}
		}
	}()
	for {
		select {
		case bb := <-b.Quit:
			// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "quitting processPackets"}
			b.Quit <- bb
			break
		default:
		}
		select {
		case p := <-b.Packets:
			log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "received packet", p}
			var bundled bool
			// fmt.Println(<-b.Incoming)
			go func() {
				// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "starting bundler"}
				for bi := range b.Incoming {
					log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "iterating bundles"}
					if bi.UUID == p.UUID {
						log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "appending packet to bundle"}
						bi.Packets = append(bi.Packets, p)
						bundled = true
						if len(bi.Packets) > 2 {
							log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "if we have 3 or more it should be possible to now assemble the message"}
							b.Incoming <- bi
							break
						}
					}
					if bi.Started.Sub(time.Now()) > LatencyMax {
						b.Trash <- bi
						log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "deleting bundle that fall outside the latency maximum"}
						continue
					} else {
						b.Returning <- bi
						log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "returned bundle to channel queue"}
					}
				}
				log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "finished looking for matching packets"}

				if !bundled {
					b.Incoming <- Bundle{p.Sender, p.UUID, p.Received, []Packet{p}, false}
					log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "bundled new packet and placing in the return queue"}
				}
			}()
			log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "triggering channel return"}
			b.DoneRet <- true
			log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "triggered channel return"}
		default:
		}
	}
}

func (b *Base) processBundles() {
	// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "starting bundle processor"}
	for {
		select {
		case <-b.Quit:
			goto processQuit
		default:
		}
		select {
		case bundle := <-b.Incoming:
			data := b.Cfg.FEC.Decode(bundle.Packets.GetShards())
			if data != nil {
				log <- cl.Debug{"message received: '" + string(data) + "' from",
					bundle.Sender.Get()}
				b.Message <- Message{
					bundle.Sender, bundle.UUID, bundle.Started, data,
				}
			}
		}
		continue
	processQuit:
		// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "quitting processBundles"}
		b.Quit <- true
		break
	}
}

// Send a message of up to maxMessageSize bytes to a given UDP address
func (b *Base) Send(data []byte, target string) (err error) {
	if len(data) > 3072 {
		err = errors.New(cl.Ine().Error() + "maximum message size is " + fmt.Sprint(MaxMessageSize) + " bytes")
	}
	go func() {
		var shards, pkg [][]byte
		// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "rs encoding data"}
		if shards = b.Cfg.FEC.Encode(data); shards != nil {
			// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "successfully encoded message '" + string(data) + "'"}
			shardsLen := len(shards[0])
			buuid := make([]byte, 4)
			_, e := rand.Read(buuid)
			if e != nil {
				panic(err)
			}
			pkg = make([][]byte, 9)
			// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "preallocating buffers for shards"}
			for i := range pkg {
				pkg[i] = make([]byte, shardsLen+20)
			}
			for i := range pkg {
				// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "assembling segment", i}
				for j, y := range NewAddr().Put(target) {
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
				// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "appending checksum"}
				hh := fek.Uint64ToBytes(highwayhash.Sum64(pkg[i][:12+shardsLen], fek.Zerokey))
				for j, y := range hh {
					pkg[i][j+12+shardsLen] = y
				}
				if addr, e := net.ResolveUDPAddr(UNet, b.Cfg.Listener); e != nil {
					// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "sub.Base.Send.ResolveUDPAddr", e}
				} else if conn, e := net.DialUDP(UNet, nil, addr); e != nil {
					// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "sub.Base.Send.DialUDP", e}
					return
				} else if n, _ := conn.Write(pkg[i][:20+shardsLen]); err != nil {
					// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "sub.Base.Send.Write", e}
					return
				} else if n != shardsLen+20 {
					// log <- cl.Debug{cl.Ine(), b.Cfg.Listener, "submitted", shardsLen + 20, "bytes but", n,"were reported sent"}
				}
			}
		}
	}()
	return
}
