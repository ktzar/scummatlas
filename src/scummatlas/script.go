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
	currOpcode := p.data[p.offset]
	currOpcodeDef, ok := opCodesDefs[currOpcode]
	if !ok {
		fmt.Printf("%x is not a known code\n", currOpcode)
		panic("Unknown code")
	}
	fmt.Println("Curr OpcodeDef: ", currOpcodeDef.name)

	if currOpcodeDef.length == varLen {
		panic("Variable length opcode, cannot proceed")
	}
	if currOpcodeDef.length == endsList {
		for {
			p.offset += 1
			if p.data[p.offset] == 0xff {
				p.offset += 1
				return
			}
		}
	}
	if currOpcodeDef.length == multi {
		currSubOpcode := p.data[p.offset+1]
		currOpcodeDef, ok := subOpcodesDefs[currOpcode][currSubOpcode]
		if !ok {
			fmt.Printf("%x %x are not know code-subcode", currOpcode, currSubOpcode)
			panic("unknown code")
		}
		fmt.Println("\t->subOpcode: ", currOpcodeDef.name)
	}
	p.offset += int(currOpcodeDef.length)
}

func parseScriptBlock(data []byte) Script {
	parser := new(ScriptParser)
	parser.data = data
	parser.offset = 0
	for parser.offset < len(data) {
		parser.parseNext()
	}
	return parser.script
}

const notDefined byte = 0xFF
const varLen byte = 0xFE
const multi byte = 0xFD
const endsList byte = 0xFC

type opDef struct {
	length byte
	name   string
}

