package blocks

import "testing"
import "io/ioutil"

func TestImage(t *testing.T) {
	data, err := ioutil.ReadFile("./testdata/objects/scummbardoor.dump")
	if err != nil {
		t.Errorf("Error reading the file")
	}
	object := NewObjectFromOBCD(data)
	/*
		if len(object.Verbs) != 2 {
			t.Errorf("Incorrect number of verbs")
		}
	*/
	if object.Id != 428 {
		t.Errorf("Incorrect id ")
	}
	if object.Name != "door" {
		t.Errorf("Incorrect name")
	}
	if object.Width != 40 {
		t.Errorf("Incorrect width")
	}
	if object.Height != 56 {
		t.Errorf("Incorrect height")
	}
	if object.X != 696 {
		t.Errorf("Incorrect X")
	}
	if object.Y != 80 {
		t.Errorf("Incorrect Y")
	}
	if object.Parent != 0 {
		t.Errorf("Incorrect Parent %x", object.Parent)
	}
	if object.Flags != 0 {
		t.Errorf("Incorrect Flags %x", object.Flags)
	}

}
