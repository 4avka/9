package txscript

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// TestOpcodeDisabled tests the opcodeDisabled function manually because all disabled opcodes result in a script execution failure when executed normally, so the function is not called under normal circumstances.
func TestOpcodeDisabled(
	t *testing.T) {

	t.Parallel()
	tests := []byte{OpCat, OpSubstr, OpLeft, OpRight, OpInvert,
		OpAnd, OpOr, Op1Mul, Op2Div, OpMul, OpDiv, OpMod,
		OpLShift, OpRShift,
	}

	for _, opcodeVal := range tests {

		pop := parsedOpcode{opcode: &opcodeArray[opcodeVal], data: nil}
		err := opcodeDisabled(&pop, nil)

		if !IsErrorCode(err, ErrDisabledOpcode) {

			t.Errorf("opcodeDisabled: unexpected error - got %v, "+
				"want %v", err, ErrDisabledOpcode)
			continue
		}
	}
}

// TestOpcodeDisasm tests the print function for all opcodes in both the oneline and full modes to ensure it provides the expected disassembly.
func TestOpcodeDisasm(
	t *testing.T) {

	t.Parallel()

	// First, test the oneline disassembly. The expected strings for the data push opcodes are replaced in the test loops below since they involve repeating bytes.  Also, the OpNoOp# and OP_UNKNOWN# are replaced below too, since it's easier than manually listing them here.
	oneBytes := []byte{0x01}
	oneStr := "01"
	expectedStrings := [256]string{0x00: "0", 0x4f: "-1",
		0x50: "OpReserved", 0x61: "OpNoOp", 0x62: "OpVer",
		0x63: "OpIf", 0x64: "OpIfNot", 0x65: "OpVerIf",
		0x66: "OpVerIfNot", 0x67: "OpElse", 0x68: "OpEndIf",
		0x69: "OpVerify", 0x6a: "OpReturn", 0x6b: "OpToAltStack",
		0x6c: "OpFromAltStack", 0x6d: "Op2Drop", 0x6e: "Op2Dup",
		0x6f: "Op3Dup", 0x70: "Op2Over", 0x71: "Op2Rot",
		0x72: "Op2Swap", 0x73: "OpIfDup", 0x74: "OpDepth",
		0x75: "OpDrop", 0x76: "OpDup", 0x77: "OpNip",
		0x78: "OpOver", 0x79: "OpPick", 0x7a: "OpRoll",
		0x7b: "OpRot", 0x7c: "OpSwap", 0x7d: "OpTuck",
		0x7e: "OpCat", 0x7f: "OpSubstr", 0x80: "OpLeft",
		0x81: "OpRight", 0x82: "OpSize", 0x83: "OpInvert",
		0x84: "OpAnd", 0x85: "OpOr", 0x86: "OpXor",
		0x87: "OpEqual", 0x88: "OpEqualVerify", 0x89: "OpReserved1",
		0x8a: "OpReserved2", 0x8b: "Op1Add", 0x8c: "Op1Sub",
		0x8d: "Op1Mul", 0x8e: "Op2Div", 0x8f: "OpNegate",
		0x90: "OpAbs", 0x91: "OpNot", 0x92: "Op0NotEqual",
		0x93: "OpAdd", 0x94: "OpSub", 0x95: "OpMul", 0x96: "OpDiv",
		0x97: "OpMod", 0x98: "OpLShift", 0x99: "OpRShift",
		0x9a: "OpBoolAnd", 0x9b: "OpBoolOr", 0x9c: "OpNumEqual",
		0x9d: "OpNumEqualVerify", 0x9e: "OpNumNotEqual",
		0x9f: "OpLessThan", 0xa0: "OpGreaterThan",
		0xa1: "OpLessThanOrEqual", 0xa2: "OpGreaterThanOrEqual",
		0xa3: "OpMin", 0xa4: "OpMax", 0xa5: "OpWithin",
		0xa6: "OpRipeMD160", 0xa7: "OpSHA1", 0xa8: "OpSHA256",
		0xa9: "OpHash160", 0xaa: "OpHash256", 0xab: "OpCodeSeparator",
		0xac: "OpCheckSig", 0xad: "OpCheckSigVerify",
		0xae: "OpCheckMultiSig", 0xaf: "OpCheckMultiSigVerify",
		0xfa: "OpSmallInteger", 0xfb: "OpPubKeys",
		0xfd: "OpPubKeyHash", 0xfe: "OpPubKey",
		0xff: "OpInvalidOpCode",
	}

	for opcodeVal, expectedStr := range expectedStrings {

		var data []byte

		switch {

		// OpData1 through OpData65 display the pushed data.
		case opcodeVal >= 0x01 && opcodeVal < 0x4c:
			data = bytes.Repeat(oneBytes, opcodeVal)
			expectedStr = strings.Repeat(oneStr, opcodeVal)
		// OpPushData1.
		case opcodeVal == 0x4c:
			data = bytes.Repeat(oneBytes, 1)
			expectedStr = strings.Repeat(oneStr, 1)
		// OpPushData2.
		case opcodeVal == 0x4d:
			data = bytes.Repeat(oneBytes, 2)
			expectedStr = strings.Repeat(oneStr, 2)
		// OpPushData4.
		case opcodeVal == 0x4e:
			data = bytes.Repeat(oneBytes, 3)
			expectedStr = strings.Repeat(oneStr, 3)
		// Op1 through Op16 display the numbers themselves.
		case opcodeVal >= 0x51 && opcodeVal <= 0x60:
			val := byte(opcodeVal - (0x51 - 1))
			data = []byte{val}
			expectedStr = strconv.Itoa(int(val))
		// OpNoOp1 through OpNoOp10.
		case opcodeVal >= 0xb0 && opcodeVal <= 0xb9:

			switch opcodeVal {

			case 0xb1:
				// OpNoOp2 is an alias of OpCheckLockTimeVerify
				expectedStr = "OpCheckLockTimeVerify"
			case 0xb2:
				// OpNoOp3 is an alias of OpCheckSequenceVerify
				expectedStr = "OpCheckSequenceVerify"
			default:
				val := byte(opcodeVal - (0xb0 - 1))
				expectedStr = "OpNoOp" + strconv.Itoa(int(val))
			}
		// OP_UNKNOWN#.
		case opcodeVal >= 0xba && opcodeVal <= 0xf9 || opcodeVal == 0xfc:
			expectedStr = "OP_UNKNOWN" + strconv.Itoa(int(opcodeVal))
		}
		pop := parsedOpcode{opcode: &opcodeArray[opcodeVal], data: data}
		gotStr := pop.print(true)

		if gotStr != expectedStr {

			t.Errorf("pop.print (opcode %x): Unexpected disasm "+
				"string - got %v, want %v", opcodeVal, gotStr,
				expectedStr)
			continue
		}
	}

	// Now, replace the relevant fields and test the full disassembly.
	expectedStrings[0x00] = "OpZero"
	expectedStrings[0x4f] = "Op1Negate"

	for opcodeVal, expectedStr := range expectedStrings {

		var data []byte

		switch {

		// OpData1 through OpData65 display the opcode followed by the pushed data.
		case opcodeVal >= 0x01 && opcodeVal < 0x4c:
			data = bytes.Repeat(oneBytes, opcodeVal)
			expectedStr = fmt.Sprintf("OpData%d 0x%s", opcodeVal,
				strings.Repeat(oneStr, opcodeVal))
		// OpPushData1.
		case opcodeVal == 0x4c:
			data = bytes.Repeat(oneBytes, 1)
			expectedStr = fmt.Sprintf("OpPushData1 0x%02x 0x%s",
				len(data), strings.Repeat(oneStr, 1))
		// OpPushData2.
		case opcodeVal == 0x4d:
			data = bytes.Repeat(oneBytes, 2)
			expectedStr = fmt.Sprintf("OpPushData2 0x%04x 0x%s",
				len(data), strings.Repeat(oneStr, 2))
		// OpPushData4.
		case opcodeVal == 0x4e:
			data = bytes.Repeat(oneBytes, 3)
			expectedStr = fmt.Sprintf("OpPushData4 0x%08x 0x%s",
				len(data), strings.Repeat(oneStr, 3))
		// Op1 through Op16.
		case opcodeVal >= 0x51 && opcodeVal <= 0x60:
			val := byte(opcodeVal - (0x51 - 1))
			data = []byte{val}
			expectedStr = "OP_" + strconv.Itoa(int(val))
		// OpNoOp1 through OpNoOp10.
		case opcodeVal >= 0xb0 && opcodeVal <= 0xb9:

			switch opcodeVal {

			case 0xb1:
				// OpNoOp2 is an alias of OpCheckLockTimeVerify
				expectedStr = "OpCheckLockTimeVerify"
			case 0xb2:
				// OpNoOp3 is an alias of OpCheckSequenceVerify
				expectedStr = "OpCheckSequenceVerify"
			default:
				val := byte(opcodeVal - (0xb0 - 1))
				expectedStr = "OpNoOp" + strconv.Itoa(int(val))
			}
		// OP_UNKNOWN#.
		case opcodeVal >= 0xba && opcodeVal <= 0xf9 || opcodeVal == 0xfc:
			expectedStr = "OP_UNKNOWN" + strconv.Itoa(int(opcodeVal))
		}
		pop := parsedOpcode{opcode: &opcodeArray[opcodeVal], data: data}
		gotStr := pop.print(false)

		if gotStr != expectedStr {

			t.Errorf("pop.print (opcode %x): Unexpected disasm "+
				"string - got %v, want %v", opcodeVal, gotStr,
				expectedStr)
			continue
		}
	}
}
