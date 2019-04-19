package config

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const checkmark = "✅"
const space = "  "
const tDEFAULT = "(default)"

func GenBool(c *CatTreeContext) {
	row0 := c.Parent.GetReference().(RowText)
	row := &row0
	var ntrue, nfalse *tview.TreeNode
	def := row.Default.Get().(bool)
	tdef, fdef := "", " "+tDEFAULT
	if def {
		tdef, fdef = fdef, tdef
	}
	handler := func(t, f *tview.TreeNode) {
		row.Put(!row.Bool())
		if row.Bool() {
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
	if row.Bool() {
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

func GenPort(c *CatTreeContext) {
	row0 := c.Parent.GetReference().(RowText)
	row := &row0
	if row.Value != nil {
		c.Parent.AddChild(tview.NewTreeNode("<unset>"))
	} else {
		c.Parent.AddChild(tview.NewTreeNode(fmt.Sprint(row.Int())))
		c.Parent.AddChild(tview.NewTreeNode("clear"))
	}
	if row.Default != nil {
		if p, ok := row.Default.Get().(int); ok {
			c.Parent.AddChild(tview.NewTreeNode("set default (" + fmt.Sprint(p) + ")"))
		} else {
			c.Parent.AddChild(tview.NewTreeNode("<unset>"))
		}
	}
}

func GenInt(c *CatTreeContext) {
	row0 := c.Parent.GetReference().(RowText)
	row := &row0
	it := tview.NewTreeNode("")
	it.SetText(fmt.Sprint(row.Int()))
	c.Parent.AddChild(it)
	if row.Default != nil {
		if p, ok := row.Default.Get().(int); ok {
			defopt := tview.NewTreeNode(
				"set default (" + fmt.Sprint(p) + ")").
				SetSelectedFunc(func() {
					row.Value.Put(row.Default.Get())
					it.SetText(fmt.Sprint(row.Int()))
					c.App.ForceDraw()
				})
			c.Parent.AddChild(defopt)
		} else {
			c.Parent.AddChild(tview.NewTreeNode("<unset>"))
		}
	}
	it.SetSelectedFunc(func() {
		it.SetText(fmt.Sprint(c.App.GetFocus().GetRect()))
		c.App.ForceDraw()
	})
}

func GenFloat(c *CatTreeContext) {
	row0 := c.Parent.GetReference().(RowText)
	row := &row0
	c.Parent.AddChild(tview.NewTreeNode(fmt.Sprint(row.Float())))
	if row.Default != nil {
		if p, ok := row.Default.Get().(float64); ok {
			c.Parent.AddChild(tview.NewTreeNode("set default (" + fmt.Sprint(p) + ")"))
		} else {
			c.Parent.AddChild(tview.NewTreeNode("<unset>"))
		}
	}
}

func GenDuration(c *CatTreeContext) {
	row0 := c.Parent.GetReference().(RowText)
	row := &row0
	c.Parent.AddChild(tview.NewTreeNode(fmt.Sprint(row.Duration())))
	if row.Default != nil {
		if p, ok := row.Default.Get().(time.Duration); ok {
			c.Parent.AddChild(tview.NewTreeNode("set default (" + fmt.Sprint(p) + ")"))
		}
	}
}

func GenString(c *CatTreeContext) {
	row0 := c.Parent.GetReference().(RowText)
	row := &row0
	istring := func() string {
		return "'" + fmt.Sprint(row.Value.Get()) + "'"
	}
	var p string
	var ok bool
	if row.Default.Get() != nil {
		if p, ok = row.Default.Get().(string); ok {
		}
	}
	var iclear *tview.TreeNode
	var gen func(ctc *CatTreeContext, rtx *RowText)
	gen = func(ctc *CatTreeContext, rtx *RowText) {
		iclear = tview.NewTreeNode("clear")
		if rtx.Row.Value.Get() != nil {
			titem := tview.NewTreeNode(istring())
			ctc.Parent.AddChild(titem)
			if rtx.Row.Value.Get() != nil {
				ctc.Parent.AddChild(iclear)
			}
		} else {
			ctc.Parent.AddChild(tview.NewTreeNode("<unset>"))
		}
		if rtx.Row.Default.Get() != nil {
			dn := tview.NewTreeNode("set default (" + fmt.Sprint(p) + ")")
			ctc.Parent.AddChild(dn)
		}
	}
	gen(c, row)
}

func GenStringSlice(c *CatTreeContext) {
	row0 := c.Parent.GetReference().(RowText)
	row := &row0
	if row.Value != nil {
		switch ss := row.Value.Get().(type) {
		case []string:
			for _, x := range ss {
				if len(x) > 0 {
					sss := tview.NewTreeNode(x).
						AddChild(tview.NewTreeNode("edit")).
						AddChild(tview.NewTreeNode("delete"))
					c.Parent.AddChild(sss.
						SetSelectedFunc(func() {
							openjump(c.SetParent(sss))
						}).Collapse())
				}
			}
		default:
		}
	}
	an := tview.NewTreeNode("add new")
	c.Parent.AddChild(an)
	an.SetSelectedFunc(func() {
		openjump(c.SetParent(an))
	}).Collapse()
}

func GenOptions(c *CatTreeContext) {
	row0 := c.Parent.GetReference().(RowText)
	row := &row0
	val := row.Tag()

	var ok bool
	var def string
	var topts []*tview.TreeNode
	if def, ok = row.Default.Get().(string); ok {
	}
	for _, x := range row.Opts {
		var itemtext string
		if x == val {
			itemtext = checkmark
		} else {
			itemtext = space
		}
		itemtext += x
		if x == def {
			itemtext += " " + tDEFAULT
		}
		cc := tview.NewTreeNode(itemtext)
		topts = append(topts, cc)
		c.Parent.AddChild(cc)
	}
	handler := func(opt string) {
		row.Value = row.Value.Put(opt)
		for i, x := range row.Opts {
			var itemtext string
			if x == opt {
				itemtext = checkmark
			} else {
				itemtext = space
			}
			itemtext += x
			if x == def {
				itemtext += " " + tDEFAULT
			}
			topts[i].SetText(itemtext)
		}
		c.App.ForceDraw()
	}
	for ii, xx := range topts {
		x := xx
		i := ii
		x.SetSelectedFunc(func() {
			handler(row.Opts[i])
		})
	}
}

func GetStringInputFunc(
	root *tview.Flex,
	treeview *tview.TreeView,
	tapp *tview.Application,
	rt *RowText,
	vnode *tview.TreeNode,
	ctc *CatTreeContext,
	gen func(ctc *CatTreeContext, rtx *RowText),
) func() (out *tview.InputField) {
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
				validated := rt.Put(currenttext)
				if !validated {
					out.SetFieldBackgroundColor(tcell.ColorRed).
						SetFieldTextColor(tcell.ColorBlack).
						// SetText("saving key " + currentlabel + currenttext).
						SetLabel(currentlabel + "error: invalid")
				} else {
					ctc.Parent.ClearChildren()
					gen(ctc, rt)
					treeview.SetCurrentNode(ctc.Parent.GetChildren()[0])
					time.Sleep(time.Second * 1)
					out.SetFieldBackgroundColor(tcell.ColorDarkGreen).
						SetLabel("")
					root.RemoveItem(out)
					treeview.SetBorderPadding(0, 0, 1, 1)
					tapp.SetFocus(treeview).
						ForceDraw()
				}
			}
			return e
		})
		return
	}
}
