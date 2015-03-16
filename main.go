package main

import (
	"fmt"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"scummatlas"
	"scummatlas/templates"
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

	game := scummatlas.NewGame(gamedir, outputdir)

	templates.WriteIndex(game.RoomNames, outputdir)

	for i, room := range game.Rooms {
		backgroundFile := fmt.Sprintf("%v/room%02d_bg.png", outputdir, i)
		fmt.Printf("\nParsing room %v, file %v", i, backgroundFile)
		pngFile, err := os.Create(backgroundFile)
		if err != nil {
			panic("Error creating " + backgroundFile)
		}
		templates.WriteRoom(room, i, outputdir)
		png.Encode(pngFile, room.Image)

	}

}
