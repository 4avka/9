package conf

import "git.parallelcoin.io/dev/tcell"

func MainColor() tcell.Color {
	return tcell.NewRGBColor(64, 64, 64)
}

func DimColor() tcell.Color {
	return tcell.NewRGBColor(48, 48, 48)
}

func PrelightColor() tcell.Color {
	return tcell.NewRGBColor(32, 32, 32)
}

func TextColor() tcell.Color {
	return tcell.NewRGBColor(216, 216, 216)
}

func BackgroundColor() tcell.Color {
	return tcell.NewRGBColor(16, 16, 16)
}
