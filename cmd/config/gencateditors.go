package config

import (
	"fmt"
	"time"

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
	if rt.Row.Value != nil {
		c.Parent.AddChild(
			tview.NewTreeNode(
				"'" + fmt.Sprint(*rt.Row.Value) + "'"))
		if *rt.Row.Value != nil {
			c.Parent.AddChild(tview.NewTreeNode("clear"))
		}
	} else {
		c.Parent.AddChild(tview.NewTreeNode("<unset>"))
	}
	if rt.Row.Default != nil {
		if p, ok := (*rt.Row.Default).(string); ok {
			c.Parent.AddChild(tview.NewTreeNode("set default (" + fmt.Sprint(p) + ")"))
		}
	}
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

	if rt.Row.Default != nil {
		if p, ok := (*rt.Row.Default).(string); ok {
			c.Parent.AddChild(tview.NewTreeNode("set default (" + fmt.Sprint(p) + ")"))
		}
	}
}
