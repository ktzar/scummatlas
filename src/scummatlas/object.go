package scummatlas

import (
	"fmt"
	"image"
	"strings"
)

type Object struct {
	Image  *image.RGBA
	Script Script
	Name   string
	Id     int
	X      int
	Y      int
	Width  int
	Height int
	//TODO Direction uint8
	Flags  uint8
	Parent uint8
}

func (self Object) IdHex() string {
	return fmt.Sprintf("%x", self.Id)
}

func NewObjectFromOBCD(data []byte) Object {
	headerOffset := 8
	if FourCharString(data, headerOffset) != "CDHD" {
		panic("No object header")
	}
	headerSize := BE32(data, headerOffset+4)

	intInOffsetTimesEight := func(offset int) int {
		return int(data[headerOffset+offset]) * 8
	}
	obj := Object{
		Id:     LE16(data, headerOffset+8),
		X:      intInOffsetTimesEight(10),
		Y:      intInOffsetTimesEight(11),
		Width:  intInOffsetTimesEight(12),
		Height: intInOffsetTimesEight(13),
		Flags:  data[headerOffset+14],
		Parent: data[headerOffset+15],
	}

	verbOffset := headerOffset + headerSize
	if FourCharString(data, verbOffset) != "VERB" {
		panic("Object with no verbs")
	}
	verbSize := BE32(data, verbOffset+4)
	objNameOffset := verbOffset + verbSize
	if FourCharString(data, objNameOffset) != "OBNA" {
		panic("Object with no name")
	}
	objNameSize := BE32(data, objNameOffset+4)
	name := data[objNameOffset+4 : objNameOffset+objNameSize]
	obj.Name = filterObjectName(name)
	return obj
}

func filterObjectName(in []byte) (out string) {
	filtered := []byte{}
	for _, v := range in {
		if v != 0x40 && v != 0x00 && v != 0x0f {
			filtered = append(filtered, v)
		}
	}
	out = strings.TrimSpace(string(filtered))
	return
}
