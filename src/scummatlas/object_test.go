package scummatlas

import "testing"
import "io/ioutil"
import "fmt"

func TestImage(t *testing.T) {
	out, err = ioutil.ReadFile("testdata/objects/scummbardoor.dump")
	if err != nil {
		t.ErrorF("Error reading the file")
	}
	fmt.Printf("%x", out)

}
