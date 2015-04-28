package image

import (
	"fmt"
	"image"
	"image/color"
	b "scummatlas/binaryutils"
	l "scummatlas/condlog"
)

func ParsePalette(data []byte) color.Palette {
	var palette = make(color.Palette, 0, 256)
	for i := 0; i < len(data); i += 3 {
		color := color.RGBA{
			data[i],
			data[i+1],
			data[i+2],
			255,
		}
		palette = append(palette, color)
	}

	return palette
}

func parseStripeTable(data []byte, offset int, stripeCount int, isZp bool) []int {
	offsets := make([]int, 0, stripeCount)
	stripeOffset := 0
	for i := 0; i < stripeCount; i++ {
		if isZp {
			stripeOffset = b.LE16(data, offset+2*i)
		} else {
			stripeOffset = b.LE32(data, offset+4*i)
		}
		offsets = append(offsets, stripeOffset)
	}
	return offsets
}

func parseStripesIntoImage(data []byte, initialOffset int, width int, height int, pal color.Palette, transpIndex uint8) *image.RGBA {
	l.Log("image", "Decoding image")
	l.Log("image", "Stripes information")
	l.Log("image", "\n#ID\tCode\tMethod\tDirect\tPalInSz\tTransp")

	isZp := string(data[initialOffset:initialOffset+3]) == "ZP0"

	stripeCount := width / 8
	offsets := parseStripeTable(data, initialOffset+8, stripeCount, isZp)
	fmt.Println("offsets", offsets)

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for i := 0; i < stripeCount; i++ {
		offset := offsets[i]
		size := len(data) - offset
		if i < stripeCount-1 {
			size = offsets[i+1] - offsets[i]
		}
		if l.Logflags["image"] {
			printStripeInfo(i, data[initialOffset+offset])
		}
		drawStripe(
			img,
			i,
			data[initialOffset+offset:initialOffset+offset+size],
			pal,
			transpIndex)
	}
	return img
}

func ParseImage(data []byte, zBuffers int, width int, height int, pal color.Palette, transpIndex uint8) (images []*image.RGBA) {
	blockName := string(data[8:12])

	if blockName == "BOMP" {
		l.Log("image", "BOMP not implemented yet")
		images := append(images, image.NewRGBA(image.Rect(0, 0, width, height)))
		return images
	}
	if blockName != "SMAP" {
		panic("No stripe table found, " + blockName + " found instead")
	}
	blockSize := b.BE32(data, 12)
	l.Log("image", "SmapSize %v", blockSize)

	images = append(images, parseStripesIntoImage(data, 8, width, height, pal, transpIndex))

	/*
		if b.FourCharString(data, 8+blockSize) == "ZP01" {
			fmt.Println("ZP01 found")
		}
		images = append(images, parseStripesIntoImage(data, 8+blockSize, width, height, pal, transpIndex))
	*/

	l.Log("image", "image decoded\n")
	return
}
