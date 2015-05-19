package templates

import (
	"fmt"
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

func (self scriptData) ScriptTable() (out []template.HTML) {
	for i, _ := range self.Scripts {
		if i%12 == 0 {
			out = append(out, template.HTML("</tr><tr>"))
		}
		out = append(out, template.HTML(fmt.Sprintf(
			"<td><a href='#script-%d'>Script %d</a></td>",
			i, i)))
	}
	return
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
