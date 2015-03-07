package scummatlas

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
)

func parsePalette(data []byte) color.Palette {
	var palette = make(color.Palette, 0, 256)
	for i := 0; i < len(data); i += 3 {
		color := color.RGBA{
			data[i],
			data[i+1],
			data[i+2],
			1,
		}
		palette = append(palette, color)
	}

	return palette
}

func parseImage(data []byte, zBuffers int, width int, height int, pal color.Palette) *image.RGBA {
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

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	fmt.Println("Decoding image")
	fmt.Println("Palette length", len(pal))
	for i := 0; i < stripeCount; i++ {
		offset := offsets[i]
		//fmt.Printf("\nOffsets of %v is %x", i, offset)
		//fmt.Print("\tHeader ", data[offset+8])
		//fmt.Print("\tCode ", int(data[offset+8])%10)

		drawStripe(img, i, data[offset:offset+height*8], pal)
	}
	fmt.Println("image decoded")

	jpegFile, err := os.Create("./image.jpg")
	if err != nil {
		panic("Error creating image.jpg")
	}
	pngFile, err := os.Create("./image.png")
	if err != nil {
		panic("Error creating image.png")
	}
	var options jpeg.Options
	options.Quality = 80
	jpeg.Encode(jpegFile, img, &options)
	png.Encode(pngFile, img)
	os.Exit(1)

	return img
}

func drawStripe(img *image.RGBA, i int, stripe []byte, pal color.Palette) {
	var x, y int
	height := img.Rect.Size().Y
	fmt.Print(".")
	if len(stripe)/8 != height {
		panic("Wrong stripe height")
	}
	for y = 0; y < height; y++ {
		for x = 0; x < 8; x++ {
			paletteIndex := stripe[y*8+x]
			color := pal[paletteIndex]
			img.Set(i*8+x, y, color)
		}
	}

}
