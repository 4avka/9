// Demo code for the InputField primitive.
package main

import (
	"git.parallelcoin.io/dev/9/pkg/util/tcell"
	"git.parallelcoin.io/dev/9/pkg/util/tview"
)

func main() {
	app := tview.NewApplication()
	inputField := tview.NewInputField().
		SetLabel("Enter a number: ").
		SetPlaceholder("E.g. 1234").
		SetFieldWidth(10).
		SetAcceptanceFunc(tview.InputFieldInteger).
		SetDoneFunc(func(key tcell.Key) {
			app.Stop()
		})
	if err := app.SetRoot(inputField, true).Run(); err != nil {
		panic(err)
	}
}
