package script

import "testing"

func TestStringOpcode(t *testing.T) {
	var actions []string
	var length int

	actions, length = parsePrintOpcode([]byte{
		0x0F,
		0x41, 0x42, 0x43, 0x44,
		0x00, 0xAA, 0xAA}, 0)
	if actions[0] != "text=\"ABCD\"" {
		t.Errorf("Action %v is not right", actions[0])
	}
	if length != 5 {
		t.Errorf("Length %d is not right", length)
	}

	//OFFSET
	actions, length = parsePrintOpcode([]byte{
		0xAA, 0xAA, 0x0F,
		0x41, 0x42, 0x43, 0x44, 0x00,
		0xAA, 0xAA}, 2)
	if actions[0] != "text=\"ABCD\"" {
		t.Errorf("Action %v is not right", actions[0])
	}
	if length != 5 {
		t.Errorf("Length %d is not right", length)
	}

	//ENCODINGS
	actions, length = parsePrintOpcode([]byte{
		0x0F,
		0x41, 0xFF, 0x01, 0x42, 0x43,
		0xFF, 0x03,
		0x44, 0x00,
	}, 0)
	if actions[0] != "text=\"A\\n BC\\wait D\"" {
		t.Errorf("Action %v is not right", actions[0])
	}
	if length != 9 {
		t.Errorf("Length %d is not right", length)
	}

}
