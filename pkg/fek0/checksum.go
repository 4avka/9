package fek

import (
	"github.com/minio/highwayhash"
)

var Zerokey = make([]byte, 32)

// func GetChecksum(in []byte) []byte {
// 	return Uint64ToBytes(highwayhash.Sum64(in, Zerokey))
// }

func AppendChecksum(in []byte) []byte {
	return append(in, Uint64ToBytes(highwayhash.Sum64(in, Zerokey))...)
}

func VerifyChecksum(in []byte) (out []byte, verified bool) {
	l := len(in)
	if l > 8 {
		computed := highwayhash.Sum64(in[:l-8], Zerokey)
		if computed == BytesToUint64(in[l-8:]) {
			out = in[:l-8]
			verified = true
		} else {
			out = make([]byte, l-9)
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
	_ = bytes[7]
	// We are taking off the seatbelt here for performance reasons. We know that
	// the hash is uint64 thus we won't check it is right. Uint64() will panic
	// either way if the slice is not at least 8 bytes
	for i, x := range bytes {
		out += uint64(x) << uint(i*8)
	}
	return
}

// func ZeroBytes(b []byte) {
// 	for i := range b {
// 		b[i] = 0
// 	}
// }