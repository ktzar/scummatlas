package image

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	b "scummatlas/binaryutils"
	l "scummatlas/condlog"
)

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

func drawStripe(img *image.RGBA, stripNumber int, data []byte, pal color.Palette, transpIndex uint8) {

	stripeType, err := getCompressionMethod(stripNumber, data[0])
	if err != nil {
		l.Log("image", "Error getting compression method "+err.Error())
		return
	}

	height := img.Rect.Size().Y
	totalPixels := 8 * height
	currentPixel := 0

	curPal := uint8(data[1])

	curSubs := uint8(1)

	setColor := func() {
		var x, y int
		if stripeType.direction == HORIZONTAL {
			x = (currentPixel) % 8
			y = (currentPixel) / 8
		} else {
			y = (currentPixel) % height
			x = (currentPixel) / height
		}
		x += 8 * stripNumber
		if curPal >= 0 {
			if stripeType.transparent == TRANSP && curPal == transpIndex {
				img.Set(x, y, color.RGBA{0, 0, 0, 0})
			} else {
				img.Set(x, y, pal[curPal])
			}
		} else {
			panic("Out of palette")
		}
		currentPixel++
	}

	bs := b.NewBitStream(data[2:])

	setColor()
	if stripeType.method == METHOD_TWO {
		for currentPixel < totalPixels {
			if bs.GetBit() == 1 {
				if bs.GetBit() == 1 {
					palShift := bs.GetBits(3) - 4
					if palShift == 0 {
						length := int(bs.GetBits(8))
						for i := 0; i < length-1; i++ {
							setColor()
						}
					} else {
						curPal += palShift
					}
				} else { // 10 Read new palette index
					curPal = bs.GetBits(stripeType.paletteLength)
				}
			} else { // 0 Draw next pixel with current palette index
			}
			setColor()
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
					curPal = bs.GetBits(stripeType.paletteLength)
				}
			} else { // 0, draw next pixel with current palette index
			}
			setColor()
		}
	}
}

type StripeType struct {
	code          byte
	method        int
	direction     int
	transparent   int
	paletteLength uint8
}

func getCompressionMethod(stripNumber int, code byte) (StripeType, error) {
	stripe := StripeType{
		direction:     HORIZONTAL,
		method:        METHOD_UNKNOWN,
		transparent:   NO_TRANSP,
		paletteLength: 255,
	}

	if code == 0x01 {
		stripe.method = METHOD_UNCOMPRESSED
	} else if code >= 0x0e && code <= 0x30 {
		stripe.method = METHOD_ONE
	} else if code >= 0x40 && code <= 0x80 {
		stripe.method = METHOD_TWO
	}

	if (code >= 0x03 && code <= 0x12) ||
		(code >= 0x22 && code <= 0x26) {
		stripe.direction = VERTICAL
	}

	if (code >= 0x22 && code <= 0x30) ||
		(code >= 0x54 && code <= 0x6c) {
		stripe.transparent = TRANSP
	}

	codes := []uint8{0x0e, 0x18, 0x22, 0x2c, 0x40, 0x54, 0x68, 0x7c}

	for _, v := range codes {
		if code >= v && code <= v+4 {
			stripe.paletteLength = code - (v - 4)
			break
		}
	}

	if stripe.method == METHOD_UNKNOWN {
		return stripe, errors.New(fmt.Sprintf("Unknown method for code %x", code))
	}

	return stripe, nil
}

func printStripeInfo(stripNumber int, code byte) {
	stripeType, err := getCompressionMethod(stripNumber, code)
	if err != nil {
		l.Log("image", fmt.Sprintf("Error with stripe %d: %v", stripNumber, err.Error()))
		return
	}

	out := fmt.Sprintf("%v\t0x%X\t", stripNumber, code)
	if stripeType.method == METHOD_ONE {
		out += "   1"
	} else if stripeType.method == METHOD_TWO {
		out += "   2"
	}
	if stripeType.direction == VERTICAL {
		out += "\tVert"
	} else {
		out += "\tHoriz"
	}
	out += "\t" + string(stripeType.paletteLength)
	if stripeType.transparent == TRANSP {
		out += "\tYes"
	} else {
		out += "\tNo"
	}
	l.Log("image", out)
}
