package main

import (
	"fmt"

	"git.parallelcoin.io/dev/9/cmd/config"
	"github.com/davecgh/go-spew/spew"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const BACK = "back"

func RunConf(args []string, tokens config.Tokens, app *config.App) int {
	const menutitle = ">>> ⓟarallelcoin configuration CLI"
	// fmt.Println("ⓟarallelcoin configuration CLI")

	// newPrimitive := func(text string) tview.Primitive {
	// 	return tview.NewTextView().
	// 		SetTextAlign(tview.AlignCenter).
	// 		SetText(text)
	// }

	treeview := tview.NewTreeView().SetGraphicsColor(tcell.ColorGreen).SetAlign(false)
	// treeview.SetGraphicsColor(tcell.ColorDimGray).
	treeview.SetBorder(false).SetBorderPadding(1, 0, 3, 3)
	treeroot := tview.NewTreeNode("9").
		SetSelectable(false).SetIndent(1)
	inputter := tview.NewInputField().
		SetTitle("select an item and press enter to edit. Esc to quit").
		SetBorderPadding(1, 1, 1, 1).
		SetBorder(true).SetTitleColor(tcell.ColorGreen).SetTitleAlign(tview.AlignCenter)
	// infopane := tview.NewTextView()
	// SetBorder(true)
	root := tview.NewGrid().
		SetRows(60, 40).
		AddItem(treeview, 0, 0, 1, 1, 1, 1, true).
		// AddItem(infopane, 1, 0, 1, 1, 1, 1, true).
		AddItem(inputter, 1, 0, 1, 1, 1, 1, true)
	// root.SetBorder(true).SetBorderPadding(1, 1, 1, 1)
	tapp := tview.NewApplication()
	tapp.SetInputCapture(
		func(event *tcell.EventKey) *tcell.EventKey {
			// root.SetText(fmt.Sprint(event.Key()))
			if event.Key() == 27 {
				tapp.Stop()
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

			tnj := tview.NewTreeNode(j + " [darkgreen][[-] [lightblue]" + V + "[-] [darkgreen]][-] [gray]" + app.Cats[x][j].Usage).
				SetReference(configbase).
				SetSelectable(true).
				SetExpanded(false)
			tnj.SetSelectedFunc(func() {
				tnj.SetExpanded(!tnj.IsExpanded())
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
