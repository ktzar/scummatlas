package scummatlas

type Object struct {
	Image  *image.RGBA
	Script Script
	Name   string
	Id     int
	X      int
	Y      int
	Width  int
	Height int
	//TODO Direction uint8
	Flags  uint8
	Parent uint8
}

func (self Object) IdHex() string {
	return fmt.Sprintf("%x", self.Id)
}

func (self *Object) addOBCDToObject(data []byte) {
	//TODO how to add cleanly the OBCD data to an object
	//Maybe have another function to get the Id from the OBCD
	//create the object in room.go and then get this function
	//to furnish the rest of the fields from the data block
	// The problem is that the image lives in a different block
}
