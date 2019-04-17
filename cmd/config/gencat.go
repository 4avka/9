package config

import (
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/rivo/tview"
)

type rowText struct {
	Row      Row
	Cat      string
	NodeText string
}
type rowTexts []rowText

func (c *Cats) GetCatTree(ta *tview.Application, tv *tview.TreeView) (out *tview.TreeNode) {
	C := *c
	// The root is configuration
	out = tview.NewTreeNode("📁configuration") //.Collapse()

	cats := c.GetSortedKeys()

	var items []rowTexts

	for _, x := range cats {
		cx := C[x]
		var it rowTexts
		for _, y := range cx.GetSortedKeys() {
			it = append(it, rowText{
				Row: cx[y],
				Cat: x,
			})
		}
		rts := it.GenRowTexts()
		for i := range it {
			it[i].NodeText = rts[i]
		}
		items = append(items, it)
	}

	for i := range items {
		co := tview.NewTreeNode("📁" + cats[i])
		for _, x := range items[i] {
			io := tview.NewTreeNode(x.NodeText)
			switch x.Row.Type {
			case "bool":
				if x.Row.Bool() {
					io.AddChild(tview.NewTreeNode("✅true"))
					io.AddChild(tview.NewTreeNode("  false"))
				} else {
					io.AddChild(tview.NewTreeNode("  true"))
					io.AddChild(tview.NewTreeNode("✅false"))
				}
				def := "false"
				if (*x.Row.Default).(bool) {
					def = "true"
				}
				io.AddChild(tview.NewTreeNode("set default (" + def + ")"))
			case "port":
				if x.Row.Value != nil {
					io.AddChild(tview.NewTreeNode("<unset>"))
				} else {
					io.AddChild(tview.NewTreeNode(fmt.Sprint(x.Row.Int())))
					io.AddChild(tview.NewTreeNode("clear"))
				}
				if x.Row.Default != nil {
					if p, ok := (*x.Row.Default).(int); ok {
						io.AddChild(tview.NewTreeNode("set default (" + fmt.Sprint(p) + ")"))
					} else {
						io.AddChild(tview.NewTreeNode("<unset>"))
					}
				}
			case "int":
				io.AddChild(tview.NewTreeNode(fmt.Sprint(x.Row.Int())))
				if x.Row.Default != nil {
					if p, ok := (*x.Row.Default).(int); ok {
						io.AddChild(tview.NewTreeNode("set default (" + fmt.Sprint(p) + ")"))
					} else {
						io.AddChild(tview.NewTreeNode("<unset>"))
					}
				}
			case "float":
				io.AddChild(tview.NewTreeNode(fmt.Sprint(x.Row.Float())))
				if x.Row.Default != nil {
					if p, ok := (*x.Row.Default).(float64); ok {
						io.AddChild(tview.NewTreeNode("set default (" + fmt.Sprint(p) + ")"))
					} else {
						io.AddChild(tview.NewTreeNode("<unset>"))
					}
				}
			case "duration":
				//
			case "string":
				if x.Row.Value != nil {
					io.AddChild(tview.NewTreeNode(fmt.Sprint(*x.Row.Value)))
				} else {
					io.AddChild(tview.NewTreeNode("<unset>"))
				}
				// if x.Row.Value == nil || *x.Row.Value != nil {
				// } else {
				// 	io.AddChild(tview.NewTreeNode(x.Row.Tag()))
				// }
				// if x.Row.Default != nil {
				// 	if p, ok := (*x.Row.Default).(int); ok {
				// 		io.AddChild(tview.NewTreeNode("set default (" + fmt.Sprint(p) + ")"))
				// 	}
				// }
			case "stringslice":
				//
			case "options":
				//
			default:
			}
			co.AddChild(io)
		}
		out.AddChild(co)
	}

	spew.Dump(items)

	return
}

func (r rowTexts) GenRowTexts() (out []string) {
	maxNameLen := 0
	// maxValueLen := 0
	for _, x := range r {
		if len(x.Row.Name) > maxNameLen {
			maxNameLen = len(x.Row.Name)
		}
	}
	for i, x := range r {
		padlen := maxNameLen - len(x.Row.Name)
		pad := strings.Repeat(" ", padlen+1)
		out = append(out, x.Row.Name+pad)
		out[i] += x.Row.Usage
	}
	return
}

