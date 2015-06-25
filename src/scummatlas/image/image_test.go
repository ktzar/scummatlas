package image

import "testing"

/*
// TODO
func TestPalette(t *testing.T) {
	t.Errorf("Not implemented")
}
*/

func TestGetCompressionMethod(t *testing.T) {

	assertCompressionMethod := func(code uint8, method int, direction int, trans int, palette uint8) {

		stripeType, err := getCompressionMethod(0, code)
		if err != nil {
			t.Errorf("Method returns error %v for %x", err.Error(), code)
		}
		if stripeType.method != method {
			t.Errorf("Method doesn't match for code %x", code)
		}
		if stripeType.direction != direction {
			t.Errorf("Direction doesn't match for code %x", code)
		}
		if stripeType.transparent != trans {
			t.Errorf("Transparency doesn't match for code %x", code)
		}
		if stripeType.paletteLength != palette {
			t.Errorf("Palette size %x doesn't match for code %x", code)
		}
	}

	assertCompressionMethod(0x01, MethodUncompressed, Horizontal, NoTransp, 255)
	assertCompressionMethod(0x0e, MethodOne, Vertical, NoTransp, 4)
	assertCompressionMethod(0x12, MethodOne, Vertical, NoTransp, 8)

	assertCompressionMethod(0x22, MethodOne, Vertical, Transp, 4)
	assertCompressionMethod(0x26, MethodOne, Vertical, Transp, 8)

	assertCompressionMethod(0x2c, MethodOne, Horizontal, Transp, 4)
	assertCompressionMethod(0x30, MethodOne, Horizontal, Transp, 8)

	assertCompressionMethod(0x40, MethodTwo, Horizontal, NoTransp, 4)
	assertCompressionMethod(0x44, MethodTwo, Horizontal, NoTransp, 8)

	assertCompressionMethod(0x54, MethodTwo, Horizontal, Transp, 4)
	assertCompressionMethod(0x58, MethodTwo, Horizontal, Transp, 8)

	assertCompressionMethod(0x68, MethodTwo, Horizontal, Transp, 4)
	assertCompressionMethod(0x6c, MethodTwo, Horizontal, Transp, 8)

	assertCompressionMethod(0x7c, MethodTwo, Horizontal, NoTransp, 4)
	assertCompressionMethod(0x80, MethodTwo, Horizontal, NoTransp, 8)
}
