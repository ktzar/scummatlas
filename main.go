package main

import (
	"fileutils"
	"fmt"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"scummatlas"
	l "scummatlas/condlog"
	"scummatlas/templates"
	"strconv"
)

func helpAndDie(msg string) {
	fmt.Println(msg)
	fmt.Println("Usage: scummatlas [options] gamedir outputdir")
	fmt.Println("Options:")
	fmt.Println("  --room roomNumber\tProcess a single room")
	os.Exit(0)
}

func main() {
	var singleRoom int

	paramsOffset := 1
	if os.Args[1] == "--room" {
		paramsOffset = 3
		room, err := strconv.Atoi(os.Args[2])
		if err != nil {
			panic("Room needs to be a number")
		}
		singleRoom = room
	}
	if len(os.Args) < paramsOffset+2 {
		helpAndDie("Not enough arguments")
	}
	gamedir := os.Args[paramsOffset]
	outputdir := os.Args[paramsOffset+1]
	log.SetFlags(0)

	_, err := ioutil.ReadDir(outputdir)
	if err != nil {
		err := os.Mkdir(outputdir, 0755)
		if err != nil {
			helpAndDie("Output directory doesn't exist and I can't create it.")
		}
	}
	os.MkdirAll(outputdir+"/img_obj", 0755)
	os.MkdirAll(outputdir+"/img_bg", 0755)

	game := scummatlas.NewGame(gamedir)
	game.ProcessIndex(outputdir)
	if singleRoom < 1 {
		game.ProcessAllRooms(outputdir)
	} else {
		game.ProcessSingleRoom(singleRoom+1, outputdir)
	}

	templates.WriteIndex(game, outputdir)
	templates.WriteTable(game, outputdir)

	copyStaticFiles(outputdir)

	processRoom := func(room scummatlas.Room) {
		fmt.Println("Generate files for room ", room.Number)
		templates.WriteRoom(room, outputdir)
		writeRoomBackground(room, outputdir)
		createRoomObjectImages(room, outputdir)
		for _, obj := range room.Objects {
			obj.PrintVerbs()
		}
	}

	if singleRoom > 0 {
		processRoom(game.Rooms[singleRoom])
	} else {
		for _, room := range game.Rooms {
			processRoom(room)
		}
	}

}

func copyStaticFiles(outputdir string) {
	fileutils.CopyDir("./static", outputdir+"/static")
}

func writeRoomBackground(room scummatlas.Room, outputdir string) {
	backgroundFile := fmt.Sprintf("%v/img_bg/room%02d_bg", outputdir, room.Number)
	l.Log("template", "\nWriting room %v background in %v\n", room.Number, backgroundFile+".png")
	pngFile, err := os.Create(backgroundFile + ".png")
	if err != nil {
		panic("Error creating " + backgroundFile + ".png")
	}
	png.Encode(pngFile, room.Image)
}

func createRoomObjectImages(r scummatlas.Room, outputdir string) {
	for _, object := range r.Objects {
		if len(object.Image.Frames) == 0 {
			//fmt.Printf("Obj %v does not have an image\n", object.Id)
			continue
		}

		for frameIndex, frame := range object.Image.Frames {
			imagePath := fmt.Sprintf(
				"%v/img_obj/room%02d_obj_%02x_%02d.png",
				outputdir,
				r.Number,
				object.Id,
				frameIndex)

			pngFile, err := os.Create(imagePath)
			if err != nil {
				panic("Error creating " + imagePath)
			}
			png.Encode(pngFile, frame)
			l.Log("template", "Obj image %v created\n", imagePath)
		}
	}
}
