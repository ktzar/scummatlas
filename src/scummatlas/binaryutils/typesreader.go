package binaryutils

import (
	"encoding/binary"
	"io/ioutil"
)

func BE32(data []byte, index int) int {
	return int(binary.BigEndian.Uint32(data[index : index+4]))
}

func LE32(data []byte, index int) int {
	return int(binary.LittleEndian.Uint32(data[index : index+4]))
}

func LE24(data []byte, index int) int {
	fourbytes := []byte{data[index], data[index+1], data[index+2], 0x00}
	return int(binary.LittleEndian.Uint32(fourbytes))
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

func ReadXoredFile(fileName string, code byte) (out []byte, err error) {
	out, err = ioutil.ReadFile(fileName)
	for i := range out {
		out[i] = out[i] ^ code
	}
	return out, err
}

func OneBitsInWord(word int) []int {
	out := make([]int, 0)
	mask := 0x01
	for i := uint8(0); i < 16; i++ {
		if (mask<<i)&word > 0 {
			out = append(out, int(i))
		}
	}
	return out
}

func CountOneBitsInWord(word int) int {
	mask := 0x01
	accum := 0
	for i := uint8(0); i < 16; i++ {
		if (mask<<i)&word > 0 {
			accum++
		}
	}
	return accum
}
