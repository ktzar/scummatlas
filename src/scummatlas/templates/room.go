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
	Boxes      [][4]scummatlas.Point
	scummatlas.Room
}

func (self RoomData) ViewBox() string {
	return fmt.Sprintf("-10 -10 %v %v", self.Width+10, self.Height+10)
}

func (self RoomData) SvgWidth() int {
	return self.Width * 2
}

func (self RoomData) SvgHeight() int {
	return self.Height * 2
}

const roomTpl = `
<html>
    <head>
        <title>Room {{.Index}}</title>
		<link href="./style.css"/>
		<script src="./static/script.js"></script>
    </head>
    <body id="room-page">
		<h1>{{.Title}}</h1>
		<h2>Background</h2>
		<img width="100%" src="{{.Background}}"/>
		<h2>Walking boxes</h2>
		<svg width="{{.SvgWidth}}" height="{{.SvgHeight}}" viewBox="{{.ViewBox}}">
		{{range .Objects}}
		  <rect id="{{.Id}}" x="{{.X}}" y="{{.Y}}" width="{{.Width}}" height="{{.Height}}" fill="rgba(255,0,0,0.5)" stroke="white" stroke-width="1"/>
		  <text id="text_{{.Id}}" x="{{.X}}" y="{{.Y}}" font-size="0.8em" fill="black" font-family="monospace">{{.Name}}</text>
		{{end}}
		{{range .Boxes}}
		  <polygon points="
		  {{range .}}{{.X}},{{.Y}} {{end}}
		  " style="fill:rgba(128,128,128,0.5);stroke:black;stroke-width:1" />
		{{end}}
		</svg>

		<h2>Objects</h2>
		<table>
			<tr>
				<th>ID</th>
				<th>Name</th>
				<th>Position</th>
				<th>Size</th>
			</tr>
		{{range .Objects}}
		<tr><td>{{.Name}}</td><td>{{.IdHex}}</td><td>{{.X}},{{.Y}}</td><td>{{.Width}}x{{.Height}}</td></tr>
		{{end}}
		</table>
		<h2>Scripts</h2>
		<h3>Enter script</h3>
		<h3>Exit script</h3>
		<h3>Local scripts</h3>
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

	var boxes [][4]scummatlas.Point

	for _, v := range room.Boxes {
		boxes = append(boxes, v.Corners())
	}

	data := RoomData{
		index,
		"A room",
		bgPath,
		boxes,
		room,
	}
	t.Execute(file, data)
}
