package conf

import (
	"git.parallelcoin.io/dev/9/pkg/util/tview"
	"git.parallelcoin.io/dev/tcell"
)

// Menu is the tview table and misc info for each menu panel
type Menu struct {
	table *tview.Table
	width int
}

// Property is a wrapper that contains all the things
// for editing a configuration variable
type Property struct {
	flex         *tview.Flex
	inputhandler func(event *tcell.EventKey) *tcell.EventKey
}

// Cursor tracks the last selected item with a page display
type Cursor struct {
	cat, itemname string
	page          Property
}

// Table is a wrapper that allows us to create theme toggling methods
type Table struct {
	*tview.Table
}

// Configurator bundles all the menus and thingies together
type Configurator struct {
	app        *tview.Application
	root       Menu
	cat        Menu
	cats       Menu
	activepage Property
	cursor     Cursor
	coverbox   *tview.TextView
	exitActive bool
}
