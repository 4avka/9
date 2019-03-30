package cmd

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

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
			"run:node",
			"run:wallet",
			"run:shell",
		}, options...)
		options = append(options, "exit")
		prompt := &survey.Select{
			Message: "please select an option, type to filter",
			Options: options,
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
	fmt.Println("running:", subsection)
	return 0
}

func ConfConf(subsection string) int {
	fmt.Println("configure:", subsection)
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
	lines = append(lines, "exit")
	prompt := &survey.Select{
		Message: "configuration:" + subsection + " ",
		Options: lines,
	}
	var name string
	err := survey.AskOne(prompt, &name, nil)
	if err != nil {
		fmt.Println("ERROR:", err)
	}
	if name == "exit" {
		return 0
	}

	// for i, x := range lines {
	// 	fmt.Println("line", i, x)
	// }
	return 0
}
