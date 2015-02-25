package scummatlas

import (
	"encoding/binary"
)

func BE32(data []byte, index int) int {
	return int(binary.BigEndian.Uint32(data[index : index+4]))
}

func LE32(data []byte, index int) int {
	return int(binary.LittleEndian.Uint32(data[index : index+4]))
}

func LE16(data []byte, index int) int {
	return int(binary.LittleEndian.Uint16(data[index : index+2]))
}
