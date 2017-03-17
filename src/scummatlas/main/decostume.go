package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"image/png"
	"image/color/palette"
	"os"
	"scummatlas/blocks"
	b "scummatlas/binaryutils"
	"scummatlas/templates"
)

func main() {
	var infile string

	flag.StringVar(&infile, "in", "REQUIRED", "File to parse")
	flag.Parse()

	if infile == "REQUIRED" {
		fmt.Println("missing compulsory parameter\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	data, err := ioutil.ReadFile(infile)
	if err != nil {
		fmt.Println("Can't read the file", infile)
		os.Exit(1)
	}

	costume, err := parseCostume(data)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	for i, limb := range costume.Limbs {
		fileName := fmt.Sprintf("limb_%v.png", i)
		pngFile, err := os.Create(fileName);
		if err != nil {
			panic("Error creating " + fileName + ".png")
		}
		if limb.Image != nil {
			png.Encode(pngFile, limb.Image)
		}
	}
	templates.WriteHexMap(costume.HexMap, "costume_bytes.html")
	costume.Debug()
}

func parseCostume(data []byte) (*blocks.Costume, error) {
	blockType := b.FourCharString(data, 0)
	if blockType != "COST" {
		return nil, errors.New(blockType + "is not a supported costume block")
	}
	fmt.Println(palette.Plan9)
	costume := blocks.NewCostume(data[8:], palette.Plan9)
	return costume, nil
}
