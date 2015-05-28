package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	b "scummatlas/binaryutils"
	s "scummatlas/script"
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

	script, err := parseScript(data)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println(script.Debug())
}

func parseScript(data []byte) (s.Script, error) {
	blockType := b.FourCharString(data, 0)
	var script s.Script
	initialOffset := -1
	switch blockType {
	case "SCRP", "VERB":
		initialOffset = 8
	case "LSCR":
		initialOffset = 9
	}
	if initialOffset == -1 {
		return nil, errors.New(blockType + "is not a supported script block")
	}
	script = s.ParseScriptBlock(data[initialOffset:])
	return script, nil
}
