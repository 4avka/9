package main

import (
	"strings"

	"git.parallelcoin.io/dev/9/cmd/config"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// pre-declare everything so we can decide the order to put things
var tapp *tview.Application
var root *tview.Flex
var titlebar *tview.TextView
var rootgrid, catsgrid *tview.Grid

const menutitle = "ⓟ 9 parallelcoin configuration CLI"

var runConf = func(args []string, tokens config.Tokens, app *config.App) int {
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

	rootgrid = tview.NewGrid()
	rootgridwidth := 17
	rootgrid.SetColumns(rootgridwidth, -1)
	catsgrid = tview.NewGrid()
	catsgrid.SetColumns(catstablewidth, -1)

	roottable := tview.NewTable()
	blackbox := tview.NewBox()
	// .SetBorder(true)

	roottable.SetBackgroundColor(tcell.ColorDarkGreen)
	roottable.SetBorders(false)
	roottable.SetCellSimple(0, 0, "<               ")
	roottable.SetCellSimple(1, 0, "  run a server  >")
	roottable.SetCellSimple(2, 0, "  configuration >")
	roottable.SetSelectedStyle(tcell.ColorBlack, tcell.ColorWhite, tcell.AttrNone)
	roottable.SetSelectable(true, true)
	roottable.GetCell(0, 0)

	catstable := GetCatsTable(app)
	catstable.SetSelectable(true, true)
	catstable.SetSelectedStyle(tcell.ColorWhite, tcell.ColorBlack, tcell.AttrDim)

	runtable := GetRuntable()
	runtable.SetSelectedStyle(tcell.ColorWhite, tcell.ColorBlack, tcell.AttrDim)
	runtable.SetSelectable(true, true)

	roottable.SetSelectedFunc(func(y, x int) {
		switch y {
		case 0:
			tapp.Stop()
		case 1:
			runtable.SetBackgroundColor(tcell.ColorDarkGreen)
			runtable.SetSelectedStyle(tcell.ColorBlack, tcell.ColorWhite, tcell.AttrNone)
			for i := runtable.GetRowCount(); i >= 0; i-- {
				runtable.GetCell(i, 0).SetTextColor(tcell.ColorWhite)
			}

			for i := roottable.GetRowCount(); i >= 0; i-- {
				roottable.GetCell(i, 0).SetTextColor(tcell.ColorWhite)
			}
			for i := roottable.GetRowCount(); i >= 0; i-- {
				rtbl := roottable.GetCell(i, 0)
				y, _ := roottable.GetSelection()
				if i != y {
					// SetTextColor(tcell.ColorBlack).
					rtbl.SetAttributes(tcell.AttrDim)
				}
			}
			roottable.SetBackgroundColor(tcell.ColorBlack)
			roottable.SetSelectedStyle(tcell.ColorBlack, tcell.ColorDarkGreen, tcell.AttrDim&tcell.AttrBold)

			tapp.SetFocus(runtable)
		case 2:
			catstable.SetBackgroundColor(tcell.ColorDarkGreen)
			catstable.SetSelectedStyle(tcell.ColorBlack, tcell.ColorWhite, tcell.AttrNone)
			for i := catstable.GetRowCount(); i >= 0; i-- {
				catstable.GetCell(i, 0).SetTextColor(tcell.ColorWhite)
			}
			catstable.SetSelectedStyle(tcell.ColorBlack, tcell.ColorWhite, tcell.AttrDim)

			for i := roottable.GetRowCount(); i >= 0; i-- {
				roottable.GetCell(i, 0).SetTextColor(tcell.ColorWhite)
			}
			for i := roottable.GetRowCount(); i >= 0; i-- {
				rtbl := roottable.GetCell(i, 0)
				y, _ := roottable.GetSelection()
				if i != y {
					// SetTextColor(tcell.ColorBlack).
					rtbl.SetAttributes(tcell.AttrDim)
				}
			}
			roottable.SetSelectedStyle(tcell.ColorBlack, tcell.ColorDarkGreen, tcell.AttrDim&tcell.AttrBold)
			roottable.SetBackgroundColor(tcell.ColorBlack)

			tapp.SetFocus(catstable)
		}
	})

	roottable.SetSelectionChangedFunc(func(y, x int) {
		switch y {
		case 0:
			rootgrid.RemoveItem(blackbox)
			rootgrid.RemoveItem(catstable)
			rootgrid.RemoveItem(runtable)
			rootgrid.AddItem(blackbox, 0, 1, 1, 1, 1, 1, false)
			rootgrid.SetColumns(rootgridwidth, -1)

			catstable.SetSelectedStyle(tcell.ColorWhite, tcell.ColorBlack, tcell.AttrDim)

			runtable.SetSelectedStyle(tcell.ColorWhite, tcell.ColorBlack, tcell.AttrDim)
			roottable.SetSelectedStyle(tcell.ColorOrange, tcell.ColorBlack, tcell.AttrBlink)
			roottable.GetCell(0, 0).SetText("< exit")
			// tapp.ForceDraw()
		case 1:
			rootgrid.RemoveItem(blackbox)
			rootgrid.RemoveItem(catstable)
			rootgrid.AddItem(runtable, 0, 1, 1, 1, 1, 1, true)
			rootgrid.AddItem(blackbox, 0, 2, 1, 1, 1, 1, false)
			rootgrid.SetColumns(rootgridwidth, runtablewidth, -1)

			roottable.SetSelectedStyle(tcell.ColorBlack, tcell.ColorWhite, tcell.AttrNone)
			for i := runtable.GetRowCount(); i >= 0; i-- {
				runtable.GetCell(i, 0).SetTextColor(tcell.ColorDarkGreen)
			}
			runtable.SetSelectedStyle(tcell.ColorDarkGreen, tcell.ColorBlack, tcell.AttrDim)
			roottable.GetCell(0, 0).SetText("<")
		case 2:
			rootgrid.RemoveItem(blackbox)
			rootgrid.RemoveItem(runtable)
			rootgrid.AddItem(catstable, 0, 1, 1, 1, 1, 1, true)
			rootgrid.AddItem(blackbox, 0, 2, 1, 1, 1, 1, false)
			rootgrid.SetColumns(rootgridwidth, catstablewidth, -1)

			roottable.SetSelectedStyle(tcell.ColorBlack, tcell.ColorWhite, tcell.AttrNone)
			for i := catstable.GetRowCount(); i >= 0; i-- {
				catstable.GetCell(i, 0).SetTextColor(tcell.ColorDarkGreen)
			}
			catstable.SetSelectedStyle(tcell.ColorDarkGreen, tcell.ColorBlack, tcell.AttrDim)
			roottable.GetCell(0, 0).SetText("<")
		}
	})

	runtable.SetSelectedFunc(func(y, x int) {
		switch y {
		case 0:
			// rootgrid.RemoveItem(blackbox)
			rootgrid.SetColumns(rootgridwidth, runtablewidth, -1)
			// rootgrid.AddItem(blackbox, 0, 2, 1, 1, 1, 1, false)

			runtable.SetBackgroundColor(tcell.ColorBlack)
			runtable.SetSelectedStyle(tcell.ColorWhite, tcell.ColorBlack, tcell.AttrDim)
			for i := runtable.GetRowCount(); i >= 0; i-- {
				runtable.GetCell(i, 0).SetTextColor(tcell.ColorDarkGreen)
			}
			runtable.SetSelectedStyle(tcell.ColorDarkGreen, tcell.ColorBlack, tcell.AttrDim)

			for i := roottable.GetRowCount(); i >= 0; i-- {
				roottable.GetCell(i, 0).SetTextColor(tcell.ColorWhite).SetAttributes(tcell.AttrNone)
			}
			roottable.GetCell(0, 0).SetText("<")
			roottable.SetBackgroundColor(tcell.ColorDarkGreen)
			roottable.SetSelectedStyle(tcell.ColorBlack, tcell.ColorWhite, tcell.AttrNone)

			tapp.SetFocus(roottable)
		}
	})

	runtable.SetSelectionChangedFunc(func(y, x int) {
	})

	catstable.SetSelectedFunc(func(y, x int) {
		switch y {
		case 0:
			// rootgrid.RemoveItem(blackbox)
			rootgrid.SetColumns(rootgridwidth, catstablewidth, -1)
			// rootgrid.AddItem(blackbox, 0, 2, 1, 1, 1, 1, false)

			catstable.SetBackgroundColor(tcell.ColorBlack)
			catstable.SetSelectedStyle(tcell.ColorWhite, tcell.ColorBlack, tcell.AttrDim)
			for i := catstable.GetRowCount(); i >= 0; i-- {
				catstable.GetCell(i, 0).SetTextColor(tcell.ColorDarkGreen)
			}
			catstable.SetSelectedStyle(tcell.ColorDarkGreen, tcell.ColorBlack, tcell.AttrDim)

			for i := roottable.GetRowCount(); i >= 0; i-- {
				roottable.GetCell(i, 0).SetTextColor(tcell.ColorWhite).SetAttributes(tcell.AttrNone)
			}
			roottable.GetCell(0, 0).SetText("<")
			roottable.SetBackgroundColor(tcell.ColorDarkGreen)
			roottable.SetSelectedStyle(tcell.ColorBlack, tcell.ColorWhite, tcell.AttrNone)

			tapp.SetFocus(roottable)
		}
	})

	catstable.SetSelectionChangedFunc(func(y, x int) {
	})

	rootgrid.AddItem(roottable, 0, 0, 1, 1, 1, 1, true)
	rootgrid.AddItem(blackbox, 0, 1, 1, 1, 1, 1, false)
	rootgrid.SetColumns(rootgridwidth, -1)

	root.AddItem(rootgrid, 0, 1, true)

	if e := tapp.SetRoot(root, true).Run(); e != nil {
		panic(e)
	}

	return 0
}

var catstablewidth int

func GetCatsTable(app *config.App) *tview.Table {
	catstable := tview.NewTable()
	catstable.SetBorders(false)
	sortedkeys := app.Cats.GetSortedKeys()
	for _, x := range sortedkeys {
		if catstablewidth < len(x) {
			catstablewidth = len(x)
		}
	}
	catstable.SetCellSimple(0, 0, "< ")
	for i, x := range sortedkeys {
		pad := strings.Repeat(" ", catstablewidth-len(x))
		catstable.SetCellSimple(i+1, 0, "  "+x+pad+" >")
	}
	catstablewidth += 4
	catstable.SetSelectedStyle(tcell.ColorWhite, tcell.ColorDarkGreen, tcell.AttrDim)
	return catstable
}

var runtablewidth = 13

func GetRuntable() (out *tview.Table) {
	out = tview.NewTable()
	out.SetBorders(false)
	out.SetSelectable(false, false)
	out.SetCellSimple(0, 0, "<            ")
	out.SetCellSimple(1, 0, "  🌱 node   >")
	out.SetCellSimple(2, 0, "  💵 wallet >")
	out.SetCellSimple(3, 0, "  🐚 shell  >")
	return
}
