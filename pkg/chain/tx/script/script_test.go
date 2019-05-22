package txscript

import (
	"bytes"
	"reflect"
	"testing"

	"git.parallelcoin.io/dev/9/pkg/chain/wire"
)

// TestParseOpcode tests for opcode parsing with bad data templates.
func TestParseOpcode(
	t *testing.T) {

	// Deep copy the array and make one of the opcodes invalid by setting it to the wrong length.
	fakeArray := opcodeArray
	fakeArray[OpPushData4] = opcode{value: OpPushData4,
		name: "OpPushData4", length: -8, opfunc: opcodePushData}

	// This script would be fine if -8 was a valid length.
	_, err := parseScriptTemplate([]byte{OpPushData4, 0x1, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00}, &fakeArray)
	if err == nil {

		t.Errorf("no error with dodgy opcode array!")
	}
}

// TestUnparsingInvalidOpcodes tests for errors when unparsing invalid parsed opcodes.
func TestUnparsingInvalidOpcodes(
	t *testing.T) {

	tests := []struct {
		name        string
		pop         *parsedOpcode
		expectedErr error
	}{
		{
			name: "OpFalse",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpFalse],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpFalse long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpFalse],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData1 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData1],
				data:   nil,
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData1",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData1],
				data:   make([]byte, 1),
			},
			expectedErr: nil,
		},
		{
			name: "OpData1 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData1],
				data:   make([]byte, 2),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData2 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData2],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData2",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData2],
				data:   make([]byte, 2),
			},
			expectedErr: nil,
		},
		{
			name: "OpData2 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData2],
				data:   make([]byte, 3),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData3 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData3],
				data:   make([]byte, 2),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData3",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData3],
				data:   make([]byte, 3),
			},
			expectedErr: nil,
		},
		{
			name: "OpData3 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData3],
				data:   make([]byte, 4),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData4 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData4],
				data:   make([]byte, 3),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData4",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData4],
				data:   make([]byte, 4),
			},
			expectedErr: nil,
		},
		{
			name: "OpData4 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData4],
				data:   make([]byte, 5),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData5 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData5],
				data:   make([]byte, 4),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData5",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData5],
				data:   make([]byte, 5),
			},
			expectedErr: nil,
		},
		{
			name: "OpData5 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData5],
				data:   make([]byte, 6),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData6 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData6],
				data:   make([]byte, 5),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData6",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData6],
				data:   make([]byte, 6),
			},
			expectedErr: nil,
		},
		{
			name: "OpData6 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData6],
				data:   make([]byte, 7),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData7 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData7],
				data:   make([]byte, 6),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData7",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData7],
				data:   make([]byte, 7),
			},
			expectedErr: nil,
		},
		{
			name: "OpData7 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData7],
				data:   make([]byte, 8),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData8 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData8],
				data:   make([]byte, 7),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData8",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData8],
				data:   make([]byte, 8),
			},
			expectedErr: nil,
		},
		{
			name: "OpData8 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData8],
				data:   make([]byte, 9),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData9 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData9],
				data:   make([]byte, 8),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData9",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData9],
				data:   make([]byte, 9),
			},
			expectedErr: nil,
		},
		{
			name: "OpData9 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData9],
				data:   make([]byte, 10),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData10 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData10],
				data:   make([]byte, 9),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData10",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData10],
				data:   make([]byte, 10),
			},
			expectedErr: nil,
		},
		{
			name: "OpData10 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData10],
				data:   make([]byte, 11),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData11 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData11],
				data:   make([]byte, 10),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData11",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData11],
				data:   make([]byte, 11),
			},
			expectedErr: nil,
		},
		{
			name: "OpData11 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData11],
				data:   make([]byte, 12),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData12 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData12],
				data:   make([]byte, 11),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData12",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData12],
				data:   make([]byte, 12),
			},
			expectedErr: nil,
		},
		{
			name: "OpData12 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData12],
				data:   make([]byte, 13),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData13 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData13],
				data:   make([]byte, 12),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData13",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData13],
				data:   make([]byte, 13),
			},
			expectedErr: nil,
		},
		{
			name: "OpData13 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData13],
				data:   make([]byte, 14),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData14 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData14],
				data:   make([]byte, 13),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData14",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData14],
				data:   make([]byte, 14),
			},
			expectedErr: nil,
		},
		{
			name: "OpData14 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData14],
				data:   make([]byte, 15),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData15 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData15],
				data:   make([]byte, 14),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData15",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData15],
				data:   make([]byte, 15),
			},
			expectedErr: nil,
		},
		{
			name: "OpData15 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData15],
				data:   make([]byte, 16),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData16 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData16],
				data:   make([]byte, 15),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData16",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData16],
				data:   make([]byte, 16),
			},
			expectedErr: nil,
		},
		{
			name: "OpData16 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData16],
				data:   make([]byte, 17),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData17 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData17],
				data:   make([]byte, 16),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData17",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData17],
				data:   make([]byte, 17),
			},
			expectedErr: nil,
		},
		{
			name: "OpData17 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData17],
				data:   make([]byte, 18),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData18 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData18],
				data:   make([]byte, 17),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData18",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData18],
				data:   make([]byte, 18),
			},
			expectedErr: nil,
		},
		{
			name: "OpData18 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData18],
				data:   make([]byte, 19),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData19 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData19],
				data:   make([]byte, 18),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData19",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData19],
				data:   make([]byte, 19),
			},
			expectedErr: nil,
		},
		{
			name: "OpData19 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData19],
				data:   make([]byte, 20),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData20 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData20],
				data:   make([]byte, 19),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData20",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData20],
				data:   make([]byte, 20),
			},
			expectedErr: nil,
		},
		{
			name: "OpData20 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData20],
				data:   make([]byte, 21),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData21 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData21],
				data:   make([]byte, 20),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData21",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData21],
				data:   make([]byte, 21),
			},
			expectedErr: nil,
		},
		{
			name: "OpData21 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData21],
				data:   make([]byte, 22),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData22 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData22],
				data:   make([]byte, 21),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData22",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData22],
				data:   make([]byte, 22),
			},
			expectedErr: nil,
		},
		{
			name: "OpData22 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData22],
				data:   make([]byte, 23),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData23 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData23],
				data:   make([]byte, 22),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData23",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData23],
				data:   make([]byte, 23),
			},
			expectedErr: nil,
		},
		{
			name: "OpData23 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData23],
				data:   make([]byte, 24),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData24 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData24],
				data:   make([]byte, 23),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData24",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData24],
				data:   make([]byte, 24),
			},
			expectedErr: nil,
		},
		{
			name: "OpData24 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData24],
				data:   make([]byte, 25),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData25 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData25],
				data:   make([]byte, 24),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData25",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData25],
				data:   make([]byte, 25),
			},
			expectedErr: nil,
		},
		{
			name: "OpData25 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData25],
				data:   make([]byte, 26),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData26 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData26],
				data:   make([]byte, 25),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData26",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData26],
				data:   make([]byte, 26),
			},
			expectedErr: nil,
		},
		{
			name: "OpData26 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData26],
				data:   make([]byte, 27),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData27 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData27],
				data:   make([]byte, 26),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData27",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData27],
				data:   make([]byte, 27),
			},
			expectedErr: nil,
		},
		{
			name: "OpData27 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData27],
				data:   make([]byte, 28),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData28 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData28],
				data:   make([]byte, 27),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData28",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData28],
				data:   make([]byte, 28),
			},
			expectedErr: nil,
		},
		{
			name: "OpData28 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData28],
				data:   make([]byte, 29),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData29 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData29],
				data:   make([]byte, 28),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData29",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData29],
				data:   make([]byte, 29),
			},
			expectedErr: nil,
		},
		{
			name: "OpData29 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData29],
				data:   make([]byte, 30),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData30 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData30],
				data:   make([]byte, 29),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData30",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData30],
				data:   make([]byte, 30),
			},
			expectedErr: nil,
		},
		{
			name: "OpData30 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData30],
				data:   make([]byte, 31),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData31 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData31],
				data:   make([]byte, 30),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData31",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData31],
				data:   make([]byte, 31),
			},
			expectedErr: nil,
		},
		{
			name: "OpData31 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData31],
				data:   make([]byte, 32),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData32 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData32],
				data:   make([]byte, 31),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData32",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData32],
				data:   make([]byte, 32),
			},
			expectedErr: nil,
		},
		{
			name: "OpData32 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData32],
				data:   make([]byte, 33),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData33 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData33],
				data:   make([]byte, 32),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData33",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData33],
				data:   make([]byte, 33),
			},
			expectedErr: nil,
		},
		{
			name: "OpData33 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData33],
				data:   make([]byte, 34),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData34 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData34],
				data:   make([]byte, 33),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData34",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData34],
				data:   make([]byte, 34),
			},
			expectedErr: nil,
		},
		{
			name: "OpData34 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData34],
				data:   make([]byte, 35),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData35 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData35],
				data:   make([]byte, 34),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData35",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData35],
				data:   make([]byte, 35),
			},
			expectedErr: nil,
		},
		{
			name: "OpData35 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData35],
				data:   make([]byte, 36),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData36 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData36],
				data:   make([]byte, 35),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData36",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData36],
				data:   make([]byte, 36),
			},
			expectedErr: nil,
		},
		{
			name: "OpData36 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData36],
				data:   make([]byte, 37),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData37 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData37],
				data:   make([]byte, 36),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData37",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData37],
				data:   make([]byte, 37),
			},
			expectedErr: nil,
		},
		{
			name: "OpData37 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData37],
				data:   make([]byte, 38),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData38 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData38],
				data:   make([]byte, 37),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData38",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData38],
				data:   make([]byte, 38),
			},
			expectedErr: nil,
		},
		{
			name: "OpData38 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData38],
				data:   make([]byte, 39),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData39 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData39],
				data:   make([]byte, 38),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData39",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData39],
				data:   make([]byte, 39),
			},
			expectedErr: nil,
		},
		{
			name: "OpData39 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData39],
				data:   make([]byte, 40),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData40 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData40],
				data:   make([]byte, 39),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData40",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData40],
				data:   make([]byte, 40),
			},
			expectedErr: nil,
		},
		{
			name: "OpData40 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData40],
				data:   make([]byte, 41),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData41 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData41],
				data:   make([]byte, 40),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData41",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData41],
				data:   make([]byte, 41),
			},
			expectedErr: nil,
		},
		{
			name: "OpData41 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData41],
				data:   make([]byte, 42),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData42 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData42],
				data:   make([]byte, 41),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData42",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData42],
				data:   make([]byte, 42),
			},
			expectedErr: nil,
		},
		{
			name: "OpData42 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData42],
				data:   make([]byte, 43),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData43 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData43],
				data:   make([]byte, 42),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData43",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData43],
				data:   make([]byte, 43),
			},
			expectedErr: nil,
		},
		{
			name: "OpData43 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData43],
				data:   make([]byte, 44),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData44 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData44],
				data:   make([]byte, 43),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData44",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData44],
				data:   make([]byte, 44),
			},
			expectedErr: nil,
		},
		{
			name: "OpData44 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData44],
				data:   make([]byte, 45),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData45 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData45],
				data:   make([]byte, 44),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData45",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData45],
				data:   make([]byte, 45),
			},
			expectedErr: nil,
		},
		{
			name: "OpData45 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData45],
				data:   make([]byte, 46),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData46 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData46],
				data:   make([]byte, 45),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData46",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData46],
				data:   make([]byte, 46),
			},
			expectedErr: nil,
		},
		{
			name: "OpData46 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData46],
				data:   make([]byte, 47),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData47 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData47],
				data:   make([]byte, 46),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData47",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData47],
				data:   make([]byte, 47),
			},
			expectedErr: nil,
		},
		{
			name: "OpData47 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData47],
				data:   make([]byte, 48),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData48 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData48],
				data:   make([]byte, 47),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData48",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData48],
				data:   make([]byte, 48),
			},
			expectedErr: nil,
		},
		{
			name: "OpData48 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData48],
				data:   make([]byte, 49),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData49 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData49],
				data:   make([]byte, 48),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData49",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData49],
				data:   make([]byte, 49),
			},
			expectedErr: nil,
		},
		{
			name: "OpData49 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData49],
				data:   make([]byte, 50),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData50 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData50],
				data:   make([]byte, 49),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData50",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData50],
				data:   make([]byte, 50),
			},
			expectedErr: nil,
		},
		{
			name: "OpData50 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData50],
				data:   make([]byte, 51),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData51 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData51],
				data:   make([]byte, 50),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData51",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData51],
				data:   make([]byte, 51),
			},
			expectedErr: nil,
		},
		{
			name: "OpData51 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData51],
				data:   make([]byte, 52),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData52 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData52],
				data:   make([]byte, 51),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData52",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData52],
				data:   make([]byte, 52),
			},
			expectedErr: nil,
		},
		{
			name: "OpData52 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData52],
				data:   make([]byte, 53),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData53 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData53],
				data:   make([]byte, 52),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData53",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData53],
				data:   make([]byte, 53),
			},
			expectedErr: nil,
		},
		{
			name: "OpData53 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData53],
				data:   make([]byte, 54),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData54 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData54],
				data:   make([]byte, 53),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData54",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData54],
				data:   make([]byte, 54),
			},
			expectedErr: nil,
		},
		{
			name: "OpData54 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData54],
				data:   make([]byte, 55),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData55 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData55],
				data:   make([]byte, 54),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData55",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData55],
				data:   make([]byte, 55),
			},
			expectedErr: nil,
		},
		{
			name: "OpData55 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData55],
				data:   make([]byte, 56),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData56 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData56],
				data:   make([]byte, 55),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData56",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData56],
				data:   make([]byte, 56),
			},
			expectedErr: nil,
		},
		{
			name: "OpData56 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData56],
				data:   make([]byte, 57),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData57 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData57],
				data:   make([]byte, 56),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData57",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData57],
				data:   make([]byte, 57),
			},
			expectedErr: nil,
		},
		{
			name: "OpData57 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData57],
				data:   make([]byte, 58),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData58 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData58],
				data:   make([]byte, 57),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData58",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData58],
				data:   make([]byte, 58),
			},
			expectedErr: nil,
		},
		{
			name: "OpData58 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData58],
				data:   make([]byte, 59),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData59 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData59],
				data:   make([]byte, 58),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData59",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData59],
				data:   make([]byte, 59),
			},
			expectedErr: nil,
		},
		{
			name: "OpData59 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData59],
				data:   make([]byte, 60),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData60 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData60],
				data:   make([]byte, 59),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData60",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData60],
				data:   make([]byte, 60),
			},
			expectedErr: nil,
		},
		{
			name: "OpData60 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData60],
				data:   make([]byte, 61),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData61 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData61],
				data:   make([]byte, 60),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData61",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData61],
				data:   make([]byte, 61),
			},
			expectedErr: nil,
		},
		{
			name: "OpData61 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData61],
				data:   make([]byte, 62),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData62 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData62],
				data:   make([]byte, 61),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData62",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData62],
				data:   make([]byte, 62),
			},
			expectedErr: nil,
		},
		{
			name: "OpData62 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData62],
				data:   make([]byte, 63),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData63 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData63],
				data:   make([]byte, 62),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData63",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData63],
				data:   make([]byte, 63),
			},
			expectedErr: nil,
		},
		{
			name: "OpData63 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData63],
				data:   make([]byte, 64),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData64 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData64],
				data:   make([]byte, 63),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData64",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData64],
				data:   make([]byte, 64),
			},
			expectedErr: nil,
		},
		{
			name: "OpData64 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData64],
				data:   make([]byte, 65),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData65 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData65],
				data:   make([]byte, 64),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData65",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData65],
				data:   make([]byte, 65),
			},
			expectedErr: nil,
		},
		{
			name: "OpData65 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData65],
				data:   make([]byte, 66),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData66 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData66],
				data:   make([]byte, 65),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData66",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData66],
				data:   make([]byte, 66),
			},
			expectedErr: nil,
		},
		{
			name: "OpData66 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData66],
				data:   make([]byte, 67),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData67 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData67],
				data:   make([]byte, 66),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData67",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData67],
				data:   make([]byte, 67),
			},
			expectedErr: nil,
		},
		{
			name: "OpData67 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData67],
				data:   make([]byte, 68),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData68 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData68],
				data:   make([]byte, 67),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData68",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData68],
				data:   make([]byte, 68),
			},
			expectedErr: nil,
		},
		{
			name: "OpData68 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData68],
				data:   make([]byte, 69),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData69 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData69],
				data:   make([]byte, 68),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData69",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData69],
				data:   make([]byte, 69),
			},
			expectedErr: nil,
		},
		{
			name: "OpData69 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData69],
				data:   make([]byte, 70),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData70 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData70],
				data:   make([]byte, 69),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData70",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData70],
				data:   make([]byte, 70),
			},
			expectedErr: nil,
		},
		{
			name: "OpData70 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData70],
				data:   make([]byte, 71),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData71 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData71],
				data:   make([]byte, 70),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData71",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData71],
				data:   make([]byte, 71),
			},
			expectedErr: nil,
		},
		{
			name: "OpData71 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData71],
				data:   make([]byte, 72),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData72 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData72],
				data:   make([]byte, 71),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData72",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData72],
				data:   make([]byte, 72),
			},
			expectedErr: nil,
		},
		{
			name: "OpData72 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData72],
				data:   make([]byte, 73),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData73 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData73],
				data:   make([]byte, 72),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData73",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData73],
				data:   make([]byte, 73),
			},
			expectedErr: nil,
		},
		{
			name: "OpData73 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData73],
				data:   make([]byte, 74),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData74 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData74],
				data:   make([]byte, 73),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData74",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData74],
				data:   make([]byte, 74),
			},
			expectedErr: nil,
		},
		{
			name: "OpData74 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData74],
				data:   make([]byte, 75),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData75 short",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData75],
				data:   make([]byte, 74),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpData75",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData75],
				data:   make([]byte, 75),
			},
			expectedErr: nil,
		},
		{
			name: "OpData75 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpData75],
				data:   make([]byte, 76),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpPushData1",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpPushData1],
				data:   []byte{0, 1, 2, 3, 4},
			},
			expectedErr: nil,
		},
		{
			name: "OpPushData2",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpPushData2],
				data:   []byte{0, 1, 2, 3, 4},
			},
			expectedErr: nil,
		},
		{
			name: "OpPushData4",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpPushData1],
				data:   []byte{0, 1, 2, 3, 4},
			},
			expectedErr: nil,
		},
		{
			name: "Op1Negate",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op1Negate],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op1Negate long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op1Negate],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpReserved",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpReserved],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpReserved long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpReserved],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpTrue",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpTrue],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpTrue long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpTrue],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op2",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op2 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op2",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op2 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op3",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op3],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op3 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op3],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op4",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op4],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op4 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op4],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op5",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op5],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op5 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op5],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op6",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op6],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op6 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op6],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op7",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op7],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op7 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op7],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op8",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op8],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op8 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op8],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op9",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op9],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op9 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op9],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op10",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op10],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op10 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op10],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op11",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op11],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op11 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op11],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op12",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op12],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op12 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op12],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op13",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op13],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op13 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op13],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op14",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op14],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op14 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op14],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op15",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op15],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op15 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op15],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op16",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op16],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op16 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op16],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpNoOp",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNoOp],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpNoOp long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNoOp],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpVer",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpVer],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpVer long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpVer],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpIf",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpIf],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpIf long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpIf],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpIfNot",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpIfNot],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpIfNot long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpIfNot],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpVerIf",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpVerIf],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpVerIf long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpVerIf],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpVerIfNot",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpVerIfNot],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpVerIfNot long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpVerIfNot],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpElse",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpElse],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpElse long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpElse],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpEndIf",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpEndIf],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpEndIf long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpEndIf],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpVerify",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpVerify],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpVerify long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpVerify],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpReturn",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpReturn],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpReturn long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpReturn],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpToAltStack",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpToAltStack],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpToAltStack long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpToAltStack],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpFromAltStack",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpFromAltStack],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpFromAltStack long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpFromAltStack],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op2Drop",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Drop],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op2Drop long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Drop],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op2Dup",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Dup],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op2Dup long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Dup],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op3Dup",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op3Dup],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op3Dup long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op3Dup],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op2Over",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Over],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op2Over long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Over],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op2Rot",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Rot],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op2Rot long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Rot],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op2Swap",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Swap],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op2Swap long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Swap],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpIfDup",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpIfDup],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpIfDup long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpIfDup],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpDepth",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpDepth],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpDepth long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpDepth],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpDrop",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpDrop],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpDrop long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpDrop],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpDup",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpDup],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpDup long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpDup],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpNip",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNip],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpNip long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNip],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpOver",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpOver],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpOver long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpOver],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpPick",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpPick],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpPick long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpPick],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpRoll",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpRoll],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpRoll long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpRoll],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpRot",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpRot],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpRot long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpRot],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpSwap",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpSwap],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpSwap long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpSwap],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpTuck",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpTuck],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpTuck long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpTuck],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpCat",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpCat],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpCat long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpCat],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpSubstr",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpSubstr],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpSubstr long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpSubstr],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpLeft",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpLeft],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpLeft long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpLeft],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpLeft",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpLeft],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpLeft long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpLeft],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpRight",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpRight],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpRight long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpRight],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpSize",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpSize],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpSize long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpSize],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpInvert",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpInvert],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpInvert long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpInvert],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpAnd",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpAnd],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpAnd long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpAnd],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpOr",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpOr],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpOr long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpOr],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpXor",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpXor],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpXor long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpXor],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpEqual",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpEqual],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpEqual long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpEqual],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpEqualVerify",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpEqualVerify],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpEqualVerify long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpEqualVerify],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpReserved1",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpReserved1],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpReserved1 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpReserved1],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpReserved2",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpReserved2],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpReserved2 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpReserved2],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op1Add",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op1Add],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op1Add long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op1Add],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op1Sub",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op1Sub],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op1Sub long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op1Sub],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op1Mul",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op1Mul],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op1Mul long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op1Mul],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op2Div",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Div],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op2Div long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op2Div],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpNegate",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNegate],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpNegate long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNegate],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpAbs",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpAbs],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpAbs long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpAbs],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpNot",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNot],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpNot long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNot],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "Op0NotEqual",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op0NotEqual],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "Op0NotEqual long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[Op0NotEqual],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpAdd",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpAdd],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpAdd long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpAdd],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpSub",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpSub],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpSub long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpSub],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpMul",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpMul],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpMul long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpMul],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpDiv",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpDiv],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpDiv long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpDiv],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpMod",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpMod],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpMod long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpMod],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpLShift",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpLShift],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpLShift long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpLShift],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpRShift",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpRShift],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpRShift long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpRShift],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpBoolAnd",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpBoolAnd],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpBoolAnd long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpBoolAnd],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpBoolOr",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpBoolOr],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpBoolOr long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpBoolOr],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpNumEqual",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNumEqual],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpNumEqual long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNumEqual],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpNumEqualVerify",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNumEqualVerify],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpNumEqualVerify long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNumEqualVerify],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpNumNotEqual",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNumNotEqual],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpNumNotEqual long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNumNotEqual],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpLessThan",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpLessThan],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpLessThan long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpLessThan],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpGreaterThan",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpGreaterThan],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpGreaterThan long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpGreaterThan],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpLessThanOrEqual",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpLessThanOrEqual],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpLessThanOrEqual long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpLessThanOrEqual],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpGreaterThanOrEqual",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpGreaterThanOrEqual],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpGreaterThanOrEqual long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpGreaterThanOrEqual],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpMin",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpMin],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpMin long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpMin],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpMax",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpMax],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpMax long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpMax],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpWithin",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpWithin],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpWithin long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpWithin],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpRipeMD160",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpRipeMD160],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpRipeMD160 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpRipeMD160],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpSHA1",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpSHA1],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpSHA1 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpSHA1],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpSHA256",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpSHA256],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpSHA256 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpSHA256],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpHash160",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpHash160],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpHash160 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpHash160],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpHash256",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpHash256],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpHash256 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpHash256],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OP_CODESAPERATOR",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpCodeSeparator],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpCodeSeparator long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpCodeSeparator],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpCheckSig",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpCheckSig],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpCheckSig long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpCheckSig],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpCheckSigVerify",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpCheckSigVerify],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpCheckSigVerify long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpCheckSigVerify],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpCheckMultiSig",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpCheckMultiSig],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpCheckMultiSig long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpCheckMultiSig],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpCheckMultiSigVerify",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpCheckMultiSigVerify],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpCheckMultiSigVerify long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpCheckMultiSigVerify],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpNoOp1",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNoOp1],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpNoOp1 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNoOp1],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpNoOp2",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNoOp2],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpNoOp2 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNoOp2],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpNoOp3",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNoOp3],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpNoOp3 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNoOp3],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpNoOp4",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNoOp4],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpNoOp4 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNoOp4],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpNoOp5",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNoOp5],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpNoOp5 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNoOp5],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpNoOp6",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNoOp6],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpNoOp6 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNoOp6],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpNoOp7",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNoOp7],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpNoOp7 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNoOp7],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpNoOp8",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNoOp8],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpNoOp8 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNoOp8],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpNoOp9",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNoOp9],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpNoOp9 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNoOp9],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpNoOp10",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNoOp10],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpNoOp10 long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpNoOp10],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpPubKeyHash",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpPubKeyHash],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpPubKeyHash long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpPubKeyHash],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpPubKey",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpPubKey],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpPubKey long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpPubKey],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
		{
			name: "OpInvalidOpCode",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpInvalidOpCode],
				data:   nil,
			},
			expectedErr: nil,
		},
		{
			name: "OpInvalidOpCode long",
			pop: &parsedOpcode{
				opcode: &opcodeArray[OpInvalidOpCode],
				data:   make([]byte, 1),
			},
			expectedErr: scriptError(ErrInternal, ""),
		},
	}

	for _, test := range tests {

		_, err := test.pop.bytes()

		if e := tstCheckScriptError(err, test.expectedErr); e != nil {

			t.Errorf("Parsed opcode test '%s': %v", test.name, e)
			continue
		}
	}
}