// This function can be used for any opener to push the view to the bottom
// of the new branch and then return to the parent node so the user sees
// when they have activated an item
// openjump := func(node *tview.TreeNode) {
// 	node.SetExpanded(!node.IsExpanded())
// 	if node.IsExpanded() {
// 		// This makes sure the user sees the group they unfold
// 		// first it jumps to the last child
// 		ii := len(node.GetChildren()) - 1
// 		if ii > 0 {
// 			tv.SetCurrentNode(
// 				node.GetChildren()[len(node.GetChildren())-1])
// 			ta.ForceDraw()
// 		}
// 		// then back to the parent node
// 		tv.SetCurrentNode(node)
// 		ta.ForceDraw()
// 	}
// }
// under root is the categories. gather the texts for each item in each
// category. they are collected first because their layout depends on the
// maximum width text in each set of tag and value
// var cats []*tview.TreeNode
// var itemtexts []rowTexts
// for _, cat := range c.GetSortedKeys() {
// 	catmap := (*c)[cat]
// 	ct := tview.NewTreeNode("📁" + cat).Collapse()
// 	ct.SetSelectedFunc(func() { openjump(ct) })
// 	cats = append(cats, ct)
// 	itemtexts = append(itemtexts, rowTexts{})
// 	for _, item := range catmap.GetSortedKeys() {
// 		row := catmap[item]
// 		// _, _, _, _ = cat, catmap, item, row
// 		itemtexts[len(itemtexts)-1] =
// 			append(itemtexts[len(itemtexts)-1], row.GetRowText())
// 	}
// }
// // generate the tree node text
// for _, x := range itemtexts {
// 	rt := x.GenRowTexts()
// 	for i, y := range rt {
// 		x[i].NodeText = y
// 	}
// }
// // generate the array of treenodes with their attached options/editors
// var items [][]*tview.TreeNode
// for _, x := range itemtexts {
// 	var it []*tview.TreeNode
// 	for _, y := range x {
// 		tt := tview.NewTreeNode(y.NodeText).Collapse()
// 		tt.SetSelectedFunc(func() { openjump(tt) })
// 		// next attach the editors and options nodes
// 		switch y.Row.Type {
// 		case "bool":
// 			ttrue, tfalse := "  true", "  false"
// 			tvalue := fmt.Sprint(*y.Row.Value)
// 			if tvalue == "true" {
// 				ttrue = "✅" + ttrue[2:]
// 			}
// 			if tvalue == "false" {
// 				tfalse = "✅" + tfalse[2:]
// 			}
// 			btrue := tview.NewTreeNode(ttrue)
// 			bfalse := tview.NewTreeNode(tfalse)
// 			btrue.SetSelectedFunc(func() {
// 				// if not selected, select
// 			})
// 			bfalse.SetSelectedFunc(func() {
// 				// if not selected, select
// 			})
// 			bdefault :=
// 				tview.NewTreeNode("set default(" + fmt.Sprint(y.Row.Default) + ")").
// 					SetSelectedFunc(func() {
// 						// if not selected, select default
// 					})
// 			tt.AddChild(btrue)
// 			tt.AddChild(bfalse)
// 			tt.AddChild(bdefault)
// 		case "port":
// 			tt.AddChild(tview.NewTreeNode(fmt.Sprint(y.Row.Value)))
// 			tt.AddChild(tview.NewTreeNode("set default (" + fmt.Sprint(y.Row.Default) + ")"))
// 			tt.AddChild(tview.NewTreeNode("clear"))
// 			tt.Collapse().SetSelectedFunc(func() { openjump(tt) })
// 			// tt.AddChild(itemnode)
// 			// add editor for ports
// 		case "int":
// 			tt.AddChild(tview.NewTreeNode(fmt.Sprint(y.Value)))
// 			tt.AddChild(tview.NewTreeNode("set default (" + fmt.Sprint(y.Row.Default) + ")"))
// 			// tt.AddChild(tview.NewTreeNode("clear"))
// 			tt.Collapse().SetSelectedFunc(func() { openjump(tt) })
// 			// tt.AddChild(itemnode)
// 			// // add editor for integers
// 		case "float":
// 			tt.AddChild(tview.NewTreeNode(fmt.Sprint(y.Value)))
// 			tt.AddChild(tview.NewTreeNode("set default (" + fmt.Sprint(y.Row.Default) + ")"))
// 			// tt.AddChild(tview.NewTreeNode("clear"))
// 			tt.Collapse().SetSelectedFunc(func() { openjump(tt) })
// 			// add editor for floats
// 		case "duration":
// 			tt.AddChild(tview.NewTreeNode(fmt.Sprint(y.Value)))
// 			tt.AddChild(tview.NewTreeNode("set default (" + fmt.Sprint(y.Row.Default) + ")"))
// 			// tt.AddChild(tview.NewTreeNode("clear"))
// 			tt.Collapse().SetSelectedFunc(func() { openjump(tt) })
// 			// add editor for durations
// 		case "string":
// 			// var itemnode *tview.TreeNode
// 			// if !(y.Value == "<nil>" || y.Value == "") {
// 			tt.AddChild(tview.NewTreeNode("[ " + fmt.Sprint(y.Value) + " ]"))
// 			tt.AddChild(tview.NewTreeNode("set default (" + fmt.Sprint(y.Row.Default) + ")"))
// 			tt.AddChild(tview.NewTreeNode("clear"))
// 			tt.Collapse().SetSelectedFunc(func() { openjump(tt) })
// 			// } else {
// 			// 	tt.AddChild(tview.NewTreeNode("edit"))
// 			// 	tt.AddChild(tview.NewTreeNode("set default (" + y.Default + ")"))
// 			// 	tt.Collapse().SetSelectedFunc(func() { openjump(tt) })
// 			// }
// 			// string editor
// 		case "stringslice":
// 			tt.AddChild(tview.NewTreeNode("add new item"))
// 			if !(y.Row.Value == "" || y.Row.Value == "<nil>") {
// 				ttt := tview.NewTreeNode(fmt.Sprint(y.Value)).
// 					AddChild(tview.NewTreeNode("edit")).
// 					AddChild(tview.NewTreeNode("delete"))
// 				ttt.Collapse().SetSelectedFunc(func() { openjump(ttt) })
// 				tt.AddChild(ttt)
// 			}
// 			tt.AddChild(tview.NewTreeNode("set default (" + y.Default + ")"))
// 			tt.Collapse().
// 				SetSelectedFunc(func() { openjump(tt) })
// 			// add 'new', and all existing items as treenodes, and a delete
// 			// option on the existing options
// 		case "options":
// 			// add all of the options and set highlight on the current one
// 			for _, x := range y.Opts {
// 				var opttext string

