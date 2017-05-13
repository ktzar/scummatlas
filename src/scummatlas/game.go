package scummatlas

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"scummatlas/utils"
	b "scummatlas/binaryutils"
	l "scummatlas/condlog"
	s "scummatlas/script"
	"scummatlas/blocks"
)

type GameMetaData struct {
	ScummVersion int
	Name         string
	Variant      string
	Language     string
}

var GamesHashes = map[string]GameMetaData{
	"2d1e891fe52df707c30185e52c50cd92": GameMetaData{5, "The Secret of Monkey Island", "CD", "en"},
	"c0c9de81fb965e6cbe77f6e5631ca705": GameMetaData{5, "The Secret of Monkey Island", "Talkie", "en"},
	"3686cf8f89e102ececf4366e1d2c8126": GameMetaData{5, "Monkey Island: Lechuck's Revenge", "Floppy", "en"},
	"182344899c2e2998fca0bebcd82aa81a": GameMetaData{5, "Indiana Jones and the Fate of Atlantis", "CD", "en"},
	"4167a92a1d46baa4f4127d918d561f88": GameMetaData{6, "The Day of the Tentacle", "CD", "en"},
	"d917f311a448e3cc7239c31bddb00dd2": GameMetaData{6, "Sam & Max Hit the Road", "CD", "en"},
	"d8323015ecb8b10bf53474f6e6b0ae33": GameMetaData{7, "The Dig", "CD", "en"},
}

type Game struct {
	RoomOffsets []RoomOffset
	CostumeIndex []IndexItem
	RoomNames   []RoomName
	RoomIndexes []int
	Rooms       []blocks.Room
	Scripts     []s.Script
	Costumes    []blocks.Costume
	gamedir     string
	indexFile   string
	mainFile    string
	mainData    *MainScummData
	GameMetaData
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

		if extension == ".000" || extension == ".LA0" {
			game.indexFile = file.Name()
		}

		if extension == ".001" || extension == ".LA1" {
			game.mainFile = file.Name()
		}
	}

	game.processIndex()
	fmt.Println("Costume\tRoom\tOffset")
	for i, c := range(game.CostumeIndex) {
		fmt.Printf("%v\t%v\t%v\n", i, c.RoomNumber, c.Offset)
	}
	game.processMainFile()
	return &game
}

func (self *Game) inferName() {
	md5, _ := utils.ComputeMd5(self.gamedir + "/" + self.indexFile)
	strMd5 := fmt.Sprintf("%x", md5)

	for hash, metaData := range GamesHashes {
		if hash == strMd5 {
			self.GameMetaData = metaData
			break
		}
	}
}

func (self *Game) ProcessAllRooms(outputdir string) {
	roomDone := make(chan int)

	for i, offset := range self.RoomOffsets {
		go func(i int, offset RoomOffset) {
			l.Log("game", "Parsing room %d with Id: %x", i, offset.Id)
			room := self.mainData.ParseRoom(offset.Offset, i)
			room.Id = offset.Id
			if len(self.RoomNames) > i {
				room.Name = self.RoomNames[i].Name
			} else {
				room.Name = ""
			}
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

	self.mainData = NewMainScummData(data)
	self.Scripts = self.mainData.GetScripts()
	self.RoomOffsets = self.mainData.GetRoomsOffset()
	self.Rooms = make([]blocks.Room, len(self.RoomOffsets))
	self.Scripts = self.mainData.GetScripts()
	self.Costumes = self.mainData.GetCostumes()

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
			//fmt.Println(len(ParseIndexBlock(currBlock)), "scripts available")

		case "DSOU":
			//fmt.Println("Parse Directory of Sounds")

		case "DCOS":
			//fmt.Println("Parse Directory of Costumes")
			self.CostumeIndex = ParseIndexBlock(currBlock)

		case "DCHR":
			//fmt.Println("Parse Directory of Charsets")

		case "DOBJ":
			//fmt.Println("Parse Directory of Objects")

		}
	}

	self.inferName()

	return nil
}
