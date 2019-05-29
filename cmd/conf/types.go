package conf

import (
	"git.parallelcoin.io/dev/9/pkg/util/tview"
)

// Menu is the tview table and misc info for each menu panel
type Menu struct {
	table *tview.Table
	width int
}

// Configurator bundles all the menus and thingies together
type Configurator struct {
	app  *tview.Application
	root Menu
	cat  Menu
	cats Menu
}
