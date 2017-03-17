package image

import (
	"fmt"
	"image"
	"image/color"
	b "scummatlas/binaryutils"
	l "scummatlas/condlog"
)

const (
	_ = iota
	MethodUnknown
	MethodUncompressed
	MethodOne //ScummVM's UnkB
	MethodTwo //ScummVM's UnkA
	Horizontal
	Vertical
	Transp
	NoTransp
)

var transparent = color.RGBA{0, 0, 0, 0}
var opaque = color.RGBA{0, 0, 0, 255}

type StripeType struct {
	code          byte
	method        int
	direction     int
	transparent   int
	paletteLength uint8
}

func drawStripe(img *image.Paletted, stripeNumber int, data []byte, transpIndex uint8) {

	stripeType, err := getCompressionMethod(stripeNumber, data[0])
	if err != nil {
		l.Log("image", "Error getting compression method "+err.Error())
		return
	}

	height := img.Rect.Size().Y
	totalPixels := 8 * height
	currentPixel := 0

	curPal := uint8(data[1])
	curSubs := uint8(1)
	bs := b.NewBitStream(data[2:])

	setColor := func() {
		var x, y int
		if stripeType.direction == Horizontal {
			x = (currentPixel) % 8
			y = (currentPixel) / 8
		} else {
			y = (currentPixel) % height
			x = (currentPixel) / height
		}
		x += 8 * stripeNumber
		if int(curPal) < len(img.Palette) {
			if stripeType.transparent == Transp && curPal == transpIndex {
				img.SetColorIndex(x, y, transpIndex)
			} else {
				img.SetColorIndex(x, y, curPal)
			}
		} else {
			panic("Out of palette")
		}
		currentPixel++
	}

	readPaletteColor := func() {
		curPal = bs.GetBits(stripeType.paletteLength)
	}

	setColor()

	if stripeType.method == MethodTwo {
		for currentPixel < totalPixels {
			if bs.GetBit() == 1 {
				if bs.GetBit() == 1 {
					palShift := bs.GetBits(3) - 4
					if palShift > 0 {
						curPal += palShift
					} else {
						length := int(bs.GetBits(8))
						for i := 0; i < length-1; i++ {
							setColor()
						}
					}
				} else {
					readPaletteColor()
				}
			}
			setColor()
		}
	}
	if stripeType.method == MethodOne {
		for currentPixel < totalPixels {
			if bs.GetBit() == 1 {
				if bs.GetBit() == 1 {
					if bs.GetBit() == 1 {
						curSubs = -curSubs
					}
					curPal -= curSubs
				} else {
					readPaletteColor()
					curSubs = 1
				}
			}
			setColor()
		}
	}
}

func getCompressionMethod(stripeNumber int, code byte) (StripeType, error) {
	stripe := StripeType{
		direction:     Horizontal,
		method:        MethodUnknown,
		transparent:   NoTransp,
		paletteLength: 255,
	}

	if code == 0x01 {
		stripe.method = MethodUncompressed
	} else if code >= 0x0e && code <= 0x30 {
		stripe.method = MethodOne
	} else if code >= 0x40 && code <= 0x80 {
		stripe.method = MethodTwo
	}

	if (code >= 0x03 && code <= 0x12) ||
		(code >= 0x22 && code <= 0x26) {
		stripe.direction = Vertical
	}

	if (code >= 0x22 && code <= 0x30) ||
		(code >= 0x54 && code <= 0x6c) {
		stripe.transparent = Transp
	}

	codes := []uint8{0x0e, 0x18, 0x22, 0x2c, 0x40, 0x54, 0x68, 0x7c}

	for _, v := range codes {
		if code >= v && code <= v+4 {
			stripe.paletteLength = code - (v - 4)
			break
		}
	}

	if stripe.method == MethodUnknown {
		return stripe, fmt.Errorf("Unknown method for code %x", code)
	}

	return stripe, nil
}

func printStripeInfo(stripeNumber int, code byte) {
	stripeType, err := getCompressionMethod(stripeNumber, code)
	if err != nil {
		l.Log("image", fmt.Sprintf("Error with stripe %d: %v", stripeNumber, err.Error()))
		return
	}

	out := fmt.Sprintf("%v\t0x%X\t", stripeNumber, code)
	if stripeType.method == MethodOne {
		out += "   1"
	} else if stripeType.method == MethodTwo {
		out += "   2"
	}
	if stripeType.direction == Vertical {
		out += "\tVert"
	} else {
		out += "\tHoriz"
	}
	out += "\t" + fmt.Sprintf("%d", stripeType.paletteLength)
	if stripeType.transparent == Transp {
		out += "\tYes"
	} else {
		out += "\tNo"
	}
	l.Log("image", out)
}

func drawStripeMask(img *image.RGBA, stripeNumber int, data []byte, offset int, height int) {
	drawSingleColorLine := func(y int, color color.RGBA) {
		for x := 0; x < 8; x++ {
			img.Set(stripeNumber*8+x, y, color)
		}
	}

	bitmap := make([]byte, 0, height)

	linesLeft := height
	for linesLeft > 0 {
		count := data[offset]
		if count&0x80 > 0 {
			count = count & 0x7f
			value := data[offset+1]
			offset += 2
			for count > 0 && linesLeft > 0 {
				bitmap = append(bitmap, value)
				linesLeft--
				count--
			}
		} else {
			offset++
			for count > 0 && linesLeft > 0 {
				value := data[offset]
				bitmap = append(bitmap, value)
				offset++
				linesLeft--
				count--
			}
		}
	}

	for y, value := range bitmap {
		if value == 0x00 {
			drawSingleColorLine(y, transparent)
		} else if value == 0xFF {
			drawSingleColorLine(y, opaque)
		} else {
			for x, bit := range b.ByteToBits(value) {
				color := transparent
				if bit > 0 {
					color = opaque
				}
				img.Set(stripeNumber*8+x, y, color)
			}
		}
	}
}
