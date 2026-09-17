package main

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
