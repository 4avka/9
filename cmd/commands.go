package cmd

type Commands map[string]Command

type Command struct {
	Usage   string
	Handler func(args ...string) error
}
