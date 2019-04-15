package main

import (
	"git.parallelcoin.io/dev/9/cmd/config"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

var runConf = func(args []string, tokens config.Tokens, app *config.App) int {
	const menutitle = "ⓟ 9 parallelcoin configuration CLI"

	// titlebar tells the user what app they are using
	titlebar := func() (out *tview.TextView) {
		out = tview.NewTextView()
		out.SetTextColor(tcell.ColorWhite).
			SetText(menutitle).
			Box.SetBackgroundColor(tcell.ColorDarkGreen)
		return
	}

	// runbranch is the menu for running servers
	runbranch := func() (out *tview.TreeNode) {
		out = tview.NewTreeNode("run a server")
		return
	}

	// treeroot is the root object that contains nothing except tree nodes
	treeroot := func() (out *tview.TreeNode) {
		out = tview.NewTreeNode("9")
		return
	}

	// configbranch is the configuration item tree, structured to follow the
	// two level tree containing config items
	configbranch := tview.NewTreeNode("configuration")

	// input field is a field that appears at the bottom of the screen when
	// a string or number is being edited
	input := func() (out *tview.InputField) {
		out = tview.NewInputField()
		out.SetFieldBackgroundColor(tcell.ColorDarkGreen).
			SetFieldTextColor(tcell.ColorBlack).
			SetLabelColor(tcell.ColorBlack).
			Box.SetBackgroundColor(tcell.ColorGreen)
		return
	}

	// treeview is the tree used to navigate the options
	treeview := func() (out *tview.TreeView) {
		out = tview.NewTreeView()
		return
	}

	// root is the canvas (the whole current terminal view)
	root := func() (out *tview.Flex) {
		out = tview.NewFlex()
		out.SetDirection(tview.FlexRow).
			AddItem(titlebar(), 1, 1, false).
			AddItem(treeview(), 0, 1, true)
		return
	}

	// tapp pulls everything together to create the configuration interface
	tapp := func() (out *tview.Application) {
		out = tview.NewApplication()
		return
	}

	_ = configbranch
	_ = runbranch
	_ = input
	_ = titlebar
	_ = treeroot
	_ = treeview
	_ = root
	_ = tapp
	return 0
}
