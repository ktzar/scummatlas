package script

import (
	"errors"
	"fmt"
	b "scummatlas/binaryutils"
	l "scummatlas/condlog"
)

type ScriptParser struct {
	data   []byte
	offset int
	Script Script
}

func NewScriptParser(data []byte, offset int) ScriptParser {
	parser := ScriptParser{
		data,
		offset,
		Script{},
	}
	return parser
}

func (p ScriptParser) getString(position int) (length int, str string) {
	offset := p.offset + position
	length = 0
	for p.data[offset+length] != 0x00 {
		str += string(p.data[offset+length])
		length++
	}
	return
}

func (p ScriptParser) dumpHex(count int) {
	fmt.Printf("Dumping %d bytes from %x:\n%x\n",
		count,
		p.offset,
		p.data[p.offset:p.offset+count])
}

func (p ScriptParser) getWord(position int) int {
	return b.LE16(p.data, p.offset+position)
}

func (p ScriptParser) getByte(position int) int {
	return int(p.data[p.offset+position])
}

func (p ScriptParser) getList(offset int) (values []int) {
	for p.getByte(offset) != 0xFF {
		value := p.getWord(1)
		values = append(values, value)
		offset += 3
		if p.offset+offset >= len(p.data) {
			break
		}
	}
	return
}

