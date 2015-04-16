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

func WriteIndex(roomNames []scummatlas.RoomName, outdir string) {

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
