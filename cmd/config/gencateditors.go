package config

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const checkmark = "✅"
const space = "  "

func GenBool(c *CatTreeContext, rt *RowText) {
	var ntrue, nfalse *tview.TreeNode
	def := (*rt.Row.Default).(bool)
	tdef, fdef := "", " (default)"
	if def {
		tdef, fdef = fdef, tdef
	}
	handler := func(t, f *tview.TreeNode) {
		*rt.Row.Value = !rt.Row.Bool()
		if rt.Row.Bool() {
			t.SetText(checkmark + "true" + tdef)
			f.SetText(space + "false" + fdef)
		} else {
			t.SetText(space + "true" + tdef)
			f.SetText(checkmark + "false" + fdef)
		}
		c.App.ForceDraw()
	}
	var ts, fs string
	fs, ts = "✅", "  "
	if rt.Row.Bool() {
		ts, fs = fs, ts
	}
	ntrue = tview.NewTreeNode(ts + "true" + tdef).
		SetSelectedFunc(func() {
			handler(ntrue, nfalse)
		})
	nfalse = tview.NewTreeNode(fs + "false" + fdef).
		SetSelectedFunc(func() {
			handler(ntrue, nfalse)
		})
	c.Parent.AddChild(ntrue).AddChild(nfalse)
}

func GenPort(c *CatTreeContext, rt *RowText) {
	if rt.Row.Value != nil {
		c.Parent.AddChild(tview.NewTreeNode("<unset>"))
	} else {
		c.Parent.AddChild(tview.NewTreeNode(fmt.Sprint(rt.Row.Int())))
		c.Parent.AddChild(tview.NewTreeNode("clear"))
	}
	if rt.Row.Default != nil {
		if p, ok := (*rt.Row.Default).(int); ok {
			c.Parent.AddChild(tview.NewTreeNode("set default (" + fmt.Sprint(p) + ")"))
		} else {
			c.Parent.AddChild(tview.NewTreeNode("<unset>"))
		}
	}
}

func GenInt(c *CatTreeContext, rt *RowText) {
	c.Parent.AddChild(tview.NewTreeNode(fmt.Sprint(rt.Row.Int())))
	if rt.Row.Default != nil {
		if p, ok := (*rt.Row.Default).(int); ok {
			c.Parent.AddChild(tview.NewTreeNode("set default (" + fmt.Sprint(p) + ")"))
		} else {
			c.Parent.AddChild(tview.NewTreeNode("<unset>"))
		}
	}
}

func GenFloat(c *CatTreeContext, rt *RowText) {
	c.Parent.AddChild(tview.NewTreeNode(fmt.Sprint(rt.Row.Float())))
	if rt.Row.Default != nil {
		if p, ok := (*rt.Row.Default).(float64); ok {
			c.Parent.AddChild(tview.NewTreeNode("set default (" + fmt.Sprint(p) + ")"))
		} else {
			c.Parent.AddChild(tview.NewTreeNode("<unset>"))
		}
	}
}

func GenDuration(c *CatTreeContext, rt *RowText) {
	c.Parent.AddChild(tview.NewTreeNode(fmt.Sprint(rt.Row.Duration())))
	if rt.Row.Default != nil {
		if p, ok := (*rt.Row.Default).(time.Duration); ok {
			c.Parent.AddChild(tview.NewTreeNode("set default (" + fmt.Sprint(p) + ")"))
		}
	}
}

