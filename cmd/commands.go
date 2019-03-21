package cmd

type Commands map[string]struct {
	Usage   string
	Args    Lines
	Handler func(args ...string) error
}
