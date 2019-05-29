package conf

import (
	"git.parallelcoin.io/dev/9/pkg/util/tview"
)

// Menu is the tview table and misc info for each menu panel
type Menu struct {
	table *tview.Table
	width int
}

// Cursor tracks the last selected item with a page display
type Cursor struct {
	cat, itemname string
	page          *tview.Flex
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
	activepage *tview.Flex
	cursor     Cursor
	coverbox   *tview.TextView
	exitActive bool
}