func (p *ScriptParser) ParseNext() (op Operation, err error) {
	if p.offset >= len(p.data) {
		return Operation{}, errors.New("Script finished")
	}
	opcode := p.data[p.offset]
	opcodeName, ok := opCodesNames[opcode]

	if !ok {
		opcodeName, ok = opCodesNames[opcode&0x7F]
		if ok {
			l.Log("script", "Code %0x not in table, using %0x instead", opcode, opcode&0x7F)
		} else {
			opcodeName, ok = opCodesNames[opcode-0xC0]
			if ok {
				l.Log("script", "Code %0x not in table, using %0x instead", opcode, opcode&0x7F)
			} else {
				l.Log("script", "0x%x is not a known code\n", opcode)
				return Operation{}, fmt.Errorf("Unknown code %02x", opcode)
			}
		}
	}
	l.Log("script", "[%04x] (%02x) %v", p.offset, opcode, opcodeName)

	var subopcode byte
	if p.offset+1 < len(p.data) {
		subopcode = p.data[p.offset+1]
	}

	opCodeLength := 1
	paramWord1 := opcode&0x80 > 0
	paramWord2 := paramWord1 && opcode&0x40 > 0
	//paramWord3 := paramWord2 && opcode&0x20 > 0

	//Default to a function call since those are the majority
	op.opType = OpCall
	op.callMethod = opcodeName
	op.callMap = make(map[string]string)
	op.opCode = opcode
	op.offset = p.offset

	getByteWord := func(byteWord bool) (data int) {
		if byteWord {
			data = p.getWord(opCodeLength)
			opCodeLength += 2
		} else {
			data = p.getByte(opCodeLength)
			opCodeLength++
		}
		return
	}

	getWord := func() int {
		return getByteWord(true)
	}

	getByte := func() int {
		return getByteWord(false)
	}

	getList := func() []int {
		list := p.getList(opCodeLength)
		opCodeLength += len(list)*3 + 1
		return list
	}

	switch opcodeName {
	case "isGreaterEqual",
		"isLess",
		"isEqual",
		"isNotEqual",
		"isGreater",
		"lessOrEqual":
		variable := varName(getByte())
		value := getWord()
		target := getWord()
		opCodeLength++
		op = Operation{
			opType: OpConditional, condDst: target, opCode: opcode,
			condOp1: fmt.Sprintf("%v", value),
			condOp2: variable,
			condOp:  condOpSymbols[opcodeName],
			offset:  p.offset,
		}
	case "notEqualZero",
		"equalZero":
		variable := varName(getWord())
		target := getWord()
		target += p.offset + opCodeLength

		if p.getByte(2) == 0xa0 {
			opCodeLength++
		}
		if p.getByte(4) == 0x20 {
			opCodeLength++
		}
		op = Operation{
			opType: OpConditional, condDst: target, opCode: opcode,
			condOp1: variable, condOp2: "0", condOp: condOpSymbols[opcodeName],
			offset: p.offset,
		}
	case "animateActor":
		actor := getByteWord(paramWord1)
		anim := getByteWord(paramWord2)
		op.addNamedParam("actor", actor)
		op.addNamedParam("anim", anim)
	case "putActor":
		op.addNamedParam("actor", getByteWord(paramWord1))
		op.addNamedParam("x", getWord())
		op.addNamedParam("y", getWord())
	case "getActorRoom":
		result := getWord()
		actor := getByteWord(paramWord1)
		op.callResult = fmt.Sprintf("0x%x", result)
		op.addNamedParam("actor", int(actor))
	case "drawObject":
		object := getWord()
		op.addNamedParam("object", object)
		action := ""
		subopcode = byte(getByte())
		switch subopcode {
		case 0x01:
			action = "drawAt"
			op.addNamedParam("x", getWord())
			op.addNamedParam("y", getWord())
		case 0x02:
			opCodeLength = 6
			action = "setState"
			op.addNamedParam("state", getWord())
		case 0xff:
			opCodeLength = 4
			action = "draw"
		default:
			fmt.Printf("Unknown drawObject subopcode %x\n", subopcode)
		}
		op.callMethod += "." + action
	case "setState":
		op.addNamedParam("object", getWord())
		op.addNamedParam("state", getByte())
	case "startScript", "chainScript":
		script := getByte()
		list := getList()
		op.addNamedParam("script", int(script))
		op.addNamedStringParam("list", fmt.Sprintf("%v", list))
	case "resourceRoutines":
		opCodeLength++
		op.callMethod += "." + resourceRoutines[subopcode]
		switch subopcode {
		case 0x11:
		case 0x14:
			op.addNamedParam("room", getByte())
			op.addNamedParam("object", getWord())
		default:
			op.addNamedParam("resId", getByte())
		}
	case "getObjectState", "getObjectOwner":
		op.addResult(varName(getWord()))
		op.addNamedParam("object", getWord())
	case "panCameraTo":
		op.addNamedParam("x", getWord())
	case "actorOps":
		actor := getByteWord(paramWord1)
		op.addNamedParam("actor", actor)
		for p.getByte(opCodeLength) != 0xFF && op.opType != OpError {
			actionCode := getByte()
			actionLong := false
			if actionCode > 0x80 {
				actionCode = actionCode - 0x80
				actionLong = true
			}
			if actionCode > 0x40 {
				actionCode = actionCode - 0x40
			}
			action := actorOps[actionCode]
			switch action {
			case "init", "ignore_boxes", "follow_boxes", "never_zclip":
				//Nothing to do
			case "dummy", "costume", "sound", "walk_animation", "stand_animation", "talk_color",
				"init_animation", "width", "always_zclip", "animation_speed", "shadow":
				param := getByteWord(actionLong)
				op.addNamedParam(action, param)
			case "palette", "scale", "step_dist", "talk_animation":
				op.addNamedStringParam(action, fmt.Sprintf("%d,%d", getByte(), getByte()))
			case "elevation":
				op.addNamedParam(action, getWord())
			case "name":
				length, str := p.getString(opCodeLength)
				opCodeLength += length + 1
				op.addNamedStringParam(action, str)
			default:
				return Operation{}, errors.New(
					fmt.Sprintf("actorOps action %v (0x%02x) not implemented", action, actionCode))
			}
		}
		opCodeLength++
	case "getRandomNumber":
		op.callResult = varName(getWord())
		op.addParam(fmt.Sprintf("%d", getByte()))
	case "jumpRelative":
		target := getWord() + 3
		op.callMethod = "goto"
		op.addParam(fmt.Sprintf("%d", target))
	case "doSentence":
		verb := getByte()
		if verb == 0xFE {
			opCodeLength = 2
			op.addParam("STOP")
		} else {
			op.addNamedParam("verb", verb)
			op.addNamedParam("objA", getWord())
			op.addNamedParam("objB", getWord())
		}
	case "move":
		opCodeLength = 5
		op.opType = OpAssignment
		result := p.getWord(1)
		// ???? from scummvm
		if result&0x2000 > 0 {
			opCodeLength += 2
		}
		value := p.getWord(3)
		op.assignDst = fmt.Sprintf("var(%d)", result)
		op.assignVal = fmt.Sprintf("%d", value)
	case "loadRoomWithEgo":
		op.addNamedParam("object", getWord())
		op.addNamedParam("room", getByte())
		op.addNamedParam("x", getWord())
		op.addNamedParam("y", getWord())
	case "pickupObject":
		op.addNamedParam("object", getWord())
		op.addNamedParam("room", getByte())
	case "stringOps":
		l.Log("script", "string subopcode: 0x%x\n", subopcode)
		opCodeLength = 2
		op.callMethod += "." + stringOps[subopcode]
		switch subopcode {
		case 0x01:
			strId := getByte()
			length, str := p.getString(opCodeLength)
			op.addNamedParam("id", strId)
			op.addNamedStringParam("string", str)
			opCodeLength += length + 1
		case 0x02, 0x03:
			op.addNamedParam("stringId", getByte())
			op.addNamedParam("index", getByte())
			op.addNamedStringParam("char", string(getByte()))
		case 0x04:
			op.assignDst = varName(getWord())
			op.addNamedParam("stringId", getByte())
			op.addNamedParam("index", getByte())
		case 0x05:
			op.addNamedParam("stringId", getByte())
			op.addNamedParam("size", getByte())
		default:
			return Operation{}, errors.New(
				fmt.Sprintf(
					"stringOps subopcode %02x not understood",
					subopcode))
		}
		// TODO
	case "cursorCommand":
		opCodeLength = varLen
		if subopcode < 0x0a {
			opCodeLength = 2
			op.callMethod += "." + cursorCommands[subopcode]
		} else {
			switch subopcode {
			case 0x0A:
				opCodeLength = 4
			case 0x0B:
				opCodeLength = 5
			case 0x0C:
				opCodeLength = 3
			case 0x0D:
				opCodeLength = 3
			case 0x0E:
				for p.data[p.offset+int(opCodeLength)] != 0xff {
					opCodeLength++
				}
				opCodeLength++
			default:
				return Operation{}, errors.New(
					fmt.Sprintf(
						"cursorCommand subopcode %02x not understood",
						subopcode))
			}
		}
	case "putActorInRoom":
		op.addNamedParam("actor", getByteWord(paramWord1))
		op.addNamedParam("room", getByteWord(paramWord2))
	case "delay":
		opCodeLength = 4
		delay := b.LE24(p.data, p.offset+1)
		op.addParam(fmt.Sprintf("%d", delay))
	case "matrixOp":
		opCodeLength = 2
		if subopcode == 0x04 {
			op.callMethod += ".createBoxMatrix"
		} else {
			op.addNamedParam("box", getByte())
			op.addNamedParam("value", getByte())
			switch subopcode {
			case 0x01:
				op.callMethod += ".setBoxFlags"
			case 0x02, 0x03:
				op.callMethod += ".setBoxScale"
			}
		}
	case "roomOps":
		opCodeLength = 2
		if subopcode&0xE0 != 0 {
			subopcode -= 0xE0
		}
		op.callMethod += "." + roomOps[subopcode]
		switch subopcode {
		case 0x01:
			op.addNamedParam("minX", getWord())
			op.addNamedParam("maxX", getWord())
		case 0x03:
			op.addNamedParam("b", getWord())
			op.addNamedParam("h", getWord())
		case 0x04, 0xe4:
			op.addNamedParam("r", getWord())
			op.addNamedParam("g", getWord())
			op.addNamedParam("b", getWord())
			op.addNamedParam("palIdx", getByte())
			opCodeLength++ //not sure why
		case 0x05, 0x06:
		case 0x07:
			op.addNamedParam("scale1", getByte())
			op.addNamedParam("y1", getByte())
			op.addNamedParam("scale2", getByte())
			op.addNamedParam("y2", getByte())
			op.addNamedParam("slot", getByte())
		case 0x08, 0x88:
			op.addNamedParam("scale", getWord())
			op.addNamedParam("startcolor", getByte())
			op.addNamedParam("endcolor", getByte())
			opCodeLength++
		case 0x09:
			op.addNamedParam("flag", getByte())
			op.addNamedParam("slot", getByte())
		case 0x0A:
			op.addParam(fmt.Sprintf("%d", getWord()))
		case 0x0B, 0x0C:
			op.addNamedParam("r", getWord())
			op.addNamedParam("g", getWord())
			op.addNamedParam("b", getWord())
			op.addNamedParam("startcolor", getByte())
			op.addNamedParam("endcolor", getByte())
		case 0x0D, 0x0E:
			op.addNamedParam("id", getByte())
			length, str := p.getString(opCodeLength)
			op.addNamedStringParam("string", str)
			opCodeLength = 4 + length
		case 0x0F: //Transform
		case 0x10: //Cycle speed
		}
	case "walkActorToObject", "putActorAtObject":
		actor := getByteWord(paramWord1)
		object := getByteWord(true)
		op.addNamedParam("actor", actor)
		op.addNamedParam("object", object)
	case "substract", "add":
		op.opType = OpAssignment
		result := getWord()
		value := getWord()
		symbol := "+"
		if opcodeName == "substract" {
			symbol = "-"
		}
		op.assignDst = fmt.Sprintf("0x%x", result)
		op.assignVal = fmt.Sprintf("0x%x %v 0x%x", result, symbol, value)
	case "drawBox":
		if p.offset+11 < len(p.data) {
			op.addNamedParam("left", getWord())
			op.addNamedParam("top", getWord())
			op.addNamedStringParam("auxopcode",
				fmt.Sprintf("0x%x", getByte()))
			op.addNamedParam("right", getWord())
			op.addNamedParam("bottom", getWord())
			op.addNamedStringParam("color",
				fmt.Sprintf("0x%x", getByte()))
		}
	case "increment", "decrement":
		variable := getWord()
		op.opType = OpAssignment
		operation := "-"
		if opcodeName == "increment" {
			operation = "+"
		}
		op.assignDst = fmt.Sprintf("Var[%d]", variable)
		op.assignVal = fmt.Sprintf("Var[%d] %v 1", operation, variable)
	case "soundKludge":
		items := getList()
		op.addParam(fmt.Sprintf("%v", items))
	case "setObjectName":
		op.addNamedParam("object", getWord())
		length, name := p.getString(opCodeLength)
		opCodeLength = opCodeLength + length + 1
		op.addNamedStringParam("text", name)
	case "expression":
		expression := ""
		op.addResult(varName(getByte()))
		for p.data[p.offset+int(opCodeLength)] != 0xff {
			subopcode := getByte()
			switch subopcode {
			case 0x01:
				expression += fmt.Sprintf(" %d ", getWord())
			case 0x02:
				expression += "+"
			case 0x03:
				expression += "-"
			case 0x04:
				expression += "*"
			case 0x05:
				expression += "/"
			case 0x06:
				expression += "nested opCode"
			}
		}
		op.addParam(expression)
		opCodeLength++
	case "pseudoRoom":
		value := getByte()
		for value != 0x00 {
			op.addParam(fmt.Sprintf("%d", value))
			value = getByte()
		}
	case "wait":
		opCodeLength = 2
		if subopcode&0x7f == 0x01 {
			param := getByte()
			op.callMethod += ".forActor"
			op.addParam(fmt.Sprintf("%d", param))
			if p.getByte(opCodeLength) == 0 {
				opCodeLength++
			}
		}
	case "cutscene":
		list := getList()
		op.addParam(fmt.Sprintf("%v", list))
	case "print", "printEgo":
		if opcodeName == "print" {
			op.addNamedParam("actor", getByteWord(paramWord1))
		}
		actions, length := parsePrintOpcode(p.data, p.offset+opCodeLength)
		opCodeLength += length + 1
		for _, action := range actions {
			op.addParam(action)
		}
	case "actorSetClass":
		op.addNamedParam("actor", getWord())
		list := getList()
		op.addParam(fmt.Sprintf("%v", list))
	case "stopScript":
		op.addNamedParam("script", getByte())
	case "getScriptRunning":
		opCodeLength = 3
		script := getByteWord(paramWord1)
		op.callResult = "VAR_RESULT"
		op.addParam(fmt.Sprintf("%d", script))
	case "ifNotState":
		opCodeLength = 6
	case "getInventoryCount":
		op.addResult(varName(getByte()))
		op.addNamedParam("actor", getWord())
		opCodeLength++
	case "setCameraAt":
		op.addNamedParam("x", getWord())
	case "setVarRange":
		from := getWord()
		count := getByte()
		size := 1
		if paramWord1 {
			size = 2
		}
		opCodeLength += size * count
		list := make([]int, count)
		for i := 0; i < count; i++ {
			if paramWord1 {
				list[i] = p.getWord(4 + size*i)
			} else {
				list[i] = p.getByte(4 + size*i)
			}
		}
		op.addNamedParam("from", from)
		op.addNamedParam("count", count)
		op.addNamedStringParam("list", fmt.Sprintf("%v", list))
	case "setOwnerOf":
		op.addNamedParam("object", getWord())
		op.addNamedParam("owner", getByte())
	case "delayVariable":
		opCodeLength = 3
	case "and":
		variable := varName(getByte())
		value := getWord()
		op.opType = OpAssignment
		op.assignDst = variable
		op.assignVal = fmt.Sprintf("%v && %x", variable, value)
		opCodeLength++
	case "getDist":
		op.addResult(varName(getByte()))
		op.addNamedParam("objectA", getWord())
		op.addNamedParam("objectB", getWord())
		opCodeLength++
	case "findObject":
		op.addResult(varName(getWord()))
		op.addNamedParam("x", getWord())
		op.addNamedParam("y", getWord())
	case "startObject":
		object := getWord()
		script := getByte()
		list := getList()
		op.addNamedParam("object", int(object))
		op.addNamedParam("script", int(script))
		op.addNamedStringParam("list", fmt.Sprintf("%v", list))
	case "startSound":
		sound := getByteWord(paramWord1)
		op.addParam(fmt.Sprintf("%d", sound))
	case "stopSound":
		op.addNamedParam("sound", getByte())
	case "getActorWalkBox":
		variable := varName(getByteWord(paramWord1))
		actor := varName(getByteWord(true))
		op.addResult(variable)
		op.addNamedStringParam("actor", actor)
	case "getActorScale", "getActorMoving", "getActorFacing", "getActorElevation",
		"getActorWidth", "getActorCostume", "getActorX", "getActorY":
		op.callResult = varName(getWord())
		op.addNamedStringParam("actor", varName(getByte()))
		opCodeLength++
	case "ifClassOfIs":
		op.addNamedParam("value", getWord())
		op.addNamedStringParam("list", fmt.Sprintf("%v", getList()))
		op.addNamedParam("target", getWord())
	case "walkActorToActor":
		walker := getByteWord(paramWord1)
		walkee := getByteWord(paramWord2)
		distance := getByteWord(false)

		op.addNamedParam("walker", walker)
		op.addNamedParam("walkee", walkee)
		op.addNamedParam("distance", distance)
	case "walkActorTo":
		actor := varName(getByteWord(paramWord1))
		x := getByteWord(true)
		y := getByteWord(true)
		op.addNamedStringParam("actor", actor)
		op.addNamedParam("x", x)
		op.addNamedParam("y", y)
	case "verbOps":
		verb := getByteWord(paramWord1)
		op.addNamedParam("verbId", verb)
		for p.getByte(opCodeLength) != int(0xFF) &&
			op.opType != OpError {
			actionCode := byte(p.getByte(opCodeLength))
			action1Word := false
			if actionCode > 0x80 {
				actionCode -= 0x80
				action1Word = true
			}
			if actionCode > 0x40 {
				actionCode -= 0x40
			}
			if false && action1Word {
				print("dummy")
			}
			action := verbOps[actionCode]
			switch action {
			case "on", "off", "delete", "new", "dim", "center":
				op.addParam(action)
				opCodeLength++
			case "color", "hicolor", "dimcolor", "key", "setBackColor":
				param := p.getByte(opCodeLength + 1)
				op.addParam(fmt.Sprintf("%v=%d", action, param))
				opCodeLength += 2
				if action1Word {
					opCodeLength++
				}
			case "image", "name_str":
				param := p.getWord(opCodeLength + 1)
				op.addParam(fmt.Sprintf("%v=%d", action, param))
				opCodeLength += 3
			case "at":
				left := p.getWord(opCodeLength + 1)
				top := p.getWord(opCodeLength + 3)
				op.addParam(fmt.Sprintf("%v[%d,%d]", action, left, top))
				opCodeLength += 5
			case "assign":
				object := p.getWord(opCodeLength + 1)
				room := p.getByte(opCodeLength + 3)
				op.addParam(fmt.Sprintf("%v[%d,%d]", action, object, room))
				opCodeLength += 4
			case "name":
				strOffset := p.offset + opCodeLength + 1
				name, length := parseString(p.data, strOffset)
				op.addNamedStringParam(action, name)
				opCodeLength += length + 1
			default:
				return Operation{}, errors.New(
					fmt.Sprintf(
						"verbOps subopcode %02x not recognised at op offset %d. 20 bytes of data are: %x\n",
						action,
						opCodeLength,
						p.data[p.offset:p.offset+40]))
			}
		}
		opCodeLength++
	case "actorFollowCamera":
		opCodeLength = 3
	case "findInventory":
		op.callResult = varName(getWord())
		op.addNamedParam("owner", getByte())
		op.addNamedParam("index", getByte())
		opCodeLength += 2 // ?
	case "or":
		opCodeLength = 5
	case "override":
		opCodeLength = 2
	case "divide":
		opCodeLength = 5
	case "oldRoomEffect":
		opCodeLength = 4
	case "freezeScripts":
		opCodeLength = 3
	case "getClosestObjActor":
		opCodeLength = 5
	case "getStringWidth":
		opCodeLength = 5
	case "debug":
		opCodeLength = 3
	case "stopObjectScript":
		opCodeLength = 2
	case "lights":
		opCodeLength = 5
	case "loadRoom":
		opCodeLength = 2
	case "isSoundRunning":
		opCodeLength = 5
	case "breakHere":
	case "systemOps":
		opCodeLength = 2
	case "stopObjectCode":
	case "dummy":
	case "saveRestoreVerbs":
		switch subopcode {
		case 0x01:
			op.callMethod += ".save"
		case 0x02:
			op.callMethod += ".restore"
		case 0x03:
			op.callMethod += ".delete"
		}
		opCodeLength = 2
		op.addNamedParam("start", getByte())
		op.addNamedParam("end", getByte())
		op.addNamedParam("mode", getByte())

	case "endCutScene":
	case "startMusic":
		opCodeLength = 3
	case "faceActor":
		op.addNamedParam("actor", getByte())
		op.addNamedParam("object", getWord())
	case "getVerbEntryPoint":
		opCodeLength = 7
	case "actorFromPos":
		op.addResult(varName(getByte()))
		op.addNamedParam("x", getWord())
		op.addNamedParam("y", getWord())
	case "multiply":
		opCodeLength = 5 // ?
	case "isActorInBox":
		opCodeLength = 7
	case "stopMusic":
	case "getAnimCounter":
		opCodeLength = 5
	}

	if opCodeLength == varLen {
		return Operation{}, errors.New(fmt.Sprintf("Action %v (%x) is varLen. Can't proceed", opcodeName, opcode))
	}

	if opCodeLength == 0 {
		panic("Opcode length can't be 0")
	}

	p.offset += int(opCodeLength)
	p.Script = append(p.Script, op)
	return
}

func ParseScriptBlock(data []byte) Script {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()
	parser := new(ScriptParser)
	parser.data = data
	parser.offset = 0
	i := 0
	for parser.offset+1 < len(data) {
		_, err := parser.ParseNext()
		if err != nil {
			parser.Script = append(parser.Script, Operation{
				opType:   OpError,
				errorMsg: "error_" + err.Error(),
			})
			return parser.Script
		}
		i++
		if i > 2000 {
			break
		}
	}
	return parser.Script
}

const notDefined byte = 0xFF
const varLen int = 0xFE
const multi byte = 0xFD
const endsList int = 0xFC

func varName(code int) (name string) {
	name = varNames[byte(code)]
	if name == "" {
		name = "var(" + fmt.Sprintf("0x%x", code) + ")"
	}
	return
}
