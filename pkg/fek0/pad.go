// Package fek is a Reed Solomon 9/3 forward error correction, intended to be 
// sent as 9 pieces where 3 uncorrupted parts allows assembly of the message
package fek

import (
	"encoding/binary"
)

// padData appends a 2 byte length prefix, and pads to a multiple of total.
// Note that the 16 bit prefix limits max chunk size to 64kb, thus creating a
// max of 64kb * number of required shares. If the data length is greater than
// this, the function returns an nil slice to indicate error
func PadData(data []byte, required, total int) []byte {
	dataLen := len(data)
	if dataLen > 65536*required || dataLen < 1 {
		return nil
	}
	prefixBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(prefixBytes, uint16(dataLen))
	data = append(prefixBytes, data...)
	dataLen = len(data)
	chunkLen := (dataLen) / total
	chunkMod := (dataLen) % total
	if chunkMod != 0 {
		chunkLen++
	}
	padLen := total*chunkLen - dataLen
	return append(data, make([]byte, padLen)...)
}

// UnpadData reverses the padding applied in PadData by reading the length
// prefix and slicing off the extra bytes
func UnpadData(data []byte) (out []byte) {
	prefixBytes := data[:2]
	payloadLen := binary.LittleEndian.Uint16(prefixBytes) //PutUint16(prefixBytes, uint16(dataLen))
	if len(data) < int(payloadLen) {
		// return empty slice to indicate error, payload is truncated
		return nil
	}
	out = data[2 : payloadLen+2]
	return out
}