// TestPushedData ensured the PushedData function extracts the expected data out of various scripts.
func TestPushedData(
	t *testing.T) {

	t.Parallel()
	var tests = []struct {
		script string
		out    [][]byte
		valid  bool
	}{
		{
			"0 IF 0 ELSE 2 ENDIF",
			[][]byte{nil, nil},
			true,
		},
		{
			"16777216 10000000",
			[][]byte{
				{0x00, 0x00, 0x00, 0x01}, // 16777216
				{0x80, 0x96, 0x98, 0x00}, // 10000000
			},
			true,
		},
		{
			"DUP HASH160 '17VZNX1SN5NtKa8UQFxwQbFeFc3iqRYhem' EQUALVERIFY CHECKSIG",
			[][]byte{
				// 17VZNX1SN5NtKa8UQFxwQbFeFc3iqRYhem
				{
					0x31, 0x37, 0x56, 0x5a, 0x4e, 0x58, 0x31, 0x53, 0x4e, 0x35,
					0x4e, 0x74, 0x4b, 0x61, 0x38, 0x55, 0x51, 0x46, 0x78, 0x77,
					0x51, 0x62, 0x46, 0x65, 0x46, 0x63, 0x33, 0x69, 0x71, 0x52,
					0x59, 0x68, 0x65, 0x6d,
				},
			},
			true,
		},
		{
			"PUSHDATA4 1000 EQUAL",
			nil,
			false,
		},
	}

	for i, test := range tests {

		script := mustParseShortForm(test.script)
		data, err := PushedData(script)

		if test.valid && err != nil {

			t.Errorf("TestPushedData failed test #%d: %v\n", i, err)
			continue
		} else if !test.valid && err == nil {

			t.Errorf("TestPushedData failed test #%d: test should "+
				"be invalid\n", i)
			continue
		}

		if !reflect.DeepEqual(data, test.out) {

			t.Errorf("TestPushedData failed test #%d: want: %x "+
				"got: %x\n", i, test.out, data)
		}
	}
}

