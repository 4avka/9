package main

import (
	"fmt"
	"strings"

	"git.parallelcoin.io/dev/9/cmd/config"
	"git.parallelcoin.io/dev/9/pkg/util/cl"
	"github.com/AlecAivazis/survey"
)

const BACK = "back"

var cursor string

func RunConf(args []string, tokens config.Tokens, app *config.App) int {
	fmt.Println("ⓟarallelcoin configuration CLI")
	runner := ConfMain(app)
	switch runner {
	case "node":
		Node(args, tokens, app)
	case "wallet":
		Wallet(args, tokens, app)
	case "shell":
		Shell(args, tokens, app)
	case "", BACK:
		return 2
	case "exit":
		return 0
	default:
		return 1
	}
	return 0
}

func ConfMain(app *config.App) string {
	cont := true
	var name string
	for cont {
		options := []string{
			"run: select a server to run",
		}
		for i := range app.Cats {
			options = append(options, "configure: "+i)
		}
		options = append(options, "exit")
		prompt := &survey.Select{
			Message:  "ⓟarallelcoin Interactive CLI",
			Options:  options,
			PageSize: 9,
		}
		err := survey.AskOne(prompt, &name, nil)
		if err != nil {
			fmt.Println("ERROR:", cl.Ine(), err)
		}
		ss := strings.Split(name, ":")[0]
		fmt.Println("ss", ss)
		switch {
		case name == "exit":
			fmt.Println("exiting")
			return name
		case ss == "run":
			fmt.Println("running a server")
			return ConfRun(app)
		case ss == "configure":
			fmt.Println("configuring a subsection")
			ConfConf(app, name)
		}
	}
	return ""
}

func ConfRun(app *config.App) string {
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

func ConfConf(app *config.App, subsection string) int {
	cursor = subsection
	// 	for {
	// 		var lines []string
	// 		re := regexp.MustCompile("(" + subsection + "[.])(.*)")
	// 		for i := range app.Cats {
	// 			if re.Match([]byte(i)) {
	// 				sects := re.FindAllStringSubmatch(i, 1)
	// 				c := app.Cats[i]
	// 				value := ""
	// 				switch t := c.Value.(type) {
	// 				case *bool:
	// 					value = fmt.Sprint(*t)
	// 				case *int:
	// 					value = fmt.Sprint(*t)
	// 				case *float64:
	// 					value = fmt.Sprintf("%.10f", *t)
	// 				case *string:
	// 					value = *t
	// 				case *[]string:
	// 					ss := *t
	// 					ll := len(ss) - 1
	// 					for i, x := range ss {
	// 						value += x
	// 						if i < ll {
	// 							value += " "
	// 						}
	// 					}
	// 				default:
	// 					// if we don't recognise it we can't print it
	// 					continue
	// 				}

	// 				item := fmt.Sprintf("%s : %v (%v) = %v", sects[0][2], c.Comment, c.Initial, value)
	// 				lines = append(lines, item)
	// 			}
	// 		}
	// 		sort.Strings(lines)
	// 		lines = append(lines, BACK)
	// 		prompt := &survey.Select{
	// 			Message:  "configuration:" + subsection + " ",
	// 			Options:  lines,
	// 			Help:     "select the variable to edit",
	// 			PageSize: 9,
	// 		}
	// 		var name string
	// 		err := survey.AskOne(prompt, &name, nil)
	// 		if err != nil {
	// 			fmt.Println("ERROR:", cl.Ine(), err)
	// 		}
	// 		name = strings.Split(name, " ")[0]
	// 		if name == BACK {
	// 			break
	// 		}
	// 		// fmt.Printf("editing %s:%s\n", subsection, name)
	// 		key := subsection + "." + name
	// 		if ConfConfEdit(key) != 0 {
	// 			break
	// 		}
	// 		datadir := *Config.DataDir
	// 		log <- cl.Info{"config location", datadir}
	// 		configFile := CleanAndExpandPath(filepath.Join(datadir, "config"))
	// 		log <- cl.Info{"configFile", configFile}
	// 		if EnsureDir(configFile) {
	// 			fmt.Println("created new directory to store data", datadir)
	// 		}
	// 		fh, err := os.Create(configFile)
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 		_, err = fmt.Fprint(fh, Config)
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 	}
	return 0
}

