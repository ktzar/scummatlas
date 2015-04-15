package scummatlas

import "io/ioutil"
import "os"
import "log"

var DBG_SCRIPT = ioutil.Discard

func resetLog() {
	log.SetOutput(os.Stdout)
}
