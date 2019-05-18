package conf

import "git.parallelcoin.io/dev/tcell"

// MainColor is the main background color for menu panels
func MainColor() tcell.Color {
	return tcell.NewRGBColor(64, 64, 64)
}

// DimColor is the colour of the most recently selected before current item
func DimColor() tcell.Color {
	return tcell.NewRGBColor(48, 48, 48)
}

// PrelightColor is the background colour of the next item ahead that is rendered
// when each item that opens it is moved onto with the cursor
func PrelightColor() tcell.Color {
	return tcell.NewRGBColor(32, 32, 32)
}

// TextColor is the color of normal text with MainColor as background
func TextColor() tcell.Color {
	return tcell.NewRGBColor(216, 216, 216)
}

// BackgroundColor is the colour of all parts not containing any widgets
func BackgroundColor() tcell.Color {
	return tcell.NewRGBColor(16, 16, 16)
}
