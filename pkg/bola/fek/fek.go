package fek

import (
	"crypto/rand"
	"encoding/binary"

	"github.com/minio/highwayhash"
	"github.com/templexxx/reedsolomon"
)

var Zerokey = make([]byte, 32)

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

func (r *RS) Encode(uuid []byte, data []byte) [][]byte {
	padded := r.pad(data)
	splitted := r.split(padded)
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
	for i := range splitted {
		splitted[i] = AppendChecksum(append(uuid, append([]byte{byte(i)}, splitted[i]...)...))
	}
	return splitted
}

func (r *RS) Decode(shards [][]byte) []byte {
	return nil
}

// pad takes a piece of data and pads it according to the total required by RS
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

// GetUUID returns a cryptographically secure 8 byte UUID
func GetUUID() []byte {
	uuid := make([]byte, 4)
	n, e := rand.Read(uuid)
	if n != 4 || e != nil {
		panic(e)
	}
	return uuid
}

// UUIDtoUint64 converts the UUID to a comparable uint64
func UUIDtoUint32(uuid []byte) uint32 {
	if len(uuid) != 4 {
		u := make([]byte, 4)
		copy(u, uuid)
	}
	return binary.LittleEndian.Uint32(uuid)
}

func Uint32toUUID(uuid uint32) (out []byte) {
	out = make([]byte, 4)
	binary.LittleEndian.PutUint32(out, uuid)
	return
}

func AppendChecksum(in []byte) []byte {
	return append(in, Uint64ToBytes(highwayhash.Sum64(in, Zerokey))...)
}

// Uint64ToBytes - returns a byte slice from uint64 - required because highwayhash takes bytes as input but returns uint32
func Uint64ToBytes(input uint64) (out []byte) {
	out = make([]byte, 8)
	for i := range out {
		out[i] = byte(input >> uint(i*8))
	}
	return
}

// BytesToUint64 - converts 4 byte slice to uint32
func BytesToUint64(bytes []byte) (out uint64) {
	_ = bytes[7]
	// We are taking off the seatbelt here for performance reasons. We know that
	// the hash is uint64 thus we won't check it is right. Uint64() will panic
	// either way if the slice is not at least 8 bytes
	for i, x := range bytes {
		out += uint64(x) << uint(i*8)
	}
	return
}
