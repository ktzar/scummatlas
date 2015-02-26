package scummatlas

import (
	"fmt"
)

type Image bool
type Script bool
type Box bool
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
	LocalScript   Script
	BoxData       []Box
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
		}

		room.nextBlock()
	}

	fmt.Println("New ROOM\n")
	room.Print()
	return room
}

func (r *Room) parseBOXD() {
	r.BoxData = append(r.BoxData, true)
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
}

func (r Room) getBlockName() string {
	return string(r.data[r.offset : r.offset+4])
}

func (r *Room) nextBlock() {
	blockSize := BE32(r.data, r.offset+4)
	r.offset += blockSize
}
