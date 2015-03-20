package templates

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"scummatlas"
)

type roomData struct {
	Index      string
	Title      string
	Background string
	Boxes      [][4]scummatlas.Point
	scummatlas.Room
}

func (self roomData) DoubleHeight() int {
	return self.Height * 2
}

func (self roomData) ViewBox() string {
	return fmt.Sprintf("-10 -10 %v %v", self.Width+10, self.Height+10)
}

func (self roomData) SvgWidth() int {
	return self.Width * 2
}

func (self roomData) SvgHeight() int {
	return self.Height * 2
}

func WriteRoom(room scummatlas.Room, index int, outputdir string) {

	roomTpl, err := ioutil.ReadFile("./templates/room.html")
	if err != nil {
		panic("No index.html in the templates directory")
	}

	t := template.Must(template.New("index").Parse(string(roomTpl)))

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

	data := roomData{
		fmt.Sprintf("%02d", index),
		"A room",
		bgPath,
		boxes,
		room,
	}
	t.Execute(file, data)
}
