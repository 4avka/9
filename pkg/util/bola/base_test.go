package bola

import (
	"fmt"
	"testing"
	"time"

	"git.parallelcoin.io/dev/9/pkg/fek"
)

var base0Address = "127.0.0.1:1111"
var base1Address = "127.0.0.1:2222"

func TestBase(t *testing.T) {
	// cl.Color = false
	base0 := NewBase(BaseCfg{
		func(message Message) {
			fmt.Println(message)
		},
		base0Address,
		fek.New(3, 9),
		4096,
	})
	base1 := NewBase(BaseCfg{
		func(message Message) {
			fmt.Println(message)
		},
		base1Address,
		fek.New(3, 9),
		4096,
	})

	if err := base0.Start(); err != nil {
		t.Fatal(err)
	}
	if err := base1.Start(); err != nil {
		t.Fatal(err)
	}

	if e := base0.Send([]byte("hello world!"), base1Address); e != nil {
		t.Fatal(e)
	}
	if e := base1.Send([]byte("hello yourself! HARRUMPH!"), base0Address); e != nil {
		t.Fatal(e)
	}
	// time.Sleep(time.Second)
	// base0.Stop()
	// base1.Stop()

	time.Sleep(time.Second) // / 10)

	_, _ = base0, base1
}
