package templates

import (
	"html/template"
	"os"
	"scummatlas"
	l "scummatlas/condlog"
	s "scummatlas/script"
)

type scriptData struct {
	Title   string
	Scripts []s.Script
}

func writeScripts(game scummatlas.Game, outdir string) {
	filename := outdir + "/scripts.html"
	file, err := os.Create(filename)
	l.Log("template", "Create "+filename)
	if err != nil {
		panic("Can't create index file")
	}

	t := template.Must(template.ParseFiles("./templates/scripts.html", "./templates/partials.html"))
	t.Execute(file, scriptData{
		"A game",
		game.Scripts,
	})
}
