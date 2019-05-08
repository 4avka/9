package fek

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func TestPadUnpad(t *testing.T) {
	rs := New(5, 9)
	testData := make([]byte, 73)
	_, _ = rand.Read(testData)
	t.Log(len(testData), testData)
	padded := rs.pad(testData)
	t.Log(len(padded), padded)
	unpadded := rs.unpad(padded)
	t.Log(len(unpadded), unpadded)
}

func TestSplitJoin(t *testing.T) {
	rs := New(3, 9)
	testData := make([]byte, 67)
	_, _ = rand.Read(testData)
	t.Log(len(testData), testData)
	padded := rs.pad(testData)
	t.Log(len(padded), padded)
	splitted := rs.split(padded)
	for i := range splitted {
		t.Log(len(splitted[i]), splitted[i])
	}
	joined := rs.join(splitted)
	t.Log(len(joined), joined)
	unpadded := rs.unpad(padded)
	t.Log(len(unpadded), unpadded)
}

func TestCodec(t *testing.T) {
	rs := New(3, 9)
	testData := make([]byte, 61)
	_, _ = rand.Read(testData)
	t.Log("original")
	t.Log(len(testData), testData)
	uuid := GetUUID()
	frags := rs.Encode(uuid, testData)
	t.Log("encoded")
	for i := range frags {
		// frags[i] = AppendChecksum(frags[i])
		t.Log(len(frags[i]), frags[i][:4], frags[i][4:5],
			frags[i][5:len(frags[i])-8],
			frags[i][len(frags[i])-8:])
	}
	// puncture the heck out of the frags and jumble them
	frags = [][]byte{frags[5], frags[3], frags[7]}
	t.Log("damaged and jumbled")
	for i := range frags {
		t.Log(len(frags[i][5:len(frags[i])-8]), frags[i][5:len(frags[i])-8])
	}
	result := rs.Decode(frags)
	// t.Log("reconstituted with only 3 valid pieces")
	// for i := range frags {
	// 	t.Log(len(frags[i][1:]), frags[i][1:])
	// }
	t.Log("result")
	t.Log(len(result), result)
	if bytes.Compare(testData, result) != 0 {
		t.Fatal("did not recover input bytes")
	} else {
		t.Log("output matches input")
	}
}
