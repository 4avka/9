package conf

import (
	"git.parallelcoin.io/dev/9/pkg/util/tcell"
)

func (table *Table) activate() {
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
func (table *Table) activated() {
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
func (table *Table) prelight() {
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
func (table *Table) lastlight() {
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
