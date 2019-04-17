package main

import (
	"fmt"
	"strings"
	"time"

	"git.parallelcoin.io/dev/9/cmd/config"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const BACK = "back"

func RunConf(args []string, tokens config.Tokens, app *config.App) int {
	const menutitle = "ⓟ 9 parallelcoin configuration CLI"

	tapp := tview.NewApplication()

	// titlebar tells the user what app they are using
	titlebar := tview.NewTextView()
	titlebar.SetTextColor(tcell.ColorWhite).
		SetText(menutitle).
		Box.SetBackgroundColor(tcell.ColorDarkGreen)

	// a treeview is used to render the options as they are hierarchically
	// structured
	treeview := tview.NewTreeView()
	treeview.SetGraphics(true).
		SetGraphicsColor(tcell.ColorDarkGreen).
		SetBorderPadding(0, 1, 1, 1)
	treeview.SetInputCapture(
		func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == 27 {
				tapp.Stop()
			}
			return event
		})
	treeview.Box.SetBackgroundColor(tcell.ColorBlack)

	// flexbox contains all the page items, titlebar and tree
	root := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(titlebar, 1, 1, false).
		AddItem(treeview, 0, 1, true)

	// treeroot is the root of the treeview node tree
	treeroot := tview.NewTreeNode("9")
	treeroot.SetSelectable(false).
		SetIndent(1)
	// add node as root of treeview
	treeview.SetRoot(treeroot).SetCurrentNode(treeroot)

	// first menu is for launching servers
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
	// input field is attached to the bottom when text (string/number) input is
	// required
	input := tview.NewInputField()
	input.SetFieldBackgroundColor(tcell.ColorDarkGreen).
		SetFieldTextColor(tcell.ColorBlack).
		SetLabelColor(tcell.ColorBlack).
		Box.SetBackgroundColor(tcell.ColorGreen)
	input.SetInputCapture(
		func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == 27 {
				root.RemoveItem(input)
				tapp.SetFocus(treeview)
				treeview.SetBorderPadding(0, 1, 1, 1)
			}
			if event.Key() == 13 {
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
			return event
		})

	treeroot.AddChild(runbranch.SetReference(treeroot))
	// configbase is the configuration item tree, structured to follow the
	// two level tree containing config items
	configbase := tview.NewTreeNode("configuration")
	configbase.SetSelectable(true).
		SetSelectedFunc(func() {
			// This toggles the branch to open or close
			configbase.SetExpanded(!configbase.IsExpanded())
		}).
		SetExpanded(false)
	// first get the keys of the first level of the tree
	for _, x := range app.Cats.GetSortedKeys() {
		tn := tview.NewTreeNode(x).
			SetReference(configbase).
			SetSelectable(true).
			SetExpanded(false)
		tn.SetSelectedFunc(func() {
			tn.SetExpanded(!tn.IsExpanded())
			if tn.IsExpanded() {
				// This makes sure the user sees the group they unfold
				// first it jumps to the last child
				treeview.SetCurrentNode(tn.GetChildren()[len(tn.GetChildren())-1])
				tapp.ForceDraw()
				// then back to the parent node
				treeview.SetCurrentNode(tn)
				tapp.ForceDraw()
			}
		})
		// we want it to be neat, so here we compute max width for the tag label
		// fields.
		maxlen := 0
		maxvaluelen := 0
		// first get the max length of the keys for this section
		for _, j := range app.Cats[x].GetSortedKeys() {
			if len(j) > maxlen {
				maxlen = len(j)
			}
			// get the max length of the values in each
			if yv := app.Cats[x][j].Value; yv != nil {
				valtext := fmt.Sprint(*yv)
				if len(valtext) > 0 {
					if len(valtext) > maxvaluelen {
						maxvaluelen = len(valtext)
					}
				}
			}
		}
		// This loop is separately run because we need the pad length for all of
		// the keys and values before we start constructing them
		for _, j := range app.Cats[x].GetSortedKeys() {
			// TODO: types for multi-bool-options
			V, X := "", j
			acxj := app.Cats[x][j]
			if yv := acxj.Value; yv != nil {
				if *yv != nil {
					V = fmt.Sprint(*yv)
				}
			}

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
			tnj := tview.NewTreeNode(
				"[:] " + keytext + " [white:darkgreen] " +
					valuetext + " [darkgray:black] " + app.Cats[x][j].Usage + "[:]")
			tnj.SetReference(configbase).
				SetSelectable(true).
				SetExpanded(false).
				SetSelectedFunc(func() {
					tnj.SetExpanded(!tnj.IsExpanded())
					treeview.SetCurrentNode(tnj)
					if tn.IsExpanded() {
						switch acxj.Type {
						case "string":
							treeview.SetBorderPadding(0, 0, 1, 1)
							root.AddItem(input, 1, 1, false)
							input.SetText(V).
								SetLabelColor(tcell.ColorWhite).
								SetLabel(acxj.Type + "'" + X + "' ").
								SetFieldBackgroundColor(tcell.ColorBlack).
								SetFieldTextColor(tcell.ColorWhite).
								Box.SetBackgroundColor(tcell.ColorDarkGreen)
							tapp.SetFocus(input)
						case "bool":
							tnj.ClearChildren()
							tt := tview.NewTreeNode("true")
							ff := tview.NewTreeNode("false")
							tnj.AddChild(
								tt,
							).AddChild(
								ff,
							)

							vv, ok := (*acxj.Value).(bool)
							if ok {
								if vv {
									tt.SetColor(tcell.ColorGreen)
									ff.SetColor(tcell.ColorWhite)
									// treeview.SetCurrentNode(tt)
								} else {
									tt.SetColor(tcell.ColorWhite)
									ff.SetColor(tcell.ColorGreen)
									// treeview.SetCurrentNode(ff)
								}
							} else {
								dd, ok := (*acxj.Default).(bool)
								if ok {
									if dd {
										treeview.SetCurrentNode(tt)
									} else {
										treeview.SetCurrentNode(ff)
									}
								}
							}
							// This makes sure the user sees the group they unfold
							// first it jumps to the last child
							treeview.SetCurrentNode(
								tnj.GetChildren()[len(tnj.GetChildren())-1])
							tapp.ForceDraw()
							// then back to the parent node
							treeview.SetCurrentNode(tnj)
							tapp.ForceDraw()

						case "options":
							tnj.ClearChildren()
							opts := make(map[string]*tview.TreeNode)
							current := (*acxj.Value).(string)
							for _, x := range acxj.Opts {
								opts[x] = tview.NewTreeNode(x)
								if x == current {
									opts[x].SetColor(tcell.ColorGreen)
								} else {
									opts[x].SetColor(tcell.ColorWhite)
								}
								tnj.AddChild(opts[x])
							}
							// This makes sure the user sees the group they unfold
							// first it jumps to the last child
							treeview.SetCurrentNode(
								tnj.GetChildren()[len(tnj.GetChildren())-1])
							tapp.ForceDraw()
							// then back to the parent node
							treeview.SetCurrentNode(tnj)
							tapp.ForceDraw()

							// vv, ok := (*acxj.Value).(bool)
							// if ok {
							// 	if vv {
							// 		treeview.SetCurrentNode(tt)
							// 	} else {
							// 		treeview.SetCurrentNode(ff)
							// 	}
							// } else {
							// 	dd, ok := (*acxj.Default).(bool)
							// 	if ok {
							// 		if dd {
							// 			treeview.SetCurrentNode(tt)
							// 		} else {
							// 			treeview.SetCurrentNode(ff)
							// 		}
							// 	}
							// }
						}
					} else {
						switch acxj.Type {
						case "string":
							root.RemoveItem(input)
							tapp.SetFocus(treeview)
							tnj.SetExpanded(false)
						case "bool":
							tnj.ClearChildren()
							// treeview.SetCurrentNode(tnj)
							// tnj.SetExpanded(false)
						case "options":
							tnj.ClearChildren()
							// treeview.SetCurrentNode(tnj)
							// tnj.SetExpanded(false)
						}
					}
				})
			tn.AddChild(tnj)
		}
		configbase.AddChild(tn)
	}
	// attach the constructed configuration tree to the main tree
	treeroot.AddChild(
		configbase.SetReference(root).SetSelectable(true),
	)
	// attach an exit option
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

	// spew.Dump(app.Cats)

	return 0
}
