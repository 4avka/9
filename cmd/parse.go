package cmd

import (
	"fmt"
	"strconv"
	"strings"
)

func Parse(args []string) int {

	GenerateLines(lines)

	return 0

	for _, x := range positivetests {
		fmt.Println("\n", x)
		xx := strings.Split(strings.TrimSpace(x), " ")
		xxx := xx[1:]
		foundmap := make(map[string]string)
		for j, y := range xxx {

			_, e := strconv.Atoi(y)
			if e == nil {
				foundmap["number"] = y
			} else {

				switch y {
				case "help", "h":
					foundmap["help"] = y

				case "conf", "c":
					foundmap["conf"] = y

				case "listcommands", "l":
					foundmap["listcommands"] = y

				case "node", "n":
					foundmap["node"] = y

				case "shell", "s":
					foundmap["shell"] = y

				case "factory", "f":
					foundmap["factory"] = y

				case "new":
					foundmap["new"] = y

				case "copy", "cp":
					foundmap["copy"] = y

				case "rpc", "r":
					foundmap["rpc"] = y

				case "test", "t":
					foundmap["test"] = y

				default:
					if j == 0 {
						foundmap["datadir"] = y
					} else {

						foundmap["WTF"] = y
					}
				}
			}
		}
		fmt.Println("\t", foundmap)

		var rpccommand []string
		if _, ok := foundmap["rpc"]; ok {
			for i, v := range xx {
				if v == foundmap["rpc"] {
					rpccommand = xx[i+1:]
				}
			}
			fmt.Println("RPCCOMMAND:", rpccommand)
		}
		if _, ok := foundmap["listcommands"]; ok {
			fmt.Println("LISTCOMMANDS")
		}
		if i, ok := foundmap["datadir"]; ok {
			fmt.Println("DATADIR", i)
		}
		if _, ok := foundmap["test"]; ok {
			fmt.Println("TEST")
			if i, ok := foundmap["number"]; ok {
				fmt.Println("NUMBER", i)
			}
			if i, ok := foundmap["WTF"]; ok {
				fmt.Println("BASENAME", i)
			}
		}
		if _, ok := foundmap["new"]; ok {
			fmt.Println("NEW")
		}
		if _, ok := foundmap["copy"]; ok {
			fmt.Println("COPY")
		}
		if _, ok := foundmap["help"]; ok {
			fmt.Println("HELP")
		}
		if _, ok := foundmap["conf"]; ok {
			fmt.Println("CONF")
		}
	}

	return 0
}

var positivetests = []string{
	"9",
	"9 ~/.local/9/ help",
	"9 help",
	"9 ~/.local/9/ h",
	"9 h",
	"9 ~/.local/9/ conf",
	"9 conf",
	"9 ~/.local/9/ c",
	"9 c",
	"9 ~/.local/9/ listcommands",
	"9 listcommands",
	"9 ~/.local/9/ l",
	"9 l",
	"9 ~/.local/9/ rpc getinfo",
	"9 rpc getinfo",
	"9 ~/.local/9/ r getinfo",
	"9 rpc getblockhash 1414",
	"9 ~/.local/9/ r listreceivedbyaddress 1 true true",
	"9 r getinfo",
	"9 ~/.local/9/ node",
	"9 node",
	"9 ~/.local/9/ n",
	"9 n",
	"9 ~/.local/9/ shell",
	"9 shell",
	"9 ~/.local/9/ s",
	"9 s",
	"9 ~/.local/9/ factory",
	"9 factory",
	"9 ~/.local/9/ new",
	"9 new",
	"9 ~/.local/9/ 4 new",
	"9 new 19",
	"9 ~/.local/9/ new data",
	"9 new datadir",
	"9 ~/.local/9/ new data",
	"9 new 4 data ",
	"9 ~/.local/9/ 4 data new",
	"9 data 4 new",
	"9 ~/.local/9/ testy copy 3",
	"9 3 copy testy",
	"9 ~/.local/9/ test",
	"9 testy",
	"9 ~/.local/9/ test 2",
	"9 9 testy",
	"9 ~/.local/9/ log/path/ 5 test",
	"9 t 10 log/path/here/",
	"9 ~/.local/9/ test testy 9",
	"9 testy test 5",
	"9 ~/.local/9/ 5 test /tmp/logpath testy ",
	"9 /tmp/log/ 5 testy test",
	"9 t 10 log/path/here/",
	"9 ~/.local/9/ test testy 9",
	"9 testy t 5",
	"9 ~/.local/9/ 5 t /tmp/logpath testy ",
	"9 /tmp/log/ 5 testy test",
}