// TestHasCanonicalPush ensures the canonicalPush function works as expected.
func TestHasCanonicalPush(
	t *testing.T) {

	t.Parallel()

	for i := 0; i < 65535; i++ {

		script, err := NewScriptBuilder().AddInt64(int64(i)).Script()

		if err != nil {

			t.Errorf("Script: test #%d unexpected error: %v\n", i,
				err)
			continue
		}

		if result := IsPushOnlyScript(script); !result {

			t.Errorf("IsPushOnlyScript: test #%d failed: %x\n", i,
				script)
			continue
		}
		pops, err := parseScript(script)

		if err != nil {

			t.Errorf("parseScript: #%d failed: %v", i, err)
			continue
		}

		for _, pop := range pops {

			if result := canonicalPush(pop); !result {

				t.Errorf("canonicalPush: test #%d failed: %x\n",
					i, script)
				break
			}
		}
	}

	for i := 0; i <= MaxScriptElementSize; i++ {

		builder := NewScriptBuilder()
		builder.AddData(bytes.Repeat([]byte{0x49}, i))
		script, err := builder.Script()

		if err != nil {

			t.Errorf("StandardPushesTests test #%d unexpected error: %v\n", i, err)
			continue
		}

		if result := IsPushOnlyScript(script); !result {

			t.Errorf("StandardPushesTests IsPushOnlyScript test #%d failed: %x\n", i, script)
			continue
		}
		pops, err := parseScript(script)

		if err != nil {

			t.Errorf("StandardPushesTests #%d failed to TstParseScript: %v", i, err)
			continue
		}

		for _, pop := range pops {

			if result := canonicalPush(pop); !result {

				t.Errorf("StandardPushesTests TstHasCanonicalPushes test #%d failed: %x\n", i, script)
				break
			}
		}
	}
}

