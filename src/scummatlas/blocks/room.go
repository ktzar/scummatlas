package blocks

import (
	"fmt"
	"image"
	"image/color"
	"os"
	b "scummatlas/binaryutils"
	l "scummatlas/condlog"
	i "scummatlas/image"
	s "scummatlas/script"
	"strings"
)

type BoxMatrix bool

type Exit struct {
	Path string
	Room int
}

type Room struct {
	data         []byte
	offset       int
	Id           int
	Name         string
	Width        int
	Height       int
	ObjCount     int
	TranspIndex  uint8
	Palette      color.Palette
	Image        *image.Paletted
	Zplanes      []*image.RGBA
	Boxes        []Box
	BoxMatrix    BoxMatrix
	ExitScript   s.Script
	EntryScript  s.Script
	LocalScripts map[int]s.Script
	Objects      map[int]Object
	//ColorCycle ColorCycle
}

func (r Room) Exits() (exits []Exit) {
	for _, object := range r.Objects {
		for _, verb := range object.Verbs {
			properties := verb.Script.Properties()
			if properties.HasExit && properties.ExitTo != r.Id {
				exits = append(exits, Exit{
					strings.Title(object.Name),
					properties.ExitTo})
			}
		}
	}
	return
}

func (r Room) TwoDigitNumber() string {
	return fmt.Sprintf("%02d", r.Id)
}

func (r Room) PaletteLength() int {
	return len(r.Palette)
}

func (r Room) BoxCount() int {
	return len(r.Boxes)
}

func (r Room) PaletteHex() []string {
	hexes := make([]string, len(r.Palette))
	for i, c := range r.Palette {
		r, g, b, _ := c.RGBA()
		hexes[i] = fmt.Sprintf("%02x%02x%02x", r & 0x00ff, g & 0xff, b & 0xff)
	}
	return hexes
}

func (r Room) LocalScriptCount() int {
	return len(r.LocalScripts)
}

