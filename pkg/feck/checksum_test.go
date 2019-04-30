package feck

import (
	"testing"
)

func TestChecksum(t *testing.T) {
	test := []byte("hello world!")
	t.Log(string(test))
	t.Log(test)
	checked := AppendChecksum(test)
	t.Log(checked)
	verified, ok := VerifyChecksum(checked)
	t.Log(ok, verified)
}
