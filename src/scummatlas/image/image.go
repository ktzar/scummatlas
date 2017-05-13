package image

import (
	"image"
	"image/color"
	b "scummatlas/binaryutils"
	l "scummatlas/condlog"
)

func ParseLimb(data []byte, width int, height int, palette color.Palette) (i *image.Paletted, length int) {
	shift := uint8(4)
	mask := byte(0xf)
	if (len(palette) > 16) {
		shift = uint8(3)
		mask = byte(0x7)
	}
	im := image.NewPaletted(image.Rect(0, 0, width, height), palette)
	cursor := 0
	pixelCount := 0
	pixelLimit := width * height

	drawNColours := func (repetition int, colour byte) {
		if int(colour) > len(palette) -1 {
			return
		}
		for i := 0 ; i < repetition ; i ++ {
			im.SetColorIndex(
				pixelCount/height,
				pixelCount%height,
				colour,
			)
			pixelCount ++
		}
	}

	for pixelCount < pixelLimit {
		if cursor + 1 > len(data) {
			break
		}
		read := data[cursor]
		color := read >> shift
		repetition := int(read & mask)
		if repetition == 0 {
			cursor ++
			repetition = int(data[cursor])
		}
		cursor ++
		if repetition + pixelCount > pixelLimit {
			break
		}
		drawNColours(repetition, color)
	}
	return im, cursor
}

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

func ParseImage(data []byte, zBuffers int, width int, height int, pal color.Palette, transpIndex uint8) (image *image.Paletted, zplanes []*image.RGBA) {
	blockName := string(data[8:12])

	if blockName == "BOMP" {
		l.Log("image", "BOMP not implemented yet")
		return nil, nil
	}
	if blockName != "SMAP" {
		panic("No stripe table found, " + blockName + " found instead")
	}
	blockSize := b.BE32(data, 12)
	l.Log("image", "SmapSize %v", blockSize)

	stripeCount := width / 8
	offsets := parseStripeTable(data, 16, stripeCount, 4)
	image = parseStripesIntoImage(data, offsets, 8, width, height, pal, transpIndex)

	zpOffset := 8 + blockSize
	for b.FourCharString(data, zpOffset)[0:2] == "ZP" {
		blockName := b.FourCharString(data, zpOffset)
		blockSize := b.BE32(data, zpOffset+4)
		l.Log("image", blockName+" found")

		offsets = parseStripeTable(data, zpOffset+8, stripeCount, 2)
		zplane := parseZplaneStripesIntoImage(data, offsets, zpOffset, height)
		zplanes = append(zplanes, zplane)
		zpOffset += blockSize
	}

	l.Log("image", "image decoded\n")
	return
}

func parseZplaneStripesIntoImage(data []byte, offsets []int, initialOffset int, height int) *image.RGBA {
	width := len(offsets) * 8
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for stripeNumber, offset := range offsets {
		drawStripeMask(img, stripeNumber, data, initialOffset+offset, height)
	}
	return img
}

func parseStripesIntoImage(data []byte, offsets []int, initialOffset int, width int, height int, pal color.Palette, transpIndex uint8) *image.Paletted {
	l.Log("image", "Decoding image")
	l.Log("image", "Stripes information")
	l.Log("image", "\n#ID\tCode\tMethod\tDirect\tPalInSz\tTransp")

	img := image.NewPaletted(image.Rect(0, 0, width, height), pal)
	for i, offset := range offsets {
		size := len(data) - offset
		if i < len(offsets)-1 {
			size = offsets[i+1] - offset
		}
		if l.Flags["image"] {
			printStripeInfo(i, data[initialOffset+offset])
		}
		stripeData := data[initialOffset+offset : initialOffset+offset+size]
		drawStripe(img, i, stripeData, transpIndex)
	}
	return img
}

func parseStripeTable(data []byte, offset int, stripeCount int, offsetSize int) []int {
	offsets := make([]int, 0, stripeCount)
	stripeOffset := 0
	for i := 0; i < stripeCount; i++ {
		if offsetSize == 4 {
			stripeOffset = b.LE32(data, offset+4*i)
		} else {
			stripeOffset = b.LE16(data, offset+2*i)
		}
		offsets = append(offsets, stripeOffset)
	}
	return offsets
}
