package scummatlas

import (
	_ "bufio"
	"fmt"
	goimage "image"
	"image/color"
	b "scummatlas/binaryutils"
	"scummatlas/image"
)

type BoxMatrix bool

type Room struct {
	data         []byte
	offset       int
	Width        int
	Height       int
	ObjCount     int
	TranspIndex  uint8
	Palette      color.Palette
	Image        *goimage.RGBA
	Boxes        []Box
	BoxMatrix    BoxMatrix
	ExitScript   Script
	EntryScript  Script
	LocalScripts map[int]Script
	Objects      map[int]Object
	//ColorCycle ColorCycle
}

func NewRoom(data []byte) *Room {
	room := new(Room)
	room.data = data
	room.offset = 0
	room.Objects = make(map[int]Object)
	room.LocalScripts = make(map[int]Script)

	blockName := room.getBlockName()
	if blockName != "ROOM" {
		panic("Can't find ROOM, found " + blockName + " instead")
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
		case "OBCD":
			room.parseOBCD()
		case "OBIM":
			room.parseOBIM()
		case "RMIM":
			room.parseRMIM()
		case "TRNS":
			room.parseTRNS()
		}
		room.nextBlock()
	}

	return room
}

func (r *Room) parseLSCR() {
	scriptId := int(r.data[r.offset+8])
	scriptBlock := r.data[r.offset+9 : r.offset+r.getBlockSize()]
	script := parseScriptBlock(scriptBlock)
	r.LocalScripts[scriptId] = script
	if len(script) == 0 {
		log("script", "DUMP from %x\n", r.offset+9)
		fmt.Printf("%x", scriptBlock)
	}
	log("script", "\nLocal ScriptID 0x%02x, size %d, script %v\n", scriptId, r.getBlockSize(), script)
}

func (r *Room) parseTRNS() {
	r.TranspIndex = r.data[r.offset+8]
	fmt.Println("Transparent index", r.TranspIndex)
}

func (r *Room) parseOBIM() {
	blockSize := b.BE32(r.data, r.offset+4)
	objImg, id := NewObjectImageFromOBIM(r.data[r.offset:r.offset+blockSize], r)
	log("image", "======================\nObject with id 0x%02X\n%+v\n", id, objImg)

	existing, ok := r.Objects[id]
	if !ok {
		existing = Object{Id: id}
	}
	existing.Image = objImg
	r.Objects[id] = existing
}

func (r *Room) parseOBCD() {
	blockSize := b.BE32(r.data, r.offset+4)
	object := NewObjectFromOBCD(r.data[r.offset : r.offset+blockSize])

	existingObject, ok := r.Objects[object.Id]
	if ok {
		object.Image = existingObject.Image
	}
	r.Objects[object.Id] = object
}

func (r *Room) parseENCD() {
	r.EntryScript = parseScriptBlock(r.data[r.offset+8 : r.offset+r.getBlockSize()])
}

func (r *Room) parseEPAL() {
	log("palette", "EGA palette, not used\n")
	return
	//TODO REMOVE
	paletteData := r.data[r.offset+8 : r.offset+r.getBlockSize()]
	fmt.Println("Palette data size ", len(paletteData))

	r.Palette = image.ParsePalette(r.data[r.offset+8 : r.offset+8+3*256])
	fmt.Println("Palette length", len(r.Palette))
	fmt.Println(r.Palette)

}

func (r *Room) parseCLUT() {
	paletteData := r.data[r.offset+8 : r.offset+r.getBlockSize()]
	log("palette", "\nPalette data size ", len(paletteData))

	r.Palette = image.ParsePalette(r.data[r.offset+8 : r.offset+8+3*256])
	log("palette", "\nPalette length", len(r.Palette))

	for _, color := range r.Palette {
		r, g, b, _ := color.RGBA()
		r8, g8, b8 := uint8(r), uint8(g), uint8(b)
		log("palette", " %x%x%x", r8, g8, b8)
	}
	log("palette", "\n")
}

func (r *Room) parseEXCD() {
	r.ExitScript = parseScriptBlock(r.data[r.offset+8 : r.offset+r.getBlockSize()])
}

func (r *Room) parseRMIM() {
	if string(r.data[r.offset+8:r.offset+12]) != "RMIH" {
		panic("Not room image header")
	}
	headerSize := b.BE32(r.data, r.offset+12)
	zBuffers := b.LE16(r.data, r.offset+16)
	log("image", "headerSize", headerSize)
	log("image", "zBuffers", zBuffers)

	if b.FourCharString(r.data, r.offset+18) != "IM00" {
		panic("Not room image found")
	}
	imageOffset := r.offset + 18
	imageSize := b.BE32(r.data, imageOffset+4)
	log("image", b.FourCharString(r.data, imageOffset), imageSize)

	r.Image = image.ParseImage(
		r.data[imageOffset:imageOffset+4+imageSize],
		zBuffers,
		r.Width,
		r.Height,
		r.Palette,
		r.TranspIndex,
		false)
}

func (r *Room) parseBOXD() {
	boxCount := b.LE16(r.data, r.offset+8)
	var boxOffset int
	log("box", "BOXCOUNT", boxCount)
	for i := 0; i < boxCount; i++ {
		boxOffset = r.offset + 10 + i*20
		box := NewBox(r.data[boxOffset : boxOffset+20])
		r.Boxes = append(r.Boxes, box)
	}
}

func (r *Room) parseRMHD() {
	log("room", "RMHD offset", r.offset)
	r.Width = b.LE16(r.data, r.offset+8)
	r.Height = b.LE16(r.data, r.offset+10)
	r.ObjCount = b.LE16(r.data, r.offset+12)
	log("room", "Room size %vx%v\n", r.Width, r.Height)
}

func (r Room) Print() {
	fmt.Println("Size: ", r.Width, r.Height)
	fmt.Println("Object count: ", r.ObjCount)
	fmt.Println("Boxes: ", len(r.Boxes))
}

func (r Room) getBlockName() string {
	return b.FourCharString(r.data, r.offset)
}

func (r Room) getBlockSize() int {
	return b.BE32(r.data, r.offset+4)
}

func (r *Room) nextBlock() {
	r.offset += r.getBlockSize()
}

func NewBox(data []byte) Box {
	box := new(Box)

	box.ulx = b.LE16(data, 0)
	box.uly = b.LE16(data, 2)
	box.urx = b.LE16(data, 4)
	box.ury = b.LE16(data, 6)
	box.lrx = b.LE16(data, 8)
	box.lry = b.LE16(data, 10)
	box.llx = b.LE16(data, 12)
	box.lly = b.LE16(data, 14)
	box.mask = data[16]
	box.flags = data[17]
	box.scale = b.LE16(data, 18)

	return *box
}
