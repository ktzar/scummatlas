package templates

import (
	"fmt"
	"html/template"
	"os"
	"scummatlas/blocks"
	l "scummatlas/condlog"
	"strings"
)

type roomData struct {
	Background string
	Boxes      [][4]blocks.Point
	blocks.Room
}

func WriteRoom(room blocks.Room, outputdir string) {
	t := template.Must(template.ParseFiles(
		htmlPath+"room.html",
		htmlPath+"partials.html"))

	bgPath := fmt.Sprintf("./img_bg/room%02d_bg.png", room.Id)
	htmlPath := fmt.Sprintf("%v/room%02d.html", outputdir, room.Id)
	file, err := os.Create(htmlPath)
	l.Log("template", "Create "+htmlPath)
	if err != nil {
		panic("Can't create room file, " + err.Error())
	}

	var boxes [][4]blocks.Point

	for _, v := range room.Boxes {
		boxes = append(boxes, v.Corners())
	}

	data := roomData{
		bgPath,
		boxes,
		room,
	}
	data.Name = strings.Title(data.Name)
	t.Execute(file, data)
}

/* Helper functions */

func (room roomData) ZplanesURL() (urls []string) {
	for i := len(room.Zplanes); i > 0; i-- {
		urls = append(urls, fmt.Sprintf("img_bg/room%02d_bg-zplane%d.png", room.Id, i))
	}
	return
}

func (room roomData) PaletteHex() []string {
	var hexes []string
	hexes = make([]string, len(room.Palette))
	for i, color := range room.Palette {
		r, g, b, _ := color.RGBA()
		hexes[i] = fmt.Sprintf("%02x%02x%02x", uint8(r), uint8(g), uint8(b))
	}
	return hexes
}

func (room roomData) DoubleHeight() int {
	return room.Height * 2
}

func (room roomData) ViewBox() string {
	return fmt.Sprintf("0 0 %v %v", room.Width, room.Height)
}

func (room roomData) DoubleWidth() int {
	return room.Width * 2
}
