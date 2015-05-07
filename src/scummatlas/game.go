package scummatlas

import (
	"encoding/binary"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	b "scummatlas/binaryutils"
	l "scummatlas/condlog"
	"strings"
)

type Game struct {
	Name        string
	RoomOffsets []RoomOffset
	RoomNames   []RoomName
	RoomIndexes []int
	Rooms       []Room
	Scripts     []Script
	Costumes    []Costume
	gamedir     string
	indexFile   string
	mainFile    string
	mainData    MainScummData
}

type Costume interface{}

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

			if extension == ".001" {
				game.mainFile = file.Name()
			}
		}
	}
	game.processIndex()
	game.processMainFile()
	return &game
}

func (self *Game) ProcessAllRooms(outputdir string) {
	roomDone := make(chan int)

	for i, offset := range self.RoomOffsets {
		go func(i int, offset RoomOffset) {
			l.Log("game", "Parsing room %d with Id: %x", i, offset.Id)
			room := self.mainData.ParseRoom(offset.Offset, i)
			room.Id = offset.Id
			room.Name = self.RoomNames[i].Name
			self.Rooms[i] = room
			roomDone <- room.Id
		}(i, offset)
	}
	roomsDone := 0
	for roomsDone < len(self.RoomOffsets) {
		i := <-roomDone
		l.Log("game", "Room %d finished processing\n", i)
		roomsDone++
	}
}

func (self *Game) ProcessSingleRoom(i int, outputdir string) {
	offset := self.RoomOffsets[i]
	l.Log("game", "Parsing room %d with Id: %x", i, offset.Id)
	room := self.mainData.ParseRoom(offset.Offset, i)
	if len(self.RoomNames) > i {
		room.Id = offset.Id
		room.Name = self.RoomNames[i].Name
	}
	self.Rooms[i] = room
}

func (self *Game) DumpDecoded(outputdir string) {
	if self.indexFile != "" {
		self.dumpFile(self.indexFile, outputdir)
		self.dumpFile(self.mainFile, outputdir)
	}
}

func (self *Game) processMainFile() {
	if self.mainFile == "" {
		panic("No main file")
	}

	data, err := b.ReadXoredFile(self.gamedir+"/"+self.mainFile, V5_KEY)
	if err != nil {
		panic("Cannot read main file")
	}

	self.mainData = MainScummData{data}
	self.Scripts = self.mainData.GetScripts()
	self.RoomOffsets = self.mainData.GetRoomsOffset()
	self.Rooms = make([]Room, len(self.RoomOffsets))

	l.Log("structure", "Room count", len(self.RoomOffsets))
}

func (self Game) dumpFile(file string, outputdir string) error {
	if file == "" {
		return errors.New("File does not exist")
	}
	data, err := b.ReadXoredFile(self.gamedir+"/"+file, V5_KEY)
	if err != nil {
		panic("Can't read " + file)
	}
	f, _ := os.Create(outputdir + "/" + file + ".decoded")
	defer f.Close()
	f.Write(data)

	return nil
}

func (self *Game) processIndex() error {
	if self.indexFile == "" {
		return errors.New("No index file")
	}

	data, err := b.ReadXoredFile(self.gamedir+"/"+self.indexFile, V5_KEY)
	if err != nil {
		return err
	}

	currIndex := 0
	for currIndex < len(data) {
		blockName := string(data[currIndex : currIndex+4])
		blockSize := int(binary.BigEndian.Uint32(data[currIndex+4 : currIndex+8]))
		currBlock := data[currIndex : currIndex+blockSize]

		currIndex += blockSize
		l.Log("structure", "Parse Block", blockName)
		switch blockName {
		case "RNAM":
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
