package fek

import (
	"crypto/rand"
	"testing"
)

func TestPad(t *testing.T) {
	rs := New(5, 9)
	testData := make([]byte, 73)
	_, _ = rand.Read(testData)
	t.Log(len(testData), testData)
	padded := rs.pad(testData)
	t.Log(len(padded), padded)
	unpadded := rs.unpad(padded)
	t.Log(len(unpadded), unpadded)
}

func TestSplit(t *testing.T) {
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
