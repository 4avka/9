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

func NewAddr() *Addr {
	return new(Addr)
}

// Get takes a string from bytes on the prefix containing an IPv4 address (it
// is passed around as a string for easy comparison) and returns the format used
// by net.Dial
func (a Addr) Get() (out string) {
	if aa := a; len(aa) == 6 {
		return fmt.Sprintf(ipv4Format,
			aa[0], aa[1], aa[2], aa[3],
			binary.LittleEndian.Uint16(aa[4:6]),
		)
	} else {
		log <- cl.Debug{cl.Ine(), "Addr is uninitialised"}
		return
	}
}

// Put takes a string in the format xxx.xxx.xxx.xxx:xxxxx and converts it to the encoded bytes format
func (a *Addr) Put(addr string) (out Addr) {
	if a == nil {
		*a = Addr{}
	}
	b := make([]byte, 4)
	o := make([]byte, 2)
	var ou16 uint16
	_, err := fmt.Sscanf(addr, ipv4Format, &b[0], &b[1], &b[2], &b[3], &ou16)
	if err != nil {
		log <- cl.Error{cl.Ine(), err}
	}
	binary.LittleEndian.PutUint16(o, ou16)
	*a = Addr(append(b, o...))
	return *a
}
