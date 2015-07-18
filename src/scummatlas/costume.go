package scummatlas

import (
	"fmt"
	"image"
	b "scummatlas/binaryutils"
)

type Costume struct {
	NumAnim       int
	PaletteSize   int
	Mirrored      bool
	Palette       []byte
	Animations    []interface{}
	Limbs         []Limb
	Commands      []interface{}
	limbOffsets   []int
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
	Image  *image.RGBA
}

type CostumeAnim struct {
	LimbMask    [16]bool
	Definitions []interface{}
}

func (c Costume) Debug() {
	fmt.Printf("COSTUME\n=======\n")
	fmt.Printf("NumAnim: %v\n", c.NumAnim)
	fmt.Printf("animCmdOffset: 0x%x\n", c.animCmdOffset)
	fmt.Printf("PaletteSize: %v\n", c.PaletteSize)
	fmt.Printf("Mirrored: %v\n", c.Mirrored)
	fmt.Printf("Palette: %v\n", c.Palette)
	fmt.Printf("limbOffsets: %v\n", c.limbOffsets)
	fmt.Printf("animOffsets: %v\n", c.animOffsets)
	fmt.Printf("Animations: %v\n", c.Animations)
	fmt.Printf("Limbs: %v\n", c.Limbs)
	fmt.Printf("Commands: %v\n", c.Commands)
}

func NewCostume(data []byte) *Costume {
	c := new(Costume)
	c.data = data

	cursor := 0

	c.AddSection(cursor, 1, "NumAnim", "")

	c.NumAnim = int(data[cursor])
	cursor++
	format := data[cursor]
	c.AddSection(cursor, 1, "Format", "")
	cursor++

	c.PaletteSize = 32
	if format&0x01 == 0 {
		c.PaletteSize = 16
	}
	if format&0x80 == 0 {
		c.Mirrored = true
	}

	c.AddSection(cursor, c.PaletteSize, "Palette", "")

	for i := 0; i < c.PaletteSize; i++ {
		c.Palette = append(c.Palette, data[cursor])
		cursor++
	}

	//TODO anim cmds offset
	c.animCmdOffset = b.LE16(data, cursor)
	c.AddSection(cursor, 2, "AnimCmdOffset", "")
	cursor += 2

	//There are always 16 limbs
	c.AddSection(cursor, 32, "LimbsOff", "")
	for i := 0; i < 16; i++ {
		c.limbOffsets = append(c.limbOffsets, b.LE16(data, cursor))
		cursor += 2
	}

	c.AddSection(cursor, 2*c.NumAnim, "DefinitionOff", "")
	for i := 0; i < c.NumAnim; i++ {
		c.animOffsets = append(c.animOffsets, b.LE16(data, cursor))
		cursor += 2
	}

	for i := 0; i < c.NumAnim; i++ {
		c.AddSection(cursor, 2, "AnimMask", "")
		limbMask := b.LE16(data, cursor)
		cursor += 2
		animLength := b.OneBitsInWord(limbMask)

		for j := 0; j < animLength; j++ {
			c.AddSection(cursor, 2, "AnimIndex", "")
			index := b.LE16(data, cursor)
			cursor += 2
			if index != 0xffff {
				c.AddSection(cursor, 1, "AnimLength", "")
				cursor++
				//length := int(data[cursor+4] & 0x7f)

				//fmt.Printf("Animation %v has commands in %x -> %x\n", i, index, index+length)
				//if c.animCmdOffset+index+length < len(data) {
				//c.AddSection(c.animCmdOffset+index, length+1, "Command", "")
				//}
				//cursor += 5
			}
		}

	}

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
	c.AddSection(c.animCmdOffset, c.NumAnim, "AnimCmd", "")

	// Process limbs
	for limbNumber, limbOffset := range c.limbOffsets {
		//fmt.Printf("limbOffset: %v > %v\n", limbOffset, len(data))
		if limbOffset > len(data) {
			fmt.Printf("Something wrong with limb %d\n", limbNumber)
			continue
		}
		//c.AddSection(limbOffset, 1, "LimbOffset", "")
		//fmt.Printf("%x\n", data[limbOffset:limbOffset+30])
		imgOffset := limbOffset
		/*
			imgOffset := b.LE16(data, limbOffset)
			fmt.Printf("imgOffset: %v > %v\n", imgOffset, len(data))
			if imgOffset+8 > len(data) {
				fmt.Printf("Something wrong with limb image %d\n", limbNumber)
				continue
			}
		*/
		c.AddSection(imgOffset, 12, "Limb", "")
		limb := Limb{
			Width:  b.LE16(data, imgOffset),
			Height: b.LE16(data, imgOffset+2),
			RelX:   b.LE16(data, imgOffset+4),
			RelY:   b.LE16(data, imgOffset+6),
			MoveX:  b.LE16(data, imgOffset+8),
			MoveY:  b.LE16(data, imgOffset+10),
		}
		c.Limbs = append(c.Limbs, limb)
	}

	return c
}
