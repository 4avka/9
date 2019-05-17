// Demo code for the Checkbox primitive.
package main

import "git.parallelcoin.io/dev/9/pkg/util/tview"

func main() {
	app := tview.NewApplication()
	checkbox := tview.NewCheckbox().SetLabel("Hit Enter to check box: ")
	if err := app.SetRoot(checkbox, true).Run(); err != nil {
		panic(err)
	}
}
