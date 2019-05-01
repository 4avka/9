package fek

import (
	"github.com/templexxx/reedsolomon"
)

const (
	ShardsTotal    = 9
	ShardsRequired = 3
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

// Encode returns a slice of the shards, each with first byte containing the
// shard number. Detecting their corruption requires
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
	for i, x := range splitted {
		splitted[i] = append([]byte{byte(i)}, x...)
		// splitted[i] = AppendChecksum(splitted[i])
	}
	return splitted
}

func (r *RS) Decode(shards [][]byte) (out []byte) {
	bytes := make(map[int][]byte)
	shardLens := make([]int, r.total)
	var ok bool
	for i, x := range shards {
		if shards[i], ok = VerifyChecksum(x); ok {
			bytes[int(shards[i][0])] = shards[i][1:]
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
