package main

import (
	"fmt"
	//"bytes"
	"encoding/binary"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"scummatlas"
	"scummatlas/templates"
	"strings"
)

func helpAndDie(msg string) {
	fmt.Println(msg)
	fmt.Println("Usage:")
	fmt.Println("scummatlas [gamedir] [outputdir]")
	os.Exit(1)
}

func main() {
	if len(os.Args) < 3 {
		helpAndDie("Not enough arguments")
	}
	gamedir := os.Args[1]
	outputdir := os.Args[2]

	fmt.Println("Gamedir: ", gamedir)
	fmt.Println("Outputdir: ", outputdir)
	_, err := ioutil.ReadDir(outputdir)
	if err != nil {
		err := os.Mkdir(outputdir, 0755)
		if err != nil {
			helpAndDie("Output directory doesn't exist and I can't create it.")
		}
	}

	filesInfo, err := ioutil.ReadDir(gamedir)
	if err != nil {
		helpAndDie("Game directory not a directory")
	}

	// Read index file
	for _, file := range filesInfo {
		fileName := gamedir + "/" + file.Name()
		extension := filepath.Ext(fileName)
		//fmt.Println(extension)
		data := []byte{}
		if strings.Contains(extension, ".00") {
			absPath, _ := filepath.Abs(gamedir + "/" + file.Name())
			fmt.Println(absPath)
			data, err = scummatlas.ReadXoredFile(absPath, 0x69)
			if err != nil {
				helpAndDie("Can't read index file")
			}

			//f, _ := os.Create(outputdir + "/" + file.Name() + ".decoded")
			//defer f.Close()
			//f.Write(data)
		}
		if extension == ".001" {
			continue
			mainScumm := new(scummatlas.MainScummData)
			mainScumm.Data = data
			fmt.Println(mainScumm.GetRoomCount())
			roomOffsets := mainScumm.GetRoomsOffset()
			fmt.Println(roomOffsets)
			for i := 0; i < mainScumm.GetRoomCount(); i++ {
				i = 53
				backgroundFile := fmt.Sprintf("%v/room%02d_bg.png", outputdir, i)
				fmt.Printf("\nParsing room %v, file %v", i, backgroundFile)
				jpegFile, err := os.Create(backgroundFile)
				if err != nil {
					panic("Error creating " + backgroundFile)
				}

				room := mainScumm.ParseRoom(roomOffsets[i].Offset)
				png.Encode(jpegFile, room.Image)
				os.Exit(1)
			}
			os.Exit(1)
		}
		if extension == ".000" {
			currIndex := 0
			for currIndex < len(data) {
				blockName := string(data[currIndex : currIndex+4])
				blockSize := int(binary.BigEndian.Uint32(data[currIndex+4 : currIndex+8]))
				currBlock := data[currIndex : currIndex+blockSize]

				//fmt.Println("Block ", blockName, "\t", blockSize, "bytes")

				currIndex += blockSize
				//TODO Remove
				//continue
				switch blockName {
				case "RNAM":
					fmt.Println("Parse Room Names")
					roomNames := scummatlas.ParseRoomNames(currBlock)
					templates.WriteIndex(roomNames, outputdir)

				case "MAXS":
					//fmt.Println("Parse Maximum Values")

				case "DROO":
					//fmt.Println("Parse Directory of Rooms")
					//fmt.Println(scummatlas.ParseRoomIndex(currBlock))

				case "DSCR":
					//fmt.Println("Parse Directory of Scripts")
					//fmt.Println(len(scummatlas.ParseScriptsIndex(currBlock)), "scripts available")

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
	}

}
