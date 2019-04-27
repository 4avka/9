package main

import (
	"os"
)

func main() {
	app := NineApp()
	rv := app.Parse(os.Args)
	os.Exit(rv)
}
