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
	shards := rs.Encode(uuid, testData)
	t.Log("encoded")
	for i := range shards {
		t.Log(len(shards[i]), shards[i][:4], shards[i][4:5],
			shards[i][5:len(shards[i])-8],
			shards[i][len(shards[i])-8:])
	}
	// puncture the heck out of the shards
	shards[0][6] = 0
	shards[1][6] = 0
	shards[2][6] = 0
	shards[4][6] = 0
	shards[6][6] = 0
	shards[8][6] = 0
	// reverse the order of the shards
	for i := 0; i < len(shards)/2; i++ {
		shards[i], shards[8-i] = shards[8-i], shards[i]
	}
	t.Log("damaged and reversed")
	for i := range shards {
		t.Log(len(shards[i]), shards[i])
	}
	result := rs.Decode(shards)
	t.Log("reconstituted with only 3 valid pieces")
	for i := range shards {
		t.Log(len(shards[i]), shards[i])
	}
	t.Log("result")
	t.Log(len(result), result)
	if bytes.Compare(testData, result) != 0 {
		t.Fatal("did not recover input bytes")
	} else {
		t.Log("output matches input")
	}
}
