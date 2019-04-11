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
	root := tview.NewTreeNode("parallelcoin interactive CLI").
		SetSelectable(false)
	tapp := tview.NewApplication()
	tapp.SetInputCapture(
		func(event *tcell.EventKey) *tcell.EventKey {
			// root.SetText(fmt.Sprint(event.Key()))
			if event.Key() == 27 {
				tapp.Stop()
			}
			return event
		})
	treeview := tview.NewTreeView().SetRoot(root).SetCurrentNode(root)
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
	root.AddChild(runbranch.SetReference(root))

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

			tnj := tview.NewTreeNode(j + " [ " + V + " ] [gray]" + app.Cats[x][j].Usage).
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
	root.AddChild(
		configbase.SetReference(root).SetSelectable(true),
	)
	root.AddChild(
		tview.NewTreeNode("exit").
			SetReference(root).
			SetSelectable(true).
			SetSelectedFunc(func() {
				tapp.Stop()
			},
			),
	)

	if e := tapp.SetRoot(treeview, true).Run(); e != nil {
		panic(e)
	}

	spew.Dump(app.Cats)

	return 0
}
