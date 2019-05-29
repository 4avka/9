package conf

import (
	"git.parallelcoin.io/dev/9/pkg/util/tcell"
	"git.parallelcoin.io/dev/9/pkg/util/tview"
)

type colors struct {
	Main       func() tcell.Color
	Dim        func() tcell.Color
	Prelight   func() tcell.Color
	Text       func() tcell.Color
	Background func() tcell.Color
}

var col = colors{
	// Main is the main background color for menu panels
	Main: func() tcell.Color {
		return tcell.NewRGBColor(64, 64, 64)
	},

	// Dim is the colour of the most recently selected before current item
	Dim: func() tcell.Color {
		return tcell.NewRGBColor(48, 48, 48)
	},

	// Prelight is the background colour of the next item ahead that is rendered
	// when each item that opens it is moved onto with the cursor
	Prelight: func() tcell.Color {
		return tcell.NewRGBColor(32, 32, 32)
	},

	// Text is the color of normal text with MainColor as background
	Text: func() tcell.Color {
		return tcell.NewRGBColor(216, 216, 216)
	},

	// BackgroundColor is the colour of all parts not containing any widgets
	Background: func() tcell.Color {
		return tcell.NewRGBColor(16, 16, 16)
	},
}

// This sets a menu to active attributes
func activateTable(table *tview.Table) {
	if table == nil {
		return
	}
	rowcount := table.GetRowCount()
	for i := 0; i < rowcount; i++ {
		table.GetCell(i, 0).
			SetAttributes(tcell.AttrNone).
			SetTextColor(TextColor()).
			SetBackgroundColor(MainColor())
		table.SetSelectedStyle(MainColor(), TextColor(), tcell.AttrBold)
		table.Box.SetBackgroundColor(MainColor())
	}
}

// This sets a menu to activated (it has a selected item active)
func activatedTable(table *tview.Table) {
	if table == nil {
		return
	}
	rowcount := table.GetRowCount()
	for i := 0; i < rowcount; i++ {
		table.GetCell(i, 0).
			SetAttributes(tcell.AttrNone).
			SetTextColor(MainColor()).
			SetBackgroundColor(DimColor())
		table.SetSelectedStyle(DimColor(), MainColor(), tcell.AttrBold)
		table.Box.SetBackgroundColor(DimColor())
	}
}

// This sets a menu to preview (when it is active but not selected yet)
func prelightTable(table *tview.Table) {
	if table == nil {
		return
	}
	rowcount := table.GetRowCount()
	for i := 0; i < rowcount; i++ {
		table.GetCell(i, 0).
			SetAttributes(tcell.AttrNone).
			SetTextColor(DimColor()).
			SetBackgroundColor(PrelightColor())
		table.SetSelectedStyle(PrelightColor(), DimColor(), tcell.AttrBold)
		table.Box.SetBackgroundColor(PrelightColor())
	}
}

// This is just for the one case of the root table with the editor active
func lastTable(table *tview.Table) {
	if table == nil {
		return
	}
	rowcount := table.GetRowCount()
	for i := 0; i < rowcount; i++ {
		table.GetCell(i, 0).
			SetAttributes(tcell.AttrNone).
			SetTextColor(PrelightColor()).
			SetBackgroundColor(BackgroundColor())
		table.SetSelectedStyle(BackgroundColor(), PrelightColor(), tcell.AttrBold)
		table.Box.SetBackgroundColor(BackgroundColor())
	}
}
