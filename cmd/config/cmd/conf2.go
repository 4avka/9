package main

import (
	"git.parallelcoin.io/dev/9/cmd/config"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

var runConf = func(args []string, tokens config.Tokens, app *config.App) int {
	const menutitle = "ⓟ 9 parallelcoin configuration CLI"

	// pre-declare everything so we can decide the order to put things
	var titlebar *tview.TextView
	var runbranch, configBranch, treeroot *tview.TreeNode
	var treeview *tview.TreeView
	var root *tview.Flex
	var tapp *tview.Application

	// tapp pulls everything together to create the configuration interface
	tapp = func() (out *tview.Application) {
		out = tview.NewApplication()
		return
	}()

	// titlebar tells the user what app they are using
	titlebar = func() (out *tview.TextView) {
		out = tview.NewTextView()
		out.SetTextColor(tcell.ColorWhite).
			SetText(menutitle).
			Box.SetBackgroundColor(tcell.ColorDarkGreen)
		return
	}()

	// treeview is the tree used to navigate the options
	treeview = func() (out *tview.TreeView) {
		out = tview.NewTreeView()
		out.SetBorderPadding(0, 1, 1, 1)
		out.SetInputCapture(
			func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == 27 {
					tapp.Stop()
				}
				return event
			})
		return
	}()
	treeview.SetBorderPadding(0, 0, 1, 1)

	// root is the canvas (the whole current terminal view)
	root = func() (out *tview.Flex) {
		out = tview.NewFlex()
		out.SetDirection(tview.FlexRow)
		return
	}()

	// runbranch is the menu for running servers
	runbranch = func() (out *tview.TreeNode) {
		out = tview.NewTreeNode("🚦run a server")
		out.SetReference(root).
			SetSelectable(true).
			SetSelectedFunc(func() {
				out.SetExpanded(!out.IsExpanded())
			}).
			SetExpanded(false).
			AddChild(tview.NewTreeNode("🌱node")).
			AddChild(tview.NewTreeNode("💵wallet")).
			AddChild(tview.NewTreeNode("🐚shell"))
		return
	}()

	// // configBranch is the configuration item tree, structured to follow the
	// // two level tree containing config items
	// configBranch = func() (out *tview.TreeNode) {
	// 	out = tview.NewTreeNode("configuration")
	// 	out.SetSelectable(true).
	// 		SetSelectedFunc(func() {
	// 			// This toggles the branch to open or close
	// 			out.SetExpanded(!out.IsExpanded())
	// 		}).
	// 		SetExpanded(false)
	// 	return
	// }()

	configBranch = app.Cats.GetCatTree(tapp, treeview, root)

	// treeroot is the root object that contains nothing except tree nodes
	treeroot = func() (out *tview.TreeNode) {
		out = tview.NewTreeNode("9")
		// attach run and config menus
		out.AddChild(runbranch).
			AddChild(configBranch)
		return
	}()

	treeview.SetRoot(treeroot).
		SetCurrentNode(configBranch)
	root.AddItem(titlebar, 1, 1, false).
		AddItem(treeview, 0, 1, true)

	if e := tapp.SetRoot(root, true).Run(); e != nil {
		panic(e)
	}

	return 0
}
