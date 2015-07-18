package binaryutils

import "testing"

func testOneBitsInWord(expected int, word int, t *testing.T) {
	result := OneBitsInWord(word)
	if result != expected {
		t.Errorf("TestOneBitsInWord expected %v bits for %02x but got %v", expected, word, result)
	}
}

func TestOneBitsInWord(t *testing.T) {
	testOneBitsInWord(1, 0x01, t)
	testOneBitsInWord(1, 0x10, t)
	testOneBitsInWord(1, 0x02, t)
	testOneBitsInWord(1, 0x20, t)
	testOneBitsInWord(2, 0x03, t)
	testOneBitsInWord(2, 0x30, t)
	testOneBitsInWord(4, 0x33, t)
	testOneBitsInWord(6, 0x77, t)
	testOneBitsInWord(5, 0xE3, t)
	testOneBitsInWord(5, 0x3E, t)
	testOneBitsInWord(4, 0xF0, t)
	testOneBitsInWord(4, 0x0F, t)
	testOneBitsInWord(8, 0xFF, t)
}
