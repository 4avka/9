package cmd

type Commands map[string]Command
type Command struct {
	Usage   string
	Args    Lines
	Handler func(args ...string) error
}
