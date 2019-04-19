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
	tapp = func() (out *tview.Application) {
		out = tview.NewApplication()
		return
	}()

	// titlebar tells the user what app they are using
	titlebar = func() (out *tview.TextView) {
		out = tview.NewTextView()
		out.SetTextColor(tcell.ColorWhite).
			SetText(menutitle).
			Box.SetBackgroundColor(tcell.ColorDarkGreen)
		return
	}()

	// root is the canvas (the whole current terminal view)
	root = func() (out *tview.Flex) {
		out = tview.NewFlex()
		out.SetDirection(tview.FlexRow)
		return
	}()

	root.AddItem(titlebar, 1, 1, false)

	rootmenu := tview.NewGrid()
	rootmenu.SetColumns(19, -1)
	catsmenu := tview.NewGrid()
	catsmenu.SetColumns(12, -1)

	roottable := tview.NewTable()
	catstable := tview.NewTable()
	filler := tview.NewBox()

	roottable.SetBorders(false)
	roottable.SetCellSimple(0, 0, " < back")
	sortedkeys := app.Cats.GetSortedKeys()
	padlen := 0
	for _, x := range sortedkeys {
		if padlen < len(x) {
			padlen = len(x)
		}
	}
	for i, x := range sortedkeys {
		pl := padlen - len(x)
		pad := strings.Repeat(" ", pl)
		roottable.SetCellSimple(i+1, 0, "   "+x+pad+" > ")
	}

	roottable.SetSelectable(true, true)
	roottable.GetCell(0, 0)
	roottable.SetSelectedFunc(func(y, x int) {
		switch y {
		case 0:
			rootmenu.RemoveItem(catsmenu)
			rootmenu.AddItem(filler, 0, 1, 1, 1, 1, 1, false)
			tapp.SetFocus(catstable)
			catstable.SetBackgroundColor(tcell.ColorDarkGreen)
			roottable.SetSelectedStyle(tcell.ColorWhite, tcell.ColorDarkGreen, tcell.AttrDim)
			catstable.SetSelectedStyle(tcell.ColorBlack, tcell.ColorWhite, tcell.AttrBold)
		}
	})
	catsmenu.AddItem(roottable, 0, 0, 1, 1, 1, 1, false)
	catstable.SetBackgroundColor(tcell.ColorDarkGreen)
	catstable.SetBorders(false)
	catstable.SetCellSimple(0, 0, " < exit           ")
	catstable.SetCellSimple(1, 0, "   run a server  >")
	catstable.SetCellSimple(2, 0, "   configuration >")
	catstable.SetSelectedFunc(func(y, x int) {
		switch y {
		case 0:
			tapp.Stop()
		case 2:
			catstable.SetBackgroundColor(tcell.ColorBlack)
			catstable.SetSelectedStyle(tcell.ColorWhite, tcell.ColorDarkGreen, tcell.AttrDim)
			roottable.SetSelectedStyle(tcell.ColorBlack, tcell.ColorWhite, tcell.AttrBold)
			rootmenu.RemoveItem(filler)
			rootmenu.AddItem(catsmenu, 0, 1, 1, 1, 1, 1, false)
			roottable.SetBackgroundColor(tcell.ColorDarkGreen)
			tapp.SetFocus(roottable)
		}
	})
	catstable.SetSelectable(true, true)
	catstable.GetCell(0, 0)

	rootmenu.AddItem(catstable, 0, 0, 1, 1, 1, 1, true)
	rootmenu.AddItem(filler, 0, 1, 1, 1, 1, 1, false)

	root.AddItem(rootmenu, 0, 1, true)

	if e := tapp.SetRoot(root, true).Run(); e != nil {
		panic(e)
	}

	return 0
}
