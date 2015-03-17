package scummatlas

import (
	_ "bufio"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

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

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (self Box) Corners() [4]Point {
	ll := Point{self.llx, self.lly}
	ul := Point{self.ulx, self.uly}
	lr := Point{self.lrx, self.lry}
	ur := Point{self.urx, self.ury}
	return [4]Point{ul, ll, lr, ur}
}

type BoxMatrix bool

type Room struct {
	data     []byte
	offset   int
	Width    int
	Height   int
	ObjCount int
	//ColorCycle ColorCycle
	TranspIndex   int
	Palette       color.Palette
	Image         *image.RGBA
	Objects       []Object
	ObjectImage   image.Paletted
	ObjectScripts []Script
	ExitScript    Script
	EntryScript   Script
	Boxes         []Box
	LocalScripts  []Script
	BoxMatrix     BoxMatrix
}

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

func NewRoom(data []byte) *Room {
	room := new(Room)
	room.data = data
	room.offset = 0

	blockName := room.getBlockName()
	if blockName != "ROOM" {
		panic("Can't find ROOM")
	}

	room.offset = 8
	fmt.Printf("Block Name\tBlock Size\n=============\n")
	for room.offset < len(data) {
		blockName := room.getBlockName()
		fmt.Printf("%v\t%v bytes\n", blockName, room.getBlockSize())

		switch blockName {
		case "RMHD":
			room.parseRMHD()
		case "BOXD":
			room.parseBOXD()
		case "EXCD":
			room.parseEXCD()
		case "ENCD":
			room.parseENCD()
		case "EPAL":
			room.parseEPAL()
		case "CLUT":
			log.SetOutput(ioutil.Discard)
			room.parseCLUT()
		case "LSCR":
			room.parseLSCR()
		case "OBCD":
			room.parseOBCD()
		case "RMIM":
			room.parseRMIM()
		case "TRNS":
			room.parseTRNS()
		}
		log.SetOutput(os.Stdout)
		room.nextBlock()
	}

	fmt.Println("New ROOM\n")
	room.Print()
	return room
}

func (r *Room) parseLSCR() {
	scriptId := int(r.data[r.offset+8])
	fmt.Println("ScriptID", scriptId)
	script := parseScriptBlock(
		r.data[r.offset+9 : r.offset+r.getBlockSize()])
	r.LocalScripts = append(r.LocalScripts, script)
}

func (r *Room) parseTRNS() {
	r.TranspIndex = int(r.data[r.offset+8])
	fmt.Println("Transparent index", r.TranspIndex)
}

func (r *Room) parseOBCD() {
	var obj Object
	headerOffset := r.offset + 8
	if FourCharString(r.data, headerOffset) != "CDHD" {
		panic("NOOO")
	}
	headerSize := BE32(r.data, headerOffset+4)
	fmt.Println("Header size", headerSize)
	obj.Id = LE16(r.data, headerOffset+8)
	obj.X = LE16(r.data, headerOffset+10)
	obj.Y = LE16(r.data, headerOffset+12)
	obj.Width = LE16(r.data, headerOffset+14)
	obj.Height = LE16(r.data, headerOffset+16)
	obj.Flags = r.data[headerOffset+18]
	obj.Parent = r.data[headerOffset+19]

	verbOffset := headerOffset + headerSize
	if FourCharString(r.data, verbOffset) != "VERB" {
		panic("Object with no verbs")
	}
	verbSize := BE32(r.data, verbOffset+4)
	objNameOffset := verbOffset + verbSize
	if FourCharString(r.data, objNameOffset) != "OBNA" {
		panic("Object with no name")
	}
	objNameSize := BE32(r.data, objNameOffset+4)
	obj.Name = ""
	name := r.data[objNameOffset+4 : objNameOffset+objNameSize]
	filtered := []byte{}
	for _, v := range name {
		if v != 0x40 && v != 0x00 && v != 0x0f {
			filtered = append(filtered, v)
		}
	}
	obj.Name = strings.TrimSpace(string(filtered))
	r.Objects = append(r.Objects, obj)
}

func (r *Room) parseENCD() {
	r.EntryScript = parseScriptBlock(r.data[r.offset+8 : r.offset+r.getBlockSize()])
}

func (r *Room) parseEPAL() {
	fmt.Println("EGA palette, not used")
	/*
		paletteData := r.data[r.offset+8 : r.offset+r.getBlockSize()]
		fmt.Println("Palette data size ", len(paletteData))

		r.Palette = parsePalette(r.data[r.offset+8 : r.offset+8+3*256])
		fmt.Println("Palette length", len(r.Palette))
		fmt.Println(r.Palette)
	*/

}

func (r *Room) parseCLUT() {
	paletteData := r.data[r.offset+8 : r.offset+r.getBlockSize()]
	log.Println("Palette data size ", len(paletteData))

	r.Palette = parsePalette(r.data[r.offset+8 : r.offset+8+3*256])
	log.Println("Palette length", len(r.Palette))

	for _, color := range r.Palette {
		r, g, b, _ := color.RGBA()
		r8, g8, b8 := uint8(r), uint8(g), uint8(b)
		log.Printf(" %x%x%x", r8, g8, b8)
	}
	log.Println()
}

func (r *Room) parseEXCD() {
	r.EntryScript = parseScriptBlock(r.data[r.offset+8 : r.offset+r.getBlockSize()])
}

func (r *Room) parseRMIM() {
	if string(r.data[r.offset+8:r.offset+12]) != "RMIH" {
		panic("Not room image header")
	}
	headerSize := BE32(r.data, r.offset+12)
	zBuffers := LE16(r.data, r.offset+16)
	fmt.Println("headerSize", headerSize)
	fmt.Println("zBuffers", zBuffers)

	if FourCharString(r.data, r.offset+18) != "IM00" {
		panic("Not room image found")
	}
	imageOffset := r.offset + 18
	imageSize := BE32(r.data, imageOffset+4)
	fmt.Println(FourCharString(r.data, imageOffset), imageSize)

	r.Image = parseImage(
		r.data[imageOffset:imageOffset+4+imageSize],
		zBuffers,
		r.Width,
		r.Height,
		r.Palette,
		uint8(r.TranspIndex))

}

func (r *Room) parseBOXD() {
	boxCount := LE16(r.data, r.offset+8)
	var boxOffset int
	fmt.Println("BOXCOUNT", boxCount)
	for i := 0; i < boxCount; i++ {
		boxOffset = r.offset + 10 + i*20
		box := NewBox(r.data[boxOffset : boxOffset+20])
		r.Boxes = append(r.Boxes, box)
	}
}

func (r *Room) parseRMHD() {
	fmt.Println("RMHD offset", r.offset)
	r.Width = LE16(r.data, r.offset+8)
	r.Height = LE16(r.data, r.offset+10)
	r.ObjCount = LE16(r.data, r.offset+12)
	fmt.Printf("Room size %vx%v\n", r.Width, r.Height)
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
