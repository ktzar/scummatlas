package script

import "fmt"
import b "scummatlas/binaryutils"
import l "scummatlas/condlog"

func parsePrintOpcode(data []byte, offset int) (actions []string, opCodeLength int) {
	originalOffset := offset
	var subinst byte
	subinst = 0x00
	for subinst != 0x0f && offset < len(data) {
		subinst = data[offset]
		if subinst == 0xff {
			break
		}
		switch subinst {
		case 0x0f:
			say, length := parseString(data, offset+1)
			offset += length
			actions = append(actions, "text=\""+say+"\"")
		case 0x00:
			x := b.LE16(data, offset+1)
			y := b.LE16(data, offset+3)
			actions = append(actions, fmt.Sprintf("pos=%d,%d", x, y))
			offset += 5
		case 0x01:
			color := data[offset+1]
			actions = append(actions, fmt.Sprintf("color=%d", color))
			offset += 2
		case 0x02:
			clipped := data[offset+1]
			actions = append(actions, fmt.Sprintf("clipright=%d", clipped))
			offset += 2
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
			l.Log("script", "Unknown print subinst %x\n", subinst)
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
				switch escapeChar {
				case 0x01:
					say += "\\n "
				case 0x02:
					say += "\\keep "
				case 0x03:
					say += "\\wait "
				}
			case 0x04 <= escapeChar && escapeChar <= 0x0e:
				val := b.LE16(data, offset+3)
				offset += 4
				switch escapeChar {
				case 0x04:
					say += fmt.Sprintf("var(%x)", val)
				case 0x05:
					say += fmt.Sprintf("verb(%x)", val)
				case 0x06:
					say += fmt.Sprintf("actorName(%x)", val)
				case 0x07:
					say += fmt.Sprintf("stringInArray(%x)", val)
				}
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
