package scummatlas

import (
	"errors"
	"fmt"
	b "scummatlas/binaryutils"
	l "scummatlas/condlog"
	"strings"
)

type Script []string

func (self Script) Print() string {
	return strings.Join(self, ";\n")
}

type ScriptParser struct {
	data   []byte
	offset int
	script Script
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
		return "", errors.New(fmt.Sprintf("Unknown code %02x", opcode))
	}
	l.Log("script", "[%04x] (%02x) %v", p.offset, opcode, opcodeName)

	instruction := opcodeName + "("
	instructionFinished := false

	var opCodeLength int

	switch opcodeName {
	case "animateActor":
		opCodeLength = 3
		actor := p.data[p.offset+1]
		anim := p.data[p.offset+2]
		instruction += fmt.Sprintf("actor=%d, anim=%d", actor, anim)
	case "putActor":
		opCodeLength = 6
		actor := p.data[p.offset+1]
		x := b.LE16(p.data, p.offset+2)
		y := b.LE16(p.data, p.offset+4)
		instruction += fmt.Sprintf("actor=0x%x, x=%d, y=%d", actor, x, y)
	case "startMusic":
		opCodeLength = 3
	case "getActorRoom":
		opCodeLength = 5
		result := b.LE16(p.data, p.offset+1)
		actor := p.data[p.offset+3]
		instruction = fmt.Sprintf("0x%x = getActorRoom(actor=0x%x)", result, actor)
		instructionFinished = true
	case "isGreaterEqual":
		opCodeLength = 7
		variable := varName(p.data[p.offset+1])
		value := b.LE16(p.data, p.offset+2)
		target := b.LE16(p.data, p.offset+4)
		instruction = fmt.Sprintf("unless (0x%x >= %v) goto 0x%x", value, variable, target)
	case "drawObject":
		opCodeLength = varLen
		object := b.LE16(p.data, p.offset+1)
		action := ""
		params := ""

		subopcode = p.data[p.offset+3]
		switch subopcode {
		case 0x01:
			opCodeLength = 8
			action = "drawAt"
			params = fmt.Sprintf(", x = %d, y = %d",
				b.LE16(p.data, p.offset+4),
				b.LE16(p.data, p.offset+6))
		case 0x02:
			opCodeLength = 6
			action = "setState"
			params = fmt.Sprintf(", state = %d",
				b.LE16(p.data, p.offset+4))
		case 0xff:
			opCodeLength = 4
			action = "draw"
		}
		instruction = fmt.Sprintf("drawObject.%v(object = %02x%v)", action, object, params)
		instructionFinished = true
	case "getActorElevation":
		opCodeLength = 5
	case "setState":
		opCodeLength = 4
		object := b.LE16(p.data, p.offset+1)
		state := p.data[p.offset+3]
		instruction += fmt.Sprintf("object = %d, state = %d", object, state)
	case "isNotEqual":
		opCodeLength = 7
		variable := varName(p.data[p.offset+1])
		value := b.LE16(p.data, p.offset+2)
		target := b.LE16(p.data, p.offset+4)
		instruction = fmt.Sprintf("unless (0x%x != %v) goto 0x%x", value, variable, target)
	case "faceActor":
		opCodeLength = 5
	case "startScript":
		opCodeLength = 2
		script := p.data[p.offset+1]

		list := p.parseList(p.offset + 2)
		opCodeLength = 2 + len(list)*3 + 1
		instruction += fmt.Sprintf("script=%d, %v", script, list)
	case "getVerbEntryPoint":
		opCodeLength = 7
	case "resourceRoutines":
		opCodeLength = 3
		switch subopcode {
		case 0x11:
			opCodeLength = 1
		case 0x14:
			opCodeLength = 4
		}
		instruction = fmt.Sprintf("resourceRoutines.%v(",
			resourceRoutines[subopcode])

	case "walkActorToActor":
		opCodeLength = 6
	case "putActorAtObject":
		opCodeLength = 5
	case "getObjectState":
		opCodeLength = 5
		object := b.LE16(p.data, p.offset+3)
		instruction += fmt.Sprintf("object=0x%x", object)
	case "getObjectOwner":
		opCodeLength = 5
		object := b.LE16(p.data, p.offset+3)
		instruction += fmt.Sprintf("object=0x%x", object)
	case "panCameraTo":
		opCodeLength = 3
	case "actorOps":
		opCodeLength = 4
		subopcode = p.data[p.offset+3]
		actor := b.LE16(p.data, p.offset+1)
		instructionFinished = true
		command := actorOps[subopcode]
		instruction = fmt.Sprintf("actorOps.%v(actor=0x%x)", command, actor)
		for p.data[p.offset+int(opCodeLength)] != 0xff {
			opCodeLength++
			if opCodeLength > 200 {
				return "", errors.New("cutscene too long")
			}
		}
		opCodeLength++
	case "actorFromPos":
		opCodeLength = 5
	case "getRandomNumber":
		opCodeLength = 4
		result := b.LE16(p.data, p.offset+1)
		seed := p.data[p.offset+3]
		instruction = fmt.Sprintf("var(%d) := getRandomNumber(%d)", result, seed)
		instructionFinished = true
	case "and":
		opCodeLength = 5
	case "jumpRelative":
		opCodeLength = 3
		target := b.LE16(p.data, p.offset+1)
		instruction += fmt.Sprintf("0x%x", target)
	case "doSentence":
		opCodeLength = 7
		verb := p.data[p.offset+1]
		objA := b.LE16(p.data, p.offset+2)
		objB := b.LE16(p.data, p.offset+4)
		instruction += fmt.Sprintf("verb=%02x, objA=%02x, objB=%02x", verb, objA, objB)
	case "move":
		opCodeLength = 5
		result := b.LE16(p.data, p.offset+1)
		value := b.LE16(p.data, p.offset+3)
		instruction = fmt.Sprintf("var(%d) := %d", result, value)
		instructionFinished = true
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
	case "loadRoomWithEgo":
		opCodeLength = 8
		object := b.LE16(p.data, p.offset+1)
		room := p.data[p.offset+3]
		x := b.LE16(p.data, p.offset+4)
		y := b.LE16(p.data, p.offset+6)
		instruction += fmt.Sprintf("object=0x%x, room=%d, x=%d, y=%d", object, room, x, y)
	case "pickupObject":
		opCodeLength = 5
		object := b.LE16(p.data, p.offset+1)
		room := p.data[p.offset+3]
		instruction += fmt.Sprintf("object=0x%x, room=0x%x", object, room)
	case "setVarRange":
		opCodeLength = endsList
	case "stringOps":
		l.Log("script", "subopcode: 0x%x\n", subopcode)
		opCodeLength = varLen
		if subopcode == 0x02 || subopcode == 0x05 {
			opCodeLength = 5
		} else if subopcode == 0x04 {
			opCodeLength = 7
		}
	case "equalZero":
		opCodeLength = 5
		variable := varName(p.data[p.offset+1])
		target := b.LE16(p.data, p.offset+2)
		instruction = fmt.Sprintf("unless (%v == 0) goto 0x%x", variable, target)
		instructionFinished = true
	case "setOwnerOf":
		opCodeLength = 5
	case "delayVariable":
		opCodeLength = 3
	case "cursorCommand":
		opCodeLength = varLen
		if subopcode < 0x0a {
			opCodeLength = 2
			instruction = cursorCommands[subopcode] + "()"
			instructionFinished = true
		}
	case "putActorInRoom":
		opCodeLength = 3
		actor := p.data[p.offset+1]
		room := p.data[p.offset+2]
		instruction += fmt.Sprintf("actor=%d, room=%d", actor, room)
	case "delay":
		opCodeLength = 4
		delay := b.LE24(p.data, p.offset+1)
		instruction += fmt.Sprintf("%d", delay)
	case "ifNotState":
		opCodeLength = 6
	case "matrixOp":
		if subopcode == 0x04 {
			opCodeLength = 2
		} else {
			opCodeLength = 6
		}
	case "getInventoryCount":
		opCodeLength = 5
	case "setCameraAt":
		opCodeLength = 3
	case "roomOps":
		opCodeLength = varLen
		switch subopcode {
		case 0x01:
			opCodeLength = 6
			instruction = fmt.Sprintf(
				"roomOps.scroll(minX = %d, maxX = %d)",
				b.LE16(p.data, p.offset+2),
				b.LE16(p.data, p.offset+4))
		case 0x03:
			opCodeLength = 6
			instruction = fmt.Sprintf(
				"roomOps.screen(b = %d, h = %d)",
				b.LE16(p.data, p.offset+2),
				b.LE16(p.data, p.offset+4))
		case 0x06:
			opCodeLength = 2
			instruction = "roomOps.ShakeOff()"
		case 0x05:
			opCodeLength = 2
			instruction = "roomOps.ShakeOn()"
		case 0x0A:
			opCodeLength = 4
			instruction = fmt.Sprintf("roomOps.effect(%v)",
				b.LE16(p.data, p.offset+2))
		}
		instructionFinished = true
	case "getDist":
		opCodeLength = 7
	case "findObject":
		opCodeLength = 7
	case "walkActorToObject":
		opCodeLength = 4
		actor := p.data[p.offset+1]
		object := b.LE16(p.data, p.offset+2)
		instruction += fmt.Sprintf("actor=0x%x, object=0x%x", actor, object)
	case "startObject":
		opCodeLength = endsList
	case "lessOrEqual":
		opCodeLength = 7
		variable := varName(p.data[p.offset+1])
		value := b.LE16(p.data, p.offset+2)
		target := b.LE16(p.data, p.offset+4)
		instruction = fmt.Sprintf("unless (%v <= 0x%x) goto 0x%x", variable, value, target)
		instructionFinished = true
	case "subtract":
		opCodeLength = 5
		result := b.LE16(p.data, p.offset+1)
		value := p.data[p.offset+3]
		instruction = fmt.Sprintf("0x%x = 0x%x - 0x%x)", result, result, value)
		instructionFinished = true
	case "getActorScale":
		opCodeLength = 5
	case "stopSound":
		opCodeLength = 3
	case "findInventory":
		opCodeLength = 7
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
	case "chainScript":
		opCodeLength = endsList
	case "getActorX":
		opCodeLength = 5
	case "isLess":
		opCodeLength = 7
		variable := varName(p.data[p.offset+1])
		value := b.LE16(p.data, p.offset+2)
		target := b.LE16(p.data, p.offset+4)
		instruction = fmt.Sprintf("unless (0x%x < %v) goto 0x%x", value, variable, target)
		instructionFinished = true
	case "increment":
		opCodeLength = 3
		variable := b.LE16(p.data, p.offset+1)
		instruction = fmt.Sprintf("Var[%d]++", variable)
		instructionFinished = true
	case "isEqual":
		opCodeLength = 7
		variable := varName(p.data[p.offset+1])
		value := b.LE16(p.data, p.offset+2)
		target := b.LE16(p.data, p.offset+4)
		instruction = fmt.Sprintf("unless (0x%x == %v) goto 0x%x", value, variable, target)
		instructionFinished = true
	case "soundKludge":
		items := p.parseList(p.offset + 1)
		opCodeLength = 2 + len(items)*3
		instruction += fmt.Sprintf("%v", items)
	case "actorFollowCamera":
		opCodeLength = 3
	case "setObjectName":
		opCodeLength = 3
		object := b.LE16(p.data, p.offset+1)
		name := ""
		for p.data[p.offset+opCodeLength] != 0x00 {
			name += string(p.data[p.offset+opCodeLength])
			opCodeLength++
		}
		instruction += fmt.Sprintf("object = %d, text = \"%v\"", object, name)
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
	case "actorSetClass":
		object := b.LE16(p.data, p.offset+1)
		list := p.parseList(p.offset + 3)
		opCodeLength = 4 + len(list)*3
		instruction += fmt.Sprintf("actor=%d, %v", object, list)
	case "freezeScripts":
		opCodeLength = 3
	case "stopScript":
		opCodeLength = 2
		instruction += fmt.Sprintf("%x", p.data[p.offset+1])
	case "getActorFacing":
		opCodeLength = 5
	case "getClosestObjActor":
		opCodeLength = 5
	case "getStringWidth":
		opCodeLength = 5
	case "getScriptRunning":
		opCodeLength = 4
		//result := p.data, p.offset + 1
		script := b.LE16(p.data, p.offset+2)
		instruction = fmt.Sprintf("VAR_RESULT = isScriptRunning(%02x);", script)
		instructionFinished = true
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
	case "isGreater":
		opCodeLength = 7
		variable := varName(p.data[p.offset+1])
		value := b.LE16(p.data, p.offset+2)
		target := b.LE16(p.data, p.offset+4)
		instruction = fmt.Sprintf("unless (0x%x > %v) goto 0x%x", value, variable, target)
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
	case "notEqualZero":
		opCodeLength = 5
		variable := varName(p.data[p.offset+1])
		target := b.LE16(p.data, p.offset+2)
		instruction = fmt.Sprintf("unless (%v != 0) goto 0x%x", variable, target)
		instructionFinished = true
	case "saveRestoreVerbs":
		opCodeLength = 4
	case "expression": //TODO properly
		opCodeLength = 1
		for p.data[p.offset+int(opCodeLength)] != 0xff {
			opCodeLength++
		}
		opCodeLength++
	case "wait":
		opCodeLength = 2
		if subopcode == 0x01 {
			opCodeLength = 3
			instruction = fmt.Sprintf("wait.forActor(%d)",
				p.data[p.offset+2])
			instructionFinished = true

		}
	case "cutscene":
		list := p.parseList(p.offset + 1)
		opCodeLength = 2 + len(list)*3
		instruction += fmt.Sprintf("%v", list)
	case "endCutScene":
		opCodeLength = 1
	case "decrement":
		opCodeLength = 3
		variable := b.LE16(p.data, p.offset+1)
		instruction = fmt.Sprintf("Var[%d]--", variable)
	case "pseudoRoom":
		opCodeLength = varLen
	case "print", "printEgo":
		if opcodeName == "print" {
			instruction += fmt.Sprintf("actor = %d, ", p.data[p.offset+1])
			opCodeLength = 2
		} else {
			opCodeLength = 1
		}
		say, length := parsePrintOpcode(p.data, p.offset+opCodeLength)
		opCodeLength += length
		instruction += fmt.Sprintf("\"%v\"", say)
	}

	if opCodeLength == varLen {
		return "", errors.New("Variable length opcode " + fmt.Sprintf("%x", opcode) + ", cannot proceed")
	}

	p.offset += int(opCodeLength)
	if instructionFinished == false {
		instruction = instruction + ")"
	}

	p.script = append(p.script, instruction)
	return opcodeName, nil
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
			parser.script = append(parser.script, "error, "+err.Error())
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
