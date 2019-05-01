package bola

import (
	"testing"

	"git.parallelcoin.io/dev/9/pkg/fek"
)

func TestBase(t *testing.T) {
	base0 := NewBase(BaseCfg{
		func(message Message) {},
		"127.0.0.1:1111",
		fek.New(3, 9),
		4096,
	})
	base1 := NewBase(BaseCfg{
		func(message Message) {},
		"127.0.0.1:2222",
		fek.New(3, 9),
		4096,
	})
	_, _ = base0, base1
}
