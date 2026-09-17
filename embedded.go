//go:build embed
// +build embed

package main

import _ "embed"

//go:embed PixelMplus10-Regular.ttf
var font10Data []byte

//go:embed PixelMplus12-Regular.ttf
var font12Data []byte

//go:embed aozora_416.txt
var textData []byte
