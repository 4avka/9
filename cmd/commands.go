package cmd

type Commands map[string]func(lines *Lines) error
