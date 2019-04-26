package main

import (
	"fmt"
	"sort"
	"strings"

	"git.parallelcoin.io/dev/9/cmd/config"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const menutitle = "ⓟ parallelcoin 9 configuration CLI"

func MainColor() tcell.Color {
	return tcell.NewRGBColor(64, 64, 64)
}

func DimColor() tcell.Color {
	return tcell.NewRGBColor(48, 48, 48)
}

func PrelightColor() tcell.Color {
	return tcell.NewRGBColor(32, 32, 32)
}

func TextColor() tcell.Color {
	return tcell.NewRGBColor(216, 216, 216)
}

func BackgroundColor() tcell.Color {
	return tcell.NewRGBColor(16, 16, 16)
}

var iteminput *tview.InputField
var toggle *tview.Table

func Run(_ []string, _ config.Tokens, app *config.App) int {
	var cattable *tview.Table
	var cattablewidth int

	var activepage *tview.Flex
	var inputhandler func(event *tcell.EventKey) *tcell.EventKey
	var cat, itemname string

	// tapp pulls everything together to create the configuration interface
	tapp := tview.NewApplication()

	// titlebar tells the user what app they are using
	titlebar := tview.NewTextView().
		SetTextColor(TextColor()).
		SetText(menutitle)
	titlebar.Box.SetBackgroundColor(MainColor())

	coverbox := tview.NewTextView()
	coverbox.SetBorder(false).SetBackgroundColor(BackgroundColor())
	coverbox.SetBorderPadding(1, 1, 1, 1)

	roottable, roottablewidth := genMenu("launch", "configure")
	activateTable(roottable)

	launchtable, launchtablewidth := genMenu("node", "wallet", "shell")
	prelightTable(launchtable)

	catstable, catstablewidth := genMenu(app.Cats.GetSortedKeys()...)
	prelightTable(catstable)

	menuflex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(roottable, roottablewidth, 1, true).
		AddItem(coverbox, 0, 1, false)
	menuflex.Box.SetBackgroundColor(BackgroundColor())

	roottable.SetSelectionChangedFunc(func(y, x int) {
		coverbox.SetText("")
		menuflex.
			RemoveItem(coverbox).
			RemoveItem(launchtable).
			RemoveItem(catstable).
			RemoveItem(cattable)
		switch y {
		case 0:
			menuflex.
				AddItem(coverbox, 0, 1, true)
		case 1:
			menuflex.
				AddItem(launchtable, launchtablewidth, 1, true).
				AddItem(coverbox, 0, 1, true)
		case 2:
			menuflex.
				AddItem(catstable, catstablewidth, 1, true)
			if cattable != nil {
				menuflex.AddItem(cattable, cattablewidth, 1, true)
			}
			menuflex.AddItem(coverbox, 0, 1, true)
		}
	})
	roottable.SetSelectedFunc(func(y, x int) {
		menuflex.RemoveItem(coverbox)
		if cattable != nil {
			menuflex.RemoveItem(cattable)
		}
		switch y {
		case 0:
			tapp.Stop()
		case 1:
			activatedTable(roottable)
			activateTable(launchtable)
			coverbox.SetTextColor(TextColor())
			tapp.SetFocus(launchtable)
		case 2:
			activatedTable(roottable)
			activateTable(catstable)
			if cattable != nil {
				menuflex.AddItem(cattable, cattablewidth, 0, false)
				prelightTable(cattable)
			}
			tapp.SetFocus(catstable)
		}
	})
	roottable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case 13:
			// titlebar.SetText("enter")
			// pressed enter
		case 27:
			// titlebar.SetText("ESCAPE")
			// pressed escape
			tapp.Stop()
		}
		return event
	})

	launchtable.SetSelectionChangedFunc(func(y, x int) {
		switch y {
		case 0:
			menuflex.RemoveItem(cattable).RemoveItem(catstable)
			coverbox.SetText("")
		case 1:
			coverbox.SetText("run a full peer to peer parallelcoin node")
		case 2:
			coverbox.SetText("\nrun a wallet server (requires a full node)")
		case 3:
			coverbox.SetText("\n\nrun a combined wallet/full node")
		}
	})
	launchtable.SetSelectedFunc(func(y, x int) {
		switch y {
		case 0:
			prelightTable(launchtable)
			activateTable(roottable)
			tapp.SetFocus(roottable)
		case 1:
		case 2:
		case 3:
		}
	})
	launchtable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case 13:
			// titlebar.SetText("enter")
			// pressed enter
		case 27:
			// titlebar.SetText("ESCAPE")
			// pressed escape
			prelightTable(launchtable)
			activateTable(roottable)
			coverbox.SetTextColor(PrelightColor()) // SetAttributes(tcell.AttrDim)
			tapp.SetFocus(roottable)
		}
		return event
	})
	var genPage func(cat, item string, active bool, app *config.App,
		editoreventhandler func(event *tcell.EventKey) *tcell.EventKey) (out *tview.Flex)
	genPage = func(cat, item string, active bool, app *config.App,
		editoreventhandler func(event *tcell.EventKey) *tcell.EventKey) (out *tview.Flex) {
		var darkness, lightness tcell.Color
		if active {
			darkness = MainColor()
			lightness = TextColor()
		} else {
			darkness = PrelightColor()
			lightness = MainColor()
		}

		out = tview.NewFlex().SetDirection(tview.FlexRow)
		heading := tview.NewTextView().
			SetText(fmt.Sprintf("%s.%s", cat, item))
		heading.
			SetTextColor(lightness).
			SetBackgroundColor(darkness).
			SetBorderPadding(0, 0, 1, 1)
		out.
			SetBorderPadding(1, 1, 1, 1).
			SetBackgroundColor(darkness)
		out.AddItem(heading, 2, 0, false)
		infoblock := tview.NewTextView()
		infoblock.
			SetWordWrap(true).
			SetTextColor(lightness).
			SetBorderPadding(1, 0, 1, 1).
			SetBackgroundColor(darkness)
		def := app.Cats[cat][item].Default
		defstring := ""
		if def != nil {
			defstring = fmt.Sprintf("default value: %v", def.Get())
		} else {
			defstring = "" //"this value has no default"
		}
		infostring := fmt.Sprintf(
			"%v\n\n%s",
			app.Cats[cat][item].Usage, defstring,
		)
		if min, ok := app.Cats[cat][item].Min.Get().(int); ok {
			infostring += fmt.Sprint("\nminimum value: ", min)
		}
		if max, ok := app.Cats[cat][item].Max.Get().(int); ok {
			infostring += fmt.Sprint("\nmaximum value: ", max)
		}

		infoblock.SetText(infostring)
		switch app.Cats[cat][item].Type {
		case "string", "int", "float", "duration", "port":
			iteminput = tview.NewInputField()
			iteminput.
				SetFieldTextColor(darkness).
				SetFieldBackgroundColor(lightness).
				SetBackgroundColor(lightness).
				SetBorderPadding(1, 1, 1, 1)
			val := app.Cats[cat][item].Value
			if val != nil {
				vv := val.Get()
				if vv != nil {
					iteminput.SetText(fmt.Sprint(vv))
				}
			}
			iteminput.SetInputCapture(editoreventhandler)
			snackbar := tview.NewTextView()
			iteminput.SetDoneFunc(func(key tcell.Key) {
				if key == tcell.KeyEnter || key == tcell.KeyTab {
					s := iteminput.GetText()
					rw := app.Cats[cat][item]
					if len(s) < 1 {
						rw.Value.Put(nil)
					} else {
						if !rw.Validate(&rw, &s) {
							snackbar.SetBackgroundColor(tcell.ColorOrange)
							snackbar.SetTextColor(tcell.ColorRed)
							snackbar.SetText("input is not valid for this field")
							out.RemoveItem(infoblock).RemoveItem(snackbar)
							out.AddItem(snackbar, 1, 1, false)
							out.AddItem(infoblock, 0, 1, false)
							return
						} else {
							rw.Value.Put(s)
							out.RemoveItem(snackbar)
						}
					}
					menuflex.
						RemoveItem(coverbox).
						RemoveItem(activepage)
					itemname = item
					activepage = genPage(cat, itemname, false, app, inputhandler)
					menuflex.AddItem(activepage, 0, 1, true)
					prelightTable(roottable)
					activatedTable(catstable)
					activateTable(cattable)
					tapp.SetFocus(cattable)
				}
			})
			out.AddItem(iteminput, 3, 0, true)
		case "bool":
			rw := app.Cats[cat][item]
			toggle = tview.NewTable()
			toggle.SetBorderPadding(1, 1, 1, 1)
			// toggle.SetBorder(true).SetBorderColor(lightness)
			toggle.SetBackgroundColor(lightness)
			def := app.Cats[cat][item].Default.Get().(bool)
			if def {
				toggle.
					SetCell(0, 0, tview.NewTableCell("false").SetTextColor(darkness)).
					SetCell(1, 0, tview.NewTableCell("true (default)").SetTextColor(darkness))
			} else {
				toggle.
					SetCell(0, 0, tview.NewTableCell("false (default)").SetTextColor(darkness)).
					SetCell(1, 0, tview.NewTableCell("true").SetTextColor(darkness))
			}
			curropt := 0
			curr := app.Cats[cat][item]
			if curr.Bool() {
				curropt = 1
			}
			toggle.
				SetSelectable(true, true).
				Select(curropt, 0).
				SetSelectedStyle(lightness, darkness, tcell.AttrNone)
			toggle.SetBackgroundColor(lightness)
			toggle.SetInputCapture(editoreventhandler)
			toggle.SetSelectedFunc(func(y, x int) {
				menuflex.
					RemoveItem(coverbox).
					RemoveItem(activepage)
				switch y {
				case 0:
					rw.Put(false)
					itemname = item
					activepage = genPage(cat, itemname, false, app, inputhandler)
					menuflex.AddItem(activepage, 0, 1, true)
					prelightTable(roottable)
					activatedTable(catstable)
					activateTable(cattable)
					tapp.SetFocus(cattable)
				case 1:
					rw.Put(true)
					itemname = item
					activepage = genPage(cat, itemname, false, app, inputhandler)
					menuflex.AddItem(activepage, 0, 1, true)
					prelightTable(roottable)
					activatedTable(catstable)
					activateTable(cattable)
					tapp.SetFocus(cattable)

				default:
				}
			})
			out.AddItem(toggle, 4, 0, true)
		case "options":
			rw := app.Cats[cat][item]
			var toggle = tview.NewTable()
			toggle.SetBorderPadding(1, 1, 1, 1)
			def := app.Cats[cat][item].Default.Get().(string)
			curr := app.Cats[cat][item].Value.Get().(string)
			curropt := 0
			sort.Strings(app.Cats[cat][item].Opts)
			for i, x := range app.Cats[cat][item].Opts {
				itemtext := x
				if x == def {
					itemtext += " (default)"
				}
				if x == curr {
					curropt = i
				}
				toggle.
					SetCell(i, 0, tview.NewTableCell(itemtext).
						SetTextColor(darkness).SetBackgroundColor(lightness))
			}
			toggle.
				SetSelectable(true, true).
				Select(curropt, 0).
				SetSelectedStyle(lightness, darkness, tcell.AttrNone)
			toggle.SetBackgroundColor(lightness)
			toggle.SetInputCapture(editoreventhandler)
			toggle.SetSelectedFunc(func(y, x int) {
				menuflex.
					RemoveItem(coverbox).
					RemoveItem(activepage)
				rw.Put(app.Cats[cat][item].Opts[y])
				itemname = item
				activepage = genPage(cat, itemname, false, app, inputhandler)
				menuflex.AddItem(activepage, 0, 1, true)
				prelightTable(roottable)
				activatedTable(catstable)
				activateTable(cattable)
				tapp.SetFocus(cattable)
			})
			out.AddItem(toggle, len(app.Cats[cat][item].Opts)+2, 0, true)
		case "stringslice":
			// rw := app.Cats[cat][item]
			var slice = tview.NewTable()
			slice.SetBorderPadding(1, 1, 1, 1)
			var def string
			defIface := app.Cats[cat][item].Default.Get()
			switch defIface.(type) {
			case string:
				def = app.Cats[cat][item].Default.Get().(string)
			case nil:
			default:
			}
			var curr string
			currIface := app.Cats[cat][item].Value.Get()
			switch dd := currIface.(type) {
			case string:
				curr = dd
			case nil:
			default:
			}
			curropt := 0
			slicevalue, ok := app.Cats[cat][item].Get().([]string)
			if ok {
				for i, x := range slicevalue {
					itemtext := x
					if x == def {
						itemtext += " (default)"
					}
					if x == curr {
						curropt = i
					}
					slice.
						SetCell(i, 0, tview.NewTableCell(itemtext).
							SetTextColor(darkness).SetBackgroundColor(lightness))
				}
			}
			slice.
				SetCell(len(slicevalue), 0, tview.NewTableCell("add new").
					SetTextColor(darkness).SetBackgroundColor(lightness))
			slice.
				SetCell(len(slicevalue)+1, 0, tview.NewTableCell("back").
					SetTextColor(darkness).SetBackgroundColor(lightness))
			slice.SetSelectedFunc(func(y, x int) {
				if y == len(slicevalue)+1 {
					menuflex.
						RemoveItem(coverbox).
						RemoveItem(activepage)
					// rw.Put(app.Cats[cat][item].Opts[y])
					itemname = item
					activepage = genPage(cat, itemname, false, app, inputhandler)
					menuflex.AddItem(activepage, 0, 1, true)
					prelightTable(roottable)
					activatedTable(catstable)
					activateTable(cattable)
					tapp.SetFocus(cattable)
				}
			})
			slice.
				SetSelectable(true, true).
				Select(curropt, 0).
				SetSelectedStyle(lightness, darkness, tcell.AttrNone)
			slice.SetBackgroundColor(lightness)
			slice.SetInputCapture(editoreventhandler)
			out.AddItem(slice, len(slicevalue)+4, 0, true)

		}
		out.AddItem(infoblock, 0, 1, false)
		return
	}
	catstable.SetSelectionChangedFunc(func(y, x int) {
		itemname = ""
		menuflex.
			RemoveItem(activepage).
			RemoveItem(coverbox).
			RemoveItem(cattable)
		if y == 0 {
			cat = strings.TrimSpace(catstable.GetCell(y, x).Text)
			menuflex.
				AddItem(coverbox, 0, 1, true)
			return
		}
		cat = app.Cats.GetSortedKeys()[y-1]
		cattable, cattablewidth = genMenu(app.Cats[cat].GetSortedKeys()...)
		prelightTable(cattable)
		cattable.SetSelectedFunc(func(y, x int) {
			menuflex.
				RemoveItem(activepage).
				RemoveItem(coverbox)
			if y == 0 {
				activatedTable(roottable)
				prelightTable(cattable)
				activateTable(catstable)
				menuflex.
					AddItem(coverbox, 0, 1, true)
				tapp.SetFocus(catstable)
			} else {
				lastTable(roottable)
				prelightTable(catstable)
				activatedTable(cattable)
				itemname = app.Cats[cat].GetSortedKeys()[y-1]
				inputhandler = func(event *tcell.EventKey) *tcell.EventKey {
					switch event.Key() {
					case 13:
						// pressed enter
					case 27:
						// pressed escape
						menuflex.
							RemoveItem(coverbox).
							RemoveItem(activepage)
						itemname = app.Cats[cat].GetSortedKeys()[y-1]
						activepage = genPage(cat, itemname, false, app, inputhandler)
						menuflex.AddItem(activepage, 0, 1, true)
						prelightTable(roottable)
						activatedTable(catstable)
						activateTable(cattable)
						tapp.SetFocus(cattable)
					}
					return event
				}
				activepage = genPage(cat, itemname, true, app, inputhandler)
				menuflex.AddItem(activepage, 0, 1, true)
				tapp.SetFocus(activepage)
			}
		})
		cattable.SetSelectionChangedFunc(func(y, x int) {
			menuflex.
				RemoveItem(coverbox).
				RemoveItem(activepage)
			if y == 0 {
				menuflex.AddItem(coverbox, 0, 1, false)
			} else {
				itemname = app.Cats[cat].GetSortedKeys()[y-1]
				activepage = genPage(cat, itemname, false, app, nil)
				menuflex.AddItem(activepage, 0, 1, true)
			}
		})
		cattable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case 13:
				// pressed enter
			case 27:
				// pressed escape
				menuflex.
					RemoveItem(activepage).
					RemoveItem(coverbox)
				activatedTable(roottable)
				prelightTable(cattable)
				activateTable(catstable)
				menuflex.
					AddItem(coverbox, 0, 1, true)
				tapp.SetFocus(catstable)
			}
			return event
		})
		menuflex.
			AddItem(cattable, cattablewidth, 1, false).
			AddItem(coverbox, 0, 1, true)
	})
	catstable.SetSelectedFunc(func(y, x int) {
		menuflex.
			RemoveItem(coverbox).
			RemoveItem(activepage)
		if y == 0 {
			itemname = ""
			prelightTable(catstable)
			activateTable(roottable)
			coverbox.SetText("")
			menuflex.
				AddItem(coverbox, 0, 1, true)
			tapp.SetFocus(roottable)
		} else {
			// itemname = strings.TrimSpace(catstable.GetCell(y, x).Text)
			prelightTable(roottable)
			activatedTable(catstable)
			activateTable(cattable)
			if !(cat == "" || itemname == "") {
				activepage = genPage(cat, itemname, false, app, nil)
				menuflex.RemoveItem(coverbox)
				menuflex.AddItem(activepage, 0, 1, true)
			}
			tapp.SetFocus(cattable)
		}
	})
	catstable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case 13:
			// pressed enter
		case 27:
			// pressed escape
			// itemname = ""
			lastTable(cattable)
			prelightTable(catstable)
			activateTable(roottable)
			tapp.SetFocus(roottable)
		}
		return event
	})
	// root is the canvas (the whole current terminal view)
	root := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(titlebar, 1, 0, false).
		AddItem(menuflex, 0, 1, true)

	if e := tapp.SetRoot(root, true).Run(); e != nil {
		panic(e)
	}

	return 0
}

func getMaxWidth(ss []string) (maxwidth int) {
	for _, x := range ss {
		if len(x) > maxwidth {
			maxwidth = len(x)
		}
	}
	return
}

func genMenu(items ...string) (table *tview.Table, menuwidth int) {
	menuwidth = getMaxWidth(items)
	table = tview.NewTable().SetSelectable(true, true)
	table.SetCell(0, 0, tview.NewTableCell("<"))
	for i, x := range items {
		pad := strings.Repeat(" ", menuwidth-len(x))
		table.SetCell(i+1, 0, tview.NewTableCell(" "+pad+x))
	}
	t, l, _, h := table.Box.GetRect()
	menuwidth += 2
	table.Box.SetRect(t, l, menuwidth, h)
	return
}

// This sets a menu to active attributes
func activateTable(table *tview.Table) {
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
