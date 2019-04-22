package main

import "git.parallelcoin.io/dev/9/cmd/config"

// const menutitle = "ⓟ 9 parallelcoin configuration CLI"

// var catstablewidth int

// var runtablewidth = 10

func rRun(args []string, tokens config.Tokens, app *config.App) int {
	// tapp pulls everything together to create the configuration interface
	// tapp := tview.NewApplication()

	// // titlebar tells the user what app they are using
	// titlebar := func() (out *tview.TextView) {
	// 	out = tview.NewTextView().
	// 		SetTextColor(tcell.ColorWhite).
	// 		SetText(menutitle)
	// 	out.Box.SetBackgroundColor(tcell.ColorDarkGreen)
	// 	return
	// }()
	// // root is the canvas (the whole current terminal view)
	// root := func() (out *tview.Flex) {
	// 	out = tview.NewFlex().
	// 		SetDirection(tview.FlexRow).
	// 		AddItem(titlebar, 1, 1, false)
	// 	return
	// }()
	// roottablewidth := 17
	// rootgrid := func() (out *tview.Grid) {
	// 	out = tview.NewGrid().
	// 		SetColumns(roottablewidth, -1)
	// 	return
	// }()

	// blackbox := tview.NewBox().SetBackgroundColor(tcell.ColorDarkGrey)
	// roottable := func() (out *tview.Table) {
	// 	out = tview.NewTable().
	// 		SetCellSimple(0, 0, "<               ").
	// 		SetCellSimple(1, 0, "  run a server  >").
	// 		SetCellSimple(2, 0, "  configuration >").
	// 		SetSelectedStyle(tcell.ColorBlack, tcell.ColorWhite, tcell.AttrNone).
	// 		SetSelectable(true, true)
	// 	out.SetBorders(false).
	// 		SetBackgroundColor(tcell.ColorDarkGreen)
	// 	return
	// }()

	// catstablewidth := 0
	// catstable := func() (out *tview.Table) {
	// 	out = tview.NewTable()
	// 	out.SetBorders(false)
	// 	sortedkeys := app.Cats.GetSortedKeys()
	// 	for _, x := range sortedkeys {
	// 		if catstablewidth < len(x) {
	// 			catstablewidth = len(x)
	// 		}
	// 	}
	// 	out.SetCellSimple(0, 0, "< ")
	// 	for i, x := range sortedkeys {
	// 		pad := strings.Repeat(" ", catstablewidth-len(x))
	// 		out.SetCellSimple(i+1, 0, "  "+x+pad+" >")
	// 	}
	// 	catstablewidth += 4
	// 	out.SetSelectedStyle(tcell.ColorWhite, tcell.ColorDarkGreen, tcell.AttrDim).
	// 		SetSelectable(true, true).
	// 		SetSelectedStyle(tcell.ColorWhite, tcell.ColorBlack, tcell.AttrDim)
	// 	return
	// }()

	// runtable := func() (out *tview.Table) {
	// 	out = tview.NewTable().
	// 		SetBorders(false).
	// 		SetSelectable(false, false).
	// 		SetCellSimple(0, 0, "<         ").
	// 		SetCellSimple(1, 0, "  node   >").
	// 		SetCellSimple(2, 0, "  wallet >").
	// 		SetCellSimple(3, 0, "  shell  >").
	// 		SetSelectedStyle(tcell.ColorWhite, tcell.ColorBlack, tcell.AttrDim).
	// 		SetSelectable(true, true)
	// 	return
	// }()

	// roottable.SetSelectedFunc(func(y, x int) {
	// 	switch y {
	// 	case 0:
	// 		tapp.Stop()
	// 	case 1:
	// 		runtable.
	// 			SetSelectedStyle(tcell.ColorBlack, tcell.ColorWhite, tcell.AttrNone).
	// 			SetBackgroundColor(tcell.ColorDarkGreen)
	// 		for i := runtable.GetRowCount(); i >= 0; i-- {
	// 			runtable.GetCell(i, 0).SetTextColor(tcell.ColorWhite)
	// 		}
	// 		for i := roottable.GetRowCount(); i >= 0; i-- {
	// 			roottable.GetCell(i, 0).SetTextColor(tcell.ColorWhite)
	// 		}
	// 		for i := roottable.GetRowCount(); i >= 0; i-- {
	// 			rtbl := roottable.GetCell(i, 0)
	// 			y, _ := roottable.GetSelection()
	// 			if i != y {
	// 				rtbl.SetAttributes(tcell.AttrDim)
	// 			}
	// 		}
	// 		roottable.
	// 			SetSelectedStyle(tcell.ColorBlack, tcell.ColorDarkGreen, tcell.AttrDim&tcell.AttrBold).
	// 			SetBackgroundColor(tcell.ColorBlack)
	// 		tapp.SetFocus(runtable)
	// 	case 2:
	// 		catstable.
	// 			SetSelectedStyle(tcell.ColorBlack, tcell.ColorWhite, tcell.AttrNone).
	// 			SetBackgroundColor(tcell.ColorDarkGreen)
	// 		for i := catstable.GetRowCount(); i >= 0; i-- {
	// 			catstable.GetCell(i, 0).SetTextColor(tcell.ColorWhite)
	// 		}
	// 		for i := roottable.GetRowCount(); i >= 0; i-- {
	// 			roottable.GetCell(i, 0).SetTextColor(tcell.ColorWhite)
	// 		}
	// 		for i := roottable.GetRowCount(); i >= 0; i-- {
	// 			rtbl := roottable.GetCell(i, 0)
	// 			y, _ := roottable.GetSelection()
	// 			if i != y {
	// 				// SetTextColor(tcell.ColorBlack).
	// 				rtbl.SetAttributes(tcell.AttrDim)
	// 			}
	// 		}
	// 		roottable.
	// 			SetSelectedStyle(tcell.ColorBlack, tcell.ColorDarkGreen, tcell.AttrDim&tcell.AttrBold).
	// 			SetBackgroundColor(tcell.ColorBlack)
	// 		tapp.SetFocus(catstable)
	// 	}
	// })

	// roottable.SetSelectionChangedFunc(func(y, x int) {
	// 	switch y {
	// 	case 0:
	// 		rootgrid.
	// 			RemoveItem(blackbox).
	// 			RemoveItem(catstable).
	// 			RemoveItem(runtable).
	// 			AddItem(blackbox, 0, 1, 1, 1, 1, 1, false).
	// 			SetColumns(roottablewidth, -1)
	// 		roottable.
	// 			SetSelectedStyle(tcell.ColorOrange, tcell.ColorDarkGreen, tcell.AttrBold).
	// 			GetCell(0, 0).SetText("< exit")
	// 	case 1:
	// 		rootgrid.
	// 			RemoveItem(blackbox).
	// 			RemoveItem(runtable).
	// 			RemoveItem(catstable).
	// 			AddItem(runtable, 0, 1, 1, 1, 1, 1, true).
	// 			AddItem(blackbox, 0, 2, 1, 1, 1, 1, false).
	// 			SetColumns(roottablewidth, runtablewidth, -1)
	// 		for i := runtable.GetRowCount(); i >= 0; i-- {
	// 			runtable.GetCell(i, 0).SetTextColor(tcell.ColorDarkGreen)
	// 		}
	// 		runtable.SetSelectedStyle(tcell.ColorDarkGreen, tcell.ColorBlack, tcell.AttrDim)
	// 		roottable.
	// 			SetSelectedStyle(tcell.ColorBlack, tcell.ColorWhite, tcell.AttrNone).
	// 			GetCell(0, 0).SetText("<")
	// 	case 2:
	// 		rootgrid.
	// 			RemoveItem(blackbox).
	// 			RemoveItem(runtable).
	// 			RemoveItem(catstable).
	// 			AddItem(catstable, 0, 1, 1, 1, 1, 1, true).
	// 			AddItem(blackbox, 0, 2, 1, 1, 1, 1, false).
	// 			SetColumns(roottablewidth, catstablewidth, -1)
	// 		for i := catstable.GetRowCount(); i >= 0; i-- {
	// 			catstable.GetCell(i, 0).SetTextColor(tcell.ColorDarkGreen)
	// 		}
	// 		catstable.SetSelectedStyle(tcell.ColorDarkGreen, tcell.ColorBlack, tcell.AttrDim)
	// 		roottable.SetSelectedStyle(tcell.ColorBlack, tcell.ColorWhite, tcell.AttrNone)
	// 		roottable.GetCell(0, 0).SetText("<")
	// 	}
	// })

	// backtoroot := func(table *tview.Table, width int) {
	// 	rootgrid.SetColumns(roottablewidth, width, -1)
	// 	table.
	// 		SetSelectedStyle(tcell.ColorWhite, tcell.ColorBlack, tcell.AttrDim).
	// 		SetBackgroundColor(tcell.ColorBlack)
	// 	for i := table.GetRowCount(); i >= 0; i-- {
	// 		table.
	// 			GetCell(i, 0).
	// 			SetTextColor(tcell.ColorDarkGreen)
	// 	}
	// 	table.SetSelectedStyle(tcell.ColorDarkGreen, tcell.ColorBlack, tcell.AttrDim)
	// 	for i := roottable.GetRowCount(); i >= 0; i-- {
	// 		roottable.
	// 			GetCell(i, 0).
	// 			SetTextColor(tcell.ColorWhite).
	// 			SetAttributes(tcell.AttrNone)
	// 	}
	// 	roottable.
	// 		SetSelectedStyle(tcell.ColorBlack, tcell.ColorWhite, tcell.AttrNone).
	// 		SetBackgroundColor(tcell.ColorDarkGreen)
	// 	roottable.
	// 		GetCell(0, 0).
	// 		SetText("<")
	// 	tapp.SetFocus(roottable)
	// }

	// runtable.SetSelectedFunc(func(y, x int) {
	// 	if y == 0 {
	// 		backtoroot(runtable, runtablewidth)
	// 	}
	// })

	// runpage := tview.NewTextView()
	// runpage.SetBorderPadding(1, 1, 1, 1)

	// runtable.SetSelectionChangedFunc(func(y, x int) {
	// 	switch y {
	// 	case 0:
	// 		rootgrid.
	// 			RemoveItem(runpage).
	// 			AddItem(blackbox, 0, 2, 1, 1, 1, 1, true)
	// 	case 1:
	// 		rootgrid.
	// 			RemoveItem(blackbox).
	// 			AddItem(runpage, 0, 2, 1, 1, 1, 1, true)
	// 		runpage.SetText("run a full node server")
	// 	case 2:
	// 		rootgrid.
	// 			RemoveItem(blackbox).
	// 			AddItem(runpage, 0, 2, 1, 1, 1, 1, true)
	// 		runpage.SetText("\nrun a wallet server")
	// 	case 3:
	// 		rootgrid.
	// 			RemoveItem(blackbox).
	// 			AddItem(runpage, 0, 2, 1, 1, 1, 1, true)
	// 		runpage.SetText("\n\nrun a combined wallet and full node server")
	// 	}
	// })

	// catstable.SetSelectedFunc(func(y, x int) {
	// 	if y == 0 {
	// 		backtoroot(catstable, catstablewidth)
	// 	}
	// })

	// cattablewidth := 0
	// var cattable *tview.Table
	// gencattable := func(cat string) (out *tview.Table) {
	// 	out = tview.NewTable()
	// 	cattablewidth = 0
	// 	var sortedkeys []string
	// 	if catslice, ok := app.Cats[cat]; !ok {
	// 		sortedkeys = catslice.GetSortedKeys()
	// 		for _, x := range sortedkeys {
	// 			if cattablewidth < len(x) {
	// 				cattablewidth = len(x)
	// 			}
	// 		}
	// 		out.SetCellSimple(0, 0, "< ")
	// 		for i, x := range sortedkeys {
	// 			pad := strings.Repeat(" ", cattablewidth-len(x))
	// 			out.SetCellSimple(i+1, 0, "  "+x+pad+" >")
	// 		}
	// 		cattablewidth += 4
	// 		out.
	// 			SetSelectedStyle(tcell.ColorWhite, tcell.ColorDarkGreen, tcell.AttrDim).
	// 			SetSelectable(true, true)
	// 		return
	// 	} else {
	// 		out = tview.NewTable()
	// 		out.SetBorder(true)
	// return
	// }
	// }

	// catstable.SetSelectionChangedFunc(func(y, x int) {
	// 	if y == 0 {
	// 		rootgrid.
	// 			SetColumns(roottablewidth, catstablewidth, -1).
	// 			RemoveItem(cattable).
	// 			RemoveItem(runpage).
	// 			RemoveItem(blackbox).
	// 			AddItem(blackbox, 0, 2, 1, 1, 1, 1, true)
	// 		// titlebar.SetText("cat return")
	// 	} else {
	// 		rootgrid.
	// 			RemoveItem(runpage).
	// 			RemoveItem(cattable).
	// 			RemoveItem(blackbox).
	// 			SetColumns(roottablewidth, catstablewidth, cattablewidth, -1).
	// 			AddItem(gencattable(
	// 				app.Cats.GetSortedKeys()[y-1]), 0, 2, 1, 1, 1, 1, true).
	// 			AddItem(blackbox, 0, 3, 1, 1, 1, 1, true)
	// 		// titlebar.SetText(app.Cats.GetSortedKeys()[y-1])
	// 	}
	// })

	// root.AddItem(
	// 	rootgrid.
	// 		AddItem(roottable, 0, 0, 1, 1, 1, 1, true).
	// 		AddItem(blackbox, 0, 1, 1, 1, 1, 1, false).
	// 		SetColumns(roottablewidth, -1),
	// 	0, 1, true)

	// if e := tapp.SetRoot(root, true).Run(); e != nil {
	// 	panic(e)
	// }

	return 0
}
