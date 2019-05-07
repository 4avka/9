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

// Encode takes a uuid and a slice of bytes, pads, splits, generates parity
// shards, then prepends UUID and shard number, and finally appends a 64 bit
// HighwayHash checksum.
// The UUID allows the packet receiver to quickly group without checking the
// checksum, which will be used by the decoder to erase invalid shards. The
// shard number prefix is used to put the shards in their correct position for
// decoding
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
		// append uuid and shard number to front of each shard and append checksum
		splitted[i] = AppendChecksum(append(uuid, append([]byte{byte(i)}, splitted[i]...)...))
	}
	return splitted
}

// Decode takes a set of shards which are assumed to have the same 4 byte UUID
// prefix (for tying batches of shards together), so this isn't checked, but if
// there is different batches together shard numbers can collide and we aren't
// going to handle this for performance reasons, since it can be cheaply avoided
func (r *RS) Decode(shards [][]byte) (out []byte) {
	erased := make([]bool, len(shards))
	for i := range shards {
		// First, check each shard's checksum is correct. Incorrect shards are
		// zeroed by the check function, and the checksum is cut from the shard
		// if it validates
		shards[i], erased[i] = DetectAndEraseIfError(shards[i])
		// Next remove the UUID prefix
		shards[i] = shards[i][4:]
	}
	work := make([][]byte, len(shards))
	var erasedShards [][]byte
	for i := range shards {
		if !erased[i] {
			work[shards[i][0]] = shards[i][1:]
		} else {
			erasedShards = append(erasedShards, shards[i][1:])
		}
	}
	counter := 0
	missing := []int{}
	found := []int{}
	for i := range shards {
		if work[i] == nil {
			missing = append(missing, i)
			shards[i] = erasedShards[counter]
			counter++
		} else {
			shards[i] = work[i]
			found = append(found, i)
		}
	}
	err := r.Reconst(shards, found[:r.required], missing)
	if err != nil {
		return nil
	}
	for i, x := range shards {
		if i > r.required {
			break
		}
		out = append(out, x...)
	}
	dl, prefixLen := binary.Uvarint(out)
	return out[prefixLen : prefixLen+int(dl)]
}

// pad takes a piece of data and pads it according to the total and required by
// the configured RS codec. The pad and split is determined by the parameters of
// the codec so these functions are part of the type
func (r *RS) pad(data []byte) (out []byte) {
	dataLen := len(data)
	prefixBytes := make([]byte, 8)
	prefixLen := binary.PutUvarint(prefixBytes, uint64(dataLen))
	prefixBytes = prefixBytes[:prefixLen]
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
		for i := range out {
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

func AppendChecksum(in []byte) []byte {
	return append(in, Uint64ToBytes(highwayhash.Sum64(in, Zerokey))...)
}

// DetectAndEraseIfError checks that the appended 64 bit checksum is correct
// If it is incorrect, an empty slice same size as the payload will be returned
// which is what the decoder needs
func DetectAndEraseIfError(in []byte) (out []byte, erased bool) {
	l := len(in)
	if l > 8 {
		var checksum uint64
		out, checksum = in[:l-8], BytesToUint64(in[l-8:])
		computed := highwayhash.Sum64(out, Zerokey)
		if computed != checksum {
			erased = true
			// zero out slice (this should be compiled to a single memset op)
			for i := range out {
				out[i] = 0
			}
		}
	}
	return
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
	for i, x := range bytes {
		out += uint64(x) << uint(i*8)
	}
	return
}

// UUIDtoUint64 converts the UUID to a comparable uint64
func UUIDtoUint32(uuid []byte) uint32 {
	if len(uuid) != 4 {
		u := make([]byte, 4)
		copy(u, uuid)
	}
	// _ = uuid[3]
	return binary.LittleEndian.Uint32(uuid)
}

func Uint32toUUID(uuid uint32) (out []byte) {
	out = make([]byte, 4)
	binary.LittleEndian.PutUint32(out, uuid)
	return
}
