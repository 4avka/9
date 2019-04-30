package sub

import (
	"encoding/binary"
	"errors"
	"fmt"
	"git.parallelcoin.io/dev/9/pkg/util/cl"
	"hash/crc32"
	"math/rand"
	"net"
	"time"
)

// Implementations of common parts for node and worker

// NewBase creates a new base listener
func NewBase(cfg BaseCfg) (b *Base) {
	log <- cl.Debug{"creating new Base"}
	b = &Base{
		cfg:       cfg,
		packets:   make(chan Packet, baseChanBufs),
		incoming:  make(chan Bundle, baseChanBufs),
		returning: make(chan Bundle, baseChanBufs),
		trash:     make(chan Bundle),
		quit:      make(chan bool),
	}
	return
}

// Start attempts to open a listener and commences receiving packets and assembling them into messages
func (b *Base) Start() (err error) {
	var addr *net.UDPAddr
	addr, err = net.ResolveUDPAddr(uNet, b.cfg.Listener)
	if err != nil {
		log <- cl.Fatal{"sub.Base.Start ResolveUDPAddr", err}
		panic(err)
	}
	b.listener, err = net.ListenUDP(uNet, addr)
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
			case <-b.quit:
				break
			default:
			}
			select {
			case msg := <-b.message:
				go b.cfg.Handler(msg)
				continue
			default:
			}
		}
	}()
	return
}

// Stop shuts down the listener
func (b *Base) Stop() {
	log <- cl.Trace{"shutting down listener"}
	b.quit <- true
	b.listener.Close()
}

func (b *Base) readFromSocket() {
	log <- cl.Debug{"reading from socket"}
	for {
		select {
		case <-b.quit:
			log <- cl.Trace{"quitting readFromSocket"}
			break
		default:
		}
		var data = make([]byte, b.cfg.BufferSize)
		count, _, err := b.listener.ReadFromUDP(data[0:])
		if err != nil {
			log <- cl.Info{"sub.Base.readFromSocket.ReadFromUDP", err}
			continue
		}
		if count > 12 {
			data = data[:count]
			sender := string(data[:6])
			body := data[:count-4]
			check := data[count-4:]
			checkSum := binary.LittleEndian.Uint32(check)
			cs := crc32.Checksum(body, crc32.MakeTable(crc32.Castagnoli))
			if cs != checkSum {
				continue
			}
			b.packets <- Packet{
				sender: sender,
				bytes:  data,
			}
		}
	}
}

func (b *Base) processPackets() {
	for {
		select {
		case <-b.quit:
			break
		default:
		}
		select {
		case p := <-b.packets:
			sender := string(p.bytes[:6])
			go func() {
				for {
					select {
					case <-b.doneRet:
						log <- cl.Trace{"returning items to channel"}
						for i := range b.returning {
							b.incoming <- i
						}
						break
					case <-b.returning:
						continue
					case <-b.trash:
						continue
					}
				}
			}()
			for bi := range b.incoming {
				if bi.sender == sender {
					log <- cl.Trace{"appending bytes to bundle"}
					bi.packets = append(bi.packets, p.bytes)
					b.returning <- bi
					break
				}
				if len(bi.packets) > 2 {
					log <- cl.Trace{"if we have 3 or more it should be possible to now assemble the message"}
					b.incoming <- bi
					continue
				}
				if bi.received.Sub(time.Now()) > latencyMax {
					log <- cl.Trace{"delete all packets that fall outside the latency maximum"}
					b.trash <- bi
					break
				} else {
					log <- cl.Trace{"accept subsequent packets before latencyMax"}
					b.incoming <- bi
				}
				b.doneRet <- true
			}
			continue
		}
	}
}

func (b *Base) processBundles() {
	for {
		select {
		case <-b.quit:
			break
		default:
		}
		var uuid int32
		select {
		case bundle := <-b.incoming:
			data, err := rsDecode(bundle.packets)
			if err == nil && bundle.uuid != uuid {
				rand.Seed(time.Now().Unix())
				uuid = rand.Int31()
				b.message <- Message{
					uuid:      bundle.uuid,
					sender:    bundle.sender,
					timestamp: bundle.received,
					bytes:     data,
				}
				uuid = bundle.uuid
				b.trash <- bundle
			}
		}
	}
}

// Send a message of up to maxMessageSize bytes to a given UDP address
func (b *Base) Send(data []byte, addr *net.UDPAddr) (err error) {
	if len(data) > 3072 {
		err = errors.New("maximum message size is " + fmt.Sprint(maxMessageSize) + " bytes")
	}
	addr, err = net.ResolveUDPAddr(uNet, addr.String())
	if err != nil {
		log <- cl.Debug{"sub.Base.Send.ResolveUDPAddr", err}
	}
	conn, err := net.DialUDP(uNet, nil, addr)
	if err != nil {
		log <- cl.Debug{"sub.Base.Send.DialUDP", err}
		return
	}
	_, err = conn.Write(data)
	if err != nil {
		log <- cl.Debug{"sub.Base.Send.Write", err}
		return
	}
	return
}
