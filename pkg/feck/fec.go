package feck

// Reed Solomon 9/3 forward error correction, intended to be sent as 9 pieces where 3 uncorrupted parts allows assembly of the message
import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"

	"github.com/vivint/infectious"
)

var (
	DefaultConfig = New(3, 9)
)

type Config struct {
	Total    int
	Required int
	Codec    *infectious.FEC
}

func New(required, total int) *Config {
	return &Config{
		Total:    total,
		Required: required,
		Codec: func() *infectious.FEC {
			fec, err := infectious.NewFEC(required, total)
			if err != nil {
				panic(fmt.Sprintf("error creating %d,%d FEC codec (this should not happen)", required, total))
			}
			return fec
		}(),
	}
}

// padData appends a 2 byte length prefix, and pads to a multiple of rsTotal.
// Note that the 16 bit prefix limits max chunk size to 64kb, thus creating a
// max of 64kb * number of required shares. If the data length is greater than
// this, the function returns an empty slice to indicate error
func (c *Config) PadData(data []byte) (out []byte) {
	dataLen := len(data)
	if dataLen > 65536*c.Required {
		return []byte{}
	}
	prefixBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(prefixBytes, uint16(dataLen))
	data = append(prefixBytes, data...)
	dataLen = len(data)
	chunkLen := (dataLen) / c.Total
	chunkMod := (dataLen) % c.Total
	if chunkMod != 0 {
		chunkLen++
	}
	padLen := c.Total*chunkLen - dataLen
	out = append(data, make([]byte, padLen)...)
	return
}

func UnpadData(data []byte) (out []byte) {
	prefixBytes := data[:2]
	payloadLen := binary.LittleEndian.Uint16(prefixBytes) //PutUint16(prefixBytes, uint16(dataLen))
	if len(data) < int(payloadLen) {
		// return empty slice to indicate error, payload is truncated
		return []byte{}
	}
	out = data[2 : payloadLen+2]
	return out
}

func (c *Config) Encode(data []byte) (chunks [][]byte) {
	// First we must pad the data
	data = c.PadData(data)
	shares := make([]infectious.Share, c.Total)
	output := func(s infectious.Share) {
		shares[s.Number] = s.DeepCopy()
	}
	err := c.Codec.Encode(data, output)
	if err != nil {
		panic(err)
	}
	for i := range shares {
		// Append the chunk number to the front of the chunk
		chunk := append([]byte{byte(shares[i].Number)}, shares[i].Data...)
		// Checksum includes chunk number byte so we know if its checksum is
		// incorrect so could the chunk number be
		checksum := crc32.Checksum(chunk, crc32.MakeTable(crc32.Castagnoli))
		checkbytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(checkbytes, checksum)
		chunk = append(chunk, checkbytes...)
		chunks = append(chunks, chunk)
	}
	return
}

func (c *Config) Decode(chunks [][]byte) (data []byte, err error) {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		fmt.Println("Recovered in f", r)
	// 	}
	// }()
	var shares []infectious.Share
	for i := range chunks {
		bodyLen := len(chunks[i])
		body := chunks[i][:bodyLen-4]
		plcheckword := binary.LittleEndian.Uint32(chunks[i][bodyLen-4:]) //PutUint16(prefixBytes, uint16(dataLen))
		checksum := crc32.Checksum(chunks[i],
			crc32.MakeTable(crc32.Castagnoli))
		if checksum != plcheckword {
			continue
		}
		share := infectious.Share{
			Number: int(body[0]),
			Data:   body[1:],
		}
		shares = append(shares, share)
	}
	data, err = c.Codec.Decode(nil, shares)
	data = UnpadData(data)
	if len(data) == 0 {
		err = errors.New("data is somehow truncated")
	}
	return
}