// 				if y.Value == x {
// 					opttext = "✅"
// 				} else {
// 					opttext = "  "
// 				}
// 				optitem := tview.NewTreeNode(opttext + x)
// 				tt.AddChild(optitem)
// 			}
// 			tt.AddChild(tview.NewTreeNode("set default (" + y.Default + ")"))
// 		default:
// 		}
// 		it = append(it, tt)
// 	}
// 	items = append(items, it)
// }
// // string all the treenodes together
// for i, x := range items {
// 	for j, y := range x {
// 		cats[i].AddChild(y).Collapse()
// 		_, _, _, _ = i, x, j, y
// 	}
// 	out.AddChild(cats[i]).Collapse()
// }
// out.SetSelectedFunc(func() { openjump(out) }).CollapseAll()

// func (r Row) GetRowText() (out rowText) {
// 	out.Name = r.Name
// 	if r.Value != nil {
// 		out.Value = fmt.Sprint(*r.Value)
// 	} else {
// 		out.Value = fmt.Sprint(r.Value)
// 	}
// 	if r.Default != nil {
// 		out.Default = fmt.Sprint(*r.Default)
// 	} else {
// 		out.Default = fmt.Sprint(r.Default)
// 	}
// 	out.Usage = r.Usage
// 	out.Type = r.Type
// 	out.Opts = r.Opts
// 	out.Put = r.Put
// 	return
// }
