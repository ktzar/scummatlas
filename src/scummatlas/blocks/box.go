package blocks

import (
	b "scummatlas/binaryutils"
)

type Box struct {
	ulx   int
	uly   int
	urx   int
	ury   int
	lrx   int
	lry   int
	llx   int
	lly   int
	mask  byte
	flags byte
	scale int
}

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (self Box) Corners() [4]Point {
	ll := Point{self.llx, self.lly}
	ul := Point{self.ulx, self.uly}
	lr := Point{self.lrx, self.lry}
	ur := Point{self.urx, self.ury}
	return [4]Point{ul, ll, lr, ur}
}

func NewBox(data []byte) (box Box) {
	box.ulx = b.LE16(data, 0)
	box.uly = b.LE16(data, 2)
	box.urx = b.LE16(data, 4)
	box.ury = b.LE16(data, 6)
	box.lrx = b.LE16(data, 8)
	box.lry = b.LE16(data, 10)
	box.llx = b.LE16(data, 12)
	box.lly = b.LE16(data, 14)
	box.mask = data[16]
	box.flags = data[17]
	box.scale = b.LE16(data, 18)
	return
}
