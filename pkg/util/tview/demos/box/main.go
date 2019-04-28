// Demo code for the Box primitive.
package main

import (
	"git.parallelcoin.io/dev/9/pkg/util/tcell"
	"git.parallelcoin.io/dev/tview"
)

func main() {
	box := tview.NewBox().
		SetBorder(true).
		SetBorderAttributes(tcell.AttrBold).
		SetTitle("A [red]c[yellow]o[green]l[darkcyan]o[blue]r[darkmagenta]f[red]u[yellow]l[white] [black:red]c[:yellow]o[:green]l[:darkcyan]o[:blue]r[:darkmagenta]f[:red]u[:yellow]l[white:] [::bu]title")
	if err := tview.NewApplication().SetRoot(box, true).Run(); err != nil {
		panic(err)
	}
}
