package scummatlas

import (
	"fmt"
	goimage "image"
	b "scummatlas/binaryutils"
	l "scummatlas/condlog"
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
	Verbs  []Verb
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

type Verb struct {
	code   uint8
	Name   string
	offset int
	Script Script
}

func (self Verb) PrintScript() string {
	return self.Script.Print()
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

func (self Object) PrintVerbs() {
	if len(self.Verbs) > 0 {
		l.Log("object", "Verbs for obj %x", self.Id)
	}
	for _, verb := range self.Verbs {
		l.Log("object", "  -> %v (%02x) : %v", verb.Name, verb.code, verb.Script)
	}
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

			img := image.ParseImage(data[imageOffset:imageOffset+imageSize], objImg.Planes, objImg.Width, objImg.Height, r.Palette, r.TranspIndex)
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

	obj.Verbs = parseVerbBlock(data[verbOffset : verbOffset+verbSize])

	objNameOffset := verbOffset + verbSize
	if b.FourCharString(data, objNameOffset) != "OBNA" {
		panic("Object with no name")
	}
	objNameSize := b.BE32(data, objNameOffset+4)
	name := data[objNameOffset+4 : objNameOffset+objNameSize]
	obj.Name = filterObjectName(name)
	return obj
}

var objCount int
var verbCount int

func parseVerbBlock(data []byte) (out []Verb) {
	//dumpBlock(fmt.Sprintf("VERB_%d", verbCount), data)
	verbCount++
	currentOffset := 8
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("ERROR: ", err)
		}
	}()
	for currentOffset <= len(data) {
		if data[currentOffset] == 0x00 {
			return
		}
		verb := Verb{
			code:   data[currentOffset],
			Name:   getVerbName(data[currentOffset]),
			offset: b.LE16(data, currentOffset+1),
		}

		parser := ScriptParser{
			data:   data,
			offset: verb.offset,
		}
		var err error
		stop := false
		ranOpcode := ""
		for ranOpcode != "stopObjectCode" && stop == false {
			ranOpcode, err = parser.parseNext()
			if err != nil {
				stop = true
			}
		}
		verb.Script = parser.script

		scriptLength := len(verb.Script)
		if scriptLength > 0 &&
			verb.Script[scriptLength-1] == "stopObjectCode()" {
			verb.Script = verb.Script[:scriptLength-1]
		}

		out = append(out, verb)
		currentOffset += 3
	}
	return
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

func getVerbName(code uint8) (name string) {

	verbNames := map[uint8]string{
		2:    "Close",
		3:    "Open",
		0x5a: "Go to",
		5:    "Pull",
		6:    "Push",
		7:    "Use",
		8:    "Look",
		9:    "Pick up",
		0xa:  "Talk to",
	}

	name = verbNames[code]
	if name == "" {
		name = fmt.Sprintf("0x%x", code)
	}
	return
}
