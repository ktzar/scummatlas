package condlog

import "fmt"

var Logflags = map[string]bool{
	"script":    false,
	"palette":   false,
	"box":       false,
	"image":     false,
	"game":      true,
	"room":      false,
	"structure": false,
	"object":    false,
	"template":  true,
}

func Log(section string, format string, v ...interface{}) {
	if Logflags[section] {
		fmt.Printf(format, v...)
		fmt.Println()
	}
}