// func ConfConfEdit(key string) int {
// 	if _, ok := (*config)[key]; !ok {
// 		return 1
// 	}
// 	name := new(string)
// 	for {
// 		switch key {
// 		case "p2p.network":
// 			return ParseNetwork(key, name)
// 		case "log.level":
// 			return ParseLogLevel(key, name)
// 		case "mining.algo":
// 			return ParseAlgo(key, name)
// 		default:
// 			cursor = strings.Split(key, ".")[0]
// 			// switch on type
// 			switch t := (*config)[key].Value.(type) {
// 			case *int:
// 				return ParseInt(t, key, name)
// 			case *bool:
// 				return ParseBool(t, key, name)
// 			case *string:
// 				return ParseString(t, key, name)
// 			case *[]string:
// 				again := true
// 				for again {
// 					// t = (*config)[key].Value.(*[]string)
// 					prompt := &survey.Select{
// 						Message: key + ">",
// 						Options: append(append([]string{"new"}, *t...), BACK),
// 					}
// 					var name string
// 					err := survey.AskOne(prompt, &name, nil)
// 					if err != nil {
// 						fmt.Println("ERROR:", cl.Ine(), err)
// 					}
// 					switch name {
// 					case BACK:
// 						again = false
// 						return 0
// 					case "new":
// 						again2 := true
// 						for again2 {
// 							var item string
// 							prompt := &survey.Input{
// 								Message: "new item on " + key + ">",
// 								Default: item,
// 							}
// 							err = survey.AskOne(prompt, &item, nil)
// 							if err != nil {
// 								fmt.Println("ERROR:", cl.Ine(), err)
// 								break
// 							}
// 							if (*config)[key].Validate((*config)[key], item) {
// 								prompt := &survey.Select{
// 									Message: "confirm '" + item + "'",
// 									Options: []string{"ok", "edit", "cancel"},
// 								}
// 								confirm := ""
// 								err = survey.AskOne(prompt, &confirm, nil)
// 								if err != nil {
// 									fmt.Println("ERROR:", cl.Ine(), err)
// 									break
// 								}
// 								if confirm == "ok" {
// 									return 0
// 								}
// 								if confirm == "edit" {
// 									again2 = true
// 								}
// 								if confirm == "cancel" {
// 									again2 = false
// 								}
// 							}
// 						}
// 					default:
// 						prompt := &survey.Select{
// 							Message: key + ">" + name,
// 							Options: []string{"delete", "edit", "cancel"},
// 						}
// 						var confirm string
// 						err = survey.AskOne(prompt, &confirm, nil)
// 						if err != nil {
// 							fmt.Println("ERROR:", cl.Ine(), err)
// 							break
// 						}
// 						if confirm == "delete" {
// 							again = true
// 							fmt.Println(cl.Ine(), (*config)[key].Value)
// 							v := (*config)[key].Value.(*[]string)
// 							for i, x := range *v {
// 								if x == name {
// 									switch {
// 									case len(*v) < 2:
// 										fmt.Println("deleting only entry")
// 										(*config)[key].Value = &[]string{}
// 									case i < len(*v)-1:
// 										fmt.Println("deleting non-terminal", len(*v))
// 										vv := append((*v)[:i], (*v)[i:]...)
// 										(*config)[key].Value = &vv
// 									default:
// 										fmt.Println("deleting terminal")
// 										vv := (*v)[:i]
// 										(*config)[key].Value = &vv
// 									}
// 									return 0
// 								}
// 							}
// 						}
// 						if confirm == "edit" {
// 							prompt := &survey.Input{
// 								Message: key + ">" + name,
// 								Default: name,
// 							}
// 							var edit string
// 							err := survey.AskOne(prompt, &edit, nil)
// 							if err != nil {
// 								fmt.Println("ERROR:", cl.Ine(), err)
// 								break
// 							}
// 							if (*config)[key].Validate((*config)[key], edit) {
// 								u := (*config)[key].Value.(*[]string)
// 								for i, x := range *u {
// 									if x == name {
// 										if i < len(*u)-1 {
// 											*u = append((*u)[:i], (*u)[i+1:]...)
// 											(*config)[key].Value = *u
// 										} else {
// 											(*config)[key].Value = (*u)[:i]
// 										}
// 									}
// 								}
// 								again = true
// 							}
// 						}
// 						if confirm == "cancel" {
// 							again = false
// 						}
// 					}
// 				}
// 			case *time.Duration:
// 				td := *t
// 				tds := fmt.Sprint(td)
// 				prompt := &survey.Input{
// 					Message: key + "> ",
// 					Default: tds,
// 				}
// 				err := survey.AskOne(prompt, &tds, nil)
// 				if err != nil {
// 					fmt.Println("ERROR:", cl.Ine(), err)
// 				}
// 				if (*config)[key].Validate((*config)[key], tds) {
// 					prompt := &survey.Select{
// 						Message: key + " set to " + tds,
// 						Options: []string{"ok", "cancel"},
// 					}
// 					var confirm string
// 					err := survey.AskOne(prompt, &confirm, nil)
// 					if err != nil {
// 						fmt.Println("ERROR:", cl.Ine(), err)
// 					}
// 					if confirm == "ok" {
// 						(*config)[key].Value = &tds
// 					}
// 				}
// 				return 0
// 			default:
// 				fmt.Println(
// 					"type not handled:",
// 					reflect.TypeOf((*config)[key].Value))
// 				return 1
// 			}
// 			break
// 		}
// 	}
// }
