package cmd

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"git.parallelcoin.io/dev/9/pkg/chain/fork"
	"git.parallelcoin.io/dev/9/pkg/util/cl"
	"github.com/AlecAivazis/survey"
)

func RunConf(args []string, tokens Tokens, cmds, all Commands) int {
	fmt.Println("ⓟarallelcoin configuration CLI")
	ConfMain()
	return 0
}

func ConfMain() int {
	for {
		var options []string
		var lines []string
		for i := range Config {
			lines = append(lines, i)
		}
		for i, x := range lines {
			lines[i] = "configure:" + strings.Split(x, ".")[0]
		}
		options = uniq(lines)
		sort.Strings(options)
		options = append([]string{
			"run: select a server to run",
		}, options...)
		options = append(options, "exit")
		prompt := &survey.Select{
			Message:  "ⓟ",
			Options:  options,
			PageSize: 9,
		}
		var name string
		err := survey.AskOne(prompt, &name, nil)
		if err != nil {
			fmt.Println("ERROR:", err)
		}
		if name == "exit" {
			return 0
		}
		prefix := strings.Split(name, ":")[0]
		suffix := strings.Split(name, ":")[1]
		// fmt.Println("section:", prefix)
		// fmt.Println("subsection:", suffix)
		switch prefix {
		case "run":
			ConfRun(suffix)
			goto out
		case "configure":
			ConfConf(suffix)
		}
	}
out:
	return 0
}

func ConfRun(subsection string) int {
	prompt := &survey.Select{
		Message: "select server to run:",
		Options: []string{"node", "wallet", "shell"},
	}
	var name string
	err := survey.AskOne(prompt, &name, nil)
	if err != nil {
		return 1
	}
	return 0
}

func ConfConf(subsection string) int {
	for {
		// fmt.Println("configure:", subsection)
		var lines []string
		re := regexp.MustCompile("(" + subsection + "[.])(.*)")
		for i := range Config {
			if re.Match([]byte(i)) {
				sects := re.FindAllStringSubmatch(i, 1)
				c := Config[i]
				item := fmt.Sprintf("%s : %v (%v) = %v", sects[0][2], c.Comment, c.Default, c.Value)
				lines = append(lines, item)
			}
		}
		sort.Strings(lines)
		lines = append(lines, "exit")
		prompt := &survey.Select{
			Message:  "configuration:" + subsection + " ",
			Options:  lines,
			Help:     "select the variable to edit",
			PageSize: 9,
		}
		var name string
		err := survey.AskOne(prompt, &name, nil)
		if err != nil {
			fmt.Println("ERROR:", err)
		}
		name = strings.Split(name, " ")[0]
		if name == "exit" {
			break
		}
		// fmt.Printf("editing %s:%s\n", subsection, name)
		key := subsection + "." + name
		return ConfConfEdit(key)
	}
	return 0
}

func ConfConfEdit(key string) int {
	if _, ok := Config[key]; !ok {
		fmt.Println("key not found:", key)
		return 1
	}
	// fmt.Println("editing key", key)
	// spew.Dump(Config)
	// fmt.Println("var type", reflect.TypeOf(Config[key].Value))
	var name string
	switch key {
	case "p2p.network":
		prompt := &survey.Select{
			Message: "editing key " + key,
			Options: Networks,
			Default: Config[key].Value.(string),
		}
		err := survey.AskOne(prompt, &name, nil)
		if err != nil {
			fmt.Println("ERROR:", err)
		}
		Config[key].Value = name
	case "log.level":
		var options []string
		for i := range cl.Levels {
			options = append(options, i)
		}
		sort.Strings(options)
		prompt := &survey.Select{
			Message: "editing key " + key,
			Options: options,
			Default: Config[key].Value.(string),
		}

		err := survey.AskOne(prompt, &name, nil)
		if err != nil {
			fmt.Println("ERROR:", err)
		}
		Config[key].Value = name
	case "mining.algo":
		options := []string{}
		for _, x := range fork.P9AlgoVers {
			options = append(options, x)
		}
		options = append(options, "random")
		sort.Strings(options)
		prompt := &survey.Select{
			Message: "editing key " + key,
			Options: options,
			Default: Config[key].Value.(string),
		}
		err := survey.AskOne(prompt, &name, nil)
		if err != nil {
			fmt.Println("ERROR:", err)
		}
		Config[key].Value = name
	default:
		// switch on type
		switch t := Config[key].Value.(type) {
		case int:
			fmt.Println("int", t)
		case bool:
			t = !t
			Config[key].Value = t
			fmt.Println("set", key, "to", t, ":", Config[key].Comment)
		case string:
			fmt.Println("string", t)
		case []string:
			fmt.Println("[]string", t)
		default:
			fmt.Println(
				"type not handled:",
				reflect.TypeOf(Config[key].Value))
		}
	}
	return 0
}
