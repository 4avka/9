package bola

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"testing"
)

func TestCodec(t *testing.T) {
	const (
		requiredshards = 3
		totalshards    = 9
	)
	testData := make([]byte, 64)
	_, _ = rand.Read(testData)
	t.Log(len(testData), testData)
	rsc := New(requiredshards, totalshards)
	encoded := rsc.Encode(testData)
	for _, x := range encoded {
		t.Log(len(x), x)
		uuidb := x[:4]
		shardnum := x[4]
		payload := x[5:]
		t.Log(binary.LittleEndian.Uint32(uuidb), shardnum, payload)
	}
	muddlemap := make(map[int][]byte)
	for i, x := range encoded {
		muddlemap[i] = x
	}
	// By copying the array into a slice and reading back out we unsort it
	var muddled [][]byte
	for _, x := range muddlemap {
		muddled = append(muddled, x)
	}
	for i := range muddled {
		// t.Log(i, len(x), x)
		if i%2 == 0 || i%3 == 0 {
			muddled[i][3] = ^muddled[i][3]
			// t.Log(i, len(x), x, "MANG")
		} else {
			// t.Log(i, len(x), x)
		}
	}
	unmuddled := rsc.Decode(muddled)
	t.Log(len(unmuddled), unmuddled)
	if string(unmuddled) != string(testData) {
		t.Fatal("encode/decode failed")
	}
}

var (
	testDataAligned   = []byte("123456789123456789123456789123456789123456789123456789123456789123456789123456789")
	testDataUnaligned = []byte("1234567891234567891234567891234567891234")
	expectedAligned   = "510031323334353637383931323334353637383931323334353637383931323334353637383931323334353637383931323334353637383931323334353637383931323334353637383931323334353637383900000000000000"
	expectedUnaligned = "280031323334353637383931323334353637383931323334353637383931323334353637383931323334000000"
)

func TestPadData(t *testing.T) {
	actualAligned := hex.EncodeToString(PadData(testDataAligned, 3, 9))
	actualUnaligned := hex.EncodeToString(PadData(testDataUnaligned, 3, 9))
	if actualAligned != expectedAligned {
		t.Fatalf("Padding did not produce expected result:\ngot      '%s'\nexpected '%s'",
			actualAligned, expectedAligned)
	}
	if actualUnaligned != expectedUnaligned {
		t.Fatalf("Padding did not produce expected result:\ngot      '%s'\nexpected '%s'",
			actualUnaligned, expectedUnaligned)
	}
	t.Log(testDataAligned)
	t.Log(PadData(testDataAligned, 3, 9))
	t.Log(UnpadData(PadData(testDataAligned, 3, 9)))
	t.Log(testDataUnaligned)
	t.Log(PadData(testDataUnaligned, 3, 9))
	t.Log(UnpadData(PadData(testDataUnaligned, 3, 9)))
	blowout := PadData(make([]byte, 65536*10), 3, 9)
	t.Log(blowout)
	t.Log(UnpadData(PadData(testDataUnaligned, 3, 9)[:10]))
}

func TestChecksum(t *testing.T) {
	test := []byte("hello world!")
	t.Log(string(test))
	t.Log(test)
	checked := AppendChecksum(test)
	t.Log(checked)
	verified, ok := VerifyChecksum(checked)
	t.Log(ok, verified)
}
