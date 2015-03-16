package scummatlas

import (
	"binaryutils"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
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

func parseImage(data []byte, zBuffers int, width int, height int, pal color.Palette, transpIndex uint8) *image.RGBA {
	// DISABLE LOGGING
	log.SetOutput(os.Stdout)
	log.SetOutput(ioutil.Discard)

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

	log.Println("Decoding image")
	log.Println("Stripes information")
	log.Println("\n#ID\tCode\tMethod\tDirect\tPalInSz\tTransp")
	for i := 0; i < stripeCount; i++ {
		offset := offsets[i]
		size := len(data) - offset
		if i < stripeCount-1 {
			size = offsets[i+1] - offsets[i]
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
			x = currentPixel % 8
			y = (currentPixel + 1) / 8
		} else {
			y = currentPixel % height
			x = currentPixel / height
		}
		if x == 7 {
			x = -1
		}
		x += 8 * stripNumber
		if curPal >= 0 {
			if transparent == TRANSP && curPal == transpIndex {
				log.Println("TRANSPARENT")
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

func GetCompressionMethod(stripNumber int, code byte) (
	method int,
	direction int,
	transparent int,
	paletteLength uint8) {
	var substraction byte

	direction = VERTICAL
	method = METHOD_UNKNOWN
	transparent = NO_TRANSP
	substraction = 0x00
	if code == 0x01 {
		method = METHOD_UNCOMPRESSED
	} else if code <= 0x40 {
		method = METHOD_ONE
	} else if code <= 0x80 {
		method = METHOD_TWO
	} else {
		panic("Unknown method " + string(method) + " in stripe " + string(stripNumber))
	}

	if (code >= 0x22 && code <= 0x40) ||
		(code >= 0x54 && code <= 0x6c) {
		transparent = TRANSP
	}

	if (code >= 0x18 && code < 0x1c) || code >= 0x2c {
		direction = HORIZONTAL
	}

	switch {
	case 0x0e <= code && code <= 0x12:
		substraction = 0x0a
	case 0x18 <= code && code <= 0x1c:
		substraction = 0x14
	case 0x22 <= code && code <= 0x26:
		substraction = 0x1e
	case 0x2c <= code && code <= 0x30:
		substraction = 0x28
	case 0x40 <= code && code <= 0x44:
		substraction = 0x3c
	case 0x54 <= code && code <= 0x58:
		substraction = 0x51
	case 0x68 <= code && code <= 0x6c:
		substraction = 0x64
	case 0x7c <= code && code <= 0x80:
		substraction = 0x78
	}
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
	out += "\t" + string(code-substraction)
	if transparent == TRANSP {
		out += "\tYes"
	} else {
		out += "\tNo"
	}
	log.Println(out)
	return method, direction, transparent, code - substraction
}
