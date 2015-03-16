package scummatlas

import "testing"

func TestGetCompressionMethod(t *testing.T) {

	assertCompressionMethod := func(code, method, direction, trans, palette) {

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

	assertCompressionMethod(0x01, METHOD_UNCOMPRESSED, VERTICAL, NO_TRANS, 0x00)
	assertCompressionMethod(0x0e, METHOD_ONE, VERTICAL, NO_TRANS, 3)
	assertCompressionMethod(0x11, METHOD_ONE, VERTICAL, NO_TRANS, 7)
	assertCompressionMethod(0x12, METHOD_ONE, VERTICAL, NO_TRANS, 8)

	assertCompressionMethod(0x22, METHOD_ONE, VERTICAL, TRANSP, 4)

}
