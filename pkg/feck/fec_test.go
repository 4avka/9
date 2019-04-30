package feck

import (
	"encoding/binary"
	"encoding/hex"
	"testing"
)

var (
	testDataAligned   = []byte("123456789123456789123456789123456789123456789123456789123456789123456789123456789")
	testDataUnaligned = []byte("1234567891234567891234567891234567891234")
	expectedAligned   = "510031323334353637383931323334353637383931323334353637383931323334353637383931323334353637383931323334353637383931323334353637383931323334353637383931323334353637383900000000000000"
	expectedUnaligned = "280031323334353637383931323334353637383931323334353637383931323334353637383931323334000000"
	fecConfig         = New(3, 9)
)

func TestPadData(t *testing.T) {
	actualAligned := hex.EncodeToString(fecConfig.PadData(testDataAligned))
	actualUnaligned := hex.EncodeToString(fecConfig.PadData(testDataUnaligned))
	if actualAligned != expectedAligned {
		t.Fatalf("Padding did not produce expected result:\ngot      '%s'\nexpected '%s'",
			actualAligned, expectedAligned)
	}
	if actualUnaligned != expectedUnaligned {
		t.Fatalf("Padding did not produce expected result:\ngot      '%s'\nexpected '%s'",
			actualUnaligned, expectedUnaligned)
	}
	t.Log(testDataAligned)
	t.Log(fecConfig.PadData(testDataAligned))
	t.Log(UnpadData(fecConfig.PadData(testDataAligned)))
	t.Log(testDataUnaligned)
	t.Log(fecConfig.PadData(testDataUnaligned))
	t.Log(UnpadData(fecConfig.PadData(testDataUnaligned)))
}

// TODO: Make this codec work!

func TestFECCodec(t *testing.T) {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		t.Log("Recovered in f", r)
	// 	}
	// }()
	chunks := fecConfig.Encode(testDataAligned)
	t.Log(testDataAligned)
	t.Log(chunks)
	// Deface one of the pieces
	chunks[4][3] = ^chunks[4][3]
	// Here we only need 3 packets
	data, err := fecConfig.
		Decode(chunks[4:7])
	t.Log(data)
	if err != nil {
		t.Error(err)
	}
	// Requires one more across the punctured chunk to recover. This would not
	// normally happen as the checksums would usually filter out incorrect chunks.
	data, err = fecConfig.Decode(chunks[3:6])
	t.Log(data)
	if err != nil {
		t.Error(err)
	}
	dataLen := binary.LittleEndian.Uint16(data)
	result := data[2 : dataLen+2]
	dataString := hex.EncodeToString(data[2 : dataLen+2])
	resultString := hex.EncodeToString(result)
	if dataString != resultString {
		t.Fatalf("FEC encode/decode failed:\ngot      '%s'\nexpected '%s'", dataString, resultString)
	}
}
