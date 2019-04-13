package main

import (
	"fmt"
	"strings"
	"time"

	"git.parallelcoin.io/dev/9/cmd/config"
	"github.com/davecgh/go-spew/spew"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const BACK = "back"

func RunConf(args []string, tokens config.Tokens, app *config.App) int {
	const menutitle = "ⓟ 9 parallelcoin configuration CLI"
	// fmt.Println("ⓟarallelcoin configuration CLI")

	// newPrimitive := func(text string) tview.Primitive {
	// 	return tview.NewTextView().
	// 		SetTextAlign(tview.AlignCenter).
	// 		SetText(text)
	// }

	tapp := tview.NewApplication()
	treeview := tview.NewTreeView()
	// treeview.SetBorder(true).SetBorderColor(tcell.ColorBlack)
	treeview.SetGraphics(true).SetGraphicsColor(tcell.ColorDarkGreen)
	treeview.SetBorderPadding(0, 1, 1, 1)
	// treeview.SetTitle(menutitle).SetTitleAlign(tview.AlignLeft)
	treeroot := tview.NewTreeNode("9")
	treeroot.SetSelectable(false)
	treeroot.SetIndent(1)
	// treeroot.SetColor(tcell.ColorRed)
	input := tview.NewInputField()
	input.SetFieldBackgroundColor(tcell.ColorDarkGreen).SetFieldTextColor(tcell.ColorBlack)
	input.SetLabelColor(tcell.ColorBlack).Box.SetBackgroundColor(tcell.ColorGreen)
	// input.SetChangedFunc(func() {tapp.Draw()})
	// input.SetTitle("arrow keys to select item, enter to open/close, and enter to edit an item")
	// input.
	// SetBorder(true).
	// SetBorderColor(tcell.ColorBlack).
	// 	SetTitleColor(tcell.ColorDarkGreen).
	// 	SetTitleAlign(tview.AlignLeft).
	// 	SetBorderPadding(0, 0, 0, 0)

	titlebar := tview.NewTextView()
	titlebar.Box.SetBackgroundColor(tcell.ColorDarkGreen)
	titlebar.SetTextColor(tcell.ColorWhite)
	titlebar.SetText(menutitle)

	root := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(titlebar, 1, 1, false).
		AddItem(treeview, 0, 1, true)
	treeview.SetInputCapture(
		func(event *tcell.EventKey) *tcell.EventKey {
			// input.SetText(fmt.Sprint(event.Rune()))
			if event.Key() == 27 {
				tapp.Stop()
			}
			return event
		})
	treeview.Box.SetBackgroundColor(tcell.ColorBlack)
	input.SetInputCapture(
		func(event *tcell.EventKey) *tcell.EventKey {
			// input.SetText(fmt.Sprint(event.Rune()))
			if event.Key() == 27 {
				// input.SetText("")
				// input.SetLabel("")
				root.RemoveItem(input)
				tapp.SetFocus(treeview)
				treeview.SetBorderPadding(0, 1, 1, 1)
			}
			if event.Key() == 13 {
				treeview.SetBorderPadding(0, 0, 1, 1)
				currentlabel := input.GetLabel()
				currenttext := input.GetText()
				input.SetFieldBackgroundColor(tcell.ColorYellow)
				input.SetFieldTextColor(tcell.ColorBlack)
				input.SetText("saving key " + currentlabel + currenttext)
				input.SetLabel("")
				tapp.ForceDraw()
				time.Sleep(time.Second * 1)
				input.SetFieldBackgroundColor(tcell.ColorDarkGreen)
				input.SetLabel("")
				root.RemoveItem(input)
				treeview.SetBorderPadding(0, 1, 1, 1)
				tapp.SetFocus(treeview)
				tapp.ForceDraw()
			}
			return event
		})
	// root.SetBorderPadding(1, 1, 3, 3)
	// root.SetTitleAlign(tview.AlignLeft).SetTitle(menutitle)
	treeview.SetRoot(treeroot).SetCurrentNode(treeroot)
	// SetBorderAttributes(tcell.AttrNone)
	runbranch := tview.NewTreeNode("run a server")
	runbranch.SetReference(root).
		SetSelectable(true).
		SetSelectedFunc(func() {
			if runbranch.IsExpanded() {
				runbranch.SetExpanded(false)
			} else {
				runbranch.SetExpanded(true)
			}
		}).
		SetExpanded(false).
		AddChild(tview.NewTreeNode("node")).
		AddChild(tview.NewTreeNode("wallet")).
		AddChild(tview.NewTreeNode("shell"))
	treeroot.AddChild(runbranch.SetReference(treeroot))

	configbase := tview.NewTreeNode("configuration")
	configbase.SetSelectable(true).
		SetSelectedFunc(func() {
			configbase.SetExpanded(!configbase.IsExpanded())
		}).
		SetExpanded(false)
	for _, x := range app.Cats.GetSortedKeys() {
		tn := tview.NewTreeNode(x).
			SetReference(configbase).
			SetSelectable(true).
			SetExpanded(false)
		tn.SetSelectedFunc(func() {
			tn.SetExpanded(!tn.IsExpanded())
			if tn.IsExpanded() {
				treeview.SetCurrentNode(tn.GetChildren()[len(tn.GetChildren())-1])
				tapp.ForceDraw()
				treeview.SetCurrentNode(tn)
				tapp.ForceDraw()
			}
		})
		maxlen := 0
		maxvaluelen := 0
		for _, j := range app.Cats[x].GetSortedKeys() {
			if len(j) > maxlen {
				maxlen = len(j)
			}
		}
		for _, j := range app.Cats[x].GetSortedKeys() {
			if yv := app.Cats[x][j].Value; yv != nil {
				valtext := fmt.Sprint(*yv)
				if len(valtext) > 0 {
					if len(valtext) > maxvaluelen {
						maxvaluelen = len(valtext)
					}
				}
			}
		}
		for _, j := range app.Cats[x].GetSortedKeys() {
			V, X := "", j
			if yv := app.Cats[x][j].Value; yv != nil {
				if *yv != nil {
					V = fmt.Sprint(*yv)
				}
			}

			// else {
			// 	V = fmt.Sprint(":", j)
			// }
			padlen := maxlen - len(j)
			keytext := j + strings.Repeat(" ", padlen)
			padusagelen := maxvaluelen + 1 - len(V)
			valuetext := V + strings.Repeat(" ", padusagelen)
			if len(V) > 24 {
				valuetext = valuetext[:21] + "..."
			}
			if len(valuetext) > 24 {
				valuetext = valuetext[:24]
			}
			tnj := tview.NewTreeNode("[:] " + keytext + " [darkgreen:black] " + valuetext + " [darkgray:] " + app.Cats[x][j].Usage + "[:]").
				SetReference(configbase).
				SetSelectable(true).
				SetExpanded(false)
			tnj.SetSelectedFunc(func() {
				tnj.SetExpanded(!tnj.IsExpanded())
				if tn.IsExpanded() {
					treeview.SetBorderPadding(0, 0, 1, 1)
					root.AddItem(input, 1, 1, false)
					input.SetText(V)
					input.Box.SetBackgroundColor(tcell.ColorDarkGreen)
					input.SetLabelColor(tcell.ColorWhite)
					input.SetLabel("'" + X + "' ")
					input.SetFieldBackgroundColor(tcell.ColorBlack)
					input.SetFieldTextColor(tcell.ColorWhite)
					tapp.SetFocus(input)
				} else {
					root.RemoveItem(input)
					tapp.SetFocus(treeview)
				}
			})
			tn.AddChild(tnj)
		}
		configbase.AddChild(tn)
	}
	treeroot.AddChild(
		configbase.SetReference(root).SetSelectable(true),
	)
	treeroot.AddChild(
		tview.NewTreeNode("exit").
			SetReference(root).
			SetSelectable(true).
			SetSelectedFunc(func() {
				tapp.Stop()
			},
			),
	)

	if e := tapp.SetRoot(root, true).Run(); e != nil {
		panic(e)
	}

	spew.Dump(app.Cats)

	return 0
}
