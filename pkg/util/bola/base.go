package bola

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
		Cfg:       cfg,
		Packets:   make(chan Packet, BaseChanBufs),
		Incoming:  make(chan Bundle, BaseChanBufs),
		Returning: make(chan Bundle, BaseChanBufs),
		Trash:     make(chan Bundle),
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
				go b.Cfg.Handler(msg)
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
		count, _, err := b.Listener.ReadFromUDP(data[0:])
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
			b.Packets <- Packet{
				Sender: sender,
				Data:   data,
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
			sender := string(p.Data[:6])
			go func() {
				for {
					select {
					case <-b.DoneRet:
						log <- cl.Trace{"returning items to channel"}
						for i := range b.Returning {
							b.Incoming <- i
						}
						break
					case <-b.Returning:
						continue
					case <-b.Trash:
						continue
					}
				}
			}()
			for bi := range b.Incoming {
				if bi.Sender == sender {
					log <- cl.Trace{"appending bytes to bundle"}
					bi.Packets = append(bi.Packets, p.Data)
					b.Returning <- bi
					break
				}
				if len(bi.Packets) > 2 {
					log <- cl.Trace{"if we have 3 or more it should be possible to now assemble the message"}
					b.Incoming <- bi
					continue
				}
				if bi.Received.Sub(time.Now()) > LatencyMax {
					log <- cl.Trace{"delete all packets that fall outside the latency maximum"}
					b.Trash <- bi
					break
				} else {
					log <- cl.Trace{"accept subsequent packets before latencyMax"}
					b.Incoming <- bi
				}
				b.DoneRet <- true
			}
			continue
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
		var uuid int32
		select {
		case bundle := <-b.Incoming:
			data, err := rsDecode(bundle.Packets)
			if err == nil && bundle.UUID != uuid {
				rand.Seed(time.Now().Unix())
				uuid = rand.Int31()
				b.Message <- Message{
					UUID:     bundle.UUID,
					Sender:   bundle.Sender,
					Received: bundle.Received,
					Data:     data,
				}
				uuid = bundle.UUID
				b.Trash <- bundle
			}
		}
	}
}

// Send a message of up to maxMessageSize bytes to a given UDP address
func (b *Base) Send(data []byte, addr *net.UDPAddr) (err error) {
	if len(data) > 3072 {
		err = errors.New("maximum message size is " + fmt.Sprint(MaxMessageSize) + " bytes")
	}
	addr, err = net.ResolveUDPAddr(UNet, addr.String())
	if err != nil {
		log <- cl.Debug{"sub.Base.Send.ResolveUDPAddr", err}
	}
	conn, err := net.DialUDP(UNet, nil, addr)
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
