package main

import (
	"fmt"
	//"bytes"
	"encoding/binary"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"scummatlas"
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


}
