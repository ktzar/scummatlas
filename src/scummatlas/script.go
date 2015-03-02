package scummatlas

import "fmt"

type Script []Opcode

type Opcode struct {
	raw    []byte
	name   string
	params []string
}

type ScriptParser struct {
	data   []byte
	offset int
	script []Opcode
}

func (p *ScriptParser) parseNext() {
	opcode := p.data[p.offset]
	subopcode := p.data[p.offset+1]
	opcodeName, ok := opCodesNames[opcode]

	if !ok {
		fmt.Printf("%x is not a known code\n", opcode)
		panic("Unknown code")
	}
	fmt.Printf("Opcode %x -> %v\n", opcode, opcodeName)

	var opCodeLength byte

	switch opcodeName {
	case "animateActor":
		opCodeLength = 5
	case "putActor":
		opCodeLength = 7
	case "startMusic":
		opCodeLength = 3
	case "getActorRoom":
		opCodeLength = 5
	case "isGreaterEqual":
		opCodeLength = 7
	case "drawObject":
		opCodeLength = varLen
	case "getActorElevation":
		opCodeLength = 5
	case "setState":
		opCodeLength = 5
	case "isNotEqual":
		opCodeLength = 7
	case "faceActor":
		opCodeLength = 5
	case "startScript":
		opCodeLength = 2
		for p.data[p.offset+int(opCodeLength)] != 0xff {
			opCodeLength += 2
		}
		opCodeLength++
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
	case "getObjectOwner":
		opCodeLength = 5
	case "panCameraTo":
		opCodeLength = 3
	case "actorOps":
		opCodeLength = varLen
	case "actorFromPos":
		opCodeLength = 5
	case "getRandomNumber":
		opCodeLength = 5
	case "and":
		opCodeLength = 5
	case "jumpRelative":
		opCodeLength = 3
	case "doSentence":
		opCodeLength = 7
	case "move":
		opCodeLength = 5
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
	case "pickupObject":
		opCodeLength = 5
	case "setVarRange":
		opCodeLength = endsList
	case "stringOps":
		fmt.Printf("subopcode: %x\n", subopcode)
		opCodeLength = varLen
		if subopcode == 0x02 || subopcode == 0x05 {
			opCodeLength = 5
		} else if subopcode == 0x04 {
			opCodeLength = 7
		}
	case "equalZero":
		opCodeLength = 5
	case "setOwnerOf":
		opCodeLength = 5
	case "delayVariable":
		opCodeLength = 3
	case "cursorCommand":
		opCodeLength = varLen
	case "putActorInRoom":
		opCodeLength = 5
	case "delay":
		opCodeLength = 6
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
		fmt.Printf(" subops %x\n", subopcode)
		opCodeLength = varLen
	case "getDist":
		opCodeLength = 7
	case "findObject":
		opCodeLength = 7
	case "walkActorToObject":
		opCodeLength = 4
	case "startObject":
		opCodeLength = endsList
	case "lessOrEqual":
		opCodeLength = 7
	case "subtract":
		opCodeLength = 5
	case "getActorScale":
		opCodeLength = 5
	case "stopSound":
		opCodeLength = 3
	case "findInventory":
		opCodeLength = 7
	case "drawBox":
		opCodeLength = 12
	case "chainScript":
		opCodeLength = endsList
	case "getActorX":
		opCodeLength = 5
	case "isLess":
		opCodeLength = 7
	case "increment":
		opCodeLength = 2
	case "isEqual":
		opCodeLength = 7
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
		opCodeLength = varLen
	case "add":
		opCodeLength = 5
	case "divide":
		opCodeLength = 5
	case "oldRoomEffect":
		opCodeLength = 4
	case "actorSetClass":
		opCodeLength = endsList
	case "freezeScripts":
		opCodeLength = 3
	case "stopScript":
		opCodeLength = 3
	case "getActorFacing":
		opCodeLength = 5
	case "getClosestObjActor":
		opCodeLength = 5
	case "getStringWidth":
		opCodeLength = 5
	case "getScriptRunning":
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
	case "isGreater":
		opCodeLength = 7
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
	case "saveRestoreVerbs":
		opCodeLength = 4
	case "expression":
		opCodeLength = varLen
	case "wait":
		opCodeLength = 2
		if subopcode == 0x01 {
			opCodeLength = 4
		}
	case "cutscene":
		opCodeLength = 2
		for p.data[p.offset+int(opCodeLength)] != 0xff {
			opCodeLength++
		}
		opCodeLength++
	case "endCutScene":
		opCodeLength = 1
	case "decrement":
		opCodeLength = 2
	case "pseudoRoom":
		opCodeLength = varLen
	case "print", "printEgo":
		opCodeLength = varLen
		switch subopcode {
		case 0x0F:
			opCodeLength = 2
			say := ""
			for {
				currChar := p.data[p.offset+int(opCodeLength)]
				if currChar == 0xff {
					escapeChar := p.data[p.offset+int(opCodeLength+1)]
					switch {
					case 0x01 <= escapeChar && escapeChar <= 0x03:
						opCodeLength += 2
					case 0x04 <= escapeChar && escapeChar <= 0x0e:
						opCodeLength += 4
					}
				} else if currChar >= 0x20 && currChar <= 0x7e { //printable ascii char
					opCodeLength++
					say += string(currChar)
				} else if currChar >= 0x00 {
					opCodeLength++
					break
				} else {
					panic("Invalid character in print")
				}
				if opCodeLength > 200 {
					break
				}
			}
			if opCodeLength > 200 {
				panic("print too long")
			}
			fmt.Printf("\tSay \"%v\"\n", say)
		case 0x01, 0x02:
			opCodeLength = 4
		case 0x00, 0x03, 0x08:
			opCodeLength = 6
		case 0x04, 0x06, 0x07:
			opCodeLength = 2
		}
	}

	if opCodeLength == varLen {
		panic("Variable length opcode, cannot proceed")
	}

	p.offset += int(opCodeLength)
}

func parseScriptBlock(data []byte) Script {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in ", r)
		}
	}()
	parser := new(ScriptParser)
	parser.data = data
	parser.offset = 0
	for parser.offset+1 < len(data) {
		parser.parseNext()
	}
	return parser.script
}

const notDefined byte = 0xFF
const varLen byte = 0xFE
const multi byte = 0xFD
const endsList byte = 0xFC

type opHandler func([]byte) (int, string)

type opDef struct {
	name    string
	handler opHandler
}

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
	0xc8: "isEqual",
	0xa3: "getActorY",
	0xc3: "getActorX",
	0xd6: "getActorMoving",
	0xe1: "putActor",
	0x6a: "startScript",
	0x91: "animateActor",
	0x93: "getInventoryCount",
	0xff: "drawBox",
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
