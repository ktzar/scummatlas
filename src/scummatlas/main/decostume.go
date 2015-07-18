package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"scummatlas"
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
	templates.WriteHexMap(costume.HexMap, "costume_bytes.html")
	costume.Debug()
}

func parseCostume(data []byte) (*scummatlas.Costume, error) {
	blockType := b.FourCharString(data, 0)
	if blockType != "COST" {
		return nil, errors.New(blockType + "is not a supported costume block")
	}
	costume := scummatlas.NewCostume(data[8:])
	return costume, nil
}
