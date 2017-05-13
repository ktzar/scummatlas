package scummatlas

import (
	"fmt"
	b "scummatlas/binaryutils"
	l "scummatlas/condlog"
	s "scummatlas/script"
	"scummatlas/blocks"
)

type MainScummData struct {
	data     []byte
	sections map[string][]int
}

func NewMainScummData(data []byte) *MainScummData {
	if !checkMainScummData(data) {
		panic("No main container in the file")
	}
	main := new(MainScummData)
	main.data = data
	main.sections = make(map[string][]int)

	loffSize := b.BE32(data, 12)
	currentOffset := 12 + loffSize + 4

	for currentOffset < len(data) {
		blockName := b.FourCharString(data, currentOffset)
		if blockName == "LFLF" {
			currentOffset += 8
			continue
		}
		if main.sections[blockName] == nil {
			main.sections[blockName] = make([]int, 1)
		}
		main.sections[blockName] = append(main.sections[blockName], currentOffset)
		blockSize := b.BE32(data, currentOffset+4)
		currentOffset += blockSize
		l.Log("structure", "%v\t%v", blockName, blockSize)
	}

	return main
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
		count := int(d.data[currentOffset])
		offset := b.LE32(d.data, currentOffset+1)
		roomOffset := RoomOffset{count, offset}
		out = append(out, roomOffset)
		currentOffset += 5
	}
	return out
}

func (d *MainScummData) getRoomCount() int {
	blockName := b.FourCharString(d.data, 8)
	if blockName != "LOFF" {
		panic("No room offset table in the file")
	}
	blockSize := b.BE32(d.data, 12)
	roomCount := int(d.data[16])
	l.Log("room", "%v (%v bytes)\t", blockName, blockSize)
	l.Log("room", "roomCount: %v", roomCount)
	return roomCount
}

func (d MainScummData) GetCostumes() (costumes []blocks.Costume) {
	firstRoom := d.ParseRoom(d.GetRoomsOffset()[3].Offset, 0)
	for i, offset := range d.sections["COST"] {
		blockSize := b.BE32(d.data, offset+4)
		l.Log("structure", "Parsing costume %d", i)
		costumes = append(costumes,
			*blocks.NewCostume(
				d.data[offset+8:offset+blockSize],
				firstRoom.Palette,
			),
		)
		blocks.DumpBlock(fmt.Sprintf("COST_%d", i), d.data[offset:offset+blockSize])
	}
	return
}

func (d MainScummData) GetScripts() (scripts []s.Script) {
	for i, offset := range d.sections["SCRP"] {
		blockSize := b.BE32(d.data, offset+4)
		l.Log("script", "Parsing global script %d", i)
		script := s.ParseScriptBlock(d.data[offset+8 : offset+blockSize])
		scripts = append(scripts, script)
		blocks.DumpBlock(fmt.Sprintf("SCRP_%d", i), d.data[offset:offset+blockSize])
	}
	return
}

func (d *MainScummData) ParseRoom(offset int, order int) blocks.Room {
	blockName := b.FourCharString(d.data, offset)
	blockSize := b.BE32(d.data, offset+4)
	if blockName != "ROOM" {
		panic("No room block found")
	}
	l.Log("room", "Room", order)

	data := d.data[offset : offset+blockSize]
	blocks.DumpBlock(fmt.Sprintf("ROOM_%d", order), data)
	room := blocks.NewRoom(data)
	return *room
}

func checkMainScummData(data []byte) bool {
	blockName := b.FourCharString(data, 0)
	l.Log("structure", blockName)
	if blockName != "LECF" {
		return false
	}
	blockName = b.FourCharString(data, 8)
	if blockName != "LOFF" {
		return false
	}
	return true
}