// TestGetPreciseSigOps ensures the more precise signature operation counting mechanism which includes signatures in P2SH scripts works as expected.
func TestGetPreciseSigOps(
	t *testing.T) {

	t.Parallel()
	tests := []struct {
		name      string
		scriptSig []byte
		nSigOps   int
	}{
		{
			name:      "scriptSig doesn't parse",
			scriptSig: mustParseShortForm("PUSHDATA1 0x02"),
		},
		{
			name:      "scriptSig isn't push only",
			scriptSig: mustParseShortForm("1 DUP"),
			nSigOps:   0,
		},
		{
			name:      "scriptSig length 0",
			scriptSig: nil,
			nSigOps:   0,
		},
		{
			name: "No script at the end",
			// No script at end but still push only.
			scriptSig: mustParseShortForm("1 1"),
			nSigOps:   0,
		},
		{
			name:      "pushed script doesn't parse",
			scriptSig: mustParseShortForm("DATA_2 PUSHDATA1 0x02"),
		},
	}

	// The signature in the p2sh script is nonsensical for the tests since this script will never be executed.  What matters is that it matches the right pattern.
	pkScript := mustParseShortForm("HASH160 DATA_20 0x433ec2ac1ffa1b7b7d0" +
		"27f564529c57197f9ae88 EQUAL")

	for _, test := range tests {

		count := GetPreciseSigOpCount(test.scriptSig, pkScript, true)

		if count != test.nSigOps {

			t.Errorf("%s: expected count of %d, got %d", test.name,
				test.nSigOps, count)
		}
	}
}

