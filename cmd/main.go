package cmd

import (
	"fmt"
)

func Start(args []string) int {
	fmt.Println("starting 9")

	fmt.Println(args)
	return 0
}

var _ = `

	pod <profile>
	
		launch gui <with different profile to default>
	

	pod <profile> help,h
	
		show help


	pod <profile> (cli)
	
		load rpc cli


	pod <profile> (conf)

		launch configuration cli
		

	pod <profile> (ctl|c)
		show listcommands


	pod <profile> (ctl,c) (getinfo/blahblah...)


	pod <profile> 

`
