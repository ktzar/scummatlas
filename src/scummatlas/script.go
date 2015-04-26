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

func (p *ScriptParser) parseList(offset int) (values []int) {
	for p.data[offset] != 0xFF {
		fmt.Printf("%x \n", p.data[offset])
		//TODO the first byte is supposed to always be 1 ???
		values = append(values, b.LE16(p.data, offset+1))
		fmt.Printf("%x\n", values)
		offset += 3
	}
	fmt.Println("length", len(values))
	return
}

func (p *ScriptParser) parseNext() (string, error) {
	opcode := p.data[p.offset]
	subopcode := p.data[p.offset+1]
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
		l.Log("script", "drawObject subcode 0x%x\n", subopcode)
		switch subopcode {
		case 0x01:
			opCodeLength = 6
		case 0x02:
			opCodeLength = 6
		case 0xff:
			opCodeLength = 4
		}
	case "getActorElevation":
		opCodeLength = 5
	case "setState":
		opCodeLength = 5
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

		list := p.parseList(p.offset + 3)
		opCodeLength = 3 + len(list)*3
		instruction += fmt.Sprintf("script=%d, %v", script, list)
	case "getVerbEntryPoint":
		opCodeLength = 7
	case "resourceRoutines":
		opCodeLength = 2
		switch subopcode {
		case 0x11:
			opCodeLength = 1
		case 0x14:
			opCodeLength = 4
		}
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
		opCodeLength = 5
	case "and":
		opCodeLength = 5
	case "jumpRelative":
		opCodeLength = 3
		target := b.LE16(p.data, p.offset+1)
		instruction += fmt.Sprintf("0x%x", target)
	case "doSentence":
		opCodeLength = 7
	case "move":
		opCodeLength = 5
		result := b.LE16(p.data, p.offset+1)
		value := b.LE16(p.data, p.offset+3)
		instruction = fmt.Sprintf("*(0x%x) := %d", result, value)
		instructionFinished = true
	case "multiply":
		opCodeLength = 5
	case "startSound":
		opCodeLength = 3
	case "ifClassOfIs":
		opCodeLength = varLen
	case "walkActorTo":
		opCodeLength = 7
	case "isActorInBox":
		opCodeLength = 7
	case "stopMusic":
		opCodeLength = 1
	case "getAnimCounter":
		opCodeLength = 5
	case "getActorY":
		opCodeLength = 5
	case "loadRoomWithEgo":
		opCodeLength = 9
		object := b.LE16(p.data, p.offset+1)
		room := p.data[p.offset+3]
		x := b.LE16(p.data, p.offset+4)
		y := b.LE16(p.data, p.offset+6)
		instruction += fmt.Sprintf("object=0x%x, room=0x%x, x=%d, y=%d", object, room, x, y)
	case "pickupObject":
		opCodeLength = 5
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
	case "putActorInRoom":
		opCodeLength = 3
		actor := p.data[p.offset+1]
		room := p.data[p.offset+2]
		instruction += fmt.Sprintf("actor=%d, room=%d", actor, room)
	case "delay":
		opCodeLength = 6
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
		l.Log("script", " subops 0x%x\n", subopcode)
		opCodeLength = varLen
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
	case "increment":
		opCodeLength = 3
		variable := b.LE16(p.data, p.offset+1)
		instruction = fmt.Sprintf("Var[%d]++", variable)
	case "isEqual":
		opCodeLength = 7
		variable := varName(p.data[p.offset+1])
		value := b.LE16(p.data, p.offset+2)
		target := b.LE16(p.data, p.offset+4)
		instruction = fmt.Sprintf("unless (0x%x == %v) goto 0x%x", value, variable, target)
		instructionFinished = true
	case "soundKludge":
		opCodeLength = endsList
	case "actorFollowCamera":
		opCodeLength = 3
	case "setObjectName":
		opCodeLength = varLen
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
		opCodeLength = 3 + len(list)*3
		instruction += fmt.Sprintf("object=%d, %v", object, list)
	case "freezeScripts":
		opCodeLength = 3
	case "stopScript":
		opCodeLength = 2
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
			opCodeLength = 4
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

var opCodesNames = map[byte]string{
	0x00: "stopObjectCode",
	0x01: "putActor",
	0x02: "startMusic",
	0x03: "getActorRoom",
	0x04: "isGreaterEqual",
	0x05: "drawObject",
	0x06: "getActorElevation",
	0x07: "setState",
	0x08: "isNotEqual",
	0x09: "faceActor",
	0x0a: "startScript",
	0x0b: "getVerbEntryPoint",
	0x0c: "resourceRoutines",
	0x0d: "walkActorToActor",
	0x0e: "putActorAtObject",
	0x0f: "getObjectState",
	0x10: "getObjectOwner",
	0x11: "animateActor",
	0x12: "panCameraTo",
	0x13: "actorOps",
	0x14: "print",
	0x15: "actorFromPos",
	0x16: "getRandomNumber",
	0x17: "and",
	0x18: "jumpRelative",
	0x19: "doSentence",
	0x1a: "move",
	0x1b: "multiply",
	0x1c: "startSound",
	0x1d: "ifClassOfIs",
	0x1e: "walkActorTo",
	0x1f: "isActorInBox",
	0x20: "stopMusic",
	0x22: "getAnimCounter",
	0x23: "getActorY",
	0x24: "loadRoomWithEgo",
	0x25: "pickupObject",
	0x26: "setVarRange",
	0x27: "stringOps",
	0x28: "equalZero",
	0x29: "setOwnerOf",
	0x2b: "delayVariable",
	0x2c: "cursorCommand",
	0x2d: "putActorInRoom",
	0x2e: "delay",
	0x2f: "ifNotState",
	0x30: "matrixOp",
	0x31: "getInventoryCount",
	0x32: "setCameraAt",
	0x33: "roomOps",
	0x34: "getDist",
	0x35: "findObject",
	0x36: "walkActorToObject",
	0x37: "startObject",
	0x38: "lessOrEqual",
	0x3a: "subtract",
	0x3b: "getActorScale",
	0x3c: "stopSound",
	0x3d: "findInventory",
	0x3f: "drawBox",
	0x40: "cutscene",
	0x42: "chainScript",
	0x43: "getActorX",
	0x44: "isLess",
	0x46: "increment",
	0x48: "isEqual",
	0x4c: "soundKludge",
	0x50: "pickupObject",
	0x52: "actorFollowCamera",
	0x54: "setObjectName",
	0x56: "getActorMoving",
	0x57: "or",
	0x58: "override",
	0x5a: "add",
	0x5b: "divide",
	0x5c: "oldRoomEffect",
	0x5d: "actorSetClass",
	0x60: "freezeScripts",
	0x62: "stopScript",
	0x63: "getActorFacing",
	0x66: "getClosestObjActor",
	0x67: "getStringWidth",
	0x68: "getScriptRunning",
	0x6b: "debug",
	0x6c: "getActorWidth",
	0x6e: "stopObjectScript",
	0x70: "lights",
	0x71: "getActorCostume",
	0x72: "loadRoom",
	0x78: "isGreater",
	0x7a: "verbOps",
	0x7b: "getActorWalkBox",
	0x7c: "isSoundRunning",
	0x80: "breakHere",
	0x98: "systemOps",
	0xa0: "stopObjectCode",
	0xa7: "dummy",
	0xa8: "notEqualZero",
	0xab: "saveRestoreVerbs",
	0xac: "expression",
	0xae: "wait",
	0xc0: "endCutScene",
	0xc6: "decrement",
	0xcc: "pseudoRoom",
	0xd8: "printEgo",

	//from ScummVM sourcecode
	0x59: "doSentence",
	0x61: "putActor",
	0x64: "loadRoomWithEgo",
	0x69: "setOwnerOf",
	0x6a: "startScript",
	0x6d: "putActorInRoom",
	0x74: "getDist",
	0x77: "startObject",
	0x8c: "resourceRoutines",
	0x8f: "getObjectState",
	0x91: "animateActor",
	0x93: "getInventoryCount",
	0x9a: "move",
	0xa3: "getActorY",
	0xba: "subtract",
	0xc3: "getActorX",
	0xc8: "isEqual",
	0xd6: "getActorMoving",
	0xe1: "putActor",
	0xff: "drawBox",
}

func varName(code uint8) (name string) {
	name = varNames[code]
	if name == "" {
		name = "var(" + fmt.Sprintf("0x%x", code) + ")"
	}
	return
}

var actorOps = map[byte]string{
	0x00: "dummy",
	0x01: "costume",
	0x02: "step_dist",
	0x03: "sound",
	0x04: "walk_animation",
	0x05: "talk_animation",
	0x06: "stand_animation",
	0x07: "animation",
	0x08: "default",
	0x09: "elevation",
	0x0a: "animation_default",
	0x0b: "palette",
	0x0c: "talk_color",
	0x0d: "actor_name",
	0x0e: "init_animation",
	0x10: "actor_width",
	0x11: "actor_scale",
	0x12: "never_zclip",
	0x13: "always_zclip",
	0x14: "ignore_boxes",
	0x15: "follow_boxes",
	0x16: "animation_speed",
	0x17: "shadow",
}

var varNames = map[byte]string{
	0:  "KEYPRESS",
	1:  "EGO",
	2:  "CAMERA_POS_X",
	3:  "HAVE_MSG",
	4:  "ROOM",
	5:  "OVERRIDE",
	6:  "MACHINE_SPEED",
	7:  "ME",
	8:  "NUM_ACTOR",
	9:  "CURRENT_LIGHTS",
	10: "CURRENTDRIVE",
	11: "TMR_1",
	12: "TMR_2",
	13: "TMR_3",
	14: "MUSIC_TIMER",
	15: "ACTOR_RANGE_MIN",
	16: "ACTOR_RANGE_MAX",
	17: "CAMERA_MIN_X",
	18: "CAMERA_MAX_X",
	19: "TIMER_NEXT",
	20: "VIRT_MOUSE_X",
	21: "VIRT_MOUSE_Y",
	22: "ROOM_RESOURCE",
	23: "LAST_SOUND",
	24: "CUTSCENEEXIT_KEY",
	25: "TALK_ACTOR",
	26: "CAMERA_FAST_X",
	27: "SCROLL_SCRIPT",
	28: "ENTRY_SCRIPT",
	29: "ENTRY_SCRIPT2",
	30: "EXIT_SCRIPT",
	31: "EXIT_SCRIPT2",
	32: "VERB_SCRIPT",
	33: "SENTENCE_SCRIPT",
	34: "INVENTORY_SCRIPT",
	35: "CUTSCENE_START_SCRIPT",
	36: "CUTSCENE_END_SCRIPT",
	37: "CHARINC",
	38: "WALKTO_OBJ",
	39: "DEBUGMODE",
	40: "HEAPSPACE",
	42: "RESTART_KEY",
	43: "PAUSE_KEY",
	44: "MOUSE_X",
	45: "MOUSE_Y",
	46: "TIMER",
	47: "TIMER_TOTAL",
	48: "SOUNDCARD",
	49: "VIDEOMODE",
	50: "MAINMENU_KEY",
	51: "FIXEDDISK",
	52: "CURSORSTATE",
	53: "USERPUT",
	54: "V5_TALK_STRING_Y",
	56: "SOUNDRESULT",
	57: "TALKSTOP_KEY",
	59: "FADE_DELAY",
	60: "NOSUBTITLES",
	64: "SOUNDPARAM",
	65: "SOUNDPARAM2",
	66: "SOUNDPARAM3",
	67: "INPUTMODE",
	68: "MEMORY_PERFORMANCE",
	69: "VIDEO_PERFORMANCE",
	70: "ROOM_FLAG",
	71: "GAME_LOADED",
	72: "NEW_ROOM",
}
