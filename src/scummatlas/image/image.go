package image

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	b "scummatlas/binaryutils"
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

func ParseImage(data []byte, zBuffers int, width int, height int, pal color.Palette, transpIndex uint8, debug bool) *image.RGBA {
	if string(data[8:12]) != "SMAP" {
		panic("No stripe table found")
	}

	stripeCount := width / 8
	log.Println("SmapSize", b.BE32(data, 12))

	offsets := make([]int, 0, stripeCount)
	stripeOffset := 0
	for i := 0; i < stripeCount; i++ {
		stripeOffset = b.LE32(data, 16+4*i)
		offsets = append(offsets, stripeOffset)
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	if debug {
		fmt.Println("Decoding image")
		fmt.Println("Stripes information")
		fmt.Println("\n#ID\tCode\tMethod\tDirect\tPalInSz\tTransp")
	}
	for i := 0; i < stripeCount; i++ {
		offset := offsets[i]
		size := len(data) - offset
		if i < stripeCount-1 {
			size = offsets[i+1] - offsets[i]
		}
		if debug {
			printStripeInfo(i, data[8+offset])
		}
		drawStripe(img, i, data[8+offset:8+offset+size], pal, transpIndex)
	}
	log.Println("image decoded")
	log.SetOutput(os.Stdout)
	return img
}
