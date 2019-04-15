package main

import (
	"time"

	"git.parallelcoin.io/dev/9/cmd/config"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

var runConf = func(args []string, tokens config.Tokens, app *config.App) int {
	const menutitle = "ⓟ 9 parallelcoin configuration CLI"

	// pre-declare everything so we can decide the order to put things
	var titlebar *tview.TextView
	var runbranch, configBranch, treeroot *tview.TreeNode
	var input *tview.InputField
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
		return
	}()

	// root is the canvas (the whole current terminal view)
	root = func() (out *tview.Flex) {
		out = tview.NewFlex()
		out.SetDirection(tview.FlexRow)
		return
	}()

	// runbranch is the menu for running servers
	runbranch = func() (out *tview.TreeNode) {
		out = tview.NewTreeNode("run a server")
		out.SetReference(root).
			SetSelectable(true).
			SetSelectedFunc(func() {
				out.SetExpanded(!out.IsExpanded())
			}).
			SetExpanded(false).
			AddChild(tview.NewTreeNode("node")).
			AddChild(tview.NewTreeNode("wallet")).
			AddChild(tview.NewTreeNode("shell"))
		return
	}()

	// configBranch is the configuration item tree, structured to follow the
	// two level tree containing config items
	configBranch = func() (out *tview.TreeNode) {
		out = tview.NewTreeNode("configuration")
		out.SetSelectable(true).
			SetSelectedFunc(func() {
				// This toggles the branch to open or close
				out.SetExpanded(!out.IsExpanded())
			}).
			SetExpanded(false)
		return
	}()

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

	inputCapture := func(e *tcell.EventKey) *tcell.EventKey {
		switch {
		case e.Key() == 27:
			// if user presses escape, switch back to the treeview
			root.RemoveItem(input)
			treeview.SetBorderPadding(0, 1, 1, 1)
			tapp.SetFocus(treeview)
		case e.Key() == 13:
			// if the user presses enter, trigger validate and save
			currenttext := input.GetText()
			currentlabel := input.GetLabel()
			treeview.SetBorderPadding(0, 0, 1, 1)
			input.SetFieldBackgroundColor(tcell.ColorYellow).
				SetFieldTextColor(tcell.ColorBlack).
				SetText("saving key " + currentlabel + currenttext).
				SetLabel("")
			tapp.ForceDraw()
			time.Sleep(time.Second * 1)
			input.SetFieldBackgroundColor(tcell.ColorDarkGreen).
				SetLabel("")
			root.RemoveItem(input)
			treeview.SetBorderPadding(0, 1, 1, 1)
			tapp.SetFocus(treeview).
				ForceDraw()
		}
		return e
	}

	// input field is a field that appears at the bottom of the screen when
	// a string or number is being edited
	input = func() (out *tview.InputField) {
		out = tview.NewInputField()
		out.SetFieldBackgroundColor(tcell.ColorDarkGreen).
			SetFieldTextColor(tcell.ColorBlack).
			SetLabelColor(tcell.ColorBlack).
			Box.SetBackgroundColor(tcell.ColorGreen)
		out.SetInputCapture(inputCapture)
		return
	}()

	// This function can be used for any opener to push the view to the bottom
	// of the new branch and then return to the parent node so the user sees
	// when they have activated an item
	openjump := func(node *tview.TreeNode) {
		node.SetExpanded(!node.IsExpanded())
		if node.IsExpanded() {
			// This makes sure the user sees the group they unfold
			// first it jumps to the last child
			treeview.SetCurrentNode(
				node.GetChildren()[len(node.GetChildren())-1])
			tapp.ForceDraw()
			// then back to the parent node
			treeview.SetCurrentNode(node)
			tapp.ForceDraw()
		}
	}

	// now we assemble the first level of the configuration categories
	catkeys := app.Cats.GetSortedKeys()

	// next the map of items in each category
	var itemkeys [][]string
	for _, x := range catkeys {
		itemkeys = append(itemkeys, app.Cats[x].GetSortedKeys())
	}

	nodemap := make(map[string]map[string]*tview.TreeNode)
	catNodes := func() (out []*tview.TreeNode) {
		out = make([]*tview.TreeNode, len(catkeys))
		// first attach new nodes from the categories
		for i, x := range catkeys {
			out[i] = tview.NewTreeNode(x).
				SetReference(configBranch).
				SetSelectable(true).
				SetExpanded(false)
			outi := out[i]
			outi.SetSelectedFunc(func() { openjump(outi) })
			nodemap[x] = make(map[string]*tview.TreeNode)
		}
		// items per category are indexed by same order, and attached to the
		// nodes thusly. We constructed one and two dimensional string slice
		// encoding this order in the indexes of each
		for i, x := range itemkeys {
			for _, y := range x {
				nodemap[catkeys[i]][y] = tview.NewTreeNode(y).
					SetReference(out[i]).
					SetSelectable(true).
					SetExpanded(false)
				out[i].AddChild(nodemap[catkeys[i]][y])
			}
		}
		return
	}()

	// attach category nodes to tree
	for _, x := range catNodes {
		configBranch.AddChild(x)
	}

	if e := tapp.SetRoot(root, true).Run(); e != nil {
		panic(e)
	}

	// spew.Config.MaxDepth = 3
	// spew.Dump(nodemap)

	_, _ = catkeys, itemkeys
	_ = catNodes
	return 0
}
