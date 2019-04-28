package main

import (
	"fmt"

	"git.parallelcoin.io/dev/9/pkg/util/tcell"
	"git.parallelcoin.io/dev/tview"
)

// End shows the final slide.
func End(nextSlide func()) (title string, content tview.Primitive) {
	textView := tview.NewTextView().SetDoneFunc(func(key tcell.Key) {
		nextSlide()
	})
	url := "https://git.parallelcoin.io/dev/tview"
	fmt.Fprint(textView, url)
	return "End", Center(len(url), 1, textView)
}
