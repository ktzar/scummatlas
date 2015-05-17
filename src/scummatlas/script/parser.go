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

func (p ScriptParser) getWord(position int) int {
	return b.LE16(p.data, p.offset+position)
}

func (p ScriptParser) getByte(position int) int {
	return int(p.data[p.offset+position])
}

func (p *ScriptParser) ParseNext() (Operation, error) {
	if p.offset >= len(p.data) {
		return Operation{}, errors.New("Script finished")
	}
	opcode := p.data[p.offset]
	opcodeName, ok := opCodesNames[opcode]

	var subopcode byte
	if p.offset+1 < len(p.data) {
		subopcode = p.data[p.offset+1]
	}
	if !ok {
		opcodeName, ok = opCodesNames[opcode&0x7F]
		if ok {
			l.Log("script", "Code %0x not in table, using %0x instead", opcode, opcode&0x7F)
		} else {
			l.Log("script", "0x%x is not a known code\n", opcode)
			return Operation{}, fmt.Errorf("Unknown code %02x", opcode)
		}
	}
	l.Log("script", "[%04x] (%02x) %v", p.offset, opcode, opcodeName)

	var opCodeLength int
	var op Operation
	paramWord1 := opcode&0x80 > 0
	paramWord2 := opcode&0x40 > 0
	//paramWord3 := opcode&0x20 > 0

	//Default to a function call since those are the majority
	op.opType = OpCall
	op.callMethod = opcodeName
	op.callMap = make(map[string]string)
	op.opCode = opcode

	switch opcodeName {
	case "isGreaterEqual",
		"isLess",
		"isEqual",
		"isNotEqual",
		"isGreater",
		"lessOrEqual":
		opCodeLength = 7
		variable := varName(p.data[p.offset+1])
		value := p.getWord(2)
		target := p.getWord(4)
		op = Operation{
			opType: OpConditional, condDst: target, opCode: opcode,
			condOp1: fmt.Sprintf("%v", value),
			condOp2: variable,
			condOp:  condOpSymbols[opcodeName],
		}
	case "notEqualZero",
		"equalZero":
		opCodeLength = 5
		variable := varName(p.data[p.offset+1])
		target := p.getWord(2)
		op = Operation{
			opType: OpConditional, condDst: target, opCode: opcode,
			condOp1: variable, condOp2: "0", condOp: condOpSymbols[opcodeName],
		}
	case "animateActor":
		opCodeLength = 3
		var actor int
		var anim int
		if paramWord1 {
			actor = p.getWord(1)
			anim = p.getByte(3)
			opCodeLength++
		} else {
			actor = p.getByte(1)
			anim = p.getByte(2)
		}
		op.addNamedParam("actor", actor)
		op.addNamedParam("anim", anim)
	case "putActor":
		opCodeLength = 6
		actor := p.getByte(1)
		x := p.getWord(2)
		y := p.getWord(4)
		if paramWord1 {
			actor = p.getWord(1)
			x = p.getWord(3)
			y = p.getWord(5)
		}
		op.addNamedParam("actor", actor)
		op.addNamedParam("x", x)
		op.addNamedParam("y", y)
	case "getActorRoom":
		opCodeLength = 4
		result := p.getWord(1)
		actor := p.data[p.offset+3]
		op.callResult = fmt.Sprintf("0x%x", result)
		op.addNamedParam("actor", int(actor))
	case "drawObject":
		opCodeLength = varLen
		object := p.getWord(1)
		op.addNamedParam("object", object)
		action := ""

		subopcode = p.data[p.offset+3]
		switch subopcode {
		case 0x01:
			opCodeLength = 8
			action = "drawAt"
			op.addNamedParam("x", p.getWord(4))
			op.addNamedParam("y", p.getWord(6))
		case 0x02:
			opCodeLength = 6
			action = "setState"
			op.addNamedParam("state", p.getWord(4))
		case 0xff:
			opCodeLength = 4
			action = "draw"
		}
		op.callMethod += "." + action
	case "setState":
		opCodeLength = 4
		op.addNamedParam("object", p.getWord(1))
		op.addNamedParam("state", p.getByte(3))
	case "startScript", "chainScript":
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
		op.callMethod += "." + resourceRoutines[subopcode]
	case "getObjectState":
		opCodeLength = 5
		object := p.getWord(3)
		op.addNamedParam("object", object)
	case "getObjectOwner":
		opCodeLength = 5
		object := p.getWord(3)
		op.addNamedParam("object", object)
	case "panCameraTo":
		opCodeLength = 3
		x := p.getWord(1)
		op.addNamedParam("x", x)
	case "actorOps":
		opCodeLength = 4
		actor := p.getWord(1)
		command := actorOps[subopcode]
		op.callMethod += "." + command
		op.addNamedParam("actor", actor)
		for p.data[p.offset+int(opCodeLength)] != 0xff {
			opCodeLength++
			if opCodeLength > 200 {
				return Operation{}, errors.New("cutscene too long")
			}
		}
		opCodeLength++
	case "getRandomNumber":
		opCodeLength = 4
		result := p.getWord(1)
		seed := p.data[p.offset+3]
		op.addParam(string(seed))
		op.callResult = fmt.Sprintf("var(%d)", result)
	case "jumpRelative":
		opCodeLength = 3
		target := p.getWord(1)
		op.addParam(fmt.Sprintf("0x%x", target))
	case "doSentence":
		opCodeLength = 7
		verb := p.getByte(1)
		objA := p.getWord(2)
		objB := p.getWord(4)
		op.addNamedParam("verb", verb)
		op.addNamedParam("objA", objA)
		op.addNamedParam("objB", objB)
	case "move":
		opCodeLength = 5
		op.opType = OpAssignment
		result := p.getWord(1)
		value := p.getWord(3)
		op.assignDst = fmt.Sprintf("var(%d)", result)
		op.assignVal = fmt.Sprintf("%d", value)
	case "loadRoomWithEgo":
		opCodeLength = 8
		object := p.getWord(1)
		room := p.data[p.offset+3]
		x := p.getWord(4)
		y := p.getWord(6)
		op.addNamedParam("object", object)
		op.addNamedParam("room", int(room))
		op.addNamedParam("x", x)
		op.addNamedParam("y", y)
	case "pickupObject":
		opCodeLength = 5
		object := p.getWord(1)
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
		opCodeLength = 3
		actor := p.getByte(2)
		room := p.getByte(3)
		op.addNamedParam("actor", actor)
		op.addNamedParam("room", int(room))
	case "delay":
		opCodeLength = 4
		delay := b.LE24(p.data, p.offset+1)
		op.addParam(fmt.Sprintf("%d", delay))
	case "matrixOp":
		if subopcode == 0x04 {
			opCodeLength = 2
			op.callMethod += ".createBoxMatrix"
		} else {
			opCodeLength = 4
			switch subopcode {
			case 0x01:
				op.callMethod += ".setBoxFlags"
			case 0x02, 0x03:
				op.callMethod += ".setBoxScale"
			}
			op.addNamedParam("box", p.getByte(2))
			op.addNamedParam("value", p.getByte(3))
		}
	case "roomOps":
		opCodeLength = varLen
		switch subopcode {
		case 0x01:
			opCodeLength = 6
			op.callMethod += ".scroll"
			op.addNamedParam("minX", p.getWord(2))
			op.addNamedParam("maxX", p.getWord(4))
		case 0x03:
			opCodeLength = 6
			op.callMethod += ".screen"
			op.addNamedParam("b", p.getWord(2))
			op.addNamedParam("h", p.getWord(4))
		case 0x04:
			opCodeLength = 10
			op.callMethod += ".setPalette"
			palette := p.data[p.offset+9]
			op.addNamedParam("r", p.getWord(2))
			op.addNamedParam("g", p.getWord(4))
			op.addNamedParam("b", p.getWord(6))
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
			op.addNamedParam("scale1", p.getByte(2))
			op.addNamedParam("y1", p.getByte(3))
			op.addNamedParam("scale2", p.getByte(4))
			op.addNamedParam("y2", p.getByte(5))
			op.addNamedParam("slot", p.getByte(6))
		case 0x08, 0x88:
			opCodeLength = 7
			op.callMethod += ".intensity"
			op.addNamedParam("scale", p.getWord(2))
			op.addNamedParam("startcolor", p.getByte(4))
			op.addNamedParam("endcolor", p.getByte(5))
		case 0x09:
			opCodeLength = 4
			op.callMethod += ".savegame"
			op.addNamedParam("flag", p.getByte(2))
			op.addNamedParam("slot", p.getByte(3))
		case 0x0A:
			opCodeLength = 4
			op.callMethod += ".effect"
			op.addParam(fmt.Sprintf("%d", p.getWord(2)))
		case 0x0B, 0x0C:
			opCodeLength = 10
			startcolor := p.data[p.offset+8]
			endcolor := p.data[p.offset+9]
			if subopcode == 0x0C {
				op.callMethod += ".shadow"
			} else {
				op.callMethod += ".intensity"
			}
			op.addNamedParam("r", p.getWord(2))
			op.addNamedParam("g", p.getWord(4))
			op.addNamedParam("b", p.getWord(6))
			op.addNamedParam("startcolor", int(startcolor))
			op.addNamedParam("endcolor", int(endcolor))
		case 0x0D: //Save string
		case 0x0E: //Load string
		case 0x0F: //Transform
		case 0x10: //Cycle speed
		}
	case "walkActorToObject":
		opCodeLength = 4
		op.addNamedParam("actor", p.getByte(1))
		op.addNamedParam("object", p.getWord(2))
	case "subtract":
		opCodeLength = 5
		op.opType = OpAssignment
		result := p.getWord(1)
		value := p.data[p.offset+3]
		op.assignDst = fmt.Sprintf("0x%x", result)
		op.assignVal = fmt.Sprintf("0x%x - 0x%x", result, value)
	case "drawBox":
		opCodeLength = 12
		if p.offset+11 < len(p.data) {
			left := p.getWord(1)
			top := p.getWord(3)
			auxopcode := p.data[p.offset+5]
			right := p.getWord(6)
			bottom := p.getWord(8)
			color := p.data[p.offset+10]
			op.addNamedParam("left", left)
			op.addNamedParam("top", top)
			op.addNamedStringParam("auxopcode", fmt.Sprintf("0x%x", auxopcode))
			op.addNamedParam("right", right)
			op.addNamedParam("bottom", bottom)
			op.addNamedStringParam("color", fmt.Sprintf("0x%x", color))
		}
	case "increment", "decrement":
		opCodeLength = 3
		variable := p.getWord(1)
		op.opType = OpAssignment
		operation := "-"
		if opcodeName == "increment" {
			operation = "+"
		}
		op.assignDst = fmt.Sprintf("Var[%d]", variable)
		op.assignVal = fmt.Sprintf("Var[%d] %v 1", operation, variable)
	case "soundKludge":
		items := p.parseList(p.offset + 1)
		opCodeLength = 2 + len(items)*3
		op.addParam(fmt.Sprintf("%v", items))
	case "setObjectName":
		opCodeLength = 3
		name := ""
		for p.data[p.offset+opCodeLength] != 0x00 {
			name += string(p.data[p.offset+opCodeLength])
			opCodeLength++
		}
		op.addNamedParam("object", p.getWord(1))
		op.addNamedStringParam("text", name)
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
			op.callMethod += ".forActor"
			op.addParam(fmt.Sprintf("%d", p.data[p.offset+2]))
		}
	case "cutscene":
		list := p.parseList(p.offset + 1)
		opCodeLength = 2 + len(list)*3
		op.addParam(fmt.Sprintf("%v", list))
	case "print", "printEgo":
		if opcodeName == "print" {
			op.addNamedParam("actor", p.getByte(1))
			opCodeLength = 2
		} else {
			opCodeLength = 1
		}
		say, length := parsePrintOpcode(p.data, p.offset+opCodeLength)
		opCodeLength += length + 1
		op.addParam(fmt.Sprintf("\"%v\"", say))
	case "actorSetClass":
		list := p.parseList(p.offset + 3)
		opCodeLength = 4 + len(list)*3
		op.addNamedParam("actor", p.getByte(1))
		op.addParam(fmt.Sprintf("%v", list))
	case "stopScript":
		opCodeLength = 2
		op.addNamedParam("script", p.getByte(1))
	case "getScriptRunning":
		opCodeLength = 4
		//result := p.data, p.offset + 1
		script := p.getByte(2)
		if paramWord1 {
			opCodeLength++
			script = p.getWord(2)
		}
		op.callResult = "VAR_RESULT"
		op.addParam(fmt.Sprintf("%d", script))
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
		objectA := p.getWord(1)
		objectB := p.getWord(2)
		variable := varName(byte(p.getByte(1)))
		op.addResult(variable)
		op.addNamedParam("objectA", objectA)
		op.addNamedParam("objectB", objectB)
	case "findObject":
		opCodeLength = 7
	case "startObject":
		object := p.getWord(1)
		script := p.getByte(3)
		list := p.parseList(p.offset + 4)
		opCodeLength = 4 + len(list)*3 + 1
		op.addNamedParam("object", int(object))
		op.addNamedParam("script", int(script))
		op.addNamedStringParam("list", fmt.Sprintf("%v", list))
	case "startSound":
		var sound int
		if paramWord1 {
			opCodeLength = 3
			sound = p.getWord(1)
		} else {
			opCodeLength = 2
			sound = p.getByte(1)
		}
		op.addParam(fmt.Sprintf("%d", sound))
	case "stopSound":
		opCodeLength = 2
		op.addNamedParam("sound", p.getByte(1))
	case "getActorWalkBox":
		opCodeLength = 5
		variable := varName(byte(p.getByte(1)))
		actor := varName(byte(p.getWord(2)))
		if paramWord1 {
			variable = varName(byte(p.getWord(1)))
			actor = varName(byte(p.getWord(3)))
		}
		op.addResult(variable)
		op.addNamedStringParam("actor", actor)
	case "getActorScale", "getActorMoving", "getActorFacing", "getActorElevation",
		"getActorWidth", "getActorCostume", "getActorX", "getActorY":
		opCodeLength = 5
		result := varName(uint8(p.getWord(1)))
		actor := varName(p.data[p.offset+3])
		op.callResult = result
		op.addNamedStringParam("actor", actor)
	case "ifClassOfIs":
		opCodeLength = varLen
		value := p.getWord(1)
		list := p.parseList(3)
		opCodeLength = 3 + len(list)*3 + 1
		target := p.getWord(opCodeLength)
		opCodeLength += 2
		op.addNamedParam("value", value)
		op.addNamedStringParam("list", fmt.Sprintf("%v", list))
		op.addNamedParam("target", target)
	case "actorFollowCamera":
		opCodeLength = 3
	case "findInventory":
		opCodeLength = 7
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
		opCodeLength = 3
	case "verbOps": //TODO
		opCodeLength = varLen
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
	case "faceActor":
		opCodeLength = 5
	case "getVerbEntryPoint":
		opCodeLength = 7
	case "walkActorToActor":
		opCodeLength = 6
		if paramWord2 {
			opCodeLength = 5
		}
	case "putActorAtObject":
		opCodeLength = 5
	case "actorFromPos":
		opCodeLength = 5
	case "multiply":
		opCodeLength = 5
	case "walkActorTo":
		opCodeLength = 6
	case "isActorInBox":
		opCodeLength = 7
	case "stopMusic":
		opCodeLength = 1
	case "getAnimCounter":
		opCodeLength = 5
	}

	if opCodeLength == varLen {
		return Operation{}, errors.New("Variable length opcode " + fmt.Sprintf("%x", opcode) + ", cannot proceed")
	}

	if opCodeLength == 0 {
		panic("Opcode length can't be 0")
	}
	p.offset += int(opCodeLength)
	p.Script = append(p.Script, op)
	return op, nil
}

func (p ScriptParser) parseList(offset int) (values []int) {
	for p.data[offset] != 0xFF {
		//TODO the first byte is supposed to always be 1 ???
		value := p.getWord(1)
		values = append(values, value)
		offset += 3
		if offset >= len(p.data) {
			break
		}
	}
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
				opType:     OpCall,
				callMethod: "error_" + err.Error(),
			})
			return parser.Script
		}
		i++
		if i > 1000 {
			break
		}
	}
	return parser.Script
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
