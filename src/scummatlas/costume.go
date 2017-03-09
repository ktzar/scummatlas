package scummatlas

import (
	"fmt"
	"image"
	i "scummatlas/image"
	b "scummatlas/binaryutils"
)

type Costume struct {
	AnimCount       int
	PaletteSize   int
	Mirrored      bool
	Palette       []byte
	Animations    []CostumeAnim
	Limbs         []Limb
	Commands      []byte
	frameOffsets  []int
	animOffsets   []int
	animCmdOffset int
	HexMap
}

type Limb struct {
	Width  byte
	Height byte
	RelX   int
	RelY   int
	MoveX  int
	MoveY  int
	Image  *image.Gray
}

func DecodeLimb(data []byte, offset int, palette []byte)(limb Limb) {
	limb.Width =  data[offset]
	limb.Height = data[offset+1]
	limb.RelX =   b.LE16(data, offset+2)
	limb.RelY =   b.LE16(data, offset+4)
	limb.MoveX =  b.LE16(data, offset+6)
	limb.MoveY =  b.LE16(data, offset+8)

	limb.Image = i.ParseLimb(
		data[offset+10:],
		int(limb.Width / 16),
		int(limb.Height / 16),
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

func NewCostume(data []byte) *Costume {

	c := new(Costume)
	c.data = data
	cursor := 0
	i := 0

	/*processAnimDefinition := func() {
		fmt.Printf("Processing anim %d at %02x\n", i, cursor)
		c.AddSection(cursor, 2, "AnimMask", fmt.Sprintf("anim %d", i))
		limbMask := b.LE16(data, cursor)
		fmt.Println("limbMask", limbMask)
		if limbMask == 0xffff {
			fmt.Println("limbMask return ")
			return
		}
		cursor += 2
		limbs := b.OneBitsInWord(limbMask)

		for j, limb := range limbs {
			fmt.Println("\nLimb ", limb, "\n------")
			c.AddSection(cursor, 2, "AnimIndex", fmt.Sprintf("anim %d - lib %d", i, j))
			index := b.LE16(data, cursor)
			fmt.Printf("index: %04x\n", index)
			cursor += 2

			if index != 0xffff {
				c.AddSection(cursor, 1, "AnimEnd", fmt.Sprintf("anim %d - lib %d", i, j))
				cursor++
			}
		}
	}
	*/

	c.AddSection(cursor, 1, "AnimCount", "")

	c.AnimCount = int(data[cursor]) + 1
	cursor++
	format := data[cursor]
	c.AddSection(cursor, 1, "Format", "")
	cursor++

	if format > 0x61 || format < 0x57 {
		panic("Not a valid costume")
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
		c.Palette = append(c.Palette, data[cursor])
		cursor++
	}

	//TODO anim cmds offset
	c.animCmdOffset = b.LE16(data, cursor)
	c.AddSection(cursor, 2, "AnimCmdOffset", "")
	cursor += 2

	c.AddSection(c.animCmdOffset, 10, "AnimCommands", "")

	//There are always 16 limbs
	for i = 0; i < 16; i++ {
		frameOffset := b.LE16(data, cursor)
		c.frameOffsets = append(c.frameOffsets, frameOffset)
		c.AddSection(cursor, 2, "FrameOffset", fmt.Sprintf("FrameOffset %d", i))
		cursor += 2
	}

	for i = 0; i < c.AnimCount; i++ {
		data := b.LE16(data, cursor)
		c.AddSection(cursor, 2, "AnimOffset", fmt.Sprintf("AnimOffset %d", i))
        if data != 0 {
            c.animOffsets = append(c.animOffsets, data)
            //c.AddSection(data, 2, "AnimMask", fmt.Sprintf("%d", i))
        }
		cursor += 2
	}

    for i, _ := range c.animOffsets {
		c.ProcessCostumeAnim(i)
    }

    /*
    // Calculating offset from the actual position
    // not sure if needed
	for i, _ := range c.animOffsets {
		c.animOffsets[i] += cursor
	}
    */

	/*
		for i, animOffset := range c.animOffsets {
			limbMask := b.LE16(data, animOffset)
			animLength := b.OneBitsInWord(limbMask)
			c.AddSection(
				animOffset,
				2+animLength*3,
				fmt.Sprintf("AnimDefinition%d", i),
				"")
		}
	*/

	//Process anim commands
	//c.AddSection(c.animCmdOffset, c.AnimCount, "AnimCmd", "")
	for i, command := range c.data[c.animCmdOffset : c.frameOffsets[0]] {
        c.Commands = append(c.Commands, command)
        c.AddSection(
            c.animCmdOffset + i, 1,
			"Command",
            fmt.Sprintf("Command %d", i))
    }

	// Process limbs
	for limbNumber, limbOffset := range c.frameOffsets {
		limbOffset &= 0x7fff
		c.AddSection(limbOffset, 2, "LimbOffset", fmt.Sprintf("%d", limbNumber))
		if limbOffset > len(data) {
			fmt.Printf("Something wrong with limb %d\n", limbNumber)
			continue
		}
		imgOffset := b.LE16(data, limbOffset)
		if imgOffset+8 > len(data) {
			fmt.Printf("Something wrong with limb image %d\n", limbNumber)
			continue
		}
		c.AddSection(imgOffset, 1, "Pict", "Width")
		c.AddSection(imgOffset+1, 1, "Pict", "Height")
		c.AddSection(imgOffset+2, 2, "Pict", "RelX")
		c.AddSection(imgOffset+4, 2, "Pict", "RelY")
		c.AddSection(imgOffset+6, 2, "Pict", "MoveX")
		c.AddSection(imgOffset+8, 2, "Pict", "MoveY")
		limb := DecodeLimb(data, imgOffset, c.Palette);
		c.Limbs = append(c.Limbs, limb)
	}

	return c
}
