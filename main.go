package main

import (
    "fmt"
    //"bytes"
    "io/ioutil"
    "path/filepath"
    "encoding/binary"
    "os"
    "scummatlas"
    //"strings"
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

    fmt.Println("Gamedir: ", gamedir)
    fmt.Println("Outputdir: ", outputdir)
    _, err := ioutil.ReadDir(outputdir)
    if err != nil {
        err := os.Mkdir(outputdir, 0755)
        if err != nil {
            helpAndDie("Output directory doesn't exist and I can't create it.")
        }
    }

    filesInfo, err := ioutil.ReadDir(gamedir)
    if err != nil {
        helpAndDie("Game directory not a directory")
    }

    // Read index file
    for _, file := range(filesInfo) {
        fileName := gamedir + "/" + file.Name()
        extension := filepath.Ext(fileName)
        //fmt.Println(extension)
        data := []byte {}
        if extension == ".000" {
            absPath, _ := filepath.Abs(gamedir + "/" + file.Name())
            fmt.Println(absPath)
            data, err = scummatlas.ReadXoredFile(absPath, 0x69) 
            if err != nil {
                helpAndDie("Can't read index file")
            }

            //f, _ := os.Create(outputdir + "/" + file.Name() + ".decoded")
            //defer f.Close()
            //f.Write(data)
        }
        if extension == ".000" {
            currIndex := 0
            for currIndex < len(data) {
                blockName := string(data[currIndex:currIndex+4])
                blockSize := int(binary.BigEndian.Uint32(data[currIndex+4:currIndex+8]))
                currBlock := data[currIndex:currIndex+blockSize]

                fmt.Println("Block ", blockName, "\t", blockSize, "bytes");

                switch blockName {
                    case "RNAM":
                        fmt.Println("Parse Room Names")
                        rooms := scummatlas.ParseRoomNames(currBlock)
                        fmt.Println(rooms)

                    case "MAXS":
                        fmt.Println("Parse Maximum Values")

                    case "DROO":
                        fmt.Println("Parse Directory of Rooms")
                        rooms := scummatlas.ParseRoomIndex(currBlock)
                        fmt.Println(rooms)

                    case "DSCR":
                        fmt.Println("Parse Directory of Scripts")

                    case "DSOU":
                        fmt.Println("Parse Directory of Sounds")

                    case "DCOS":
                        fmt.Println("Parse Directory of Costumes")

                    case "DCHR":
                        fmt.Println("Parse Directory of Charsets")

                    case "DOBJ":
                        fmt.Println("Parse Directory of Objects")

                }

                currIndex += blockSize
            }
        }
    }

}
