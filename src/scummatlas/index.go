package scummatlas

import (
	"fmt"
	b "scummatlas/binaryutils"
	l "scummatlas/condlog"
)

type IndexFile struct {
	RoomNames      []RoomName
	MaximumValues  map[string]int
	RoomIndexes    []RoomIndex
	SoundIndexes   []IndexItem
	CostumeIndexes []IndexItem
	CharsetIndexes []IndexItem
	ObjectOwners   []ObjectOwner
}

type RoomName struct {
	Id   int
	Name string
}

type IndexItem struct {
	RoomNumber int
	RoomOffset int
}

type RoomIndex IndexItem

type ObjectOwner struct {
	Owner int
	State int
	Class int
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

func ParseRoomIndex(data []byte) (index []RoomIndex) {
	var out []RoomIndex
	numEntries := b.LE16(data, 8)
	l.Log("structure", "Num entries: ", numEntries)

	currentIndex := 10
	for currentIndex < len(data) {
		roomNumber := int(data[currentIndex])
		roomOffset := b.LE16(data, 1)
		out = append(out, RoomIndex{roomNumber, roomOffset})
		currentIndex += 5
	}
	return out
}

func ParseScriptsIndex(data []byte) (index []IndexItem) {
	var out []IndexItem
	numEntries := b.LE16(data, 8)
	l.Log("structure", "Num entries: ", numEntries)

	currentIndex := 10
	for currentIndex < len(data) {
		roomNumber := int(data[currentIndex])
		roomOffset := b.LE16(data, 1)
		out = append(out, IndexItem{roomNumber, roomOffset})
		currentIndex += 5
	}
	return out
}
