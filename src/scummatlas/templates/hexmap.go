package templates

import "html/template"
import "scummatlas/blocks"
import "os"
import "fmt"

type HexCol struct {
	Value byte
	Class string
	Title string
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

func WriteHexMap(hexmap blocks.HexMap, outputfile string) {
	t := template.Must(template.ParseFiles(
		htmlPath+"hexdump.html",
		htmlPath+"partials.html"))

	colours := []int{
		/* Nice pastel */
		0x68b0b0,
		0xf47d7e,
		0xb5d045,
		0xfb8335,
		0x81c0c5,
		0xe0c7a8,
		/* Variations */
		0xee9999,
		0x99ee99,
		0x9999ee,
		0xeeee99,
		0x99eeee,
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
				column.Title = section.Type + " " + section.Description
				break
			}
		}
		rows[curRow][curCol] = column
	}
	for _, section := range sections {
		start := section.Start
		end := section.Start + section.Length - 1

		row := start / HexDumpRowSize
		col := start % HexDumpRowSize
		//fmt.Printf("Start of section %v in [%v,%v]\n", start, row, col)
		rows[row][col].Class += " start"

		row = end / HexDumpRowSize
		col = end % HexDumpRowSize
		//fmt.Printf("End of section %v in [%v,%v]\n", end, row, col)
		rows[row][col].Class += " end"
	}
	data.Rows = rows

	t.Execute(file, data)
}
