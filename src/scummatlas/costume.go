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
	i := 0

	processAnimDefinition := func() {
		return
		c.AddSection(cursor, 2, "AnimMask", fmt.Sprintf("anim %d", i))
		limbMask := b.LE16(data, cursor)
		cursor += 2
		limbs := b.OneBitsInWord(limbMask)
		fmt.Println(limbs)

		for j, limb := range limbs {
			fmt.Println("\nLimb ", limb, "\n------")
			c.AddSection(cursor, 2, "AnimIndex", fmt.Sprintf("anim %d - lib %d", i, j))
			index := b.LE16(data, cursor)
			fmt.Printf("index: %04x\n", index)
			cursor += 2

			if index != 0xffff {
				length := int(data[cursor] & 0x7f)
				fmt.Printf("length: %d\n", length)
				c.AddSection(cursor, 1, "AnimLength", fmt.Sprintf("anim %d - lib %d", i, j))
				cursor++

				//loop := index & 0x8000 > 0
				index = index & 0x7fff
				//cmdOffset := int(data[c.animCmdOffset+index*2])
				//fmt.Printf("cmdOffset: %d\n", cmdOffset)
				//c.AddSection(c.animCmdOffset+cmdOffset, length+1, "Command", "")

				//fmt.Printf("Animation %v has commands in %x -> %x\n", i, index, index+length)
				//if c.animCmdOffset+index+length < len(data) {
				//c.AddSection(c.animCmdOffset+index, length+1, "Command", "")
				//}
				//cursor += 5
			}
		}
	}

	c.AddSection(cursor, 1, "NumAnim", "")

	c.NumAnim = int(data[cursor]) + 1
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

	//There are always 16 limbs
	c.AddSection(cursor, 32, "LimbsOff", "")
	for i = 0; i < 16; i++ {
		c.limbOffsets = append(c.limbOffsets, b.LE16(data, cursor))
		cursor += 2
	}

	for i = 0; i < c.NumAnim; i++ {
		c.animOffsets = append(c.animOffsets, b.LE16(data, cursor))
		c.AddSection(cursor, 2, "DefinitionOff", fmt.Sprintf("%d", i))
		cursor += 2
	}

	for i, animOffset := range c.animOffsets {
		if animOffset == 0 {
			fmt.Println("Animation offset set to 0")
			//continue
		} else {
			//cursor = animOffset
			fmt.Println("\n\nAnimation ", i, "\n==========")
			processAnimDefinition()
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
	//c.AddSection(c.animCmdOffset, c.NumAnim, "AnimCmd", "")

	// Process limbs
	for limbNumber, limbOffset := range c.limbOffsets {
		if limbOffset > len(data) {
			fmt.Printf("Something wrong with limb %d\n", limbNumber)
			continue
		}
		c.AddSection(limbOffset, 2, "Limb", "")
		imgOffset := b.LE16(data, limbOffset)
		fmt.Printf("imgOffset: %v > %v\n", imgOffset, len(data))
		if imgOffset+8 > len(data) {
			fmt.Printf("Something wrong with limb image %d\n", limbNumber)
			continue
		}
		c.AddSection(imgOffset, 2, "Pict", fmt.Sprintf("%d", limbNumber))
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
