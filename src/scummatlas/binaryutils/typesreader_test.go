package binaryutils

import "testing"

func testCountOneBitsInWord(expected int, word int, t *testing.T) {
	result := CountOneBitsInWord(word)
	if result != expected {
		t.Errorf("testCountOneBitsInWord expected %v bits for %02x but got %v", expected, word, result)
	}

	resultSlice := OneBitsInWord(word)
	if len(resultSlice) != expected {
		t.Errorf("testCountOneBitsInWord expected %v bits for %02x but got %v", expected, word, len(resultSlice))
	}
}

func TestCountOneBitsInWord(t *testing.T) {
	testCountOneBitsInWord(1, 0x01, t)
	testCountOneBitsInWord(1, 0x10, t)
	testCountOneBitsInWord(1, 0x02, t)
	testCountOneBitsInWord(1, 0x20, t)
	testCountOneBitsInWord(2, 0x03, t)
	testCountOneBitsInWord(2, 0x30, t)
	testCountOneBitsInWord(4, 0x33, t)
	testCountOneBitsInWord(6, 0x77, t)
	testCountOneBitsInWord(5, 0xE3, t)
	testCountOneBitsInWord(5, 0x3E, t)
	testCountOneBitsInWord(4, 0xF0, t)
	testCountOneBitsInWord(4, 0x0F, t)
	testCountOneBitsInWord(8, 0xFF, t)
}
