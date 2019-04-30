package bola

import (
	"encoding/binary"
	"fmt"
	"git.parallelcoin.io/dev/9/pkg/util/cl"
)

const (
	ipv4Format = "%d.%d.%d.%d:%d"
)

type Addr []byte

// Get takes a string from bytes on the prefix containing an IPv4 address (it is passed around as a string for easy comparison) and returns the format used by net.Dial
func (a *Addr) Get(encoded string) (out string) {
	if len(encoded) != 6 {
		return
	}
	e := []byte(encoded)
	out = fmt.Sprintf(ipv4Format,
		e[0], e[1], e[2], e[3],
		binary.LittleEndian.Uint16(e[4:6]),
	)
	return
}

// Put takes a string in the format xxx.xxx.xxx.xxx:xxxxx and converts it to the encoded bytes format
func (a *Addr) Put(addr string) {
	b := make([]byte, 4)
	o := make([]byte, 2)
	var ou16 uint16
	_, err := fmt.Sscanf(addr, ipv4Format, b[0], b[1], b[2], b[3], ou16)
	if err != nil {
		log <- cl.Error{err}
	}
	binary.LittleEndian.PutUint16(o, ou16)
	b = append(b, o...)
	return
}
