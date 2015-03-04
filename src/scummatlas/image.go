package scummatlas

import "fmt"
import "os"

type Image bool

func parseImage(data []byte, zBuffers int, width int, height int) Image {
	fmt.Println("PARSE ROOM IMG")

	if string(data[8:12]) != "SMAP" {
		panic("No stripe table found")
	}
	smapSize := BE32(data, 12)
	stripeCount := width / 8
	fmt.Println("SmapSize", smapSize)

	fmt.Println("There should be ", stripeCount, "stripes")
	offsets := map[int]int{}
	for i := 0; i < width/8; i++ {
		stripeOffset := LE32(data, 16+4*i)
		fmt.Printf("\nOffsets of %v is %v", i, stripeOffset)
		offsets[i] = stripeOffset
	}

	fmt.Println(offsets)

	os.Exit(255)
	return false
}