// TestGetWitnessSigOpCount tests that the sig op counting for p2wkh, p2wsh, nested p2sh, and invalid variants are counted properly.
func TestGetWitnessSigOpCount(
	t *testing.T) {

	t.Parallel()
	tests := []struct {
		name      string
		sigScript []byte
		pkScript  []byte
		witness   wire.TxWitness
		numSigOps int
	}{
		// A regualr p2wkh witness program. The output being spent should only have a single sig-op counted.
		{
			name: "p2wkh",
			pkScript: mustParseShortForm("OpZero DATA_20 " +
				"0x365ab47888e150ff46f8d51bce36dcd680f1283f"),
			witness: wire.TxWitness{
				hexToBytes("3045022100ee9fe8f9487afa977" +
					"6647ebcf0883ce0cd37454d7ce19889d34ba2c9" +
					"9ce5a9f402200341cb469d0efd3955acb9e46" +
					"f568d7e2cc10f9084aaff94ced6dc50a59134ad01"),
				hexToBytes("03f0000d0639a22bfaf217e4c9428" +
					"9c2b0cc7fa1036f7fd5d9f61a9d6ec153100e"),
			},
			numSigOps: 1,
		},
		// A p2wkh witness program nested within a p2sh output script. The pattern should be recognized properly and attribute only a single sig op.
		{
			name: "nested p2sh",
			sigScript: hexToBytes("160014ad0ffa2e387f07" +
				"e7ead14dc56d5a97dbd6ff5a23"),
			pkScript: mustParseShortForm("HASH160 DATA_20 " +
				"0xb3a84b564602a9d68b4c9f19c2ea61458ff7826c EQUAL"),
			witness: wire.TxWitness{
				hexToBytes("3045022100cb1c2ac1ff1d57d" +
					"db98f7bdead905f8bf5bcc8641b029ce8eef25" +
					"c75a9e22a4702203be621b5c86b771288706be5" +
					"a7eee1db4fceabf9afb7583c1cc6ee3f8297b21201"),
				hexToBytes("03f0000d0639a22bfaf217e4c9" +
					"4289c2b0cc7fa1036f7fd5d9f61a9d6ec153100e"),
			},
			numSigOps: 1,
		},
		// A p2sh script that spends a 2-of-2 multi-sig output.
		{
			name:      "p2wsh multi-sig spend",
			numSigOps: 2,
			pkScript: hexToBytes("0020e112b88a0cd87ba387f" +
				"449d443ee2596eb353beb1f0351ab2cba8909d875db23"),
			witness: wire.TxWitness{
				hexToBytes("522103b05faca7ceda92b493" +
					"3f7acdf874a93de0dc7edc461832031cd69cbb1d1e" +
					"6fae2102e39092e031c1621c902e3704424e8d8" +
					"3ca481d4d4eeae1b7970f51c78231207e52ae"),
			},
		},
		// A p2wsh witness program. However, the witness script fails to parse after the valid portion of the script. As a result, the valid portion of the script should still be counted.
		{
			name:      "witness script doesn't parse",
			numSigOps: 1,
			pkScript: hexToBytes("0020e112b88a0cd87ba387f44" +
				"9d443ee2596eb353beb1f0351ab2cba8909d875db23"),
			witness: wire.TxWitness{
				mustParseShortForm("DUP HASH160 " +
					"'17VZNX1SN5NtKa8UQFxwQbFeFc3iqRYhem'" +
					" EQUALVERIFY CHECKSIG DATA_20 0x91"),
			},
		},
	}

	for _, test := range tests {

		count := GetWitnessSigOpCount(test.sigScript, test.pkScript,
			test.witness)

		if count != test.numSigOps {

			t.Errorf("%s: expected count of %d, got %d", test.name,
				test.numSigOps, count)
		}
	}
}

