package blocks

import "testing"

func TestBoxCorners(t *testing.T) {

	aBox := Box{
		ulx: 0,
		uly: 10,
		urx: 250,
		ury: 20,
		lrx: 20,
		lry: 210,
		llx: 250,
		lly: 250,
	}

	corners := aBox.Corners()

	var exp []Point;
	exp = append(exp, Point{ 0,10})
	exp = append(exp, Point{250,250})
	exp = append(exp, Point{20, 210})
	exp = append(exp, Point{250, 20})

	for i := 0; i < 4 ; i ++ {
		if corners[i] != exp[i] {
			t.Errorf("Corner %v does not match", i)
		}
	}

}
