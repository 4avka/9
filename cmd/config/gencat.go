package config

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
)

type rowText struct {
	Name     string
	Value    string
	Usage    string
	NodeText string
	Type     string
	Opts     []string
	Put      func(interface{}) bool
}
type rowTexts []rowText

func (c *Cats) GetCatTree(ta *tview.Application, tv *tview.TreeView) (out *tview.TreeNode) {
	// This function can be used for any opener to push the view to the bottom
	// of the new branch and then return to the parent node so the user sees
	// when they have activated an item
	openjump := func(node *tview.TreeNode) {
		node.SetExpanded(!node.IsExpanded())
		if node.IsExpanded() {
			// This makes sure the user sees the group they unfold
			// first it jumps to the last child
			// TODO: bounds error here about 0 child nodes
			tv.SetCurrentNode(
				node.GetChildren()[len(node.GetChildren())-1])
			ta.ForceDraw()
			// then back to the parent node
			tv.SetCurrentNode(node)
			ta.ForceDraw()
		}
	}
	// The root is configuration
	out = tview.NewTreeNode("📁configuration")
	// under root is the categories. gather the texts for each item in each
	// category. they are collected first because their layout depends on the
	// maximum width text in each set of tag and value
	var cats []*tview.TreeNode
	var itemtexts []rowTexts
	for cat, catmap := range *c {
		ct := tview.NewTreeNode("📁" + cat)
		ct.SetSelectedFunc(func() { openjump(ct) })
		cats = append(cats, ct)
		itemtexts = append(itemtexts, rowTexts{})
		for item, row := range catmap {
			_, _, _, _ = cat, catmap, item, row
			itemtexts[len(itemtexts)-1] =
				append(itemtexts[len(itemtexts)-1], row.GetRowText())
		}
	}
	// generate the tree node text
	for _, x := range itemtexts {
		rt := x.GenRowTexts()
		for i, y := range rt {
			x[i].NodeText = y
		}
	}
	// generate the array of treenodes with their attached options/editors
	var items [][]*tview.TreeNode
	for _, x := range itemtexts {
		var it []*tview.TreeNode
		for _, y := range x {
			symbol := "📄"
			switch y.Type {
			case "string":
				symbol = "🆎"
			case "int", "float":
				symbol = "1⃣ "
			case "stringslice":
				symbol = "📝"
			case "options":
				symbol = "✨"
			case "bool":
				symbol = "🚫"
			case "duration":
				symbol = "⏰"
			}
			tt := tview.NewTreeNode(symbol + y.NodeText)
			tt.SetSelectedFunc(func() { openjump(tt) })
			// next attach the editors and options nodes
			switch y.Type {
			case "bool":
				ttrue, tfalse := "  true", "  false"
				tvalue := strings.ToLower(y.Value)
				if tvalue == "true" {
					ttrue = "✅" + ttrue[2:]
				}
				if tvalue == "false" {
					tfalse = "✅" + tfalse[2:]
				}
				tt.AddChild(tview.NewTreeNode(ttrue))
				tt.AddChild(tview.NewTreeNode(tfalse))
				// add true and false option treenodes
			case "int":
				// add editor for integers
			case "float":
				// add editor for floats
			case "duration":
				// add editor for durations
			case "string":
				// string editor
			case "stringslice":
				// add 'new', and all existing items as treenodes, and a delete
				// option the existing options
			case "options":
				for _, x := range y.Opts {
					var opttext string
					if y.Value == x {
						opttext = "✅"
					} else {
						opttext = "  "
					}
					optitem := tview.NewTreeNode(opttext + x)
					tt.AddChild(optitem)
				}
				// add all of the options and set highlight on the current one
			default:
			}
			it = append(it, tt)
		}
		items = append(items, it)
	}
	// string all the treenodes together
	for i, x := range items {
		for j, y := range x {
			cats[i].AddChild(y)
			_, _, _, _ = i, x, j, y
		}
		out.AddChild(cats[i])
	}

	return
}

func (r Row) GetRowText() (out rowText) {
	out.Name = r.Name
	if r.Value != nil {
		out.Value = fmt.Sprint(*r.Value)
	} else {
		out.Value = ""
	}
	out.Usage = r.Usage
	out.Type = r.Type
	out.Opts = r.Opts
	out.Put = r.Put
	return
}

func (r rowTexts) GenRowTexts() (out []string) {
	maxNameLen, maxValueLen := 0, 0
	for _, x := range r {
		if len(x.Name) > maxNameLen {
			maxNameLen = len(x.Name)
		}
		if len(x.Value) > maxValueLen {
			maxValueLen = len(x.Value)
		}
	}
	for i, x := range r {
		padlen := maxNameLen - len(x.Name)
		pad := strings.Repeat(" ", padlen+1)
		out = append(out, x.Name+pad)
		padlen = maxValueLen - len(x.Value)
		pad = strings.Repeat(" ", padlen+1)
		out[i] += " " + x.Value + pad + x.Usage
	}
	return
}
