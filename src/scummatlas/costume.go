package scummatlas

import (
	"fmt"
	"image"
	"image/color"
	i "scummatlas/image"
	b "scummatlas/binaryutils"
)

type Costume struct {
	AnimCount       int
	PaletteSize   int
	Mirrored      bool
	Palette       color.Palette
	Animations    []CostumeAnim
	Limbs         []Limb
	Commands      []byte
	frameOffsets  []int
	animOffsets   []int
	animCmdOffset int
	HexMap
}

type Limb struct {
	Width  int
	Height int
	RelX   int
	RelY   int
	MoveX  int
	MoveY  int
	Image  *image.Paletted
}

func DecodeLimb(data []byte, offset int, palette color.Palette)(limb Limb, length int) {
	limb.Width =  b.LE16(data, offset)
	limb.Height = b.LE16(data, offset+2)
	limb.RelX =   b.LE16(data, offset+4)
	limb.RelY =   b.LE16(data, offset+6)
	limb.MoveX =  b.LE16(data, offset+8)
	limb.MoveY =  b.LE16(data, offset+10)

	if limb.Width > 1024 || limb.Height > 768 {
		fmt.Println("Image too big")
		return
	}
	limb.Image, length = i.ParseLimb(
		data[offset+10:],
		int(limb.Width),
		int(limb.Height),
		palette,
	)
	return
}

type CostumeAnim struct {
	LimbMask    []int
	Definitions []AnimDefinition
}

type AnimDefinition struct {
    start int
    length int
    loop bool
}

func (c Costume) Debug() {
	fmt.Printf("COSTUME\n=======\n")
	fmt.Printf("AnimCount: %v\n", c.AnimCount)
	fmt.Printf("animCmdOffset: 0x%x\n", c.animCmdOffset)
	fmt.Printf("PaletteSize: %v\n", c.PaletteSize)
	fmt.Printf("Mirrored: %v\n", c.Mirrored)
	fmt.Printf("Palette: %v\n", c.Palette)
	fmt.Printf("frameOffsets: %x\n", c.frameOffsets)
	fmt.Printf("animOffsets: %x\n", c.animOffsets)
	fmt.Printf("Animations: %v\n", c.Animations)
	fmt.Printf("Limbs: %v\n", c.Limbs)
	fmt.Printf("Commands: %v\n", c.Commands)
	fmt.Printf("First Image: %v\n", c.Limbs[0].Image)
}

func (c* Costume) ProcessCostumeAnim(i int) {

    cursor := c.animOffsets[i]

    anim := CostumeAnim{}

    fmt.Printf("Processing anim at %02x\n", cursor)

    limbMask := b.LE16(c.data, cursor)

    c.AddSection(cursor, 2, "AnimMask", fmt.Sprintf("anim %d", i))

    fmt.Println("limbMask", limbMask)
    if limbMask == 0xffff {
        fmt.Println("limbMask return ")
        return
    }
    cursor += 2
    anim.LimbMask = b.OneBitsInWord(limbMask)

    for j, limb := range anim.LimbMask {
        fmt.Println("\nLimb ", limb, "\n------")
        c.AddSection(cursor, 2, "AnimIndex", fmt.Sprintf("anim %d - lib %d", i, j))
        index := b.LE16(c.data, cursor)
        length := 0
        loop := false

        fmt.Printf("index: %04x\n", index)
        cursor += 2

        if index != 0xffff {
            length = int(c.data[cursor] & 0x7f)
            loop = c.data[cursor] & 0x80 > 0
            c.AddSection(cursor, 1, "AnimLenLoop", fmt.Sprintf("anim %d - lib %d", i, j))
            cursor++
        }
        anim.Definitions = append(anim.Definitions, AnimDefinition{
            index,
            length,
            loop,
        })
    }
}

func NewCostume(data []byte, roomPalette color.Palette) *Costume {

	c := new(Costume)
	c.data = data
	cursor := 0
	i := 0

	c.AddSection(cursor, 1, "AnimCount", "")

	c.AnimCount = int(data[cursor]) + 1
	cursor++
	format := data[cursor] & 0x7f
	c.AddSection(cursor, 1, "Format", "")
	cursor++

	if format > 0x61 || format < 0x57 {
		fmt.Println("Not a valid costume")
		return c
	}
	c.PaletteSize = 32
	if format&0x01 == 0 {
		c.PaletteSize = 16
	}
	if format&0x80 == 0 {
		c.Mirrored = true
	}

	c.AddSection(cursor, c.PaletteSize, "Palette", "")

	for i = 0; i < c.PaletteSize; i++ {
		c.Palette = append(c.Palette, roomPalette[data[cursor]])
		cursor++
	}

	//TODO anim cmds offset
	c.animCmdOffset = b.LE16(data, cursor) - 6
	c.AddSection(cursor, 2, "AnimCmdOffset", "")
	cursor += 2

	//There are always 16 limbs
	for i = 0; i < 16; i++ {
		frameOffset := b.LE16(data, cursor) -6
		c.frameOffsets = append(c.frameOffsets, frameOffset)
		c.AddSection(cursor, 2, "FrameOffset", fmt.Sprintf("FrameOffset %d", i))
		cursor += 2
	}

	for i = 0; i < c.AnimCount; i++ {
		data := b.LE16(data, cursor)
		c.AddSection(cursor, 2, "AnimOffset", fmt.Sprintf("AnimOffset %d", i))
        if data != 0 {
            c.animOffsets = append(c.animOffsets, data - 6)
            //c.AddSection(data, 2, "AnimMask", fmt.Sprintf("%d", i))
        }
		cursor += 2
	}

    for i, _ := range c.animOffsets {
		c.ProcessCostumeAnim(i)
    }

	for i, command := range c.data[c.animCmdOffset : c.frameOffsets[0]] {
        c.Commands = append(c.Commands, command)
        c.AddSection(
            c.animCmdOffset + i, 1,
			"Command",
            fmt.Sprintf("Command %d", i))
    }

	numPictures := 0
	for cmd := range c.Commands {
		if cmd < 0x71 && cmd > numPictures {
			numPictures = cmd
		}
	}

	fmt.Println("numPictures", numPictures)

	for picNumber := 0 ; picNumber < numPictures ; picNumber ++ {
		limbOffset := c.frameOffsets[0] + picNumber * 2
		fmt.Printf("picNumber %v in %x\n", picNumber, limbOffset)
		c.AddSection(limbOffset, 2, "LimbOffset", fmt.Sprintf("%d", picNumber))
		if limbOffset > len(data) {
			fmt.Printf("Something wrong with limb %d\n", picNumber)
			continue
		}
		imgOffset := b.LE16(data, limbOffset) - 6
		if imgOffset+18 > len(data) || imgOffset < 0 {
			fmt.Printf("Something wrong with limb image %d\n", picNumber)
			continue
		}
		c.AddSection(imgOffset, 2, "Pict", "Width")
		c.AddSection(imgOffset+2, 2, "Pict", "Height")
		c.AddSection(imgOffset+4, 2, "Pict", "RelX")
		c.AddSection(imgOffset+6, 2, "Pict", "RelY")
		c.AddSection(imgOffset+8, 2, "Pict", "MoveX")
		c.AddSection(imgOffset+10, 2, "Pict", "MoveY")
		fmt.Println("Img Offset", imgOffset)
		limb, length := DecodeLimb(data, imgOffset, c.Palette);
		c.AddSection(imgOffset+12, length, "Pict", "RLE")
		c.Limbs = append(c.Limbs, limb)
	}

	return c
}
