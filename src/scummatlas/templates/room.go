package templates

import (
	"fmt"
	"html/template"
	_ "io/ioutil"
	"os"
	"scummatlas"
)

type RoomData struct {
	Index      int
	Title      string
	Background string
	scummatlas.Room
}

const roomTpl = `
<html>
    <head>
        <title>Room {{.Index}}</title>
    </head>
    <body>
		<h1>{{.Title}}</h1>
		<h2>Background</h2>
		<img width="100%" src="{{.Background}}"/>
		<h2>Walking boxes</h2>
		<h2>Scripts</h2>
		<h2>Objects</h2>
		<h2>Palette</h2>
    </body>
</html>`

func WriteRoom(room scummatlas.Room, index int, outputdir string) {

	t := template.Must(template.New("index").Parse(roomTpl))

	bgPath := fmt.Sprintf("./room%02d_bg.png", index)
	htmlPath := fmt.Sprintf("%v/room%02d.html", outputdir, index)
	file, err := os.Create(htmlPath)
	if err != nil {
		panic("Can't create room file")
	}

	data := RoomData{
		index,
		"A room",
		bgPath,
		room,
	}
	fmt.Println(room.Boxes)
	t.Execute(file, data)
}
