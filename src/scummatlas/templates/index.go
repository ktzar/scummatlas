package templates

import (
	"html/template"
	"io/ioutil"
	"os"
	"scummatlas"
	l "scummatlas/condlog"
)

type IndexData struct {
	Title string
	Rooms []scummatlas.RoomName
}

type TableData struct {
	Title string
	Rooms []scummatlas.Room
}

func WriteIndex(game *scummatlas.Game, outdir string) {

	roomNames := game.RoomNames

	//TODO Cache that for the future
	indexTpl, err := ioutil.ReadFile("./templates/index.html")
	if err != nil {
		panic("No index.html in the templates directory")
	}

	data := IndexData{
		"A game",
		roomNames,
	}
	t := template.Must(template.New("index").Parse(string(indexTpl)))

	filename := outdir + "/index.html"
	file, err := os.Create(filename)
	l.Log("template", "Create "+filename)
	if err != nil {
		panic("Can't create index file")
	}

	t.Execute(file, data)
}

func WriteTable(game *scummatlas.Game, outdir string) {

	//roomNames := game.RoomNames
	rooms := game.Rooms

	//TODO Cache that for the future
	tableTpl, err := ioutil.ReadFile("./templates/table.html")
	if err != nil {
		panic("No table.html in the templates directory")
	}

	data := TableData{
		"A game",
		rooms,
	}
	t := template.Must(template.New("table").Parse(string(tableTpl)))

	filename := outdir + "/table.html"
	file, err := os.Create(filename)
	l.Log("template", "Create "+filename)
	if err != nil {
		panic("Can't create table file")
	}

	t.Execute(file, data)
}
