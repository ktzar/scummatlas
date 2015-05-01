package binaryutils

import "testing"

func TestByteToBits(t *testing.T) {
	expected := [8]uint8{0, 1, 1, 0, 0, 1, 1, 0}
	calculated := ByteToBits(0x66)
	if calculated != expected {
		t.Errorf("%v not equals %v", calculated, expected)
	}

	expected = [8]uint8{0, 0, 1, 0, 0, 1, 1, 1}
	calculated = ByteToBits(0x27)
	if calculated != expected {
		t.Errorf("%v not equals %v", calculated, expected)
	}
}

func TestOneByte(t *testing.T) {
	expected := "01100110"
	data := []byte{0x66}
	bs := NewBitStream(data)
	stream := ""
	for i := 0; i < len(expected); i++ {
		if bs.GetBit() == 1 {
			stream += "1"
		} else {
			stream += "0"
		}
	}
	if stream != expected {
		t.Errorf("%v not equals %v", stream, expected)
	}

}

func TestStream(t *testing.T) {
	expected := "101101010000000000010110"
	data := []byte{0xad, 0x00, 0x68}
	bs := NewBitStream(data)
	stream := ""
	for i := 0; i < len(expected); i++ {
		if bs.GetBit() == 1 {
			stream += "1"
		} else {
			stream += "0"
		}
	}
	if stream != expected {
		t.Errorf("%v not equals %v", stream, expected)
	}
}

func TestGetMultiple(t *testing.T) {
	data := []byte{0xad, 0x00, 0x68}
	expected := []uint8{0xd, 0xa, 0x0, 0x0, 0x8, 0x6}
	bs := NewBitStream(data)
	for i := range expected {
		num := bs.GetBits(4)
		if num != expected[i] {
			t.Errorf("Value not correct. Got %x and expected %x", num, expected[i])
		}
	}
}

func TestGetMultipleByByte(t *testing.T) {
	data := []byte{0xad, 0x00, 0x68}
	expected := []uint8{0xad, 0x00, 0x68}
	bs := NewBitStream(data)
	for i := range expected {
		num := bs.GetBits(8)
		if num != expected[i] {
			t.Errorf("Value not correct. Got %x and expected %x", num, expected[i])
		}
	}
}

func TestGetMultipleByTwo(t *testing.T) {
	data := []byte{0xad}
	expected := []uint8{1, 3, 2, 2}
	bs := NewBitStream(data)
	for i := range expected {
		num := bs.GetBits(2)
		if num != expected[i] {
			t.Errorf("Value not correct. Got %x and expected %x", num, expected[i])
		}
	}
}

func TestGetMultipleByThree(t *testing.T) {
	data := []byte{0xad}
	expected := []uint8{5, 5, 2, 0}
	bs := NewBitStream(data)
	for i := range expected {
		num := bs.GetBits(3)
		if num != expected[i] {
			t.Errorf("Value not correct. Got %x and expected %x", num, expected[i])
		}
	}
}
