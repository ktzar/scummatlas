package templates

import (
	"html/template"
	_ "ioutil"
	"os"
	"scummatlas"
)

type IndexData struct {
	Title string
	Rooms []scummatlas.RoomName
}

const indexTpl = `
<html>
    <head>
        <title>Room Index</title>
    </head>
    <body>
	<h1>{{.Title}}</h1>
    <h2>List of rooms</h2>
	<ul>
    {{range .Rooms}}
		<li>
			<a href="./room{{.IndexNumber}}.html">
				<span class="roomnumber">{{.Number}}</span>
				<span class="roomname">{{.Name}}</span>
				<img src="room{{.IndexNumber}}_bg.png"/>
			</a>
		</li>{{end}}
	</ul>
    </body>
</html>`

func WriteIndex(roomNames []scummatlas.RoomName, outdir string) {
	data := IndexData{
		"A game",
		roomNames,
	}
	t := template.Must(template.New("index").Parse(indexTpl))

	file, err := os.Create(outdir + "/index.html")
	if err != nil {
		panic("Can't create index file")
	}

	t.Execute(file, data)
}
