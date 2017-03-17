package templates

import (
	"fmt"
	"html/template"
	"os"
	"scummatlas/blocks"
	l "scummatlas/condlog"
)

type costumeData struct {
	CostumeIndex int
	Title string
	blocks.Costume
}

func WriteCostume(id int, costume blocks.Costume, outputdir string) {
	t := template.Must(template.ParseFiles(
		htmlPath+"costume.html",
		htmlPath+"partials.html"))

	htmlPath := fmt.Sprintf("%v/costume_%v.html", outputdir, id)
	file, err := os.Create(htmlPath)
	l.Log("template", "Create "+htmlPath)
	if err != nil {
		panic("Can't create costume file, " + err.Error())
	}

	data := costumeData{
		id,
		"Costume",
		costume,
	}
	t.Execute(file, data)
}

/* Helper functions */
