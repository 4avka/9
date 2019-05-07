package fek

import (
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
	testData := make([]byte, 70)
	_, _ = rand.Read(testData)
	t.Log(len(testData), testData)
	uuid := GetUUID()
	shards := rs.Encode(uuid, testData)
	for i := range shards {
		t.Log(shards[i])
	}
}
