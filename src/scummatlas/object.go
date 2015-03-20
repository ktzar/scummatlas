package scummatlas

import (
	"fmt"
	goimage "image"
	b "scummatlas/binaryutils"
	"scummatlas/image"
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
	Frames   []*goimage.RGBA
}

func (self ObjectImage) FramesIndexes() (out []string) {
	for i := 0; i < len(self.Frames); i++ {
		out = append(out, fmt.Sprintf("%02d", i))
	}
	return
}

func (self Object) IdHex() string {
	return fmt.Sprintf("%x", self.Id)
}

func NewObjectImageFromOBIM(data []byte, r *Room) (objImg ObjectImage, id int) {
	headerName := b.FourCharString(data, 8)
	if headerName != "IMHD" {
		panic("Image header not present")
	}
	headerSize := b.BE32(data, 12)
	header := data[16 : 16+headerSize-8]

	id = b.LE16(header, 0)

	objImg = ObjectImage{
		States: b.LE16(header, 2),
		Planes: b.LE16(header, 4),
		X:      b.LE16(header, 8),
		Y:      b.LE16(header, 10),
		Width:  b.LE16(header, 12),
		Height: b.LE16(header, 14),
	}

	if objImg.States > 0 {
		imageOffset := 8 + headerSize

		for state := 1; state <= objImg.States; state++ {
			expectedHeader := imageStateHeader(state)
			if b.FourCharString(data, imageOffset) != expectedHeader {
				panic("Not " + expectedHeader + " found!, found " + b.FourCharString(data, imageOffset) + " instead")
			}
			imageSize := b.BE32(data, imageOffset+4)

			log := false
			img := image.ParseImage(data[imageOffset:imageOffset+imageSize], objImg.Planes, objImg.Width, objImg.Height, r.Palette, r.TranspIndex, log)
			objImg.Frames = append(objImg.Frames, img)
			imageOffset += imageSize
		}

	}

	return
}

func imageStateHeader(state int) string {
	return fmt.Sprintf("IM%02X", state)
}

func NewObjectFromOBCD(data []byte) Object {
	headerOffset := 8
	if b.FourCharString(data, headerOffset) != "CDHD" {
		panic("No object header")
	}
	headerSize := b.BE32(data, headerOffset+4)

	intInOffsetTimesEight := func(offset int) int {
		return int(data[headerOffset+offset]) * 8
	}
	obj := Object{
		Id:     b.LE16(data, headerOffset+8),
		X:      intInOffsetTimesEight(10),
		Y:      intInOffsetTimesEight(11),
		Width:  intInOffsetTimesEight(12),
		Height: intInOffsetTimesEight(13),
		Flags:  data[headerOffset+14],
		Parent: data[headerOffset+15],
	}

	verbOffset := headerOffset + headerSize
	if b.FourCharString(data, verbOffset) != "VERB" {
		panic("Object with no verbs")
	}
	verbSize := b.BE32(data, verbOffset+4)
	objNameOffset := verbOffset + verbSize
	if b.FourCharString(data, objNameOffset) != "OBNA" {
		panic("Object with no name")
	}
	objNameSize := b.BE32(data, objNameOffset+4)
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
