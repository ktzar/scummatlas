package script

import "fmt"
import b "scummatlas/binaryutils"

func parsePrintOpcode(data []byte, offset int) (actions []string, opCodeLength int) {
	originalOffset := offset
	var subinst byte
	subinst = 0x00
	for subinst != 0x0f && offset < len(data) {
		subinst = data[offset]
		switch subinst {
		case 0x0f:
			say, length := parseString(data, offset+1)
			offset += length
			actions = append(actions, "text=\""+say+"\"")
		case 0x01:
			color := data[offset+1]
			actions = append(actions, fmt.Sprintf("color=%d", color))
			offset += 2
		case 0x02:
			clipped := data[offset+1]
			actions = append(actions, fmt.Sprintf("clipright=%d", clipped))
			offset += 2
		case 0x00:
			x := b.LE32(data, offset)
			y := b.LE32(data, offset+2)
			actions = append(actions, fmt.Sprintf("pos=%d,%d", x, y))
			offset += 4
		case 0x04:
			actions = append(actions, "center")
			offset++
		case 0x06:
			actions = append(actions, "left")
			offset++
		case 0x07:
			actions = append(actions, "overhead")
			offset++
		default:
			panic(fmt.Sprintf("Unknown print subinst %x", subinst))
			offset++
			break
		}
	}
	opCodeLength = offset - originalOffset
	return
}

func parseString(data []byte, offset int) (say string, length int) {
	originalOffset := offset
	i := 0
	say = ""
	for i <= 200 {
		i++
		currChar := data[offset]
		if currChar == 0xFF {
			escapeChar := data[offset+1]
			switch {
			case 0x01 <= escapeChar && escapeChar <= 0x03:
				offset += 2
			case 0x04 <= escapeChar && escapeChar <= 0x0e:
				offset += 4
			}
		} else if currChar >= 0x20 && currChar <= 0x7e { //printable ascii char
			say += string(currChar)
			offset++
		} else if currChar >= 0x0f {
			say += ""
			offset++
		} else if currChar >= 0x00 {
			offset++
			break
		} else {
			panic("Invalid character in print")
		}
	}
	return say, offset - originalOffset
}
