package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("initialising App")
	app := NineApp()
	// j, e := json.MarshalIndent(app, "", "\t")
	// if e != nil {
	// 	panic(e)
	// }
	// fmt.Println(string(j))

	fmt.Println("linking configuration")
	app.Config = MakeConfig(app)
	// spew.Dump(cfg)
	fmt.Println("parsing CLI args")
	rv := app.Parse(os.Args)
	// spew.Dump(cfg)
	fmt.Println("done")
	os.Exit(rv)
}
