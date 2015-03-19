package scummatlas

import (
	"fmt"
	"image"
	"strings"
)

type Object struct {
	Id     int
	Name   string
	Flags  uint8
	Parent uint8
	Script Script
	Image  ObjectImage
	X      int
	Y      int
	Width  int
	Height int
	//TODO Direction uint8
}

type ObjectImage struct {
	X        int
	Y        int
	Width    int
	Height   int
	States   int
	Planes   int
	Hotspots int
	Image    *image.RGBA
}

func (self Object) IdHex() string {
	return fmt.Sprintf("%x", self.Id)
}

func NewObjectImageFromOBIM(data []byte, r *Room) (objImg ObjectImage, id int) {
	headerName := FourCharString(data, 8)
	if headerName != "IMHD" {
		panic("Image header not present")
	}
	headerSize := BE32(data, 12)
	header := data[16 : 16+headerSize-8]

	id = LE16(header, 0)

	objImg = ObjectImage{
		States: LE16(header, 2),
		Planes: LE16(header, 4),
		X:      LE16(header, 8),
		Y:      LE16(header, 10),
		Width:  LE16(header, 12),
		Height: LE16(header, 14),
	}

	if objImg.States > 0 {
		//TODO parse next block IM01
		currentState := 1
		imageOffset := 8 + headerSize
		//TODO parse all states for currentState <= objImg.States {}
		expectedHeader := fmt.Sprintf("IM%02d", currentState)
		if FourCharString(data, imageOffset) != expectedHeader {
			panic("Not " + expectedHeader + " found!, found " + FourCharString(data, imageOffset) + " instead")
		}
		imageSize := BE32(data, imageOffset+4)

		img := parseImage(data[imageOffset:imageOffset+imageSize], objImg.Planes, objImg.Width, objImg.Height, r.Palette, r.TranspIndex)
		objImg.Image = img
	}

	return
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
