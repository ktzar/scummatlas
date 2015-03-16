package scummatlas

type Game struct {
	Name      string
	RoomCount int
	Rooms     []Room
}

const DEBUG_SAVE_DECODED = true

func NewGame(gamedir string) *Game {

	var game Game

	filesInfo, err := ioutil.ReadDir(gamedir)
	if err != nil {
		helpAndDie("Game directory not a directory")
	}

	// Read index file
	for _, file := range filesInfo {
		fileName := gamedir + "/" + file.Name()
		extension := filepath.Ext(fileName)
		data := []byte{}

		if strings.Contains(extension, ".00") {
			absPath, _ := filepath.Abs(gamedir + "/" + file.Name())
			data, err = scummatlas.ReadXoredFile(absPath, 0x69)
			if err != nil {
				helpAndDie("Can't read index file")
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

		if extension == ".001" { // MAIN FILE
			mainScumm := scummatlas.MainScummData {data}

			game.RoomCount = mainScumm.GetRoomCount()
			roomOffsets := mainScumm.GetRoomsOffset()

			fmt.Println("Room count", game.RoomCount)

			for i := 1; i < mainScumm.GetRoomCount(); i++ {
				backgroundFile := fmt.Sprintf("%v/room%02d_bg.png", outputdir, i)
				fmt.Printf("\nParsing room %v, file %v", i, backgroundFile)
				pngFile, err := os.Create(backgroundFile)
				if err != nil {
					panic("Error creating " + backgroundFile)
				}

				room := mainScumm.ParseRoom(roomOffsets[i-1].Offset)
				game.Rooms = append(game.Rooms, room)
				templates.WriteRoom(room, i, outputdir)
				png.Encode(pngFile, room.Image)
			}
			//os.Exit(1)
		}
	}

	return game
}
