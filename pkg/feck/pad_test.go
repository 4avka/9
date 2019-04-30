package feck

import (
	"encoding/hex"
	"testing"
)

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
	blowout:=PadData(make([]byte, 65536*10), 3, 9)
	t.Log(blowout)
	t.Log(UnpadData(PadData(testDataUnaligned, 3, 9)[:10]))
}
