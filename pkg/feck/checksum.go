package feck
import (
	"encoding/binary"
	"github.com/minio/highwayhash"
)
var zerokey = make([]byte, 32)
func AppendChecksum(in []byte) []byte {
	return append(in, Uint64ToBytes(highwayhash.Sum64(in, zerokey))...)
}
func VerifyChecksum(in []byte) (out []byte, verified bool) {
	l := len(in)
	if l > 8 {
		computed := highwayhash.Sum64(in[:l-8], zerokey)
		if computed == BytesToUint64(in[l-8:]) {
			out = in[:l-8]
			verified = true
		} else {
			out = make([]byte, l-9)
		}
	}
	return
}
// Uint64ToBytes - returns a byte slice from uint64 - required because Murmur3 takes bytes as input but returns uint32
func Uint64ToBytes(input uint64) []byte {
	p := make([]byte, 8)
	binary.LittleEndian.PutUint64(p, input)
	return p
}
// BytesToUint64 - converts 4 byte slice to uint32
func BytesToUint64(bytes []byte) uint64 {
	// We are taking off the seatbelt here for performance reasons. We know that
	// the hash is uint64 thus we won't check it is right. Uint64() will panic
	// either way if the slice is not at least 8 bytes
	// if len(bytes) != 8 {
	// 	log.Fatal("Byte slice is not 8 bytes long")
	// }
	return binary.LittleEndian.Uint64(bytes)
}
// func ZeroBytes(b []byte) {
// 	for i := range b {
// 		b[i] = 0
// 	}
// }
