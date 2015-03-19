package templates

import (
	"html/template"
	"io/ioutil"
	"os"
	"scummatlas"
)

type IndexData struct {
	Title string
	Rooms []scummatlas.RoomName
}

func WriteIndex(roomNames []scummatlas.RoomName, outdir string) {

	indexTpl, err := ioutil.ReadFile("./templates/index.html")
	if err != nil {
		panic("No index.html in the templates directory")
	}

	data := IndexData{
		"A game",
		roomNames,
	}
	t := template.Must(template.New("index").Parse(string(indexTpl)))

	file, err := os.Create(outdir + "/index.html")
	if err != nil {
		panic("Can't create index file")
	}

	t.Execute(file, data)
}
