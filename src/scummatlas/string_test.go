package scummatlas

import "testing"

func TestStringOpcode(t *testing.T) {
	str, length := parsePrintOpcode([]byte{
		0x0F, 0x41, 0x42, 0x43, 0x44, 0x00, 0xAA, 0xAA}, 0)
	if str != "ABCD" {
		t.Errorf("String %v is not right", str)
	}
	if length != 5 {
		t.Errorf("Length %d is not right", length)
	}

	//OFFSET
	str, length = parsePrintOpcode([]byte{
		0xAA, 0xAA, 0x0F, 0x41, 0x42, 0x43, 0x44, 0x00, 0xAA, 0xAA}, 2)
	if str != "ABCD" {
		t.Errorf("String %v is not right", str)
	}
	if length != 5 {
		t.Errorf("Length %d is not right", length)
	}

	//ENCODINGS
	//TODO

}
