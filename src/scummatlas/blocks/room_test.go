package blocks

import "testing"
import "io/ioutil"

func TestRoom(t *testing.T) {
	data, err := ioutil.ReadFile("./testdata/rooms/someroom.dump")
	if err != nil {
		t.Errorf("Error reading the file")
	}
	room := NewRoom(data)
	if len(room.ExitScript) == 0 {
		t.Errorf("Did not parse Exit Script")
	}
	if len(room.EntryScript) == 0 {
		t.Errorf("Did not parse Entry Script")
	}
	if len(room.LocalScripts) != 6 {
		t.Errorf("Bad number of local scripts found")
	}
	if room.ObjCount != 20 {
		t.Errorf("Bad number of objects %d", room.ObjCount)
	}
	if room.Width != 320 {
		t.Errorf("Bad width %d", room.Width)
	}
	if room.Height != 144 {
		t.Errorf("Bad height %d", room.Height)
	}

}
