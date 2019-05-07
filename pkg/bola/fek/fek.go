package fek

import (
	"encoding/binary"

	"github.com/templexxx/reedsolomon"
)

type RS struct {
	*reedsolomon.RS
	required int
	total    int
}

func New(required, total int) *RS {
	rsc, err := reedsolomon.New(required, total-required)
	if err != nil {
		// Only error condition is total < 0 && total > 256
		return nil
	}
	return &RS{rsc, required, total}
}

func (r *RS) Encode(data []byte) [][]byte {
	return nil
}

func (r *RS) Decode(shards [][]byte) []byte {
	return nil
}

// pad takes a piece of data and pads it according to the total set in RS
func (r *RS) pad(data []byte) (out []byte) {
	dataLen := len(data)
	prefixBytes := make([]byte, 8)
	prefixLen := binary.PutUvarint(prefixBytes, uint64(dataLen))
	dataLen += prefixLen
	remainder := dataLen % r.required
	quotient := dataLen / r.required
	if remainder != 0 {
		quotient++
	}
	padLen := quotient * r.required
	out = make([]byte, padLen)
	copy(out, prefixBytes[:prefixLen])
	copy(out[prefixLen:], data)
	return
}

// unpad takes a padded piece of data and returns the unpadded content
func (r *RS) unpad(data []byte) []byte {
	dl, prefixLen := binary.Uvarint(data)
	dataLen := int(dl)
	if len(data) < dataLen+prefixLen {
		return nil
	}
	return data[prefixLen : dataLen+prefixLen]
}

// split returns a slice of byte slices split into r.required pieces
// the remaining empty shards are allocated in preparation for Encode to
// populate the remainder
func (r *RS) split(data []byte) (out [][]byte) {
	if len(data)%r.required == 0 {
		shardLen := len(data) / r.required
		out = make([][]byte, r.total)
		cursor := 0
		for i := 0; i < r.total; i++ {
			if i < r.required {
				out[i] = data[cursor : cursor+shardLen]
				cursor += shardLen
			} else {
				out[i] = make([]byte, shardLen)
			}
		}
	}
	return
}

func (r *RS) join(shards [][]byte) (out []byte) {
	// Only the first required shards are joined, as created by split, and
	// as is resultant from Decode (it only regenerates the data shards)
	for i := 0; i < r.required; i++ {
		out = append(out, shards[i]...)
	}
	return
}
