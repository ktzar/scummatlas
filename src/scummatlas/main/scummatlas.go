package main

import (
	"flag"
	"fmt"
	"image/png"
	"image"
	"io/ioutil"
	"os"
	"scummatlas"
	"scummatlas/utils"
	"scummatlas/blocks"
	l "scummatlas/condlog"
	"scummatlas/templates"
	"strings"
	"sync"
)

const REQUIRED string = "REQUIRED"

var gamedir string
var outputdir string
var singleRoom int
var noimages bool
var multicpu bool
var dumpdecoded bool

func main() {
	loadOptions()

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
	os.MkdirAll(outputdir+"/img_cost", 0755)

	game := scummatlas.NewGame(gamedir)
	if dumpdecoded {
		game.DumpDecoded(outputdir)
	}

	if singleRoom > 0 {
		game.ProcessSingleRoom(singleRoom, outputdir)
	} else {
		game.ProcessAllRooms(outputdir)
	}

	templates.WriteGameFiles(*game, outputdir)

	copyStaticFiles()

	if singleRoom > 0 {
		processRoom(game.Rooms[singleRoom])
	} else {
		var wg sync.WaitGroup
		for _, room := range game.Rooms {
			if multicpu {
				wg.Add(1)
				go func(room blocks.Room) {
					processRoom(room)
					fmt.Printf("Room %v processed\n", room.Id)
					if multicpu {
						wg.Done()
					}
				}(room)
			} else {
				processRoom(room)
			}
		}
		wg.Wait()
	}

	for costumeId, costume := range game.Costumes {
		templates.WriteCostume(costumeId, costume, outputdir)
		for limbId, limb := range costume.Limbs {
			if limb.Image != nil {
				writeCostumeLimb(costumeId, limbId, limb.Image, outputdir)
			}
		}
		fmt.Printf("Files for costume %d generated\n", costumeId)
	}
}

func loadOptions() {
	var logflags string

	logkeys := make([]string, 0, len(l.Flags))
	for k := range l.Flags {
		logkeys = append(logkeys, k)
	}
	flag.StringVar(&gamedir, "gamedir", REQUIRED, "Directory with the game files")
	flag.StringVar(&outputdir, "outputdir", REQUIRED, "Directory to put the generated files in")
	flag.IntVar(&singleRoom, "room", 0, "Only parse one room")
	flag.BoolVar(&noimages, "noimages", false, "Don't create images")
	flag.BoolVar(&multicpu, "multicpu", false, "Use multiple processes")
	flag.BoolVar(&dumpdecoded, "dumpdecoded", false, "Dump decoded .000 and .001")
	flag.StringVar(&logflags, "logflags", "", "Comma separated list of log flags. Available flags: "+strings.Join(logkeys, ", "))
	flag.Parse()

	if outputdir == REQUIRED || gamedir == REQUIRED {
		fmt.Println("outputdir and gamedir are not optional\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	flags := strings.Split(logflags, ",")
	for _, flag := range flags {
		l.Flags[flag] = true
	}
}

func processRoom(room blocks.Room) {
	templates.WriteRoom(room, outputdir)
	if noimages {
		return
	}
	writeRoomBackground(room, outputdir)
	createRoomObjectImages(room)
	for _, obj := range room.Objects {
		obj.PrintVerbs()
	}
	fmt.Printf("Files for room %d generated\n", room.Id)
}

func copyStaticFiles() {
	utils.CopyDir("./static", outputdir+"/static")
}

func writeCostumeLimb(costume int, limb int, img *image.Paletted, outputdir string) {
	fileName := fmt.Sprintf("%v/img_cost/%v_%v.png", outputdir, costume, limb)
	pngFile, err := os.Create(fileName);
	if err != nil {
		panic("Error creating " + fileName + ".png")
	}
	if img != nil {
		png.Encode(pngFile, img)
	}
}

func writeRoomBackground(room blocks.Room, outputdir string) {
	backgroundFile := fmt.Sprintf("%v/img_bg/room%02d_bg", outputdir, room.Id)
	l.Log("template", "\nWriting room %v background in %v\n", room.Id, backgroundFile+".png")
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

func createRoomObjectImages(r blocks.Room) {
	for _, object := range r.Objects {
		if len(object.Image.Frames) == 0 {
			//fmt.Printf("Obj %v does not have an image\n", object.Id)
			continue
		}

		for frameIndex, frame := range object.Image.Frames {
			if frame == nil {
				continue
			}
			imagePath := fmt.Sprintf(
				"%v/img_obj/room%02d_obj_%02x_%02d.png",
				outputdir,
				r.Id,
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
