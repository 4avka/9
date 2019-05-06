package fek

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"github.com/templexxx/reedsolomon"
)

type RS struct {
	*reedsolomon.RS
	required int
	total    int
}

func Split(data []byte, required, total int) [][]byte {
	b := make([][]byte, total)
	shardSize := len(data) / required
	cursor := 0
	for i := 0; i < total; i++ {
		if i < required {
			b[i] = data[cursor : cursor+shardSize]
		} else {
			b[i] = make([]byte, shardSize, shardSize)
		}
		cursor += shardSize
	}
	return b
}
func New(required, total int) *RS {
	rsc, err := reedsolomon.New(required, total-required)
	if err != nil {
		return nil
	}
	return &RS{rsc, required, total}
}

// Encode returns a slice of the shards, each with the same initial 4 byte UUID,
// the shard number, payload, and final 8 byte HighwayHash checksum
// By including a UUID, consumers of this library can identify shards of the same
// original packet without verifying the packet (that is all done in decode)
func (r *RS) Encode(data []byte) [][]byte {
	padded := PadData(data, r.required, r.total)
	splitted := Split(padded, r.required, r.total)
	var have, missing []int
	for i := 0; i < r.total; i++ {
		if i < r.required {
			have = append(have, i)
		} else {
			missing = append(missing, i)
		}
	}
	err := r.Reconst(splitted, have, missing)
	if err != nil {
		return nil
	}
	UUID := make([]byte, 4)
	_, err = rand.Read(UUID)
	if err != nil {
		// this is an event indicating the apocalypse is in process
		panic(err)
	}
	for i, x := range splitted {
		splitted[i] = append(UUID, append([]byte{byte(i)}, x...)...)
		splitted[i] = AppendChecksum(splitted[i])
	}
	return splitted
}

// Decode reverses the transformation from Encode. The shards must have the same
// UUID prefix or the bundle will be rejected.
func (r *RS) Decode(shards [][]byte) (out []byte) {
	bytes := make(map[int][]byte)
	shardLens := make([]int, r.total)
	var ok bool
	var uuid uint32
	for i, x := range shards {
		fmt.Println(x)
		if shards[i], ok = VerifyChecksum(x); ok {
			if len(shards[i]) < 6 {
				// A minimal payload of 1 byte takes 6 bytes, if the shard is
				// smaller than this the following code will cause bounds errors
				return nil
			}
			uuidb := shards[i][:4]
			shardnum := shards[i][4]
			payload := shards[i][5:]
			// the first 4 bytes of each shard should be the same
			u := binary.LittleEndian.Uint32(uuidb)
			if i == 0 {
				uuid = u
			} else if u != uuid {

				// UUID does not match, these can't be from the same set
				return nil
			}
			bytes[int(shardnum)] = payload
		}
	}
	for i, x := range bytes {
		shardLens[i] = len(x)
	}
	var length int
	for i, x := range shardLens {
		if i > 0 {
			if x > 0 {
				if length > 0 {
					if x != length {
						return nil
					}
				} else {
					length = x
				}
			}
		}
	}
	outSlice := make([][]byte, r.total)
	var have, missing []int
	for i := range outSlice {
		if y, ok := bytes[i]; ok {
			outSlice[i] = y
			have = append(have, i)
		} else {
			if i < r.required {
				missing = append(missing, i)
			}
			outSlice[i] = make([]byte, length)
		}
	}
	err := r.Reconst(outSlice, have[:r.required], missing)
	if err != nil {
		return nil
	}
	for i, x := range outSlice {
		if i > r.required {
			break
		}
		out = append(out, x...)
	}
	return UnpadData(out)
}
