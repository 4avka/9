package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"git.parallelcoin.io/dev/9/pkg/chain/fork"
	"git.parallelcoin.io/dev/9/pkg/util/cl"
	"github.com/AlecAivazis/survey"
)

const BACK = "back"

func RunConf(args []string, tokens Tokens, cmds, all Commands) int {
	// fmt.Println("ⓟarallelcoin configuration CLI")
	runner := ConfMain()
	switch runner {
	case "node":
		Node(args, tokens, cmds, all)
	case "wallet":
		Wallet(args, tokens, cmds, all)
	case "shell":
		Shell(args, tokens, cmds, all)
	case "", BACK:
		return 2
	case "exit":
		return 0
	default:
		return 1
	}
	return 0
}

func ConfMain() string {
	cont := true
	for cont {
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
			Message:  "ⓟarallelcoin Interactive CLI",
			Options:  options,
			PageSize: 9,
		}
		var name string
		err := survey.AskOne(prompt, &name, nil)
		if err != nil {
			fmt.Println("ERROR:", err)
		}
		if name == "exit" {
			return name
		}
		split := strings.Split(name, ":")
		prefix := split[0]
		var suffix string
		if len(split) > 1 {
			suffix = split[1]
		}
		switch prefix {
		case "run":
			return ConfRun()
		case "configure":
			if suffix != "" {
				ConfConf(suffix)
			}
		}
	}
	return ""
}

func ConfRun() string {
	prompt := &survey.Select{
		Message: "select server to run:",
		Options: []string{"node", "wallet", "shell", BACK},
	}
	var name string
	err := survey.AskOne(prompt, &name, nil)
	if err != nil {
		return err.Error()
	}
	if name == BACK {
		return ""
	}
	return name
}

var cursor string

func ConfConf(subsection string) int {
	cursor = subsection
	for {
		// fmt.Println("configure:", subsection)
		var lines []string
		re := regexp.MustCompile("(" + subsection + "[.])(.*)")
		for i := range Config {
			if re.Match([]byte(i)) {
				sects := re.FindAllStringSubmatch(i, 1)
				c := Config[i]
				value := ""
				switch t := c.Value.(type) {
				case *bool:
					value = fmt.Sprint(*t)
				case *int:
					value = fmt.Sprint(*t)
				case *float64:
					value = fmt.Sprintf("%.10f", *t)
				case *string:
					value = *t
				case *[]string:
					ss := *t
					ll := len(ss) - 1
					for i, x := range ss {
						value += x
						if i < ll {
							value += " "
						}
					}
				default:
					// if we don't recognise it we can't print it
					continue
				}

				item := fmt.Sprintf("%s : %v (%v) = %v", sects[0][2], c.Comment, c.Default, value)
				lines = append(lines, item)
			}
		}
		sort.Strings(lines)
		lines = append(lines, BACK)
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
		if name == BACK {
			break
		}
		// fmt.Printf("editing %s:%s\n", subsection, name)
		key := subsection + "." + name
		if ConfConfEdit(key) != 0 {
			break
		}
		datadir := Config["app.datadir"].Value.(string)
		configFile := CleanAndExpandPath(filepath.Join(datadir, "config"))
		if EnsureDir(configFile) {
			fmt.Println("created new directory to store data", datadir)
		}
		fh, err := os.Create(configFile)
		if err != nil {
			panic(err)
		}
		_, err = fmt.Fprint(fh, Config)
		if err != nil {
			panic(err)
		}
	}
	return 0
}

