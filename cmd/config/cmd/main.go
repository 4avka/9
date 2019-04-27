package main

import (
	"os"
)

func main() {
	app := NineApp()
	app.Config = MakeConfig(app)
	rv := app.Parse(os.Args)
	os.Exit(rv)
}
