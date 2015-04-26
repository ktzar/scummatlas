package binaryutils

import (
	"encoding/binary"
)

func BE32(data []byte, index int) int {
	return int(binary.BigEndian.Uint32(data[index : index+4]))
}

func LE32(data []byte, index int) int {
	return int(binary.LittleEndian.Uint32(data[index : index+4]))
}

func LE24(data []byte, index int) int {
	threebytes := data[index : index+3]
	onebyte := []byte{0x00}
	threebytes = append(onebyte, threebytes...)
	return int(binary.LittleEndian.Uint32(threebytes))
}

func BE16(data []byte, index int) int {
	return int(binary.BigEndian.Uint16(data[index : index+2]))
}

func LE16(data []byte, index int) int {
	return int(binary.LittleEndian.Uint16(data[index : index+2]))
}

func FourCharString(data []byte, index int) string {
	return string(data[index : index+4])
}
