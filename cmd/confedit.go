package cmd

import (
	"fmt"
	"sort"
	"strconv"

	"git.parallelcoin.io/dev/9/pkg/chain/fork"
	"git.parallelcoin.io/dev/9/pkg/util/cl"
	"github.com/AlecAivazis/survey"
)

func ParseNetwork(key string, name *string) int {
	k := (*config)[key].Initial.(string)
	prompt := &survey.Select{
		Message: "editing key " + key,
		Options: Networks,
		Default: k,
	}
	err := survey.AskOne(prompt, name, nil)
	if err != nil {
		fmt.Println("ERROR:", cl.Ine(), err)
	}
	*(*config)[key].Value.(*string) = *name
	cursor = "p2p"
	return 0
}

func ParseLogLevel(key string, name *string) int {
	options := make([]string, len(cl.Levels))
	j := 0
	for i := range cl.Levels {
		options[j] = i
		j++
	}
	sort.Strings(options)
	k := (*config)[key].Value.(*string)
	prompt := &survey.Select{
		Message: "editing key " + key,
		Options: options,
		Default: *k,
	}

	err := survey.AskOne(prompt, name, nil)
	if err != nil {
		fmt.Println("ERROR:", cl.Ine(), err)
	}
	*k = *name
	cursor = "log"
	return 0
}

func ParseAlgo(key string, name *string) int {
	options := []string{}
	for _, x := range fork.P9AlgoVers {
		options = append(options, x)
	}
	options = append(options, "random")
	sort.Strings(options)
	k := (*config)[key].Value.(*string)
	prompt := &survey.Select{
		Message: "editing key " + key,
		Options: options,
		Default: *k,
	}
	err := survey.AskOne(prompt, &name, nil)
	if err != nil {
		fmt.Println("ERROR:", cl.Ine(), err)
	}
	*k = *name
	cursor = "mining"
	return 0
}

func ParseInt(t *int, key string, name *string) int {
	for {
		name := fmt.Sprint(*t)
		prompt := &survey.Input{
			Message: key + ">",
			Default: name,
		}
		err := survey.AskOne(prompt, &name, nil)
		if err != nil {
			fmt.Println("ERROR:", cl.Ine(), err)
		}
		if (*config)[key].Validate((*config)[key], name) {
			prompt := &survey.Select{
				Message: key + " set to " + name,
				Options: []string{"ok", "cancel"},
			}
			var confirm string
			err := survey.AskOne(prompt, &confirm, nil)
			if err != nil {
				fmt.Println("ERROR:", cl.Ine(), err)
				continue
			}
			if confirm == "ok" {
				n, e := strconv.Atoi(name)
				if e != nil {
					return 0
				}
				*t = n
				return 0
			}
		} else {
			fmt.Println("value", name, "did not validate (out of bounds)")
			continue
		}
	}
}

func ParseBool(t *bool, key string, name *string) int {
	k := fmt.Sprint(!*(*config)[key].Value.(*bool))
	prompt := &survey.Select{
		Message: key + " set to " + k,
		Options: []string{"ok", "cancel"},
	}
	var confirm string
	err := survey.AskOne(prompt, &confirm, nil)
	if err != nil {
		fmt.Println("*bool ERROR:", err)
	}
	if confirm == "ok" {
		*t = !*t
	}
	return 0
}

func ParseString(t *string, key string, name *string) int {
	k := *t
	prompt := &survey.Input{
		Message: key + "> ",
		Default: k,
	}
	err := survey.AskOne(prompt, &k, nil)
	if err != nil {
		fmt.Println("ERROR:", cl.Ine(), err)
	}
	if (*config)[key].Validate((*config)[key], *name) {
		prompt := &survey.Select{
			Message: key + " set to " + k,
			Options: []string{"ok", "cancel"},
		}
		var confirm string
		err := survey.AskOne(prompt, &confirm, nil)
		if err != nil {
			fmt.Println("ERROR:", cl.Ine(), err)
		}
		if confirm == "ok" {
			(*config)[key].Value = &k
			// kk := (*config)[key].Value.(*string)
			// *kk = k
		}
	}
	return 0
}
