package script

import "testing"

func TestPrintingConditionalOperations(t *testing.T) {
	a := Operation{
		opType:  OpConditional,
		condOp1: "Var[1]",
		condOp2: "Var[2]",
		condOp:  "==",
		condDst: 0x2345,
	}
	if a.String() != "unless (Var[1] == Var[2]) goto 2345" {
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
		callParams: []string{
			"123", "456",
		},
	}

	if a.String() != "myMethod(123, 456)" {
		t.Errorf("conditional operation `%v` is not properly formatted", a.String())
	}

	b := Operation{
		opType:     OpCall,
		callMethod: "myMethod",
		callMap: map[string]string{
			"x": "1234",
			"y": "6789",
		},
	}

	if b.String() != "myMethod(x=1234, y=6789)" &&
		b.String() != "myMethod(y=6789, x=1234)" {
		t.Errorf("conditional operation `%v` is not properly formatted", b.String())
	}
}
