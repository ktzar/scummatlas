package scummatlas

import (
	"encoding/binary"
	"errors"
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
	gamedir   string
	indexFile string
	mainFile  string
}

const DEBUG_SAVE_DECODED = true
const V5_KEY = 0x69

func NewGame(gamedir string) *Game {

	game := Game{
		gamedir: gamedir,
	}

	filesInfo, err := ioutil.ReadDir(gamedir)
	if err != nil {
		panic("Game directory not a directory")
	}

	// Read index file
	for _, file := range filesInfo {
		fileName := gamedir + "/" + file.Name()
		extension := filepath.Ext(fileName)

		if strings.Contains(extension, ".00") {
			if extension == ".000" {
				game.indexFile = file.Name()
			}

			if extension == ".001" { // MAIN FILE
				game.mainFile = file.Name()
			}
		}
	}

	return &game
}

func (self *Game) ProcessIndex(outputdir string) error {
	if self.indexFile == "" {
		return errors.New("No index file")
	}

	data, err := ReadXoredFile(self.gamedir+"/"+self.indexFile, V5_KEY)
	if DEBUG_SAVE_DECODED {
		if err != nil {
			panic("Can't read " + self.indexFile)
		}
		f, _ := os.Create(outputdir + "/" + self.indexFile + ".decoded")
		defer f.Close()
		f.Write(data)
	}

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
			self.RoomNames = ParseRoomNames(currBlock)

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

	return nil
}

func (self *Game) processMainFile(outputdir string) MainScummData {
	if self.mainFile == "" {
		panic("No main file")
	}

	data, err := ReadXoredFile(self.gamedir+"/"+self.mainFile, V5_KEY)
	if DEBUG_SAVE_DECODED {
		if err != nil {
			panic("Can't read " + self.indexFile)
		}
		f, _ := os.Create(outputdir + "/" + self.indexFile + ".decoded")
		defer f.Close()
		f.Write(data)
	}

	mainScumm := MainScummData{data}

	self.RoomCount = mainScumm.GetRoomCount()
	fmt.Println("Room count", self.RoomCount)

	self.Rooms = make([]Room, self.RoomCount)

	return mainScumm
}

func (self *Game) ProcessAllRooms(outputdir string) {
	mainScumm := self.processMainFile(outputdir)
	roomOffsets := mainScumm.GetRoomsOffset()
	for i := 0; i < mainScumm.GetRoomCount()-1; i++ {
		room := mainScumm.ParseRoom(roomOffsets[i].Offset)
		self.Rooms[i] = room
	}
}

func (self *Game) ProcessSingleRoom(room int, outputdir string) {
	mainScumm := self.processMainFile(outputdir)
	roomOffsets := mainScumm.GetRoomsOffset()
	roomData := mainScumm.ParseRoom(roomOffsets[room-1].Offset)
	self.Rooms[room-1] = roomData
}
