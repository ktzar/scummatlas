package script

import "testing"

func TestPrint(t *testing.T) {
	methodName := "abcdef"
	op := Operation{
		opCode:     0x22,
		opType:     OpCall,
		callMethod: methodName,
		callParams: []string{"one", "two"},
		offset:     0x234,
	}

	script := Script{}
	script = append(script, op)
	script = append(script, op)
	result := script.Print()
	expected := "abcdef(one, two)\nabcdef(one, two)"

	if result != expected {
		t.Errorf("Was expecting %v, but got %v", expected, result)
	}
}

func TestDebug(t *testing.T) {
	methodName := "abcdef"
	op := Operation{
		opCode:     0x22,
		opType:     OpCall,
		callMethod: methodName,
		offset:     0x234,
	}

	script := Script{}
	script = append(script, op)
	script = append(script, op)
	result := script.Debug()
	expected := "[0234] (22) abcdef()\n"

	if result != expected+expected {
		t.Errorf("Was expecting %v, but got %v", expected+expected, result)
	}
}

func TestGetMethod(t *testing.T) {
	methodName := "abcdeg"
	op := Operation{
		opType:     OpCall,
		callMethod: methodName,
	}

	if op.GetMethod() != methodName {
		t.Errorf("GetMethod() should not be %v, but "+methodName, op.GetMethod())
	}

	op = Operation{
		opType:     OpAssignment,
		callMethod: "abcdefg",
	}

	if op.GetMethod() != "" {
		t.Errorf("GetMethod() should not be %v but an empty string", op.GetMethod())
	}
}

func TestPrintingConditionalOperations(t *testing.T) {
	a := Operation{
		opType:  OpConditional,
		condOp1: "Var[1]",
		condOp2: "Var[2]",
		condOp:  "isEqual",
		condDst: 0x2345,
	}
	if a.String() != "if (Var[1] == Var[2])" {
		t.Errorf("conditional operation `%v` is not properly formatted", a.String())
	}

	if a.Debug() != "unless (Var[1] == Var[2]) goto 2345" {
		t.Errorf("conditional operation `%v` is not properly formatted", a.String())
	}
}

func TestPrintingAssignmentOperations(t *testing.T) {
	a := Operation{
		opType:    OpAssignment,
		assignDst: "Var[123]",
		assignVal: "999",
	}
	if a.String() != "Var[123] = 999" {
		t.Errorf("operation `%v` is not properly formatted", a.String())
	}
}

func TestPrintingCallOperations(t *testing.T) {
	a := Operation{
		opType:     OpCall,
		callMethod: "myMethod",
		callResult: "result",
		callParams: []string{
			"123", "456",
		},
	}

	if a.String() != "result = myMethod(123, 456)" {
		t.Errorf("call operation `%v` is not properly formatted", a.String())
	}

	b := Operation{
		opType:     OpCall,
		callMethod: "myMethod",
		callMap: map[string]string{
			"x": "1234",
			"y": "6789",
		},
	}

	opPrint := b.String()
	if opPrint != "myMethod(x=1234, y=6789)" &&
		opPrint != "myMethod(y=6789, x=1234)" {
		t.Errorf("conditional operation `%v` is not properly formatted", opPrint)
	}
}
