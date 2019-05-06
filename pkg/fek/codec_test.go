package fek

import (
	"crypto/rand"
	"testing"
)

func TestCodec(t *testing.T) {
	const (
		requiredshards = 3
		totalshards    = 9
	)
	testData := make([]byte, 64)
	_, _ = rand.Read(testData)
	// t.Log(len(testData), testData)
	rsc := New(requiredshards, totalshards)
	encoded := rsc.Encode(testData)

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
		}
	}
	unmuddled := rsc.Decode(muddled)
	// t.Log(len(unmuddled), unmuddled)
	if string(unmuddled) != string(testData) {
		t.Fatal("encode/decode failed")
	}
}
