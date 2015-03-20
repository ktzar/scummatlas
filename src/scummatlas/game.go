package scummatlas

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Game struct {
	Name      string
	RoomCount int
	RoomNames []RoomName
	Rooms     []Room
}

const DEBUG_SAVE_DECODED = true

func NewGame(gamedir string, outputdir string) *Game {

	var game Game

	filesInfo, err := ioutil.ReadDir(gamedir)
	if err != nil {
		panic("Game directory not a directory")
	}

	// Read index file
	for _, file := range filesInfo {
		fileName := gamedir + "/" + file.Name()
		extension := filepath.Ext(fileName)
		data := []byte{}

		if strings.Contains(extension, ".00") {
			absPath, _ := filepath.Abs(gamedir + "/" + file.Name())
			data, err = ReadXoredFile(absPath, 0x69)
			if err != nil {
				panic("Can't read index file")
			}

			if DEBUG_SAVE_DECODED {
				f, _ := os.Create(outputdir + "/" + file.Name() + ".decoded")
				defer f.Close()
				f.Write(data)
			}
		}

		if extension == ".000" {
			currIndex := 0
			for currIndex < len(data) {
				blockName := string(data[currIndex : currIndex+4])
				blockSize := int(binary.BigEndian.Uint32(data[currIndex+4 : currIndex+8]))
				currBlock := data[currIndex : currIndex+blockSize]

				currIndex += blockSize
				//TODO Remove
				//continue
				switch blockName {
				case "RNAM":
					fmt.Println("Parse Room Names")
					game.RoomNames = ParseRoomNames(currBlock)

				case "MAXS":
					//fmt.Println("Parse Maximum Values")

				case "DROO":
					//fmt.Println("Parse Directory of Rooms")
					//fmt.Println(ParseRoomIndex(currBlock))

				case "DSCR":
					//fmt.Println("Parse Directory of Scripts")
					//fmt.Println(len(ParseScriptsIndex(currBlock)), "scripts available")

				case "DSOU":
					//fmt.Println("Parse Directory of Sounds")

				case "DCOS":
					//fmt.Println("Parse Directory of Costumes")

				case "DCHR":
					//fmt.Println("Parse Directory of Charsets")

				case "DOBJ":
					//fmt.Println("Parse Directory of Objects")

				}

			}
		}

		if extension == ".001" { // MAIN FILE
			mainScumm := MainScummData{data}

			game.RoomCount = mainScumm.GetRoomCount()
			roomOffsets := mainScumm.GetRoomsOffset()

			fmt.Println("Room count", game.RoomCount)

			//singleRoom := 55
			//for i := singleRoom; i < singleRoom+1; i++ {
			for i := 1; i < mainScumm.GetRoomCount(); i++ {
				room := mainScumm.ParseRoom(roomOffsets[i-1].Offset)
				game.Rooms = append(game.Rooms, room)
			}
		}
	}

	return &game
}
