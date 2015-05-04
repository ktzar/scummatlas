package main

import (
	"fileutils"
	"flag"
	"fmt"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"scummatlas"
	l "scummatlas/condlog"
	"scummatlas/templates"
)

func main() {
	var gamedir string
	var outputdir string
	var singleRoom int
	var noimages bool
	flag.StringVar(&gamedir, "gamedir", "", "Directory with the game files")
	flag.StringVar(&outputdir, "outputdir", "", "Directory to put the generated files in")
	flag.IntVar(&singleRoom, "room", 0, "Only parse one room")
	flag.BoolVar(&noimages, "noimages", false, "Don't create images")
	flag.Parse()

	//Check for values
	if outputdir == "" || gamedir == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	log.SetFlags(0)

	_, err := ioutil.ReadDir(outputdir)
	if err != nil {
		err := os.Mkdir(outputdir, 0755)
		if err != nil {
			fmt.Println("Output directory doesn't exist and I can't create it.")
			os.Exit(1)
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
		if noimages {
			return
		}
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

	for i, zplane := range room.Zplanes {
		filename := fmt.Sprintf("%v-zplane%d.png", backgroundFile, i+1)
		pngFile, err := os.Create(filename)
		if err != nil {
			panic("Error creating " + filename)
		}
		png.Encode(pngFile, zplane)
	}
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
