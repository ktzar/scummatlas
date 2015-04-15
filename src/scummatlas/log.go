package scummatlas

import "fmt"

var Logflags = map[string]bool{
	"script":  false,
	"palette": false,
	"box":     false,
	"image":   false,
	"room":    false,
	"object":  false,
}

func log(section string, format string, v ...interface{}) {
	if Logflags[section] {
		fmt.Printf(format, v...)
	}
}
