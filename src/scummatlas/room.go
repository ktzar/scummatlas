package scummatlas

import (
	"fmt"
)

type Image bool
type Box struct {
	ulx   int
	uly   int
	urx   int
	ury   int
	lrx   int
	lry   int
	llx   int
	lly   int
	mask  byte
	flags byte
	scale int
}
type Script string
type BoxMatrix bool

type Room struct {
	data     []byte
	offset   int
	Width    int
	Height   int
	ObjCount int
	//ColorCycle ColorCycle
	//TranspColor TranspColor
	//Palette Palette
	Image         Image
	ObjectImage   Image
	ObjectScripts []Script
	ExitScript    Script
	EntryScript   Script
	Boxes         []Box
	LocalScripts  []Script
	BoxMatrix     BoxMatrix
}

func NewRoom(data []byte) *Room {
	room := new(Room)
	room.data = data
	room.offset = 0

	blockName := room.getBlockName()
	if blockName != "ROOM" {
		panic("Can't find ROOM")
	}

	room.offset = 8
	for room.offset < len(data) {
		blockName := room.getBlockName()
		fmt.Println("Parsing", blockName)

		switch blockName {
		case "RMHD":
			room.parseRMHD()
		case "BOXD":
			room.parseBOXD()
		case "EXCD":
			room.parseEXCD()
		case "ENCD":
			room.parseENCD()
		case "LSCR":
			room.parseLSCR()
		case "RMIM":
			room.parseRMIM()
		}

		room.nextBlock()
	}

	fmt.Println("New ROOM\n")
	room.Print()
	return room
}

func parseScriptBlock(data []byte) Script {
	fmt.Println("Script size", BE32(data, 4))
	return ""
}

func (r *Room) parseLSCR() {
	script := parseScriptBlock(
		r.data[r.offset : r.offset+r.getBlockSize()])
	r.LocalScripts = append(r.LocalScripts, script)
}

func (r *Room) parseENCD() {
	r.EntryScript = parseScriptBlock(r.data[r.offset : r.offset+r.getBlockSize()])
}

func (r *Room) parseEXCD() {
	r.EntryScript = parseScriptBlock(r.data[r.offset : r.offset+r.getBlockSize()])
}

func (r *Room) parseBOXD() {
	boxCount := LE16(r.data, r.offset+8)
	var boxOffset int
	for i := 0; i < boxCount; i++ {
		boxOffset = 10 + i*20
		box := NewBox(r.data[boxOffset : boxOffset+20])
		r.Boxes = append(r.Boxes, box)
	}
}

func (r *Room) parseRMHD() {
	fmt.Println("RMHD offset", r.offset)
	r.Width = LE16(r.data, r.offset+8)
	r.Height = LE16(r.data, r.offset+10)
	r.ObjCount = LE16(r.data, r.offset+12)
}

func (r Room) Print() {
	fmt.Println("Size: ", r.Width, r.Height)
	fmt.Println("Object count: ", r.ObjCount)
	fmt.Println("Boxes: ", len(r.Boxes))
}

func (r Room) getBlockName() string {
	return string(r.data[r.offset : r.offset+4])
}

func (r Room) getBlockSize() int {
	return BE32(r.data, r.offset+4)
}

func (r *Room) nextBlock() {
	r.offset += r.getBlockSize()
}

func NewBox(data []byte) Box {
	box := new(Box)

	box.ulx = LE16(data, 0)
	box.uly = LE16(data, 2)
	box.urx = LE16(data, 4)
	box.ury = LE16(data, 6)
	box.lrx = LE16(data, 8)
	box.lry = LE16(data, 10)
	box.llx = LE16(data, 12)
	box.lly = LE16(data, 14)
	box.mask = data[16]
	box.flags = data[17]
	box.scale = LE16(data, 18)

	return *box
}
