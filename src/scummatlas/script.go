package scummatlas

import (
	"errors"
	"fmt"
	b "scummatlas/binaryutils"
	l "scummatlas/condlog"
)

type Script []Operation

const (
	_ = iota
	OpCall
	OpConditional
	OpAssignment
)

type Operation struct {
	opType     int
	assignDst  string
	assignVal  string
	condOp1    string
	condOp     string
	condOp2    string
	condDst    int
	callMethod string
	callResult string
	callMap    map[string]string
	callParams []string
}

func (op *Operation) addNamedStringParam(paramName string, value string) {
	op.callMap[paramName] = value
}

func (op *Operation) addNamedParam(paramName string, value int) {
	op.callMap[paramName] = fmt.Sprintf("%d", value)
}

func (op *Operation) addParam(param string) {
	op.callParams = append(op.callParams, param)
}

func (op Operation) String() string {
	if op.opType == OpCall {
		params := ""
		for _, param := range op.callParams {
			if params != "" {
				params += ", "
			}
			params += param
		}
		for paramName := range op.callMap {
			if params != "" {
				params += ", "
			}
			params += paramName + "=" + op.callMap[paramName]
		}
		callResult := ""
		if op.callResult != "" {
			callResult += fmt.Sprintf("%v = ", op.callResult)
		}
		return fmt.Sprintf("%v%v(%v)", callResult, op.callMethod, params)
	} else if op.opType == OpAssignment {
		return fmt.Sprintf("%v = %v", op.assignDst, op.assignVal)
	} else if op.opType == OpConditional {
		return fmt.Sprintf("unless (%v %v %v) goto %x", op.condOp1, op.condOp, op.condOp2, op.condDst)
	}
	return ""
}

func (script Script) Print() string {
	out := ""
	for i, op := range script {
		if i > 0 {
			out += ";\n"
		}
		out += op.String() + ";"
	}
	return out
}

type ScriptParser struct {
	data   []byte
	offset int
	script Script
}

