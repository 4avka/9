package main

import (
	. "git.parallelcoin.io/dev/9/cmd/tmp"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	a := NewApp("9",
		Version("v1.9.9"),
		Group("app"),
	)
	spew.Dump(a)
}