// TestRemoveOpcodes ensures that removing opcodes from scripts behaves as expected.
func TestRemoveOpcodes(
	t *testing.T) {

	t.Parallel()
	tests := []struct {
		name   string
		before string
		remove byte
		err    error
		after  string
	}{
		{
			// Nothing to remove.
			name:   "nothing to remove",
			before: "NOP",
			remove: OpCodeSeparator,
			after:  "NOP",
		},
		{
			// Test basic opcode removal.
			name:   "codeseparator 1",
			before: "NOP CODESEPARATOR TRUE",
			remove: OpCodeSeparator,
			after:  "NOP TRUE",
		},
		{
			// The opcode in question is actually part of the data in a previous opcode.
			name:   "codeseparator by coincidence",
			before: "NOP DATA_1 CODESEPARATOR TRUE",
			remove: OpCodeSeparator,
			after:  "NOP DATA_1 CODESEPARATOR TRUE",
		},
		{
			name:   "invalid opcode",
			before: "CAT",
			remove: OpCodeSeparator,
			after:  "CAT",
		},
		{
			name:   "invalid length (instruction)",
			before: "PUSHDATA1",
			remove: OpCodeSeparator,
			err:    scriptError(ErrMalformedPush, ""),
		},
		{
			name:   "invalid length (data)",
			before: "PUSHDATA1 0xff 0xfe",
			remove: OpCodeSeparator,
			err:    scriptError(ErrMalformedPush, ""),
		},
	}

	// tstRemoveOpcode is a convenience function to parse the provided raw script, remove the passed opcode, then unparse the result back into a raw script.
	tstRemoveOpcode := func(script []byte, opcode byte) ([]byte, error) {

		pops, err := parseScript(script)

		if err != nil {

			return nil, err
		}
		pops = removeOpcode(pops, opcode)
		return unparseScript(pops)
	}

	for _, test := range tests {

		before := mustParseShortForm(test.before)
		after := mustParseShortForm(test.after)
		result, err := tstRemoveOpcode(before, test.remove)

		if e := tstCheckScriptError(err, test.err); e != nil {

			t.Errorf("%s: %v", test.name, e)
			continue
		}

		if !bytes.Equal(after, result) {

			t.Errorf("%s: value does not equal expected: exp: %q"+
				" got: %q", test.name, after, result)
		}
	}
}

