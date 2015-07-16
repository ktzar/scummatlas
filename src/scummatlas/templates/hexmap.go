package templates

import "html/template"
import "scummatlas"
import "os"
import "fmt"

type HexCol struct {
	Value byte
	Class string
}

type HexMapData struct {
	Title   string
	Colours map[string]int
	Rows    [][]HexCol
}

func (h HexMapData) RowAddress(row int) string {
	return fmt.Sprintf("0x%04x", row*HexDumpRowSize)
}

const HexDumpRowSize = 32

func WriteHexMap(hexmap scummatlas.HexMap, outputfile string) {
	t := template.Must(template.ParseFiles(
		htmlPath+"hexdump.html",
		htmlPath+"partials.html"))

	colours := []int{
		/* Nice pastel */
		0x389090,
		0xf47d7e,
		0xb5d045,
		0xfb8335,
		0x81c0c5,
		0xe0c7a8,
		/* Variations */
		0xffdddd,
		0xddffdd,
		0xddddff,
		0xffffdd,
		0xddffff,
	}

	file, err := os.Create(outputfile)
	if err != nil {
		panic("Can't create hexdump file, " + err.Error())
	}

	data := HexMapData{Title: "Some hex"}
	rowsCount := (len(hexmap.Data()) / HexDumpRowSize) + 1
	rows := make([][]HexCol, rowsCount)
	for i := range rows {
		rows[i] = make([]HexCol, HexDumpRowSize)
	}

	classes := make(map[string]int)

	curColour := 0
	curRow := 0

	sections := hexmap.Sections()

	for _, section := range sections {
		_, colourExists := classes[section.Type]
		if !colourExists {
			classes[section.Type] = colours[curColour%len(colours)]
			curColour++
		}
	}
	data.Colours = classes

	for i, octet := range hexmap.Data() {
		curCol := i % HexDumpRowSize
		if i > 0 && curCol == 0 {
			curRow++
		}
		column := HexCol{Value: octet}
		for _, section := range sections {
			if section.IncludesOffset(i) {
				column.Class = section.Type
				break
			}
		}
		rows[curRow][curCol] = column
	}
	fmt.Println(rows)
	data.Rows = rows

	t.Execute(file, data)
}
