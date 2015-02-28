package scummatlas

import (
	"fmt"
)

type MainScummData struct {
	Data []byte
}

func (d *MainScummData) fourChars(offset int) string {
	return string(d.Data[offset : offset+4])
}

type RoomOffset struct {
	Number int
	Offset int
}

func (d *MainScummData) GetRoomCount() int {
	blockName := d.fourChars(0)
	blockSize := BE32(d.Data, 4)
	fmt.Println(blockName)
	if blockName != "LECF" {
		panic("No main container in the file")
	}
	fmt.Printf("%v (%v bytes)\t", blockName, blockSize)

	blockName = d.fourChars(8)
	if blockName != "LOFF" {
		panic("No room offset table in the file")
	}
	blockSize = BE32(d.Data, 12)
	roomCount := int(d.Data[16])
	fmt.Printf("%v (%v bytes)\t", blockName, blockSize)
	fmt.Printf("roomCount: %v\n", roomCount)
	return roomCount
}

func (d *MainScummData) GetRoomsOffset() (offsets []RoomOffset) {
	count := d.GetRoomCount()
	currentOffset := 17
	var out []RoomOffset
	for i := 0; i < count; i++ {
		count := int(d.Data[currentOffset])
		offset := LE32(d.Data, currentOffset+1)
		roomOffset := RoomOffset{count, offset}
		out = append(out, roomOffset)
		currentOffset += 5
	}
	return out
}

func (d *MainScummData) ParseRoom(offset int) Room {
	blockName := d.fourChars(offset)
	blockSize := BE32(d.Data, offset+4)
	if blockName != "ROOM" {
		panic("No room block found")
	}
	fmt.Println("Room of size", blockSize)

	room := NewRoom(d.Data[offset : offset+blockSize])
	return *room
}