// TestRemoveOpcodeByData ensures that removing data carrying opcodes based on the data they contain works as expected.
func TestRemoveOpcodeByData(
	t *testing.T) {

	t.Parallel()
	tests := []struct {
		name   string
		before []byte
		remove []byte
		err    error
		after  []byte
	}{
		{
			name:   "nothing to do",
			before: []byte{OpNoOp},
			remove: []byte{1, 2, 3, 4},
			after:  []byte{OpNoOp},
		},
		{
			name:   "simple case",
			before: []byte{OpData4, 1, 2, 3, 4},
			remove: []byte{1, 2, 3, 4},
			after:  nil,
		},
		{
			name:   "simple case (miss)",
			before: []byte{OpData4, 1, 2, 3, 4},
			remove: []byte{1, 2, 3, 5},
			after:  []byte{OpData4, 1, 2, 3, 4},
		},
		{
			// padded to keep it canonical.
			name: "simple case (pushdata1)",
			before: append(append([]byte{OpPushData1, 76},
				bytes.Repeat([]byte{0}, 72)...),
				[]byte{1, 2, 3, 4}...),
			remove: []byte{1, 2, 3, 4},
			after:  nil,
		},
		{
			name: "simple case (pushdata1 miss)",
			before: append(append([]byte{OpPushData1, 76},
				bytes.Repeat([]byte{0}, 72)...),
				[]byte{1, 2, 3, 4}...),
			remove: []byte{1, 2, 3, 5},
			after: append(append([]byte{OpPushData1, 76},
				bytes.Repeat([]byte{0}, 72)...),
				[]byte{1, 2, 3, 4}...),
		},
		{
			name:   "simple case (pushdata1 miss noncanonical)",
			before: []byte{OpPushData1, 4, 1, 2, 3, 4},
			remove: []byte{1, 2, 3, 4},
			after:  []byte{OpPushData1, 4, 1, 2, 3, 4},
		},
		{
			name: "simple case (pushdata2)",
			before: append(append([]byte{OpPushData2, 0, 1},
				bytes.Repeat([]byte{0}, 252)...),
				[]byte{1, 2, 3, 4}...),
			remove: []byte{1, 2, 3, 4},
			after:  nil,
		},
		{
			name: "simple case (pushdata2 miss)",
			before: append(append([]byte{OpPushData2, 0, 1},
				bytes.Repeat([]byte{0}, 252)...),
				[]byte{1, 2, 3, 4}...),
			remove: []byte{1, 2, 3, 4, 5},
			after: append(append([]byte{OpPushData2, 0, 1},
				bytes.Repeat([]byte{0}, 252)...),
				[]byte{1, 2, 3, 4}...),
		},
		{
			name:   "simple case (pushdata2 miss noncanonical)",
			before: []byte{OpPushData2, 4, 0, 1, 2, 3, 4},
			remove: []byte{1, 2, 3, 4},
			after:  []byte{OpPushData2, 4, 0, 1, 2, 3, 4},
		},
		{
			// This is padded to make the push canonical.
			name: "simple case (pushdata4)",
			before: append(append([]byte{OpPushData4, 0, 0, 1, 0},
				bytes.Repeat([]byte{0}, 65532)...),
				[]byte{1, 2, 3, 4}...),
			remove: []byte{1, 2, 3, 4},
			after:  nil,
		},
		{
			name:   "simple case (pushdata4 miss noncanonical)",
			before: []byte{OpPushData4, 4, 0, 0, 0, 1, 2, 3, 4},
			remove: []byte{1, 2, 3, 4},
			after:  []byte{OpPushData4, 4, 0, 0, 0, 1, 2, 3, 4},
		},
		{
			// This is padded to make the push canonical.
			name: "simple case (pushdata4 miss)",
			before: append(append([]byte{OpPushData4, 0, 0, 1, 0},
				bytes.Repeat([]byte{0}, 65532)...), []byte{1, 2, 3, 4}...),
			remove: []byte{1, 2, 3, 4, 5},
			after: append(append([]byte{OpPushData4, 0, 0, 1, 0},
				bytes.Repeat([]byte{0}, 65532)...), []byte{1, 2, 3, 4}...),
		},
		{
			name:   "invalid opcode ",
			before: []byte{OpUnknown187},
			remove: []byte{1, 2, 3, 4},
			after:  []byte{OpUnknown187},
		},
		{
			name:   "invalid length (instruction)",
			before: []byte{OpPushData1},
			remove: []byte{1, 2, 3, 4},
			err:    scriptError(ErrMalformedPush, ""),
		},
		{
			name:   "invalid length (data)",
			before: []byte{OpPushData1, 255, 254},
			remove: []byte{1, 2, 3, 4},
			err:    scriptError(ErrMalformedPush, ""),
		},
	}

	// tstRemoveOpcodeByData is a convenience function to parse the provided raw script, remove the passed data, then unparse the result back into a raw script.
	tstRemoveOpcodeByData := func(script []byte, data []byte) ([]byte, error) {

		pops, err := parseScript(script)

		if err != nil {

			return nil, err
		}
		pops = removeOpcodeByData(pops, data)
		return unparseScript(pops)
	}

	for _, test := range tests {

		result, err := tstRemoveOpcodeByData(test.before, test.remove)

		if e := tstCheckScriptError(err, test.err); e != nil {

			t.Errorf("%s: %v", test.name, e)
			continue
		}

		if !bytes.Equal(test.after, result) {

			t.Errorf("%s: value does not equal expected: exp: %q"+
				" got: %q", test.name, test.after, result)
		}
	}
}

