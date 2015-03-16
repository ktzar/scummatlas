package scummatlas

import "testing"

func TestGetCompressionMethod(t *testing.T) {

	assertCompressionMethod := func(code uint8, method int, direction int, trans int, palette uint8) {

		_method, _direction, _trans, _palette := getCompressionMethod(0, code)
		if _method != method {
			t.Errorf("Method doesn't match for code %x", code)
		}
		if _direction != direction {
			t.Errorf("Direction doesn't match for code %x", code)
		}
		if _trans != trans {
			t.Errorf("Transparency doesn't match for code %x", code)
		}
		if _palette != palette {
			t.Errorf("Palette size doesn't match for code %x", code)
		}
	}

	assertCompressionMethod(0x01, METHOD_UNCOMPRESSED, HORIZONTAL, NO_TRANSP, 0x01)
	assertCompressionMethod(0x0e, METHOD_ONE, VERTICAL, NO_TRANSP, 4)
	assertCompressionMethod(0x12, METHOD_ONE, VERTICAL, NO_TRANSP, 8)

	assertCompressionMethod(0x22, METHOD_ONE, VERTICAL, TRANSP, 4)
	assertCompressionMethod(0x26, METHOD_ONE, VERTICAL, TRANSP, 8)

	assertCompressionMethod(0x2c, METHOD_ONE, HORIZONTAL, TRANSP, 4)
	assertCompressionMethod(0x30, METHOD_ONE, HORIZONTAL, TRANSP, 8)

	assertCompressionMethod(0x40, METHOD_TWO, HORIZONTAL, NO_TRANSP, 4)
	assertCompressionMethod(0x44, METHOD_TWO, HORIZONTAL, NO_TRANSP, 8)

	assertCompressionMethod(0x54, METHOD_TWO, HORIZONTAL, TRANSP, 4)
	assertCompressionMethod(0x58, METHOD_TWO, HORIZONTAL, TRANSP, 8)

	assertCompressionMethod(0x68, METHOD_TWO, HORIZONTAL, TRANSP, 4)
	assertCompressionMethod(0x6c, METHOD_TWO, HORIZONTAL, TRANSP, 8)

	assertCompressionMethod(0x7c, METHOD_TWO, HORIZONTAL, NO_TRANSP, 4)
	assertCompressionMethod(0x80, METHOD_TWO, HORIZONTAL, NO_TRANSP, 8)
}