func GenString(c *CatTreeContext, rt *RowText) {
	istring := func() string {
		return "'" + fmt.Sprint(*rt.Row.Value) + "'"
	}
	var p string
	var ok bool
	// var defiface interface{}
	if rt.Row.Default != nil {
		if p, ok = (*rt.Row.Default).(string); ok {
			// 		defiface = p
		}
	}
	iclear := tview.NewTreeNode("clear")
	var gen func(ctc *CatTreeContext, rtx *RowText)
	gen = func(ctc *CatTreeContext, rtx *RowText) {
		if rtx.Row.Value != nil {
			titem := tview.NewTreeNode(istring())
			titem.SetSelectedFunc(func() {
				ifunc := GetStringInputFunc(ctc.Root, ctc.TreeView, ctc.App)()
				ifunc.SetText(rtx.Row.Tag()).
					SetLabelColor(tcell.ColorWhite).
					SetLabel(rtx.Row.Type + "'" + rtx.Row.Name + "' ").
					SetFieldBackgroundColor(tcell.ColorBlack).
					SetFieldTextColor(tcell.ColorWhite).
					Box.SetBackgroundColor(tcell.ColorDarkGreen)
				ctc.Root.AddItem(ifunc, 1, 0, false)
				ctc.App.SetFocus(ifunc)
			})
			ctc.Parent.AddChild(titem)
			if *rtx.Row.Value != nil {
				ctc.Parent.AddChild(iclear.SetSelectedFunc(func() {
					ctc.Parent.ClearChildren()
					*rtx.Row.Value = nil
					gen(ctc, rtx)
					ctc.TreeView.SetCurrentNode(ctc.Parent.GetChildren()[0])
				}))
			}
		} else {
			ctc.Parent.AddChild(tview.NewTreeNode("<unset>"))
		}
		if rtx.Row.Default != nil {
			dn := tview.NewTreeNode("set default (" + fmt.Sprint(p) + ")")
			ctc.Parent.AddChild(dn)
			dn.SetSelectedFunc(func() {
				*rtx.Row.Value = (*rtx.Row.Default).(string)
				c.Parent.ClearChildren()
				gen(ctc, rtx)
				children := ctc.Parent.GetChildren()
				ll := len(children) - 1
				ctc.TreeView.SetCurrentNode(children[ll])
			})
		}
	}
	gen(c, rt)
}

func GenStringSlice(c *CatTreeContext, rt *RowText) {
	if rt.Row.Value != nil {
		switch ss := (*rt.Row.Value).(type) {
		case []string:
			for _, x := range ss {
				if len(x) > 0 {
					sss := tview.NewTreeNode(x).
						AddChild(tview.NewTreeNode("edit")).
						AddChild(tview.NewTreeNode("delete"))
					sss.SetSelectedFunc(func() {
						openjump(c.SetParent(sss))
					}).Collapse()
					c.Parent.AddChild(sss)
				}
			}
		default:
		}
	}
	c.Parent.AddChild(tview.NewTreeNode("add new"))
}

func GenOptions(c *CatTreeContext, rt *RowText) {
	val := rt.Row.Tag()

	var ok bool
	var def string
	var topts []*tview.TreeNode
	if def, ok = (*rt.Row.Default).(string); ok {
	}
	for _, x := range rt.Row.Opts {
		var itemtext string
		if x == val {
			itemtext = checkmark
		} else {
			itemtext = space
		}
		itemtext += x
		if x == def {
			itemtext += " (default)"
		}
		cc := tview.NewTreeNode(itemtext)
		topts = append(topts, cc)
		c.Parent.AddChild(cc)
	}
	handler := func(opt string) {
		*rt.Row.Value = opt
		for i, x := range rt.Row.Opts {
			var itemtext string
			if x == opt {
				itemtext = checkmark
			} else {
				itemtext = space
			}
			itemtext += x
			if x == def {
				itemtext += " (default)"
			}
			topts[i].SetText(itemtext)
		}
		c.App.ForceDraw()
	}
	for ii, xx := range topts {
		x := xx
		i := ii
		x.SetSelectedFunc(func() {
			handler(rt.Row.Opts[i])
		})
	}
}

func GetStringInputFunc(root *tview.Flex, treeview *tview.TreeView,
	tapp *tview.Application) func() (out *tview.InputField) {
	// input field is a field that appears at the bottom of the screen when
	// a string or number is being edited
	return func() (out *tview.InputField) {
		out = tview.NewInputField()
		out.SetFieldBackgroundColor(tcell.ColorDarkGreen).
			SetFieldTextColor(tcell.ColorBlack).
			SetLabelColor(tcell.ColorBlack).
			Box.SetBackgroundColor(tcell.ColorGreen)
		out.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
			switch {
			case e.Key() == 27:
				// if user presses escape, switch back to the treeview
				root.RemoveItem(out)
				treeview.SetBorderPadding(0, 1, 1, 1)
				tapp.SetFocus(treeview)
			case e.Key() == 13:
				// if the user presses enter, trigger validate and save
				currenttext := out.GetText()
				currentlabel := out.GetLabel()
				treeview.SetBorderPadding(0, 0, 1, 1)
				out.SetFieldBackgroundColor(tcell.ColorYellow).
					SetFieldTextColor(tcell.ColorBlack).
					SetText("saving key " + currentlabel + currenttext).
					SetLabel("")
				tapp.ForceDraw()
				time.Sleep(time.Second * 1)
				out.SetFieldBackgroundColor(tcell.ColorDarkGreen).
					SetLabel("")
				root.RemoveItem(out)
				treeview.SetBorderPadding(0, 0, 1, 1)
				tapp.SetFocus(treeview).
					ForceDraw()
			}
			return e
		})
		return
	}
}
