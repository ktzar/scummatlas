package scummatlas

import (
	"encoding/binary"
	"fmt"
)

type MainScummData struct {
	Data []byte
}

func (d *MainScummData) GetRoomCount() int {
	blockName := string(d.Data[0:4])
	blockSize := int(binary.BigEndian.Uint32(d.Data[4 : 4+4]))
	if blockName != "LECF" {
		panic("No main container in the file")
	}
	fmt.Println(blockName, blockSize)

	blockName = string(d.Data[8 : 8+4])
	blockSize = int(binary.BigEndian.Uint32(d.Data[12 : 12+4]))
	if blockName != "LOFF" {
		panic("No room offset table in the file")
	}
	fmt.Println(blockName, blockSize)
	roomCount := int(d.Data[16])
	fmt.Println("roomCount", roomCount)
	return roomCount
}
