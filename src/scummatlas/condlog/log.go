package condlog

import "fmt"

var Flags = map[string]bool{
	"script":    false,
	"palette":   false,
	"box":       false,
	"image":     false,
	"game":      false,
	"room":      false,
	"structure": false,
	"object":    false,
	"template":  false,
}

func Log(section string, format string, v ...interface{}) {
	if Flags[section] {
		fmt.Printf(format, v...)
		fmt.Println()
	}
}