func ConfConfEdit(key string) int {
	if _, ok := Config[key]; !ok {
		// fmt.Println("key not found:", key)
		return 1
	}
	// fmt.Println("editing key", key)
	// spew.Dump(Config)
	// fmt.Println("var type", reflect.TypeOf(Config[key].Value))
	for {
		name := new(string)
		switch key {
		case "p2p.network":
			k := Config[key].Value.(*string)
			prompt := &survey.Select{
				Message: "editing key " + key,
				Options: Networks,
				Default: *k,
			}
			err := survey.AskOne(prompt, &name, nil)
			if err != nil {
				fmt.Println("ERROR:", err)
			}
			*k = *name
			cursor = "p2p"
			return 0
		case "log.level":
			var options []string
			for i := range cl.Levels {
				options = append(options, i)
			}
			sort.Strings(options)
			k := Config[key].Value.(*string)
			prompt := &survey.Select{
				Message: "editing key " + key,
				Options: options,
				Default: *k,
			}

			err := survey.AskOne(prompt, &name, nil)
			if err != nil {
				fmt.Println("ERROR:", err)
			}
			*k = *name
			cursor = "log"
			return 0
		case "mining.algo":
			options := []string{}
			for _, x := range fork.P9AlgoVers {
				options = append(options, x)
			}
			options = append(options, "random")
			sort.Strings(options)
			k := Config[key].Value.(*string)
			prompt := &survey.Select{
				Message: "editing key " + key,
				Options: options,
				Default: *k,
			}
			err := survey.AskOne(prompt, &name, nil)
			if err != nil {
				fmt.Println("ERROR:", err)
			}
			*k = *name
			cursor = "mining"
			return 0
		default:
			cursor = strings.Split(key, ".")[0]
			// switch on type
			switch t := Config[key].Value.(type) {
			case *int:
				for {
					name := fmt.Sprint(*t)
					prompt := &survey.Input{
						Message: key + ">",
						Default: name,
					}
					err := survey.AskOne(prompt, &name, nil)
					if err != nil {
						fmt.Println("ERROR:", err)
					}
					if Config[key].Validator(name) {
						prompt := &survey.Select{
							Message: key + " set to " + name,
							Options: []string{"ok", "cancel"},
						}
						var confirm string
						err := survey.AskOne(prompt, &confirm, nil)
						if err != nil {
							fmt.Println("ERROR:", err)
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
			case *bool:
				k := fmt.Sprint(!*Config[key].Value.(*bool))
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
			case *string:
				k := fmt.Sprint(*Config[key].Value.(*string))
				prompt := &survey.Input{
					Message: key + "> ",
					Default: k,
				}
				err := survey.AskOne(prompt, &k, nil)
				if err != nil {
					fmt.Println("ERROR:", cl.Ine(), err)
				}
				if Config[key].Validator(*name) {
					prompt := &survey.Select{
						Message: key + " set to " + k,
						Options: []string{"ok", "cancel"},
					}
					var confirm string
					err := survey.AskOne(prompt, &confirm, nil)
					if err != nil {
						fmt.Println("ERROR:", err)
					}
					if confirm == "ok" {
						Config[key].Value = &k
						// kk := Config[key].Value.(*string)
						// *kk = k
					}
				}
				return 0
			case *[]string:
				again := true
				for again {
					// t = Config[key].Value.(*[]string)
					prompt := &survey.Select{
						Message: key + ">",
						Options: append(append([]string{"new"}, *t...), BACK),
					}
					var name string
					err := survey.AskOne(prompt, name, nil)
					if err != nil {
						fmt.Println("ERROR:", err)
					}
					switch name {
					case BACK:
						again = false
						return 0
					case "new":
						again2 := true
						for again2 {
							var item string
							prompt := &survey.Input{
								Message: "new item on " + key + ">",
								Default: item,
							}
							err = survey.AskOne(prompt, &item, nil)
							if err != nil {
								fmt.Println("ERROR:", err)
								break
							}
							if Config[key].Validator(item) {
								prompt := &survey.Select{
									Message: "confirm '" + item + "'",
									Options: []string{"ok", "edit", "cancel"},
								}
								confirm := ""
								err = survey.AskOne(prompt, &confirm, nil)
								if err != nil {
									fmt.Println("ERROR:", err)
									break
								}
								if confirm == "ok" {
									return 0
								}
								if confirm == "edit" {
									again2 = true
								}
								if confirm == "cancel" {
									again2 = false
								}
							}
						}
					default:
						prompt := &survey.Select{
							Message: key + ">" + name,
							Options: []string{"delete", "edit", "cancel"},
						}
						var confirm string
						err = survey.AskOne(prompt, &confirm, nil)
						if err != nil {
							fmt.Println("ERROR:", err)
							break
						}
						if confirm == "delete" {
							again = true
							v := Config[key].Value.([]string)
							for i, x := range v {
								if x == name {
									if i < len(v)-1 {
										v = append(v[:i], v[i+1:]...)
										Config[key].Value = v
									} else {
										Config[key].Value = v[:i]
									}
								}
							}
						}
						if confirm == "edit" {
							prompt := &survey.Input{
								Message: key + ">" + name,
								Default: name,
							}
							var edit string
							err := survey.AskOne(prompt, &edit, nil)
							if err != nil {
								fmt.Println("ERROR:", err)
								break
							}
							if Config[key].Validator(edit) {
								u := Config[key].Value.([]string)
								for i, x := range u {
									if x == name {
										if i < len(u)-1 {
											u = append(u[:i], u[i+1:]...)
											Config[key].Value = u
										} else {
											Config[key].Value = u[:i]
										}
									}
								}
								again = true
							}
						}
						if confirm == "cancel" {
							again = false
						}
					}
				}
			case *time.Duration:
				td := *t
				tds := fmt.Sprint(td)
				prompt := &survey.Input{
					Message: key + "> ",
					Default: tds,
				}
				err := survey.AskOne(prompt, &tds, nil)
				if err != nil {
					fmt.Println("ERROR:", err)
				}
				if Config[key].Validator(tds) {
					prompt := &survey.Select{
						Message: key + " set to " + tds,
						Options: []string{"ok", "cancel"},
					}
					var confirm string
					err := survey.AskOne(prompt, &confirm, nil)
					if err != nil {
						fmt.Println("ERROR:", err)
					}
					if confirm == "ok" {
						Config[key].Value = &tds
					}
				}
				return 0
			default:
				fmt.Println(
					"type not handled:",
					reflect.TypeOf(Config[key].Value))
				return 1
			}
			break
		}
	}
}