func (p *ScriptParser) parseNext() (string, error) {
	if p.offset >= len(p.data) {
		return "", errors.New("Script finished")
	}
	opcode := p.data[p.offset]
	var subopcode byte
	if p.offset+1 < len(p.data) {
		subopcode = p.data[p.offset+1]
	}
	opcodeName, ok := opCodesNames[opcode]

	if !ok {
		l.Log("script", "0x%x is not a known code\n", opcode)
		return "", fmt.Errorf("Unknown code %02x", opcode)
	}
	l.Log("script", "[%04x] (%02x) %v", p.offset, opcode, opcodeName)

	var opCodeLength int
	var op Operation

	//Default to a function call since those are the majority
	op.opType = OpCall
	op.callMethod = opcodeName

	switch opcodeName {
	case "isGreaterEqual":
	case "isLess":
	case "isEqual":
	case "isNotEqual":
	case "isGreater":
	case "lessOrEqual":
		opCodeLength = 7
		variable := varName(p.data[p.offset+1])
		value := b.LE16(p.data, p.offset+2)
		target := b.LE16(p.data, p.offset+4)
		op = Operation{
			opType: OpConditional, condDst: target,
			condOp1: string(value), condOp2: variable, condOp: condOpSymbols[opcodeName],
		}
	case "notEqualZero":
	case "equalZero":
		opCodeLength = 5
		variable := varName(p.data[p.offset+1])
		target := b.LE16(p.data, p.offset+2)
		op = Operation{
			opType: OpConditional, condDst: target,
			condOp1: variable, condOp2: "0", condOp: condOpSymbols[opcodeName],
		}
	case "animateActor":
		opCodeLength = 3
		actor := int(p.data[p.offset+1])
		anim := int(p.data[p.offset+2])
		op.addNamedParam("actor", actor)
		op.addNamedParam("anim", anim)
	case "putActor":
		opCodeLength = 6
		actor := int(p.data[p.offset+1])
		x := b.LE16(p.data, p.offset+2)
		y := b.LE16(p.data, p.offset+4)
		if opcode&0x80 > 0 {
			actor = b.LE16(p.data, p.offset+1)
			x = b.LE16(p.data, p.offset+3)
			y = b.LE16(p.data, p.offset+5)
		}
		op.addNamedParam("actor", actor)
		op.addNamedParam("x", x)
		op.addNamedParam("y", y)
	case "getActorRoom":
		opCodeLength = 4
		result := b.LE16(p.data, p.offset+1)
		actor := p.data[p.offset+3]
		op.callResult = fmt.Sprintf("0x%x", result)
		op.addNamedParam("actor", int(actor))
	case "drawObject":
		opCodeLength = varLen
		object := b.LE16(p.data, p.offset+1)
		action := ""

		subopcode = p.data[p.offset+3]
		switch subopcode {
		case 0x01:
			opCodeLength = 8
			action = "drawAt"
			op.addNamedParam("x", b.LE16(p.data, p.offset+4))
			op.addNamedParam("y", b.LE16(p.data, p.offset+6))
		case 0x02:
			opCodeLength = 6
			action = "setState"
			op.addNamedParam("state", b.LE16(p.data, p.offset+4))
		case 0xff:
			opCodeLength = 4
			action = "draw"
		}
		op.callMethod += "." + action
	case "setState":
		opCodeLength = 4
		op.addNamedParam("object", b.LE16(p.data, p.offset+1))
		op.addNamedParam("state", int(p.data[p.offset+3]))
	case "startScript":
		opCodeLength = 2
		script := p.data[p.offset+1]
		list := p.parseList(p.offset + 2)
		opCodeLength = 2 + len(list)*3 + 1
		op.addNamedParam("script", int(script))
		op.addNamedStringParam("list", fmt.Sprintf("%v", list))
	case "resourceRoutines":
		opCodeLength = 3
		switch subopcode {
		case 0x11:
			opCodeLength = 1
		case 0x14:
			opCodeLength = 4
		}
		op.callMethod += resourceRoutines[subopcode]
	case "getObjectState":
		opCodeLength = 5
		object := b.LE16(p.data, p.offset+3)
		op.addNamedParam("object", object)
	case "getObjectOwner":
		opCodeLength = 5
		object := b.LE16(p.data, p.offset+3)
		op.addNamedParam("object", object)
	case "panCameraTo":
		opCodeLength = 3
		x := b.LE16(p.data, p.offset+1)
		op.addNamedParam("x", x)
	case "actorOps":
		opCodeLength = 4
		actor := b.LE16(p.data, p.offset+1)
		command := actorOps[subopcode]
		op.callMethod += "." + command
		op.addNamedParam("actor", actor)
		for p.data[p.offset+int(opCodeLength)] != 0xff {
			opCodeLength++
			if opCodeLength > 200 {
				return "", errors.New("cutscene too long")
			}
		}
		opCodeLength++
	case "getRandomNumber":
		opCodeLength = 4
		result := b.LE16(p.data, p.offset+1)
		seed := p.data[p.offset+3]
		op.addParam(string(seed))
		op.callResult = fmt.Sprintf("var(%d)", result)
	case "jumpRelative":
		opCodeLength = 3
		target := b.LE16(p.data, p.offset+1)
		op.addParam(fmt.Sprintf("0x%x", target))
	case "doSentence":
		opCodeLength = 7
		verb := int(p.data[p.offset+1])
		objA := b.LE16(p.data, p.offset+2)
		objB := b.LE16(p.data, p.offset+4)
		op.addNamedParam("verb", verb)
		op.addNamedParam("objA", objA)
		op.addNamedParam("objB", objB)
	case "move":
		opCodeLength = 5
		op.opType = OpAssignment
		result := b.LE16(p.data, p.offset+1)
		value := b.LE16(p.data, p.offset+3)
		op.assignDst = fmt.Sprintf("var(%d)", result)
		op.assignVal = fmt.Sprintf("%d", value)
	case "loadRoomWithEgo":
		opCodeLength = 8
		object := b.LE16(p.data, p.offset+1)
		room := p.data[p.offset+3]
		x := b.LE16(p.data, p.offset+4)
		y := b.LE16(p.data, p.offset+6)
		op.addNamedParam("object", object)
		op.addNamedParam("room", int(room))
		op.addNamedParam("x", x)
		op.addNamedParam("y", y)
	case "pickupObject":
		opCodeLength = 5
		object := b.LE16(p.data, p.offset+1)
		room := p.data[p.offset+3]
		op.addNamedParam("object", object)
		op.addNamedParam("room", int(room))
	case "stringOps":
		l.Log("script", "string subopcode: 0x%x\n", subopcode)
		opCodeLength = varLen
		if subopcode == 0x02 || subopcode == 0x05 {
			opCodeLength = 5
		} else if subopcode == 0x04 {
			opCodeLength = 7
		}
		// TODO
	case "cursorCommand":
		opCodeLength = varLen
		if subopcode < 0x0a {
			opCodeLength = 2
			op.callMethod += cursorCommands[subopcode]
		}
	case "putActorInRoom":
		opCodeLength = 4
		actor := int(p.data[p.offset+2])
		room := p.data[p.offset+3]
		if opcode&0x80 > 0 {
			opCodeLength++
			actor = b.LE16(p.data, p.offset+1)
			room = p.data[p.offset+3]
		}
		op.addNamedParam("actor", actor)
		op.addNamedParam("room", int(room))
	case "delay":
		opCodeLength = 4
		delay := b.LE24(p.data, p.offset+1)
		op.addParam(fmt.Sprintf("%d", delay))
	case "matrixOp":
		if subopcode == 0x04 {
			opCodeLength = 2
		} else {
			opCodeLength = 4
		}
	case "roomOps":
		opCodeLength = varLen
		switch subopcode {
		case 0x01:
			opCodeLength = 6
			op.callMethod += ".scroll"
			op.addNamedParam("minX", b.LE16(p.data, p.offset+2))
			op.addNamedParam("maxX", b.LE16(p.data, p.offset+4))
		case 0x03:
			opCodeLength = 6
			op.callMethod += ".screen"
			op.addNamedParam("b", b.LE16(p.data, p.offset+2))
			op.addNamedParam("h", b.LE16(p.data, p.offset+4))
		case 0x04:
			opCodeLength = 10
			op.callMethod += ".setPalette"
			r := b.LE16(p.data, p.offset+2)
			g := b.LE16(p.data, p.offset+4)
			b := b.LE16(p.data, p.offset+6)
			palette := p.data[p.offset+9]
			op.addNamedParam("r", r)
			op.addNamedParam("g", g)
			op.addNamedParam("b", b)
			op.addNamedParam("paletteIndex", int(palette))
		case 0x05:
			opCodeLength = 2
			op.callMethod += ".shakeOn"
		case 0x06:
			opCodeLength = 2
			op.callMethod += ".shakeOff"
		case 0x07:
			opCodeLength = 7
			op.callMethod += ".scale"
			scale1 := int(p.data[p.offset+2])
			y1 := int(p.data[p.offset+3])
			scale2 := int(p.data[p.offset+4])
			y2 := int(p.data[p.offset+5])
			slot := int(p.data[p.offset+6])
			op.addNamedParam("scale1", scale1)
			op.addNamedParam("y1", y1)
			op.addNamedParam("scale2", scale2)
			op.addNamedParam("y2", y2)
			op.addNamedParam("slot", slot)
		case 0x08:
		case 0x88:
			opCodeLength = 7
			op.callMethod += ".intensity"
			scale := b.LE16(p.data, p.offset+2)
			startcolor := p.data[p.offset+4]
			endcolor := p.data[p.offset+5]
			op.addNamedParam("scale", scale)
			op.addNamedParam("startcolor", int(startcolor))
			op.addNamedParam("endcolor", int(endcolor))
		case 0x09:
			opCodeLength = 4
			op.callMethod += ".savegame"
			flag := p.data[p.offset+2]
			slot := p.data[p.offset+3]
			op.addNamedParam("flag", int(flag))
			op.addNamedParam("slot", int(slot))
		case 0x0A:
			opCodeLength = 4
			op.callMethod += ".effect"
			op.addParam(fmt.Sprintf("%d", b.LE16(p.data, p.offset+2)))
		case 0x0B:
		case 0x0C:
			opCodeLength = 10
			r := b.LE16(p.data, p.offset+2)
			g := b.LE16(p.data, p.offset+4)
			b := b.LE16(p.data, p.offset+6)
			startcolor := p.data[p.offset+8]
			endcolor := p.data[p.offset+9]
			subinstruction := "intensity"
			if subopcode == 0x0C {
				subinstruction = "shadow"
				op.callMethod += ".shadow"
			} else {
				op.callMethod += ".intensity"
			}
			instruction += fmt.Sprintf(".%v(r=%d, g=%d, b=%d, "+
				"startColor=%02x, endColor=%02x)",
				subinstruction, r, g, b, startcolor, endcolor)
		case 0x0D: //Save string
		case 0x0E: //Load string
		case 0x0F: //Transform
		case 0x10: //Cycle speed
		}
	case "walkActorToObject":
		opCodeLength = 4
		actor := p.data[p.offset+1]
		object := b.LE16(p.data, p.offset+2)
		instruction += fmt.Sprintf("actor=0x%x, object=0x%x", actor, object)
	case "subtract":
		opCodeLength = 5
		result := b.LE16(p.data, p.offset+1)
		value := p.data[p.offset+3]
		instruction = fmt.Sprintf("0x%x=0x%x - 0x%x)", result, result, value)
	case "drawBox":
		opCodeLength = 12
		if p.offset+11 < len(p.data) {
			left := b.LE16(p.data, p.offset+1)
			top := b.LE16(p.data, p.offset+3)
			auxopcode := p.data[p.offset+5]
			right := b.LE16(p.data, p.offset+6)
			bottom := b.LE16(p.data, p.offset+8)
			color := p.data[p.offset+10]
			instruction += fmt.Sprintf(
				"left=%d, top=%d, auxopcode=0x%x, right=%d, bottom=%d, color=0x%x",
				left, top, auxopcode, right, bottom, color)
		}
	case "increment":
		opCodeLength = 3
		variable := b.LE16(p.data, p.offset+1)
		instruction = fmt.Sprintf("Var[%d]++", variable)
	case "soundKludge":
		items := p.parseList(p.offset + 1)
		opCodeLength = 2 + len(items)*3
		instruction += fmt.Sprintf("%v", items)
	case "setObjectName":
		opCodeLength = 3
		object := b.LE16(p.data, p.offset+1)
		name := ""
		for p.data[p.offset+opCodeLength] != 0x00 {
			name += string(p.data[p.offset+opCodeLength])
			opCodeLength++
		}
		instruction += fmt.Sprintf("object=%d, text=\"%v\"", object, name)
	case "expression": //TODO properly
		opCodeLength = 1
		for p.data[p.offset+int(opCodeLength)] != 0xff {
			opCodeLength++
		}
		opCodeLength++
	case "pseudoRoom":
		opCodeLength = varLen
	case "wait":
		opCodeLength = 2
		if subopcode == 0x01 {
			opCodeLength = 3
			instruction = fmt.Sprintf("wait.forActor(%d)",
				p.data[p.offset+2])

		}
	case "cutscene":
		list := p.parseList(p.offset + 1)
		opCodeLength = 2 + len(list)*3
		instruction += fmt.Sprintf("%v", list)
	case "decrement":
		opCodeLength = 3
		variable := b.LE16(p.data, p.offset+1)
		instruction = fmt.Sprintf("Var[%d]--", variable)
	case "print", "printEgo":
		if opcodeName == "print" {
			instruction += fmt.Sprintf("actor=%d, ", p.data[p.offset+1])
			opCodeLength = 2
		} else {
			opCodeLength = 1
		}
		say, length := parsePrintOpcode(p.data, p.offset+opCodeLength)
		opCodeLength += length
		instruction += fmt.Sprintf("\"%v\"", say)
	case "actorSetClass":
		object := b.LE16(p.data, p.offset+1)
		list := p.parseList(p.offset + 3)
		opCodeLength = 4 + len(list)*3
		instruction += fmt.Sprintf("actor=%d, %v", object, list)
	case "stopScript":
		opCodeLength = 2
		instruction += fmt.Sprintf("%d", p.data[p.offset+1])
	case "getScriptRunning":
		opCodeLength = 4
		//result := p.data, p.offset + 1
		script := b.LE16(p.data, p.offset+2)
		instruction = fmt.Sprintf("VAR_RESULT = isScriptRunning(%02x);", script)
	case "ifNotState":
		opCodeLength = 6
	case "getInventoryCount":
		opCodeLength = 5
	case "setCameraAt":
		opCodeLength = 3
	case "setVarRange":
		opCodeLength = endsList
	case "setOwnerOf":
		opCodeLength = 5
	case "delayVariable":
		opCodeLength = 3
	case "and":
		opCodeLength = 5
	case "getDist":
		opCodeLength = 7
	case "findObject":
		opCodeLength = 7
	case "startObject":
		opCodeLength = endsList
	case "actorFollowCamera":
		opCodeLength = 3
	case "getActorScale":
		opCodeLength = 5
	case "stopSound":
		opCodeLength = 3
	case "findInventory":
		opCodeLength = 7
	case "chainScript":
		opCodeLength = endsList
	case "getActorX":
		opCodeLength = 5
	case "getActorMoving":
		opCodeLength = 5
	case "or":
		opCodeLength = 5
	case "override":
		opCodeLength = 2
	case "add":
		opCodeLength = 5
	case "divide":
		opCodeLength = 5
	case "oldRoomEffect":
		opCodeLength = 4
	case "freezeScripts":
		opCodeLength = 3
	case "getActorFacing":
		opCodeLength = 5
	case "getClosestObjActor":
		opCodeLength = 5
	case "getStringWidth":
		opCodeLength = 5
	case "debug":
		opCodeLength = 3
	case "getActorWidth":
		opCodeLength = 5
	case "stopObjectScript":
		opCodeLength = 2
	case "lights":
		opCodeLength = 5
	case "getActorCostume":
		opCodeLength = 5
	case "loadRoom":
		opCodeLength = 3
	case "verbOps":
		opCodeLength = varLen
	case "getActorWalkBox":
		opCodeLength = 5
	case "isSoundRunning":
		opCodeLength = 5
	case "breakHere":
		opCodeLength = 1
	case "systemOps":
		opCodeLength = 2
	case "stopObjectCode":
		opCodeLength = 1
	case "dummy":
		opCodeLength = 1
	case "saveRestoreVerbs":
		opCodeLength = 4
	case "endCutScene":
		opCodeLength = 1
	case "startMusic":
		opCodeLength = 3
	case "getActorElevation":
		opCodeLength = 5
	case "faceActor":
		opCodeLength = 5
	case "getVerbEntryPoint":
		opCodeLength = 7
	case "walkActorToActor":
		opCodeLength = 6
	case "putActorAtObject":
		opCodeLength = 5
	case "actorFromPos":
		opCodeLength = 5
	case "multiply":
		opCodeLength = 5
	case "startSound":
		opCodeLength = 3
	case "ifClassOfIs":
		opCodeLength = varLen
	case "walkActorTo":
		opCodeLength = 6
	case "isActorInBox":
		opCodeLength = 7
	case "stopMusic":
		opCodeLength = 1
	case "getAnimCounter":
		opCodeLength = 5
	case "getActorY":
		opCodeLength = 5
	}

	if opCodeLength == varLen {
		return "", errors.New("Variable length opcode " + fmt.Sprintf("%x", opcode) + ", cannot proceed")
	}

	p.offset += int(opCodeLength)

	p.script = append(p.script, instruction)
	return opcodeName, nil
}

func (p ScriptParser) parseList(offset int) (values []int) {
	for p.data[offset] != 0xFF {
		//TODO the first byte is supposed to always be 1 ???
		value := b.LE16(p.data, offset+1)
		values = append(values, value)
		offset += 3
		if offset > len(p.data) {
			break
		}
	}
	return
}

func parseScriptBlock(data []byte) Script {
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
		_, err := parser.parseNext()
		if err != nil {
			parser.script = append(parser.script, callOp{
				"error",
				nil,
				[]string{
					err.Error()},
			})
			return parser.script
		}
		i++
		if i > 1000 {
			break
		}
	}
	return parser.script
}

const notDefined byte = 0xFF
const varLen int = 0xFE
const multi byte = 0xFD
const endsList int = 0xFC

func varName(code uint8) (name string) {
	name = varNames[code]
	if name == "" {
		name = "var(" + fmt.Sprintf("0x%x", code) + ")"
	}
	return
}
