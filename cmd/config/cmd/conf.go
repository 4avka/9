package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

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
	coverbox.
		SetTextColor(TextColor())
	coverbox.Box.
		SetBorder(false).
		SetBackgroundColor(BackgroundColor())
	coverbox.SetBorderPadding(1, 1, 2, 2)
	// coverbox.SetBorder(true)

	roottable, roottablewidth := genMenu("launch", "configure", "reinitialize")
	activateTable(roottable)

	launchmenutexts := []string{"node", "wallet", "shell"}
	launchtable, launchtablewidth := genMenu(launchmenutexts...)
	prelightTable(launchtable)

	catstable, catstablewidth := genMenu(app.Cats.GetSortedKeys()...)
	prelightTable(catstable)

	menuflex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(roottable, roottablewidth, 1, true).
		AddItem(coverbox, 0, 1, false)
	menuflex.Box.SetBackgroundColor(BackgroundColor())

	var leftExitActive bool
	var confirm *tview.Flex
	roottable.SetSelectionChangedFunc(func(y, x int) {
		leftExitActive = false
		coverbox.SetText(
			"",
		)
		menuflex.
			RemoveItem(coverbox).
			RemoveItem(launchtable).
			RemoveItem(catstable).
			RemoveItem(cattable).
			RemoveItem(confirm)
		switch y {
		case 0, 3:
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
				lastTable(cattable)
				menuflex.AddItem(cattable, cattablewidth, 1, true)
			}
			menuflex.AddItem(coverbox, 0, 1, true)
		}
	})
	var resetbutton int
	var toggleResetButton = func() int {
		if resetbutton == 0 {
			resetbutton = 1
		} else {
			resetbutton = 0
		}
		return resetbutton
	}
	var factoryResetFunc = func() {
		confirm = tview.NewFlex()
		confirm.SetDirection(tview.FlexRow)
		confirm.SetBorderPadding(1, 1, 2, 2)
		resettext := tview.NewTextView()
		resettext.SetText("all custom configurations will be lost, are you sure?")
		resettext.SetBorderPadding(1, 1, 2, 2)
		resettext.SetWordWrap(true)
		resettext.SetTextAlign(tview.AlignCenter)
		resettext.Box.SetBackgroundColor(MainColor())
		resetform := tview.NewForm()
		resetform.Box.SetBackgroundColor(MainColor())
		resetform.SetButtonsAlign(tview.AlignCenter)
		resetform.SetButtonBackgroundColor(MainColor())
		resetform.SetButtonTextColor(TextColor())
		eventcap := func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyTab:
				tapp.SetFocus(resetform.GetButton(toggleResetButton()))
			case tcell.KeyRight, tcell.KeyLeft:
				tapp.SetFocus(resetform.GetButton(toggleResetButton()))
			case tcell.KeyEsc:
				resetform.Blur()
				roottable.Select(3, 0)
				tapp.SetFocus(roottable)
				menuflex.RemoveItem(confirm)
				menuflex.AddItem(coverbox, 0, 1, false)
				return &tcell.EventKey{}
			}
			return event
		}
		resetform.AddButton("cancel", func() {
			menuflex.RemoveItem(confirm)
			tapp.SetFocus(roottable)
		})
		resetform.AddButton("reset to factory settings", func() {
			for _, x := range app.Cats {
				for _, z := range x {
					z.Init(z)
				}
			}
			resettext.SetText("CONFIRMED\n\nfactory reset completed")
			confirm.RemoveItem(resetform)
			// resetform.RemoveButton(1)
			tapp.ForceDraw()
			time.Sleep(time.Second)
			menuflex.RemoveItem(confirm)
			tapp.SetFocus(roottable)
		})
		resetform.SetInputCapture(eventcap)
		resetform.GetButton(0).SetInputCapture(eventcap)
		resetform.GetButton(1).SetInputCapture(eventcap)
		confirm.AddItem(resettext, 5, 0, false)
		confirm.AddItem(resetform, 3, 0, true)
		menuflex.AddItem(confirm, 0, 1, true)
		tapp.SetFocus(confirm)
	}
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
			menuflex.AddItem(coverbox, 0, 1, true)
			tapp.SetFocus(launchtable)
		case 2:
			activatedTable(roottable)
			activateTable(catstable)
			if cattable != nil {
				menuflex.AddItem(cattable, cattablewidth, 0, false)
				prelightTable(cattable)
			}
			menuflex.AddItem(coverbox, 0, 1, true)
			tapp.SetFocus(catstable)
		case 3:
			factoryResetFunc()
		}
	})
	roottable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		menuflex.RemoveItem(coverbox)
		roottable.GetCell(0, 0).SetText("<")
		menuflex.
			RemoveItem(cattable).
			RemoveItem(coverbox)
		switch event.Key() {
		case tcell.KeyRight, tcell.KeyTab:
			leftExitActive = false
			y, _ := roottable.GetSelection()
			switch y {
			case 1:
				activatedTable(roottable)
				activateTable(launchtable)
				menuflex.AddItem(coverbox, 0, 1, true)
				tapp.SetFocus(launchtable)
			case 2:
				activatedTable(roottable)
				activateTable(catstable)
				if cattable != nil {
					menuflex.AddItem(cattable, cattablewidth, 0, false)
					prelightTable(cattable)
				}
				menuflex.AddItem(coverbox, 0, 1, true)
				tapp.SetFocus(catstable)
			case 3:
				factoryResetFunc()
			}
		case tcell.KeyLeft, tcell.KeyEsc:
			y, _ := roottable.GetSelection()
			if y == 0 {
				if !leftExitActive {
					roottable.GetCell(0, 0).SetText("< exit")
					leftExitActive = true
				} else {
					tapp.Stop()
				}
			} else {
				roottable.Select(0, 0)
			}
		}
		return event
	})

	launchtable.SetSelectionChangedFunc(func(y, x int) {
		switch y {
		case 0:
			menuflex.
				RemoveItem(coverbox).
				RemoveItem(cattable).
				RemoveItem(catstable).
				AddItem(coverbox, 0, 1, false)
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
			return
		case 1:
			tapp.Stop()
			fmt.Println("starting up", launchmenutexts[y-1])
		case 2:
			tapp.Stop()
			fmt.Println("starting up", launchmenutexts[y-1])
		case 3:
			tapp.Stop()
			fmt.Println("starting up", launchmenutexts[y-1])
		}
	})
	launchtable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyLeft, tcell.KeyEsc:
			prelightTable(launchtable)
			activateTable(roottable)
			tapp.SetFocus(roottable)
		}
		return event
	})

	saveConfig := func() {
		ddir, ok := app.Cats["app"]["datadir"].Get().(string)
		if ok {
			configFile := config.CleanAndExpandPath(filepath.Join(
				ddir, "config"), "")
			if config.EnsureDir(configFile) {
			}
			fh, err := os.Create(configFile)
			if err != nil {
				panic(err)
			}
			j, e := json.MarshalIndent(app, "", "\t")
			if e != nil {
				panic(e)
			}
			_, err = fmt.Fprint(fh, string(j))
			if err != nil {
				panic(err)
			}
		}
	}

	var genPage func(cat, item string, active bool, app *config.App,
		editoreventhandler func(event *tcell.EventKey) *tcell.EventKey, idx int) (out *tview.Flex)

	inputhandler = func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			menuflex.
				RemoveItem(coverbox).
				RemoveItem(activepage)
			activepage = genPage(cat, itemname, false, app, inputhandler, 0)
			menuflex.AddItem(activepage, 0, 1, true)
			prelightTable(roottable)
			activatedTable(catstable)
			activateTable(cattable)
			tapp.SetFocus(cattable)
		default:
		}
		return event
	}

	genPage = func(cat, item string, active bool, app *config.App,
		editoreventhandler func(event *tcell.EventKey) *tcell.EventKey, idx int) (out *tview.Flex) {
		currow := app.Cats[cat][item]
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
		def := currow.Default
		defstring := ""
		if def != nil {
			defstring = fmt.Sprintf("default value: %v", def.Get())
		} else {
			defstring = "" //"this value has no default"
		}
		infostring := fmt.Sprintf(
			"%v\n\n%s",
			currow.Usage, defstring,
		)
		if min, ok := currow.Min.Get().(int); ok {
			infostring += fmt.Sprint("\nminimum value: ", min)
		}
		if max, ok := currow.Max.Get().(int); ok {
			infostring += fmt.Sprint("\nmaximum value: ", max)
		}
		itemtype := currow.Type
		infostring =
			"<esc>     to cancel\n\n" + infostring
		switch currow.Type {
		case "int", "float", "duration":
			infostring =
				"<ctrl-z>  to reset to default\n" +
					infostring
		case "string", "port":
			infostring =
				"<ctrl-u>  to clear\n" +
					"<ctrl-z>  to reset to default\n" +
					infostring
		default:
		}
		infoblock.SetText(infostring)
		switch itemtype {
		case "string", "int", "float", "duration", "port":

			iteminput = tview.NewInputField()
			iteminput.
				SetFieldTextColor(darkness).
				SetFieldBackgroundColor(lightness).
				SetBackgroundColor(lightness).
				SetBorderPadding(1, 1, 1, 1)
			val := currow.Value
			if val != nil {
				vv := val.Get()
				outstring := ""
				if vv != nil {
					switch ov := vv.(type) {
					case int:
						outstring = fmt.Sprintf("%8d", ov)
					case float64:
						switch itemtype {
						case "float":
							os := fmt.Sprintf("%0f", ov)
							os = strings.TrimRight(os, "0")
							if strings.HasSuffix(os, ".") {
								os += "0"
							}
							outstring = os
						case "int", "port":
							outint := int(ov)
							outstring = fmt.Sprintf("%8d", outint)
						case "duration":
							outdur := time.Duration(int(ov))
							outstring = fmt.Sprintf("%v", outdur)
						}
					case time.Duration:
						outstring = fmt.Sprintf("%v", ov)
					default:
						outstring = fmt.Sprint(ov)
					}
					iteminput.SetText(strings.TrimSpace(outstring))
				}
			}
			var canceller func(rw *config.Row) func(event *tcell.EventKey) *tcell.EventKey
			canceller = func(rw *config.Row) func(event *tcell.EventKey) *tcell.EventKey {
				return func(event *tcell.EventKey) *tcell.EventKey {
					switch {
					case event.Key() == tcell.KeyCtrlU:
						switch itemtype {
						case "int":
							rw.Value.Put(rw.Default.Get())
						case "float":
							rw.Value.Put(rw.Default.Get())
						case "duration":
							rw.Value.Put(rw.Default.Get())
						default:
							rw.Value.Put(nil)
						}
					case event.Key() == tcell.KeyCtrlZ:
						rw.Value.Put(rw.Default.Get())
					default:
						return editoreventhandler(event)
					}
					menuflex.
						RemoveItem(coverbox).
						RemoveItem(activepage)
					itemname = item
					activepage = genPage(cat, itemname, false, app, canceller(rw), 0)
					menuflex.AddItem(activepage, 0, 1, true)
					prelightTable(roottable)
					activatedTable(catstable)
					activateTable(cattable)
					tapp.SetFocus(cattable)
					saveConfig()
					return event
				}
			}
			iteminput.SetInputCapture(canceller(currow))
			snackbar := tview.NewTextView()
			iteminput.SetDoneFunc(func(key tcell.Key) {
				rrr := currow
				rw := rrr
				if key == tcell.KeyEnter || key == tcell.KeyTab {
					s := iteminput.GetText()
					if s == "" {
						switch itemtype {
						case "int":
							rw.Value.Put(0)
						case "float":
							rw.Value.Put(0.0)
						case "duration":
							rw.Value.Put(0 * time.Second)
						default:
							rw.Value.Put(nil)
						}
						saveConfig()
					} else {
						isvalid := rw.Validate(rw, &s)
						if !isvalid {
							snackbar.SetBackgroundColor(tcell.ColorOrange)
							snackbar.SetTextColor(tcell.ColorRed)
							snackbar.SetText("input is not valid for this field")
							out.RemoveItem(infoblock).RemoveItem(snackbar)
							out.AddItem(snackbar, 1, 1, false)
							out.AddItem(infoblock, 0, 1, false)
							return
						} else {
							// rw.Validate(rw, s)
							// rw.Value.Put(s)
							saveConfig()
							out.RemoveItem(snackbar)
						}
					}
					menuflex.
						RemoveItem(coverbox).
						RemoveItem(activepage)
					itemname = item
					activepage = genPage(cat, itemname, false, app, inputhandler, 0)
					menuflex.AddItem(activepage, 0, 1, true)
					prelightTable(roottable)
					activatedTable(catstable)
					activateTable(cattable)
					tapp.SetFocus(cattable)
				}
			})
			out.AddItem(iteminput, 3, 0, true)
		case "bool":
			rw := currow
			toggle = tview.NewTable()
			toggle.SetBorderPadding(1, 1, 2, 2)
			toggle.SetBackgroundColor(lightness)
			def := currow.Default.Get().(bool)
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
			curr := currow
			if curr.Bool() {
				curropt = 1
			}
			toggle.
				SetSelectable(true, true).
				Select(curropt, 0)
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
					activepage = genPage(cat, itemname, false, app, inputhandler, y)
					menuflex.AddItem(activepage, 0, 1, true)
					prelightTable(roottable)
					activatedTable(catstable)
					activateTable(cattable)
					tapp.SetFocus(cattable)
				case 1:
					rw.Put(true)
					itemname = item
					activepage = genPage(cat, itemname, false, app, inputhandler, y)
					menuflex.AddItem(activepage, 0, 1, true)
					prelightTable(roottable)
					activatedTable(catstable)
					activateTable(cattable)
					tapp.SetFocus(cattable)

				default:
				}
				saveConfig()
			})
			out.AddItem(toggle, 4, 0, true)
		case "options":
			rw := currow
			var toggle = tview.NewTable()
			toggle.SetBorderPadding(1, 1, 1, 1)
			def := currow.Default.Get().(string)
			curr := currow.Value.Get().(string)
			curropt := 0
			sort.Strings(currow.Opts)
			for i, x := range currow.Opts {
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
				Select(curropt, 0)
			toggle.SetBackgroundColor(lightness)
			toggle.SetInputCapture(editoreventhandler)
			toggle.SetSelectedFunc(func(y, x int) {
				menuflex.
					RemoveItem(coverbox).
					RemoveItem(activepage)
				rw.Put(currow.Opts[y])
				saveConfig()
				itemname = item
				activepage = genPage(cat, itemname, false, app, inputhandler, y)
				menuflex.AddItem(activepage, 0, 1, true)
				prelightTable(roottable)
				activatedTable(catstable)
				activateTable(cattable)
				tapp.SetFocus(cattable)
			})
			out.AddItem(toggle, len(currow.Opts)+2, 0, true)
		case "stringslice":
			var slice = tview.NewTable()
			slice.SetBorderPadding(1, 1, 1, 1)
			var def string
			defIface := currow.Default.Get()
			switch defIface.(type) {
			case string:
				def = currow.Default.Get().(string)
			case nil:
			default:
			}
			var curr string
			currIface := currow.Value.Get()
			switch currIface.(type) {
			case string:
				curr = currIface.(string)
			case nil:
			default:
			}
			curropt := 0
			slicevalue, ok := currow.Get().([]string)
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
						SetCell(i, 0, tview.NewTableCell("⌦").
							SetTextColor(darkness).SetBackgroundColor(lightness))
					slice.
						SetCell(i, 1, tview.NewTableCell(itemtext).
							SetTextColor(darkness).SetBackgroundColor(lightness))
				}
			}
			slice.
				SetCell(len(slicevalue), 1, tview.NewTableCell("add new").
					SetTextColor(darkness).SetBackgroundColor(lightness))
			slice.
				SetCell(len(slicevalue), 0, tview.NewTableCell("").
					SetTextColor(darkness).SetBackgroundColor(lightness).
					SetSelectable(false))
			slice.
				SetCell(len(slicevalue)+1, 1, tview.NewTableCell("set defaults").
					SetTextColor(darkness).SetBackgroundColor(lightness))
			slice.
				SetCell(len(slicevalue)+1, 0, tview.NewTableCell("").
					SetTextColor(darkness).SetBackgroundColor(lightness).
					SetSelectable(false))
			slice.
				SetCell(len(slicevalue)+2, 1, tview.NewTableCell("back").
					SetTextColor(darkness).SetBackgroundColor(lightness))
			slice.
				SetCell(len(slicevalue)+2, 0, tview.NewTableCell("").
					SetTextColor(darkness).SetBackgroundColor(lightness).
					SetSelectable(false))
			input := tview.NewInputField()
			snackbar := tview.NewTextView()
			inputDoneGen := func(idx int) func(key tcell.Key) {
				return func(key tcell.Key) {
					rrr := currow
					rw := rrr
					rwv, ok := rw.Value.Get().([]string)
					if !ok { // rwv = []string{}
					}
					if key == tcell.KeyEnter || key == tcell.KeyTab {
						s := input.GetText()
						if len(s) < 1 {
							// rw.Value.Put(nil)
						} else {
							if rw.Validate(rw, s) {
								// if idx >= len(rwv) {
								// 	rwv = append(rwv, s)
								// } else {
								// 	rwv[idx] = s
								// }
								// rw.Value.Put(rwv)
							} else {
								snackbar.SetBackgroundColor(tcell.ColorOrange)
								snackbar.SetTextColor(tcell.ColorRed)
								snackbar.SetText("input is not valid for this field")
								out.RemoveItem(infoblock).RemoveItem(snackbar)
								out.AddItem(snackbar, 1, 1, false)
								out.AddItem(infoblock, 0, 1, false)
								return
							}
							saveConfig()
							out.RemoveItem(snackbar)
						}

						// itemname = item
						// inputhandler = func(event *tcell.EventKey) *tcell.EventKey {
						// 	switch event.Key() {
						// 	case 13:
						// 		// pressed enter
						// 	case 27:
						// 		// pressed escape
						// 		menuflex.
						// 			RemoveItem(coverbox).
						// 			RemoveItem(activepage)
						// 		// itemname = item
						// 		activepage = genPage(cat, itemname, false, app, inputhandler, idx)
						// 		menuflex.AddItem(activepage, 0, 1, true)
						// 		prelightTable(roottable)
						// 		activatedTable(catstable)
						// 		activateTable(cattable)
						// 		tapp.SetFocus(cattable)
						// 	}
						// 	return event
						// }

						menuflex.
							RemoveItem(coverbox).
							RemoveItem(activepage)
						itemname = item
						activepage = genPage(cat, itemname, true, app, inputhandler, idx)
						menuflex.AddItem(activepage, 0, 1, true)
						lastTable(roottable)
						prelightTable(catstable)
						activatedTable(cattable)
						slice.Select(idx, 1)
						tapp.SetFocus(activepage)
					}
					if key == tcell.KeyEsc {
						menuflex.
							RemoveItem(coverbox).
							RemoveItem(activepage)
						itemname = item
						activepage = genPage(cat, itemname, true, app, inputhandler, len(rwv))
						menuflex.AddItem(activepage, 0, 1, true)
						lastTable(roottable)
						prelightTable(catstable)
						activatedTable(cattable)
						tapp.SetFocus(activepage)
						// return event //&tcell.EventKey{}
					}

				}
			}
			slice.SetSelectedFunc(func(y, x int) {
				switch {
				// create new
				case y == len(slicevalue):
					// pop up the new item editor
					out.RemoveItem(infoblock)
					input.SetBackgroundColor(lightness)
					input.SetLabel("new> ")
					input.SetLabelColor(darkness)
					input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
						if event.Key() == 27 {
							menuflex.
								RemoveItem(coverbox).
								RemoveItem(activepage)
							itemname = item
							activepage = genPage(cat, itemname, true, app, inputhandler, y)
							menuflex.AddItem(activepage, 0, 1, true)
							lastTable(roottable)
							prelightTable(catstable)
							activatedTable(cattable)
							tapp.SetFocus(activepage)
							// return event // &tcell.EventKey{}
						}
						return event
					})
					input.SetDoneFunc(inputDoneGen(y))
					out.AddItem(input, 1, 0, true).
						AddItem(infoblock, 0, 1, false)
					tapp.SetFocus(input)

				// set default
				case y == len(slicevalue)+1:
					currow.Init(currow)

					menuflex.
						RemoveItem(coverbox).
						RemoveItem(activepage)
					itemname = item
					activepage = genPage(cat, itemname, false, app, inputhandler, y)
					menuflex.AddItem(activepage, 0, 1, true)
					prelightTable(roottable)
					activatedTable(catstable)
					activateTable(cattable)
					tapp.SetFocus(cattable)

				// back
				case y == len(slicevalue)+2:
					menuflex.
						RemoveItem(coverbox).
						RemoveItem(activepage)
					itemname = item
					activepage = genPage(cat, itemname, false, app, inputhandler, y)
					menuflex.AddItem(activepage, 0, 1, true)
					prelightTable(roottable)
					activatedTable(catstable)
					activateTable(cattable)
					tapp.SetFocus(cattable)

					//existing
				default:
					rw := currow
					rwv, ok := rw.Value.Get().([]string)
					// column 0 is delete column 1 is edit
					// TODO: consolidate editor code from above with this
					if x == 0 {
						if ok {
							// deleted := rwv[y]
							rwv = append(rwv[:y], rwv[y+1:]...)
							rw.Value.Put(rwv)
							saveConfig()
							menuflex.
								RemoveItem(coverbox).
								RemoveItem(activepage)
							itemname = item
							activepage = genPage(cat, itemname, true, app, inputhandler, y)
							menuflex.AddItem(activepage, 0, 1, true)
							lastTable(roottable)
							prelightTable(catstable)
							activatedTable(cattable)
							tapp.SetFocus(activepage)
						} else {
							// rw.Value.Put([]string{})
						}
					} else {
						// pop up the item editor
						out.RemoveItem(infoblock)
						input.SetBackgroundColor(lightness)
						input.SetLabel("edit> ")
						input.SetLabelColor(darkness)
						if len(rwv) >= y {
							input.SetText(rwv[y])
						}
						input.SetDoneFunc(inputDoneGen(y))
						out.AddItem(input, 1, 0, true).
							AddItem(infoblock, 0, 1, false)
						tapp.SetFocus(input)

					}
				}
			})
			slice.
				SetSelectable(true, true).
				Select(curropt, 1)
			slice.SetBackgroundColor(lightness)
			slice.SetInputCapture(editoreventhandler)
			slice.Select(len(slicevalue), 1)
			out.AddItem(slice, len(slicevalue)+5, 0, true)
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
		ckeys := app.Cats[cat].GetSortedKeys()
		var catkeys []string
		for _, x := range ckeys {
			if !(cat == "app" && x == "datadir") {
				catkeys = append(catkeys, x)
			}
		}
		cattable, cattablewidth = genMenu(catkeys...)
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
				var catkeys []string
				for _, x := range app.Cats[cat].GetSortedKeys() {
					if !(cat == "app" && x == "datadir") {
						catkeys = append(catkeys, x)
					}
				}
				itemname = catkeys[y-1]
				activepage = genPage(cat, itemname, true, app, inputhandler, 0)
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
				var catkeys []string
				for _, x := range app.Cats[cat].GetSortedKeys() {
					if !(cat == "app" && x == "datadir") {
						catkeys = append(catkeys, x)
					}
				}
				itemname = catkeys[y-1]
				activepage = genPage(cat, itemname, false, app, nil, y)
				menuflex.AddItem(activepage, 0, 1, true)
			}
		})
		cattable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyRight, tcell.KeyTab:
				menuflex.
					RemoveItem(activepage).
					RemoveItem(coverbox)
				y, _ := cattable.GetSelection()
				if y == 0 {
					break
				}
				lastTable(roottable)
				prelightTable(catstable)
				activatedTable(cattable)
				var catkeys []string
				for _, x := range app.Cats[cat].GetSortedKeys() {
					if !(cat == "app" && x == "datadir") {
						catkeys = append(catkeys, x)
					}
				}
				itemname = catkeys[y-1]
				activepage = genPage(cat, itemname, true, app, inputhandler, 0)
				menuflex.AddItem(activepage, 0, 1, true)

				tapp.SetFocus(activepage)
			case tcell.KeyEsc, tcell.KeyLeft:
				// pressed escape
				menuflex.
					RemoveItem(activepage).
					RemoveItem(coverbox)
				activatedTable(roottable)
				prelightTable(cattable)
				activateTable(catstable)
				menuflex.AddItem(coverbox, 0, 1, true)
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
				activepage = genPage(cat, itemname, false, app, nil, y)
				menuflex.RemoveItem(coverbox)
				menuflex.AddItem(activepage, 0, 1, true)
			}
			tapp.SetFocus(cattable)
		}
	})
	catstable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyLeft, tcell.KeyEsc:
			menuflex.
				RemoveItem(coverbox).
				RemoveItem(cattable).
				RemoveItem(activepage)
			itemname = ""
			coverbox.SetText("")
			menuflex.
				AddItem(coverbox, 0, 1, true)
			lastTable(cattable)
			prelightTable(catstable)
			activateTable(roottable)
			tapp.SetFocus(roottable)
		case tcell.KeyRight, tcell.KeyTab:
			y, _ := catstable.GetSelection()
			if y == 0 {
				break
			}
			prelightTable(roottable)
			activatedTable(catstable)
			activateTable(cattable)
			if !(cat == "" || itemname == "") {
				activepage = genPage(cat, itemname, false, app, nil, y)
				menuflex.RemoveItem(coverbox)
				menuflex.AddItem(activepage, 0, 1, true)
			}
			tapp.SetFocus(
				cattable)
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
