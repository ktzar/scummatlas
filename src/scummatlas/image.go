package scummatlas

import (
	"binaryutils"
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
	stripeOffset := 0
	for i := 0; i < stripeCount; i++ {
		stripeOffset = LE32(data, 16+4*i)
		offsets = append(offsets, stripeOffset)
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	fmt.Println("Decoding image")
	fmt.Println("Stripes information")
	fmt.Println("#ID\tCode\tMethod\tTransp\tDirect\tPalInSz")
	for i := 0; i < stripeCount; i++ {
		offset := offsets[i]
		size := len(data) - offset
		if i < stripeCount-1 {
			size = offsets[i+1] - offsets[i]
		}
		drawStripe(img, i, data[8+offset:8+offset+size], pal)
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
	options.Quality = 100
	jpeg.Encode(jpegFile, img, &options)
	png.Encode(pngFile, img)
	os.Exit(1)

	return img
}

const (
	_ = iota
	METHOD_UNKNOWN
	METHOD_UNCOMPRESSED
	METHOD_ONE
	METHOD_TWO
	HORIZONTAL
	VERTICAL
	TRANSP
	NO_TRANSP
)

func getCompressionMethod(code byte) (
	method int,
	direction int,
	transparent int,
	paletteLength uint8) {
	var substraction byte

	direction = HORIZONTAL
	method = METHOD_UNKNOWN
	transparent = NO_TRANSP
	substraction = 0x00
	switch {
	case code == 0x01:
		method = METHOD_UNCOMPRESSED
	case 0x0e <= code && code <= 0x12:
		direction = VERTICAL
		method = METHOD_ONE
		substraction = 0x0a
	case 0x18 <= code && code <= 0x1c:
		method = METHOD_ONE
		substraction = 0x14
	case 0x22 <= code && code <= 0x26:
		method = METHOD_ONE
		direction = VERTICAL
		transparent = TRANSP
		substraction = 0x1e
	case 0x2c <= code && code <= 0x30:
		method = METHOD_ONE
		transparent = TRANSP
		substraction = 0x28
	case 0x40 <= code && code <= 0x44:
		method = METHOD_TWO
		substraction = 0x3c
	case 0x54 <= code && code <= 0x58:
		transparent = TRANSP
		method = METHOD_TWO
		substraction = 0x51
	case 0x68 <= code && code <= 0x6c:
		transparent = TRANSP
		method = METHOD_TWO
		substraction = 0x64
	case 0x7c <= code && code <= 0x80:
		method = METHOD_TWO
		substraction = 0x78
	}
	if method == METHOD_ONE {
		fmt.Print("1")
	} else if method == METHOD_TWO {
		fmt.Print("2")
	}
	if transparent == TRANSP {
		fmt.Print("\tYes")
	} else {
		fmt.Print("\tNo")
	}
	if direction == VERTICAL {
		fmt.Print("\tVert")
	} else {
		fmt.Print("\tHoriz")
	}
	fmt.Print("\t", code-substraction)
	fmt.Print("\n")
	return method, direction, transparent, code - substraction
}

func drawStripe(img *image.RGBA, stripNumber int, data []byte, pal color.Palette) {

	fmt.Printf("%v\t0x%X\t", stripNumber, data[0])
	method, direction, _, paletteLength := getCompressionMethod(data[0])

	height := img.Rect.Size().Y
	totalPixels := 8 * height
	currentPixel := 0

	curPal := uint8(data[1])
	curSubs := uint8(1)

	setColor := func() {
		var x, y int
		if direction == HORIZONTAL {
			x = 8*stripNumber + currentPixel%8
			y = (currentPixel + 1) / 8
		} else {
			y = currentPixel % height
			x = 8*stripNumber + currentPixel/height
		}
		if curPal >= 0 {
			img.Set(x, y, pal[curPal])
		} else {
			panic("Out of palette")
		}
		currentPixel++
	}

	bs := binaryutils.NewBitStream(data[2:])

	if method == METHOD_TWO {
		for currentPixel < totalPixels {
			if bs.GetBit() == 1 {
				if bs.GetBit() == 1 {
					palShift := 4 - bs.GetBits(3)
					if palShift == 0 {
						length := int(bs.GetBits(8))
						for i := 0; i < length; i++ {
							setColor()
						}
					} else {
						curPal += palShift
						setColor()
					}
				} else { //10
					//Read new palette index
					curPal = bs.GetBits(paletteLength)
					setColor()
				}
			} else {
				//Draw next pixel with current palette index
				setColor()
			}
		}
	} else {
		for currentPixel < totalPixels {
			if bs.GetBit() == 1 {
				if bs.GetBit() == 1 {
					if bs.GetBit() == 1 {
						//Negate the subtraction variable. Subtract it from the palette index, and draw the next pixel.
						curSubs = -curSubs
					}
					//Subtract the subtraction variable from the palette index, and draw the next pixel.
					curPal -= curSubs
				} else { //10, read new palette index
					curPal = bs.GetBits(paletteLength)
				}
			} else { // 0, draw next pixel with current palette index
			}
			setColor()
		}
	}
}
