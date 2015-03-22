package scummatlas

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
