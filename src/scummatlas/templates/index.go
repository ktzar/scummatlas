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
	Rooms []scummatlas.Room
}

type TableData struct {
	Title string
	Rooms []scummatlas.Room
}

type MapData struct {
	Title string
	Nodes []MapNode
	Edges []MapEdge
}

type MapNode struct {
	Id   int
	Name string
}

type MapEdge struct {
	Source int
	Target int
}

func NewMapData(game scummatlas.Game) *MapData {
	data := new(MapData)

	data.Title = "Map"

	existingRooms := make(map[int]string)
	for _, r := range game.Rooms {
		existingRooms[r.Id] = r.Name
	}
	for _, r := range game.Rooms {
		for _, e := range r.Exits() {
			if existingRooms[e.Room] != "" {
				data.Edges = append(data.Edges, MapEdge{
					r.Id, e.Room})
			}
		}
	}

	nodesHash := make(map[int]bool)
	for _, e := range data.Edges {
		nodesHash[e.Source] = true
		nodesHash[e.Target] = true
	}
	for i, _ := range nodesHash {
		data.Nodes = append(data.Nodes, MapNode{
			i, existingRooms[i]})
	}
	return data
}

func WriteGameFiles(game scummatlas.Game, outdir string) {
	writeIndex(game, outdir)
	writeTable(game, outdir)
	writeMap(game, outdir)
}

func writeIndex(game scummatlas.Game, outdir string) {

	//TODO Cache that for the future
	indexTpl, err := ioutil.ReadFile("./templates/index.html")
	if err != nil {
		panic("No index.html in the templates directory")
	}

	data := IndexData{
		"A game",
		game.Rooms,
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

func writeTable(game scummatlas.Game, outdir string) {

	//roomNames := game.RoomNames
	rooms := game.Rooms

	//TODO Cache that for the future
	tableTpl, err := ioutil.ReadFile("./templates/table.html")
	if err != nil {
		panic("No table.html in the templates directory")
	}

	data := TableData{
		"A game",
		rooms,
	}
	t := template.Must(template.New("table").Parse(string(tableTpl)))

	filename := outdir + "/table.html"
	file, err := os.Create(filename)
	l.Log("template", "Create "+filename)
	if err != nil {
		panic("Can't create table file")
	}

	t.Execute(file, data)
}

func writeMap(game scummatlas.Game, outdir string) {

	//TODO Cache that for the future
	mapTpl, err := ioutil.ReadFile("./templates/map.html")
	if err != nil {
		panic("No map.html in the templates directory")
	}

	data := NewMapData(game)
	t := template.Must(template.New("map").Parse(string(mapTpl)))

	filename := outdir + "/map.html"
	file, err := os.Create(filename)
	l.Log("template", "Create "+filename)
	if err != nil {
		panic("Can't create map file")
	}

	t.Execute(file, data)
}