// varLen variable length
// notDefined not processed yet
var opCodesDefs = map[byte]opDef{
	0x00: opDef{3, "stopObjectCode"},
	0x01: opDef{6, "putActor"},
	0x02: opDef{2, "startMusic"},
	0x03: opDef{4, "getActorRoom"},
	0x04: opDef{7, "isGreaterEqual"},
	0x05: opDef{varLen, "drawObject"},
	0x06: opDef{4, "getActorElevation"},
	0x07: opDef{4, "setState"},
	0x08: opDef{7, "isNotEqual"},
	0x09: opDef{4, "faceActor"},
	0x0A: opDef{endsList, "startScript"},
	0x0B: opDef{7, "getVerbEntryPoint"},
	0x0C: opDef{multi, "resourceRoutines"},
	0x0D: opDef{4, "walkActorToActor"},
	0x0E: opDef{4, "putActorAtObject"},
	0x0F: opDef{5, "getObjectState"},
	//0x0F: opDef{6, "ifState"}, //not in v5
	0x10: opDef{5, "getObjectOwner"},
	0x12: opDef{3, "panCameraTo"},
	0x13: opDef{varLen, "actorOps"},
	0x14: opDef{varLen, "print"},
	0x15: opDef{5, "actorFromPos"},
	0x16: opDef{4, "getRandomNumber"},
	0x17: opDef{5, "and"},
	0x18: opDef{3, "jumpRelative"},
	0x19: opDef{6, "doSentence"},
	0x1A: opDef{5, "move"},
	0x1B: opDef{5, "multiply"},
	0x1C: opDef{2, "startSound"},
	0x1D: opDef{varLen, "ifClassOfIs"},
	0x1E: opDef{6, "walkActorTo"},
	0x1F: opDef{5, "isActorInBox"},
	0x20: opDef{1, "stopMusic"},
	0x22: opDef{4, "getAnimCounter"},
	//0x22: opDef{notDefined, "saveLoadGame"}, //Not in V5
	0x23: opDef{5, "getActorY"},
	0x24: opDef{8, "loadRoomWithEgo"},
	0x25: opDef{4, "pickupObject"},
	0x26: opDef{endsList, "setVarRange"},
	0x27: opDef{varLen, "stringOps"},
	0x28: opDef{5, "equalZero"},
	0x29: opDef{4, "setOwnerOf"},
	0x2b: opDef{3, "delayVariable"},
	0x2c: opDef{varLen, "cursorCommand"},
	0x2D: opDef{3, "putActorInRoom"},
	0x2e: opDef{4, "delay"},
	0x2F: opDef{6, "ifNotState"},
	0x30: opDef{varLen, "matrixOp"},
	0x31: opDef{4, "getInventoryCount"},
	0x32: opDef{3, "setCameraAt"},
	0x33: opDef{varLen, "roomOps"},
	0x34: opDef{7, "getDist"},
	0x35: opDef{5, "findObject"},
	0x36: opDef{4, "walkActorToObject"},
	0x37: opDef{endsList, "startObject"},
	0x38: opDef{7, "lessOrEqual"},
	0x3A: opDef{5, "subtract"},
	0x3B: opDef{4, "getActorScale"},
	0x3C: opDef{2, "stopSound"},
	0x3D: opDef{5, "findInventory"},
	0x3F: opDef{11, "drawBox"},
	0x40: opDef{endsList, "cutScene"},
	0x42: opDef{endsList, "chainScript"},
	0x43: opDef{5, "getActorX"},
	0x44: opDef{7, "isLess"},
	0x46: opDef{2, "increment"},
	0x48: opDef{7, "isEqual"},
	0x4C: opDef{endsList, "soundKludge"},
	0x50: opDef{3, "pickupObject"},
	0x52: opDef{2, "actorFollowCamera"},
	0x54: opDef{varLen, "setObjectName"},
	0x56: opDef{4, "getActorMoving"},
	0x57: opDef{5, "or"},
	0x58: opDef{varLen, "override"},
	0x5a: opDef{5, "add"},
	0x5B: opDef{5, "divide"},
	0x5C: opDef{4, "oldRoomEffect"},
	0x5d: opDef{endsList, "actorSetClass"},
	0x60: opDef{2, "freezeScripts"},
	0x62: opDef{2, "stopScript"},
	0x63: opDef{4, "getActorFacing"},
	0x66: opDef{5, "getClosestObjActor"},
	0x67: opDef{3, "getStringWidth"},
	0x68: opDef{4, "getScriptRunning"},
	0x6B: opDef{3, "debug"},
	0x6C: opDef{4, "getActorWidth"},
	0x6E: opDef{2, "stopObjectScript"},
	0x70: opDef{4, "lights"},
	0x71: opDef{4, "getActorCostume"},
	0x72: opDef{2, "loadRoom"},
	0x78: opDef{7, "isGreater"},
	0x7A: opDef{varLen, "verbOps"},
	0x7B: opDef{4, "getActorWalkBox"},
	0x7C: opDef{4, "isSoundRunning"},
	0x80: opDef{1, "breakHere"},
	0x98: opDef{2, "systemOps"},
	0xA0: opDef{1, "stopObjectCode"},
	0xA7: opDef{1, "dummy"},
	//0xA7: opDef{notDefined, "saveLoadVars"}, //Not in V5
	0xA8: opDef{5, "notEqualZero"},
	0xAB: opDef{4, "saveRestoreVerbs"},
	0xAC: opDef{varLen, "expression"},
	0xAE: opDef{varLen, "wait"},
	0xC0: opDef{1, "endCutScene"},
	0xc6: opDef{2, "decrement"},
	0xCC: opDef{varLen, "pseudoRoom"},
	0xD8: opDef{varLen, "printEgo"},
}

var subOpcodesDefs = map[byte]map[byte]opDef{
	0x0c: {
		0x01: opDef{3, "load_script"},
		0x02: opDef{3, "load_sound"},
		0x03: opDef{3, "load_costume"},
		0x04: opDef{3, "load_room"},
		0x05: opDef{3, "nuke_script"},
		0x06: opDef{3, "nuke_sound"},
		0x07: opDef{3, "nuke_costume"},
		0x08: opDef{3, "nuke_room"},
		0x09: opDef{3, "lock_script"},
		0x0a: opDef{3, "lock_sound"},
		0x0b: opDef{3, "lock_costume"},
		0x0c: opDef{3, "lock_room"},
		0x0d: opDef{3, "unlock_script"},
		0x0e: opDef{3, "unlock_sound"},
		0x0f: opDef{3, "unlock_costume"},
		0x10: opDef{3, "unlock_room"},
		0x11: opDef{3, "clear_heap"},
		0x12: opDef{3, "load_charset"},
		0x13: opDef{3, "nuke_charset"},
		0x14: opDef{3, "load_object"},
	},
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
