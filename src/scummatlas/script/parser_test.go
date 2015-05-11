package script

import "testing"
import "io/ioutil"

func readScriptOrDie(filename string, t *testing.T) []byte {
	data, err := ioutil.ReadFile("../testdata/scripts/" + filename + ".dump")
	if err != nil {
		t.Errorf("Error reading the file")
	}
	return data
}

func checkLengthAndOpcodes(script Script, expectedOpcodes []string, t *testing.T) {
	if len(script) != len(expectedOpcodes) {
		t.Errorf("Length mismatch")
	}
	for i, _ := range script {
		if script[i].callMethod != expectedOpcodes[i] {
			t.Errorf("Expecting opcode %v in position %d, but got %v",
				expectedOpcodes[i], i, script[i].callMethod)
		}
	}

}

func TestRoomScript1(t *testing.T) {
	data := readScriptOrDie("monkey2_10_200", t)
	script := ParseScriptBlock(data)
	checkLengthAndOpcodes(script,
		[]string{
			"animateCostume",
			"putActorInRoom",
			"putActor",
			"startScript",
			"stopObjectCode",
		}, t)

}
