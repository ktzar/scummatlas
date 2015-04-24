package scummatlas

import (
	"fmt"
	"io/ioutil"
	b "scummatlas/binaryutils"
)

type RoomName struct {
	Number int
	Name   string
}

type RoomIndex struct {
	number int
	offset []byte
}

type ScriptIndex struct {
	roomNumber int
	roomOffset int
}

func ReadXoredFile(fileName string, code byte) (out []byte, err error) {
	out, err = ioutil.ReadFile(fileName)
	for i, _ := range out {
		out[i] = out[i] ^ 0x69
	}
	return out, err
}

func complementaryString(in []byte) string {
	var out string
	for i, _ := range in {
		in[i] = in[i] ^ 0xFF
		if in[i] > 0x40 && in[i] < 0x80 {
			out = out + string(in[i])
		}
	}
	return out
}

func ParseRoomNames(data []byte) []RoomName {
	var out []RoomName
	if string(data[0:4]) == "RNAM" {
		currentIndex := 8
		for currentIndex < len(data) {
			roomNumber := int(data[currentIndex])
			if roomNumber == 0 {
				break
			}
			name := data[currentIndex+1 : currentIndex+10]
			roomName := RoomName{
				roomNumber,
				fmt.Sprintf("%v", complementaryString(name)),
			}
			out = append(out, roomName)
			currentIndex += 10
		}
	}
	return out
}

func ParseRoomIndex(data []byte) (index []ScriptIndex) {
	var out []ScriptIndex
	numEntries := b.LE16(data, 8)
	fmt.Println("Num entries: ", numEntries)

	currentIndex := 10
	for currentIndex < len(data) {
		roomNumber := int(data[currentIndex])
		roomOffset := b.LE16(data, 1)
		out = append(out, ScriptIndex{roomNumber, roomOffset})
		currentIndex += 5
	}
	return out
}

func ParseScriptsIndex(data []byte) (index []ScriptIndex) {
	var out []ScriptIndex
	numEntries := b.LE16(data, 8)
	fmt.Println("Num entries: ", numEntries)

	currentIndex := 10
	for currentIndex < len(data) {
		roomNumber := int(data[currentIndex])
		roomOffset := b.LE16(data, 1)
		out = append(out, ScriptIndex{roomNumber, roomOffset})
		currentIndex += 5
	}
	return out
}
