package scummatlas

import (
	"binaryutils"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
)

const debugImage = false

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

func parsePalette(data []byte) color.Palette {
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

func parseImage(data []byte, zBuffers int, width int, height int, pal color.Palette, transpIndex uint8, debug bool) *image.RGBA {
	if string(data[8:12]) != "SMAP" {
		panic("No stripe table found")
	}

	stripeCount := width / 8
	log.Println("SmapSize", BE32(data, 12))

	log.Println("There should be ", stripeCount, "stripes")
	offsets := make([]int, 0, stripeCount)
	stripeOffset := 0
	for i := 0; i < stripeCount; i++ {
		stripeOffset = LE32(data, 16+4*i)
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

func drawStripe(img *image.RGBA, stripNumber int, data []byte, pal color.Palette, transpIndex uint8) {

	method, direction, transparent, paletteLength := getCompressionMethod(stripNumber, data[0])

	height := img.Rect.Size().Y
	totalPixels := 8 * height
	currentPixel := 0

	curPal := uint8(data[1])
	curSubs := uint8(1)

	setColor := func() {
		var x, y int
		if direction == HORIZONTAL {
			x = (currentPixel + 1) % 8
			y = (currentPixel + 1) / 8
		} else {
			y = currentPixel % height
			x = currentPixel / height
		}
		x += 8 * stripNumber
		if curPal >= 0 {
			if transparent == TRANSP && curPal == transpIndex {
				img.Set(x, y, color.RGBA{0, 0, 0, 0})
			} else {
				img.Set(x, y, pal[curPal])
			}
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
					palShift := bs.GetBits(3) - 4
					if palShift == 0 {
						length := int(bs.GetBits(8))
						for i := 0; i < length; i++ {
							setColor()
						}
					} else {
						curPal += palShift
						setColor()
					}
				} else { // 10 Read new palette index
					curPal = bs.GetBits(paletteLength)
					setColor()
				}
			} else { // 0 Draw next pixel with current palette index
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

func getCompressionMethod(stripNumber int, code byte) (
	method int,
	direction int,
	transparent int,
	paletteLength uint8) {

	direction = HORIZONTAL
	method = METHOD_UNKNOWN
	transparent = NO_TRANSP
	var substraction uint8

	if code == 0x01 {
		method = METHOD_UNCOMPRESSED
	} else if code >= 0x0e && code <= 0x30 {
		method = METHOD_ONE
	} else if code >= 0x40 && code <= 0x80 {
		method = METHOD_TWO
	}

	if (code >= 0x03 && code <= 0x12) ||
		(code >= 0x22 && code <= 0x26) {
		direction = VERTICAL
	}

	if (code >= 0x22 && code <= 0x30) ||
		(code >= 0x54 && code <= 0x6c) {
		transparent = TRANSP
	}

	codes := []uint8{0x0e, 0x18, 0x22, 0x2c, 0x40, 0x54, 0x68, 0x7c}

	for _, v := range codes {
		if code >= v && code <= v+4 {
			substraction = v - 4
			break
		}
	}

	paletteLength = code - substraction
	return
}

func printStripeInfo(stripNumber int, code byte) {
	method, direction, transparent, paletteLength := getCompressionMethod(stripNumber, code)

	out := fmt.Sprintf("%v\t0x%X\t", stripNumber, code)
	if method == METHOD_ONE {
		out += "   1"
	} else if method == METHOD_TWO {
		out += "   2"
	}
	if direction == VERTICAL {
		out += "\tVert"
	} else {
		out += "\tHoriz"
	}
	out += "\t" + string(paletteLength)
	if transparent == TRANSP {
		out += "\tYes"
	} else {
		out += "\tNo"
	}
	log.Println(out)
}
