package scummatlas

import (
	_ "bufio"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	_ "os"
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

type BoxMatrix bool

type Room struct {
	data     []byte
	offset   int
	Width    int
	Height   int
	ObjCount int
	//ColorCycle ColorCycle
	//TranspColor TranspColor
	Palette       color.Palette
	Image         *image.RGBA
	ObjectImage   image.Paletted
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
			room.parseCLUT()
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

func (r *Room) parseLSCR() {
	scriptId := int(r.data[r.offset+8])
	fmt.Println("ScriptID", scriptId)
	script := parseScriptBlock(
		r.data[r.offset+9 : r.offset+r.getBlockSize()])
	r.LocalScripts = append(r.LocalScripts, script)
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
	fmt.Println("Palette data size ", len(paletteData))

	r.Palette = parsePalette(r.data[r.offset+8 : r.offset+8+3*256])
	fmt.Println("Palette length", len(r.Palette))

	for _, color := range r.Palette {
		r, g, b, _ := color.RGBA()
		r8, g8, b8 := uint8(r), uint8(g), uint8(b)
		fmt.Printf(" %x%x%x", r8, g8, b8)
	}
	fmt.Println()

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

	image := parseImage(r.data[imageOffset:imageOffset+4+imageSize], zBuffers, r.Width, r.Height, r.Palette)

	/*
		f, err := os.Create("image.png")
		if err != nil {
			panic("Error creating image.png")
		}
		w := bufio.NewWriter(f)
		png.Encode(w, image)
		w.Flush()
	*/

	r.Image = image
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
