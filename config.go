package main

import "image/color"

const (
	ScreenWidth  = 256
	ScreenHeight = 256

	BoxX = 8
	BoxY = 8
	BoxW = 240
	BoxH = 240

	Padding  = 8
	FooterH  = 16
	MaxTextW = BoxW - Padding*2

	CharInterval = 2
	FastInterval = 1

	FontSizeDefault = 12
)

var FontConfig = map[int]string{
	10: "PixelMplus10-Regular.ttf",
	12: "PixelMplus12-Regular.ttf",
}

const FilePath = "aozora_416.txt"

// Pyxel デフォルト16色パレット
var (
	PyxelBlack     = color.RGBA{0x00, 0x00, 0x00, 0xFF} // 0
	PyxelDarkBlue  = color.RGBA{0x1D, 0x2B, 0x53, 0xFF} // 1
	PyxelPurple    = color.RGBA{0x7E, 0x25, 0x53, 0xFF} // 2
	PyxelGreen     = color.RGBA{0x00, 0x87, 0x51, 0xFF} // 3
	PyxelBrown     = color.RGBA{0xAB, 0x52, 0x36, 0xFF} // 4
	PyxelDarkGray  = color.RGBA{0x5F, 0x57, 0x4F, 0xFF} // 5
	PyxelLightGray = color.RGBA{0xC2, 0xC3, 0xC7, 0xFF} // 6
	PyxelWhite     = color.RGBA{0xFF, 0xF1, 0xE8, 0xFF} // 7
	PyxelRed       = color.RGBA{0xFF, 0x00, 0x4D, 0xFF} // 8
	PyxelOrange    = color.RGBA{0xFF, 0xA3, 0x00, 0xFF} // 9
	PyxelYellow    = color.RGBA{0xFF, 0xEC, 0x27, 0xFF} // 10
	PyxelLime      = color.RGBA{0x00, 0xE4, 0x36, 0xFF} // 11
	PyxelCyan      = color.RGBA{0x29, 0xAD, 0xFF, 0xFF} // 12
	PyxelIndigo    = color.RGBA{0x83, 0x76, 0x9C, 0xFF} // 13
	PyxelPink      = color.RGBA{0xFF, 0x77, 0xA8, 0xFF} // 14
	PyxelPeach     = color.RGBA{0xFF, 0xCC, 0xAA, 0xFF} // 15
)
