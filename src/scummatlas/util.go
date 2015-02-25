package scummatlas

import (
	"encoding/binary"
)

func BE32(data []byte) int {
	return int(binary.BigEndian.Uint32(data))
}

func LE32(data []byte) int {
	return int(binary.LittleEndian.Uint32(data))
}

func LE16(data []byte) int {
	return int(binary.LittleEndian.Uint16(data))
}
