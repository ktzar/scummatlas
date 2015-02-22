package scummatlas

import (
    "io/ioutil"
    "fmt"
    "encoding/binary"
)

type RoomIndex struct {
    number int
    offset []byte
}

func ReadXoredFile(fileName string, code byte) (out []byte, err error) {
    out, err = ioutil.ReadFile(fileName)
    for i, _ := range(out) {
        out[i] = out[i] ^ 0x69
    }
    return out, err
}

func complementaryByte(in []byte) (out []byte) {
    for i, _ := range(in) {
        in[i] = in[i]  ^ 0xFF
    }
    return in
}

func ParseRoomNames(data []byte) []string {
    var names []string
    if string(data[0:4]) == "RNAM" {
        currentIndex := 8
        for currentIndex < len(data){
            roomNumber := int(data[currentIndex])
            if roomNumber == 0 {
                break
            }
            name := data[currentIndex+1:currentIndex+10]
            names = append(names, string(complementaryByte(name)))
            currentIndex += 10
        }
    }
    return names
}

func ParseRoomIndex(data []byte) (index []RoomIndex) {
    aa := make([]RoomIndex, 1)
    if string(data[0:4]) == "DROO" {
        roomCount := int(binary.LittleEndian.Uint16(data[8:10]))
        fmt.Println("Number of rooms ", roomCount)

    }
    return aa
}

