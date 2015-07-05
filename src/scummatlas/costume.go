package scummatlas

import (
	"fmt"
	"image/color"
	b "scummatlas/binaryutils"
)

type Costume struct {
	numAnim     int
	paletteSize int
	mirrored    bool
	palette     color.Palette
	animations  []interface{}
	limbs       []interface{}
	commands    []interface{}
}

func (c Costume) Debug() {
	fmt.Printf("%+v", c)
}

func NewCostume(data []byte) *Costume {
	c := new(Costume)
	return c

	numAnim := int(data[8])
	fmt.Printf("Num anim: %d\n", numAnim)

	format := data[9]
	numColors := 32
	fmt.Printf("Format: %x\n", format)
	if format&0x01 == 0 {
		numColors = 16
	}
	paletteSize := numColors

	palette := []byte{}
	for i := 0; i < numColors; i++ {
		palette = append(palette, data[10+i])
	}
	fmt.Printf("palette: %v\n", palette)

	limbsTableOffset := paletteSize + 12
	limbsOffsets := []int{}
	//There are 16 limbs
	for i := 0; i < 16; i++ {
		limbsOffsets = append(
			limbsOffsets,
			b.LE16(data, limbsTableOffset+i*2))
	}

	fmt.Printf("limbsOffset: %X\n", limbsOffsets)

	return c
}
