package scummatlas

import (
	"fmt"
	b "scummatlas/binaryutils"
	l "scummatlas/condlog"
)

type MainScummData struct {
	Data []byte
}

func (d *MainScummData) fourChars(offset int) string {
	return string(d.Data[offset : offset+4])
}

type RoomOffset struct {
	Id     int
	Offset int
}

func (d *MainScummData) GetRoomsOffset() (offsets []RoomOffset) {
	count := d.getRoomCount()
	currentOffset := 17
	var out []RoomOffset
	for i := 0; i < count; i++ {
		count := int(d.Data[currentOffset])
		offset := b.LE32(d.Data, currentOffset+1)
		roomOffset := RoomOffset{count, offset}
		out = append(out, roomOffset)
		currentOffset += 5
	}
	return out
}

func (d *MainScummData) getRoomCount() int {
	blockName := d.fourChars(0)
	blockSize := b.BE32(d.Data, 4)
	l.Log("structure", blockName)
	if blockName != "LECF" {
		panic("No main container in the file")
	}
	l.Log("room", "%v (%v bytes)\t", blockName, blockSize)

	blockName = d.fourChars(8)
	if blockName != "LOFF" {
		panic("No room offset table in the file")
	}
	blockSize = b.BE32(d.Data, 12)
	roomCount := int(d.Data[16])
	l.Log("room", "%v (%v bytes)\t", blockName, blockSize)
	l.Log("room", "roomCount: %v", roomCount)
	return roomCount
}

func (d MainScummData) GetScripts() []Script {
	return []Script{}
}

func (d *MainScummData) ParseRoom(offset int, order int) Room {
	blockName := d.fourChars(offset)
	blockSize := b.BE32(d.Data, offset+4)
	if blockName != "ROOM" {
		panic("No room block found")
	}
	l.Log("room", "Room of size", blockSize)

	data := d.Data[offset : offset+blockSize]
	dumpBlock(fmt.Sprintf("ROOM_%d", order), data)
	room := NewRoom(data)
	return *room
}
