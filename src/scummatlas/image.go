package scummatlas

import "fmt"
import "os"
import "image/color"
import "image"

func parsePalette(data []byte) color.Palette {
	var palette = make(color.Palette, 0, 256)
	for i := 0; i < len(data)/3; i += 3 {
		color := color.RGBA{
			data[i*3],
			data[i*3+1],
			data[i*3+2],
			1,
		}
		palette = append(palette, color)
	}

	return palette
}

func parseImage(data []byte, zBuffers int, width int, height int, pal color.Palette) *image.Paletted {
	if string(data[8:12]) != "SMAP" {
		panic("No stripe table found")
	}

	smapSize := BE32(data, 12)
	stripeCount := width / 8
	fmt.Println("SmapSize", smapSize)

	fmt.Println("There should be ", stripeCount, "stripes")
	offsets := make([]int, 0, stripeCount)
	for i := 0; i < stripeCount; i++ {
		stripeOffset := LE32(data, 16+4*i)
		offsets = append(offsets, stripeOffset)
	}

	for i := 0; i < stripeCount; i++ {
		fmt.Printf("\nOffsets of %v is %x", i, offsets[i])
		fmt.Print("\tHeader ", data[offsets[i]+8])
		fmt.Print("\tCode ", int(data[offsets[i]+8])%10)
	}

	image := image.NewPaletted(image.Rect(0, 0, width, height), pal)

	os.Exit(255)
	return image

}
