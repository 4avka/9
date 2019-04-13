package main

import (
	"fmt"
	"time"

	"git.parallelcoin.io/dev/9/cmd/config"
	"github.com/davecgh/go-spew/spew"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const BACK = "back"

func RunConf(args []string, tokens config.Tokens, app *config.App) int {
	const menutitle = "[ ⓟarallelcoin configuration CLI ]"
	// fmt.Println("ⓟarallelcoin configuration CLI")

	// newPrimitive := func(text string) tview.Primitive {
	// 	return tview.NewTextView().
	// 		SetTextAlign(tview.AlignCenter).
	// 		SetText(text)
	// }

	tapp := tview.NewApplication()
	treeview := tview.NewTreeView()
	treeview.SetBorder(true).SetBorderColor(tcell.ColorBlack)
	treeview.SetGraphics(false).SetTitleColor(tcell.ColorGreen)
	treeview.SetTitle(menutitle).SetTitleAlign(tview.AlignLeft)
	treeroot := tview.NewTreeNode("")
	// treeroot.SetColor(tcell.ColorGreen)
	treeroot.SetSelectable(false)
	treeroot.SetIndent(1)
	input := tview.NewInputField()
	// input.SetChangedFunc(func() {tapp.Draw()})
	// input.SetTitle("arrow keys to select item, enter to open/close, and enter to edit an item")
	// input.
	// SetBorder(true).
	// SetBorderColor(tcell.ColorBlack).
	// 	SetTitleColor(tcell.ColorGreen).
	// 	SetTitleAlign(tview.AlignLeft).
	// 	SetBorderPadding(0, 0, 0, 0)

	root := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(treeview, 0, 1, true).
		AddItem(input, 1, 1, true)
	treeview.SetInputCapture(
		func(event *tcell.EventKey) *tcell.EventKey {
			// input.SetText(fmt.Sprint(event.Rune()))
			if event.Key() == 27 {
				tapp.Stop()
			}
			return event
		})
	input.SetInputCapture(
		func(event *tcell.EventKey) *tcell.EventKey {
			// input.SetText(fmt.Sprint(event.Rune()))
			if event.Key() == 27 {
				input.SetText("")
				input.SetLabel("")
				tapp.SetFocus(treeview)
			}
			if event.Key() == 13 {
				input.SetText("saving")
				tapp.SetFocus(treeview)
				go func() {
					time.Sleep(time.Second * 1)
					input.SetText("")
					input.SetLabel("")
					tapp.ForceDraw()
				}()
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
		})
		for _, j := range app.Cats[x].GetSortedKeys() {
			V := ""
			if yv := app.Cats[x][j].Value; yv != nil {
				if *yv != nil {
					V = fmt.Sprint(*yv)
				}
			}
			// else {
			// 	V = fmt.Sprint(":", j)
			// }

			tnj := tview.NewTreeNode("" + j + " [ " + V + " ] " + app.Cats[x][j].Usage).
				SetReference(configbase).
				SetSelectable(true).
				SetExpanded(false)
			tnj.SetSelectedFunc(func() {
				tnj.SetExpanded(!tnj.IsExpanded())
				input.SetText(V)
				input.SetLabel("> ")
				tapp.SetFocus(input)
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
