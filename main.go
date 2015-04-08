package main

import (
	"fileutils"
	"fmt"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"scummatlas"
	"scummatlas/templates"
	"strconv"
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
	paramsOffset := 1
	singleRoom := -1
	if os.Args[1] == "--room" {
		paramsOffset = 3
		room, err := strconv.Atoi(os.Args[2])
		if err != nil {
			panic("Room needs to be a number")
		}
		singleRoom = room
	}
	gamedir := os.Args[paramsOffset]
	outputdir := os.Args[paramsOffset+1]
	log.SetFlags(0)

	fmt.Println("Gamedir: ", gamedir)
	fmt.Println("Outputdir: ", outputdir)
	_, err := ioutil.ReadDir(outputdir)
	if err != nil {
		err := os.Mkdir(outputdir, 0755)
		if err != nil {
			helpAndDie("Output directory doesn't exist and I can't create it.")
		}
	}

	game := scummatlas.NewGame(gamedir)
	game.ProcessIndex(outputdir)
	game.ProcessAllRooms(outputdir)

	templates.WriteIndex(game.RoomNames, outputdir)

	copyStaticFiles(outputdir)

	processRoom := func(i int, room scummatlas.Room) {
		templates.WriteRoom(room, i, outputdir)
		writeRoomBackground(i, room, outputdir)
		createRoomObjectImages(i, room, outputdir)
		for _, obj := range room.Objects {
			obj.PrintVerbs()
		}
	}

	if singleRoom != -1 {
		fmt.Println("Processing room ", singleRoom)
		processRoom(singleRoom, game.Rooms[singleRoom])
	} else {
		for i, room := range game.Rooms {
			processRoom(i, room)
		}
	}

}

func copyStaticFiles(outputdir string) {
	fileutils.CopyDir("./static", outputdir+"/static")
}

func writeRoomBackground(id int, room scummatlas.Room, outputdir string) {
	backgroundFile := fmt.Sprintf("%v/room%02d_bg.png", outputdir, id)
	fmt.Printf("\nWriting room %v background in %v", id, backgroundFile)
	pngFile, err := os.Create(backgroundFile)
	if err != nil {
		panic("Error creating " + backgroundFile)
	}
	png.Encode(pngFile, room.Image)
}

func createRoomObjectImages(id int, r scummatlas.Room, outputdir string) {
	for _, object := range r.Objects {
		if len(object.Image.Frames) == 0 {
			//fmt.Printf("Obj %v does not have an image\n", object.Id)
			continue
		}

		for frameIndex, frame := range object.Image.Frames {
			imagePath := fmt.Sprintf(
				"%v/room%02d_obj_%02x_%02d.png",
				outputdir,
				id,
				object.Id,
				frameIndex)

			pngFile, err := os.Create(imagePath)
			if err != nil {
				panic("Error creating " + imagePath)
			}
			png.Encode(pngFile, frame)
			fmt.Printf("Obj image %v created\n", imagePath)
		}
	}
}