func NewRoom(data []byte) *Room {
	room := new(Room)
	room.data = data
	room.offset = 0
	room.Objects = make(map[int]Object)
	room.LocalScripts = make(map[int]s.Script)

	blockName := room.getBlockName()
	if blockName != "ROOM" {
		panic("Can't find ROOM, found " + blockName + " instead")
	}

	room.offset = 8

	l.Log("structure", "Block Name\tBlock Size\n=============")
	for room.offset < len(data) {
		blockName := room.getBlockName()
		l.Log("structure", "%v\t%v bytes", blockName, room.getBlockSize())

		switch blockName {
		case "RMHD":
			room.parseRMHD()
		case "BOXD":
			room.parseBOXD()
		case "EXCD":
			room.parseEXCD()
		case "ENCD":
			room.parseENCD()
		case "PALS":
			room.parsePALS()
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
	l.Log("script", "\nLocal ScriptID %02x, size %d", scriptId, r.getBlockSize())
	DumpBlock("LSCR_"+fmt.Sprintf("%d", scriptId),
		r.data[r.offset:r.offset+r.getBlockSize()])
	script := s.ParseScriptBlock(scriptBlock)

	r.LocalScripts[scriptId] = script
	if len(script) == 0 {
		l.Log("script", "DUMP from %x", r.offset+9)
		l.Log("script", "%x", scriptBlock)
	}
	l.Log("script", "\nScript: %v", script)
}

func (r *Room) parseENCD() {
	l.Log("script", "\nENCD")
	r.EntryScript = s.ParseScriptBlock(r.data[r.offset+8 : r.offset+r.getBlockSize()])
	DumpBlock("ENCD", r.data[r.offset:r.offset+r.getBlockSize()])
}

func (r *Room) parseEXCD() {
	l.Log("script", "\nEXCD")
	r.ExitScript = s.ParseScriptBlock(r.data[r.offset+8 : r.offset+r.getBlockSize()])
	DumpBlock("EXCD", r.data[r.offset:r.offset+r.getBlockSize()])
}

func DumpBlock(name string, data []byte) {
	f, _ := os.Create("./out/" + name + ".dump")
	defer f.Close()
	f.Write(data)
}

func (r *Room) parseTRNS() {
	r.TranspIndex = r.data[r.offset+8]
	l.Log("image", "Transparent index", r.TranspIndex)
}

func (r *Room) parseOBIM() {
	blockSize := b.BE32(r.data, r.offset+4)
	objImg, id := NewObjectImageFromOBIM(r.data[r.offset:r.offset+blockSize], r)
	l.Log("image", "======================\nObject with id 0x%02X\n%+v", id, objImg)

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

func (r *Room) parsePALS() {
	offName := b.FourCharString(r.data, r.offset+16)
	if offName != "OFFS" {
		l.Log("palette", "Wrong PALS structure. Couldn't find OFFS")
		return
	}
	offSize := b.BE32(r.data, r.offset+20)

	aPalName := b.FourCharString(r.data, r.offset+16+offSize)

	if aPalName != "APAL" {
		l.Log("palette", "Wrong PALS structure. Couldn't find APAL")
		return
	}
	paletteOffset := r.offset + 16 + offSize + 8

	r.Palette = i.ParsePalette(r.data[paletteOffset : paletteOffset+3*256])
	l.Log("palette", "Palette length %d", len(r.Palette))
}

func (r *Room) parseEPAL() {
	l.Log("palette", "EGA palette, not used")
	//TODO Implement?
	return
}

func (r *Room) parseCLUT() {
	paletteData := r.data[r.offset+8 : r.offset+r.getBlockSize()]
	l.Log("palette", "Palette data size ", len(paletteData))

	r.Palette = i.ParsePalette(r.data[r.offset+8 : r.offset+8+3*256])
	l.Log("palette", "Palette length", len(r.Palette))

	for _, color := range r.Palette {
		r, g, b, _ := color.RGBA()
		r8, g8, b8 := uint8(r), uint8(g), uint8(b)
		l.Log("palette", " %x%x%x", r8, g8, b8)
	}
	l.Log("palette", "")
}

func (r *Room) parseRMIM() {
	if b.FourCharString(r.data, r.offset+8) != "RMIH" {
		panic("Not room image header")
	}
	headerSize := b.BE32(r.data, r.offset+12)
	zBuffers := b.LE16(r.data, r.offset+16)
	l.Log("image", "headerSize", headerSize)
	l.Log("image", "zBuffers", zBuffers)

	if b.FourCharString(r.data, r.offset+18) != "IM00" {
		panic("Not room image found")
	}
	imageOffset := r.offset + 18
	imageSize := b.BE32(r.data, imageOffset+4)
	l.Log("image", b.FourCharString(r.data, imageOffset), imageSize)

	image, zplanes := i.ParseImage(
		r.data[imageOffset:imageOffset+4+imageSize],
		zBuffers,
		r.Width,
		r.Height,
		r.Palette,
		r.TranspIndex)

	r.Image = image
	r.Zplanes = zplanes
}

func (r *Room) parseBOXD() {
	boxCount := b.LE16(r.data, r.offset+8)
	var boxOffset int
	l.Log("box", "BOXCOUNT", boxCount)
	for i := 0; i < boxCount; i++ {
		boxOffset = r.offset + 10 + i*20
		box := NewBox(r.data[boxOffset : boxOffset+20])
		r.Boxes = append(r.Boxes, box)
	}
}

func (r *Room) parseRMHD() {
	l.Log("room", "RMHD offset", r.offset)
	r.Width = b.LE16(r.data, r.offset+8)
	r.Height = b.LE16(r.data, r.offset+10)
	r.ObjCount = b.LE16(r.data, r.offset+12)
	l.Log("room", "Room size %vx%v", r.Width, r.Height)
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
