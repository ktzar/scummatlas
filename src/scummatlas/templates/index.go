package templates

import (
	"html/template"
	"os"
	"scummatlas"
	"scummatlas/blocks"
	l "scummatlas/condlog"
)

type indexData struct {
	Title string
	Rooms []blocks.Room
}

type tableData struct {
	Title string
	Rooms []blocks.Room
}

type mapData struct {
	Title string
	Nodes []MapNode
	Edges []MapEdge
}

type CostumeIndex struct {
	Title   string
	Costumes []blocks.Costume
}

type MapNode struct {
	Id   int
	Name string
}

type MapEdge struct {
	Source int
	Target int
}

func newMapData(game scummatlas.Game) *mapData {
	data := new(mapData)

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
	writeScripts(game, outdir)
	writeCostumes(game, outdir)
}


func writeCostumes(game scummatlas.Game, outdir string) {
	filename := outdir + "/costumes.html"
	file, err := os.Create(filename)
	l.Log("template", "Create "+filename)
	if err != nil {
		panic("Can't create costumes file")
	}

	t := template.Must(template.ParseFiles(
		htmlPath+"costumes.html",
		htmlPath+"partials.html"))
	t.Execute(file, CostumeIndex{
		game.Name,
		game.Costumes,
	})
}

func writeIndex(game scummatlas.Game, outdir string) {
	filename := outdir + "/index.html"
	file, err := os.Create(filename)
	l.Log("template", "Create "+filename)
	if err != nil {
		panic("Can't create index file")
	}

	t := template.Must(template.ParseFiles(htmlPath+"index.html", htmlPath+"partials.html"))
	t.Execute(file, indexData{game.Name, game.Rooms})
}

func writeTable(game scummatlas.Game, outdir string) {
	t := template.Must(template.ParseFiles(htmlPath+"table.html", htmlPath+"partials.html"))

	filename := outdir + "/table.html"
	file, err := os.Create(filename)
	l.Log("template", "Create "+filename)
	if err != nil {
		panic("Can't create table file")
	}

	t.Execute(file, indexData{game.Name, game.Rooms})
}

func writeMap(game scummatlas.Game, outdir string) {
	data := newMapData(game)
	t := template.Must(template.ParseFiles(
		htmlPath+"map.html",
		htmlPath+"partials.html"))

	filename := outdir + "/map.html"
	file, err := os.Create(filename)
	l.Log("template", "Create "+filename)
	if err != nil {
		panic("Can't create map file")
	}

	t.Execute(file, data)
}
