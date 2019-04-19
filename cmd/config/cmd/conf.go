package main

import (
	"strings"

	"git.parallelcoin.io/dev/9/cmd/config"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

var runConf = func(args []string, tokens config.Tokens, app *config.App) int {
	const menutitle = "ⓟ 9 parallelcoin configuration CLI"

	// pre-declare everything so we can decide the order to put things
	var titlebar *tview.TextView
	var root *tview.Flex
	var tapp *tview.Application

	// tapp pulls everything together to create the configuration interface
	tapp = tview.NewApplication()

	// titlebar tells the user what app they are using
	titlebar = tview.NewTextView()
	titlebar.SetTextColor(tcell.ColorWhite).
		SetText(menutitle).
		Box.SetBackgroundColor(tcell.ColorDarkGreen)

	// root is the canvas (the whole current terminal view)
	root = tview.NewFlex()
	root.SetDirection(tview.FlexRow)

	root.AddItem(titlebar, 1, 1, false)

	rootmenu := tview.NewGrid()
	rootmenu.SetColumns(19, -1)
	catsmenu := tview.NewGrid()
	catsmenu.SetColumns(11, -1)

	roottable := tview.NewTable()
	filler := tview.NewBox()

	// catsmenu.AddItem(catstable, 0, 0, 1, 1, 1, 1, false)
	roottable.SetBackgroundColor(tcell.ColorDarkGreen)
	roottable.SetBorders(false)
	roottable.SetCellSimple(0, 0, "<                ")
	roottable.SetCellSimple(1, 0, "  run a server  >")
	roottable.SetCellSimple(2, 0, "  configuration >")

	roottable.SetSelectable(true, true)
	roottable.GetCell(0, 0)
	catstable := GetCatsTable(app)

	runmenu := GetRunMenu()
	// runmenu.SetColumns(11, -1)

	roottable.SetSelectedFunc(func(y, x int) {
		switch y {
		case 0:
			tapp.Stop()
		case 1:
			roottable.SetSelectable(false, false)
			runmenu.SetSelectable(true, true)
			runmenu.SetBackgroundColor(tcell.ColorDarkGreen)
			roottable.SetBackgroundColor(tcell.ColorBlack)
			tapp.SetFocus(runmenu)
		case 2:
			roottable.SetSelectable(false, false)
			catstable.SetSelectable(true, true)
			roottable.SetBackgroundColor(tcell.ColorBlack)
			catstable.SetBackgroundColor(tcell.ColorDarkGreen)
			tapp.SetFocus(catstable)
		}
	})

	runmenu.SetSelectedFunc(func(y, x int) {
		switch y {
		case 0:
			rootmenu.RemoveItem(filler)
			// rootmenu.RemoveItem(catstable)
			// rootmenu.RemoveItem(runmenu)
			rootmenu.AddItem(filler, 0, 1, 1, 1, 1, 1, false)
			rootmenu.SetColumns(17, 13, -1)
			roottable.SetSelectable(true, true)
			runmenu.SetBackgroundColor(tcell.ColorBlack)
			runmenu.SetSelectable(false, false)
			roottable.SetBackgroundColor(tcell.ColorDarkGreen)
			rootmenu.AddItem(filler, 0, 2, 1, 1, 1, 1, false)
			tapp.SetFocus(roottable)
		}
	})

	catstable.SetSelectedFunc(func(y, x int) {
		switch y {
		case 0:
			rootmenu.RemoveItem(filler)
			// rootmenu.RemoveItem(catstable)
			// rootmenu.RemoveItem(runmenu)
			rootmenu.AddItem(filler, 0, 1, 1, 1, 1, 1, false)
			rootmenu.SetColumns(17, 10, -1)
			roottable.SetSelectable(true, true)
			catstable.SetBackgroundColor(tcell.ColorBlack)
			catstable.SetSelectable(false, false)
			roottable.SetBackgroundColor(tcell.ColorDarkGreen)
			rootmenu.AddItem(filler, 0, 2, 1, 1, 1, 1, false)
			tapp.SetFocus(roottable)
		}
	})

	roottable.SetSelectionChangedFunc(func(y, x int) {
		// titlebar.SetText(fmt.Sprint("y", y, " x", x))
		switch y {
		case 1:
			rootmenu.RemoveItem(filler)
			rootmenu.RemoveItem(catstable)
			rootmenu.AddItem(runmenu, 0, 1, 1, 1, 1, 1, true)
			rootmenu.AddItem(filler, 0, 2, 1, 1, 1, 1, false)
			rootmenu.SetColumns(17, 13, -1)
		case 2:
			rootmenu.RemoveItem(filler)
			rootmenu.RemoveItem(runmenu)
			rootmenu.AddItem(catstable, 0, 1, 1, 1, 1, 1, true)
			rootmenu.AddItem(filler, 0, 2, 1, 1, 1, 1, false)
			rootmenu.SetColumns(17, 10, -1)
		default:
			rootmenu.RemoveItem(filler)
			rootmenu.RemoveItem(catstable)
			rootmenu.RemoveItem(runmenu)
			rootmenu.AddItem(filler, 0, 1, 1, 1, 1, 1, false)
			rootmenu.SetColumns(17, -1)
			// tapp.ForceDraw()
		}
	})

	rootmenu.AddItem(roottable, 0, 0, 1, 1, 1, 1, true)
	// rootmenu.AddItem(runmenu, 0, 1, 1, 1, 1, 1, true)
	rootmenu.AddItem(filler, 0, 1, 1, 1, 1, 1, false)
	rootmenu.SetColumns(17, -1)

	root.AddItem(rootmenu, 0, 1, true)

	if e := tapp.SetRoot(root, true).Run(); e != nil {
		panic(e)
	}

	return 0
}

func GetCatsTable(app *config.App) *tview.Table {
	catstable := tview.NewTable()
	catstable.SetBorders(false)
	sortedkeys := app.Cats.GetSortedKeys()
	padlen := 0
	for _, x := range sortedkeys {
		if padlen < len(x) {
			padlen = len(x)
		}
	}
	catstable.SetCellSimple(0, 0, "< ")
	for i, x := range sortedkeys {
		pad := strings.Repeat(" ", padlen-len(x))
		catstable.SetCellSimple(i+1, 0, "  "+x+pad+" >")
	}
	catstable.SetSelectable(false, false)
	return catstable
}

func GetRunMenu() (out *tview.Table) {
	out = tview.NewTable()
	out.SetBorders(false)
	out.SetSelectable(false, false)
	out.SetCellSimple(0, 0, "<            ")
	out.SetCellSimple(1, 0, "  🌱 node   >")
	out.SetCellSimple(2, 0, "  💵 wallet >")
	out.SetCellSimple(3, 0, "  🐚 shell  >")
	return
}
