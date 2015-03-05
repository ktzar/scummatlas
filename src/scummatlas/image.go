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
	offsets := make([]int, 0, stripeCount)
	for i := 0; i < stripeCount; i++ {
		stripeOffset := LE32(data, 16+4*i)
		offsets = append(offsets, stripeOffset)
	}

	for i := 0; i < stripeCount; i++ {
		fmt.Printf("\nOffsets of %v is %x", i, offsets[i])
		fmt.Print("\tHeader ", data[offsets[i]+8])
		fmt.Print("\tCode ", int(data[offsets[i]+8])%10)
	}

	os.Exit(255)
	return false
}