// TestIsPayToScriptHash ensures the IsPayToScriptHash function returns the expected results for all the scripts in scriptClassTests.
func TestIsPayToScriptHash(
	t *testing.T) {

	t.Parallel()

	for _, test := range scriptClassTests {

		script := mustParseShortForm(test.script)
		shouldBe := (test.class == ScriptHashTy)
		p2sh := IsPayToScriptHash(script)

		if p2sh != shouldBe {

			t.Errorf("%s: expected p2sh %v, got %v", test.name,
				shouldBe, p2sh)
		}
	}
}

// TestIsPayToWitnessScriptHash ensures the IsPayToWitnessScriptHash function returns the expected results for all the scripts in scriptClassTests.
func TestIsPayToWitnessScriptHash(
	t *testing.T) {

	t.Parallel()

	for _, test := range scriptClassTests {

		script := mustParseShortForm(test.script)
		shouldBe := (test.class == WitnessV0ScriptHashTy)
		p2wsh := IsPayToWitnessScriptHash(script)

		if p2wsh != shouldBe {

			t.Errorf("%s: expected p2wsh %v, got %v", test.name,
				shouldBe, p2wsh)
		}
	}
}

// TestIsPayToWitnessPubKeyHash ensures the IsPayToWitnessPubKeyHash function returns the expected results for all the scripts in scriptClassTests.
func TestIsPayToWitnessPubKeyHash(
	t *testing.T) {

	t.Parallel()

	for _, test := range scriptClassTests {

		script := mustParseShortForm(test.script)
		shouldBe := (test.class == WitnessV0PubKeyHashTy)
		p2wkh := IsPayToWitnessPubKeyHash(script)

		if p2wkh != shouldBe {

			t.Errorf("%s: expected p2wkh %v, got %v", test.name,
				shouldBe, p2wkh)
		}
	}
}

// TestHasCanonicalPushes ensures the canonicalPush function properly determines what is considered a canonical push for the purposes of removeOpcodeByData.
func TestHasCanonicalPushes(
	t *testing.T) {

	t.Parallel()
	tests := []struct {
		name     string
		script   string
		expected bool
	}{
		{
			name: "does not parse",
			script: "0x046708afdb0fe5548271967f1a67130b7105cd6a82" +
				"8e03909a67962e0ea1f61d",
			expected: false,
		},
		{
			name:     "non-canonical push",
			script:   "PUSHDATA1 0x04 0x01020304",
			expected: false,
		},
	}

	for i, test := range tests {

		script := mustParseShortForm(test.script)
		pops, err := parseScript(script)

		if err != nil {

			if test.expected {

				t.Errorf("TstParseScript #%d failed: %v", i, err)
			}
			continue
		}

		for _, pop := range pops {

			if canonicalPush(pop) != test.expected {

				t.Errorf("canonicalPush: #%d (%s) wrong result"+
					"\ngot: %v\nwant: %v", i, test.name,
					true, test.expected)
				break
			}
		}
	}
}

// TestIsPushOnlyScript ensures the IsPushOnlyScript function returns the expected results.
func TestIsPushOnlyScript(
	t *testing.T) {

	t.Parallel()
	test := struct {
		name     string
		script   []byte
		expected bool
	}{
		name: "does not parse",
		script: mustParseShortForm("0x046708afdb0fe5548271967f1a67130" +
			"b7105cd6a828e03909a67962e0ea1f61d"),
		expected: false,
	}
	if IsPushOnlyScript(test.script) != test.expected {

		t.Errorf("IsPushOnlyScript (%s) wrong result\ngot: %v\nwant: "+
			"%v", test.name, true, test.expected)
	}
}

// TestIsUnspendable ensures the IsUnspendable function returns the expected results.
func TestIsUnspendable(
	t *testing.T) {

	t.Parallel()
	tests := []struct {
		name     string
		pkScript []byte
		expected bool
	}{
		{
			// Unspendable
			pkScript: []byte{0x6a, 0x04, 0x74, 0x65, 0x73, 0x74},
			expected: true,
		},
		{
			// Spendable
			pkScript: []byte{0x76, 0xa9, 0x14, 0x29, 0x95, 0xa0,
				0xfe, 0x68, 0x43, 0xfa, 0x9b, 0x95, 0x45,
				0x97, 0xf0, 0xdc, 0xa7, 0xa4, 0x4d, 0xf6,
				0xfa, 0x0b, 0x5c, 0x88, 0xac},
			expected: false,
		},
	}

	for i, test := range tests {

		res := IsUnspendable(test.pkScript)

		if res != test.expected {

			t.Errorf("TestIsUnspendable #%d failed: got %v want %v",
				i, res, test.expected)
			continue
		}
	}
}
