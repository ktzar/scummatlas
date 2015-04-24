package image

import (
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

func ParseImage(data []byte, zBuffers int, width int, height int, pal color.Palette, transpIndex uint8) *image.RGBA {
	blockName := string(data[8:12])

	if blockName == "BOMP" {
		l.Log("image", "BOMP not implemented yet")
		return image.NewRGBA(image.Rect(0, 0, width, height))
	}
	if blockName != "SMAP" {
		panic("No stripe table found, " + blockName + " found instead")
	}

	stripeCount := width / 8
	l.Log("image", "SmapSize %v", b.BE32(data, 12))

	offsets := make([]int, 0, stripeCount)
	stripeOffset := 0
	for i := 0; i < stripeCount; i++ {
		stripeOffset = b.LE32(data, 16+4*i)
		offsets = append(offsets, stripeOffset)
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	l.Log("image", "Decoding image")
	l.Log("image", "Stripes information")
	l.Log("image", "\n#ID\tCode\tMethod\tDirect\tPalInSz\tTransp")
	for i := 0; i < stripeCount; i++ {
		offset := offsets[i]
		size := len(data) - offset
		if i < stripeCount-1 {
			size = offsets[i+1] - offsets[i]
		}
		if l.Logflags["image"] {
			printStripeInfo(i, data[8+offset])
		}
		drawStripe(img, i, data[8+offset:8+offset+size], pal, transpIndex)
	}
	l.Log("image", "image decoded\n")
	return img
}
