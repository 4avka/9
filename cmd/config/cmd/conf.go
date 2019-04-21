package main

import (
	"strings"

	"git.parallelcoin.io/dev/9/cmd/config"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const menutitle = "ⓟ 9 parallelcoin configuration CLI"

var catstablewidth int

var runtablewidth = 13

func Run(args []string, tokens config.Tokens, app *config.App) int {
	// tapp pulls everything together to create the configuration interface
	tapp := tview.NewApplication()

	// titlebar tells the user what app they are using
	titlebar := func() (out *tview.TextView) {
		out = tview.NewTextView().
			SetTextColor(tcell.ColorWhite).
			SetText(menutitle)
		out.Box.SetBackgroundColor(tcell.ColorDarkGreen)
		return
	}
	// root is the canvas (the whole current terminal view)
	root := func() (out *tview.Flex) {
		out = tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(titlebar(), 1, 1, false)
		return
	}()

	rootgridwidth := 17
	rootgrid := func() (out *tview.Grid) {
		out = tview.NewGrid().
			SetColumns(rootgridwidth, -1)
		return
	}()
	// catsgrid := func() (out *tview.Grid) {
	// 	out = tview.NewGrid().SetColumns(catstablewidth, -1)
	// 	return
	// }()

	blackbox := tview.NewBox()
	roottable := func() (out *tview.Table) {
		out = tview.NewTable().
			SetCellSimple(0, 0, "<               ").
			SetCellSimple(1, 0, "  run a server  >").
			SetCellSimple(2, 0, "  configuration >").
			SetSelectedStyle(tcell.ColorBlack, tcell.ColorWhite, tcell.AttrNone).
			SetSelectable(true, true)
		out.SetBorders(false).
			SetBackgroundColor(tcell.ColorDarkGreen)
		return
	}()

	catstablewidth := 0
	catstable := func() (out *tview.Table) {
		out = tview.NewTable()
		out.SetBorders(false)
		sortedkeys := app.Cats.GetSortedKeys()
		for _, x := range sortedkeys {
			if catstablewidth < len(x) {
				catstablewidth = len(x)
			}
		}
		out.SetCellSimple(0, 0, "< ")
		for i, x := range sortedkeys {
			pad := strings.Repeat(" ", catstablewidth-len(x))
			out.SetCellSimple(i+1, 0, "  "+x+pad+" >")
		}
		catstablewidth += 4
		out.SetSelectedStyle(tcell.ColorWhite, tcell.ColorDarkGreen, tcell.AttrDim).
			SetSelectable(true, true).
			SetSelectedStyle(tcell.ColorWhite, tcell.ColorBlack, tcell.AttrDim)
		return
	}()

	runtable := func() (out *tview.Table) {
		out = tview.NewTable().
			SetBorders(false).
			SetSelectable(false, false).
			SetCellSimple(0, 0, "<            ").
			SetCellSimple(1, 0, "  🌱 node   >").
			SetCellSimple(2, 0, "  💵 wallet >").
			SetCellSimple(3, 0, "  🐚 shell  >").
			SetSelectedStyle(tcell.ColorWhite, tcell.ColorBlack, tcell.AttrDim).
			SetSelectable(true, true)
		return
	}()

	roottable.SetSelectedFunc(func(y, x int) {
		switch y {
		case 0:
			tapp.Stop()
		case 1:
			runtable.
				SetSelectedStyle(tcell.ColorBlack, tcell.ColorWhite, tcell.AttrNone).
				SetBackgroundColor(tcell.ColorDarkGreen)
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
					rtbl.SetAttributes(tcell.AttrDim)
				}
			}
			roottable.
				SetSelectedStyle(tcell.ColorBlack, tcell.ColorDarkGreen, tcell.AttrDim&tcell.AttrBold).
				SetBackgroundColor(tcell.ColorBlack)

			tapp.SetFocus(runtable)
		case 2:
			catstable.
				SetSelectedStyle(tcell.ColorBlack, tcell.ColorWhite, tcell.AttrNone).
				SetBackgroundColor(tcell.ColorDarkGreen)
			for i := catstable.GetRowCount(); i >= 0; i-- {
				catstable.GetCell(i, 0).SetTextColor(tcell.ColorWhite)
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
			roottable.
				SetSelectedStyle(tcell.ColorBlack, tcell.ColorDarkGreen, tcell.AttrDim&tcell.AttrBold).
				SetBackgroundColor(tcell.ColorBlack)

			tapp.SetFocus(catstable)
		}
	})

	roottable.SetSelectionChangedFunc(func(y, x int) {
		switch y {
		case 0:
			rootgrid.
				RemoveItem(blackbox).
				RemoveItem(catstable).
				RemoveItem(runtable).
				AddItem(blackbox, 0, 1, 1, 1, 1, 1, false).
				SetColumns(rootgridwidth, -1)

			catstable.SetSelectedStyle(tcell.ColorWhite, tcell.ColorBlack, tcell.AttrDim)

			runtable.SetSelectedStyle(tcell.ColorWhite, tcell.ColorBlack, tcell.AttrDim)
			roottable.
				SetSelectedStyle(tcell.ColorOrange, tcell.ColorDarkGreen, tcell.AttrBold).
				GetCell(0, 0).SetText("< exit")
		case 1:
			rootgrid.
				RemoveItem(blackbox).
				RemoveItem(catstable).
				AddItem(runtable, 0, 1, 1, 1, 1, 1, true).
				AddItem(blackbox, 0, 2, 1, 1, 1, 1, false).
				SetColumns(rootgridwidth, runtablewidth, -1)

			for i := runtable.GetRowCount(); i >= 0; i-- {
				runtable.GetCell(i, 0).SetTextColor(tcell.ColorDarkGreen)
			}
			runtable.SetSelectedStyle(tcell.ColorDarkGreen, tcell.ColorBlack, tcell.AttrDim)
			roottable.
				SetSelectedStyle(tcell.ColorBlack, tcell.ColorWhite, tcell.AttrNone).
				GetCell(0, 0).SetText("<")
		case 2:
			rootgrid.RemoveItem(blackbox).
				RemoveItem(runtable).
				AddItem(catstable, 0, 1, 1, 1, 1, 1, true).
				AddItem(blackbox, 0, 2, 1, 1, 1, 1, false).
				SetColumns(rootgridwidth, catstablewidth, -1)

			for i := catstable.GetRowCount(); i >= 0; i-- {
				catstable.GetCell(i, 0).SetTextColor(tcell.ColorDarkGreen)
			}
			catstable.SetSelectedStyle(tcell.ColorDarkGreen, tcell.ColorBlack, tcell.AttrDim)
			roottable.
				SetSelectedStyle(tcell.ColorBlack, tcell.ColorWhite, tcell.AttrNone).
				GetCell(0, 0).SetText("<")
		}
	})

	backtoroot := func(table *tview.Table, width int) {
		// rootgrid.RemoveItem(blackbox)
		rootgrid.SetColumns(rootgridwidth, width, -1)
		// rootgrid.AddItem(blackbox, 0, 2, 1, 1, 1, 1, false)

		table.
			SetSelectedStyle(tcell.ColorWhite, tcell.ColorBlack, tcell.AttrDim).
			SetBackgroundColor(tcell.ColorBlack)
		for i := table.GetRowCount(); i >= 0; i-- {
			table.GetCell(i, 0).SetTextColor(tcell.ColorDarkGreen)
		}
		table.SetSelectedStyle(tcell.ColorDarkGreen, tcell.ColorBlack, tcell.AttrDim)

		for i := roottable.GetRowCount(); i >= 0; i-- {
			roottable.GetCell(i, 0).SetTextColor(tcell.ColorWhite).SetAttributes(tcell.AttrNone)
		}
		roottable.
			SetSelectedStyle(tcell.ColorBlack, tcell.ColorWhite, tcell.AttrNone).
			SetBackgroundColor(tcell.ColorDarkGreen)
		roottable.GetCell(0, 0).SetText("<")

		tapp.SetFocus(roottable)
	}

	runtable.SetSelectedFunc(func(y, x int) {
		switch y {
		case 0:
			backtoroot(runtable, runtablewidth)
		}
	})

	runtable.SetSelectionChangedFunc(func(y, x int) {
	})

	catstable.SetSelectedFunc(func(y, x int) {
		switch y {
		case 0:
			backtoroot(catstable, catstablewidth)
		}
	})

	catstable.SetSelectionChangedFunc(func(y, x int) {
	})

	root.AddItem(
		rootgrid.
			AddItem(
				roottable, 0, 0, 1, 1, 1, 1, true).
			AddItem(
				blackbox, 0, 1, 1, 1, 1, 1, false).
			SetColumns(
				rootgridwidth, -1),
		0, 1, true)

	if e := tapp.SetRoot(root, true).Run(); e != nil {
		panic(e)
	}

	return 0
}
