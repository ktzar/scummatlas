package script

import "fmt"
import "strconv"

//Operation types
const (
	_ = iota
	OpCall
	OpConditional
	OpAssignment
	OpError
)

type Operation struct {
	opType     int
	offset     int
	opCode     byte
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
	errorMsg   string
}

func (op Operation) GetMethod() string {
	if op.opType == OpCall {
		return op.callMethod
	} else {
		return ""
	}
}

type Script []Operation

func (script Script) Exit() (room int, hasExit bool) {
	room = 0
	hasExit = false
	for _, op := range script {
		if op.callMethod == "loadRoomWithEgo" ||
			op.callMethod == "putActorInRoom" {
			room, _ = strconv.Atoi(op.callMap["room"])
			hasExit = true
			break
		}
	}
	return
}

func (script Script) Debug() string {
	out := ""
	for _, op := range script {
		out += fmt.Sprintf("[%04x] (%02x) %v\n",
			op.offset, op.opCode, op.String())
	}
	return out
}

func (script Script) Print() string {
	out := ""
	for _, op := range script {
		out += op.String() + "\n"
	}
	return out
}

func (op *Operation) addNamedStringParam(paramName string, value string) {
	op.callMap[paramName] = "\"" + value + "\""
}

func (op *Operation) addNamedParam(paramName string, value int) {
	op.callMap[paramName] = fmt.Sprintf("%d", value)
}

func (op *Operation) addParam(param string) {
	op.callParams = append(op.callParams, param)
}

func (op *Operation) addResult(result string) {
	op.callResult = result
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
	} else if op.opType == OpError {
		return fmt.Sprintf("%v", op.errorMsg)
	}
	return ""
}
