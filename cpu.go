package main

import (
	"fmt"
	"reflect"
)

func fetchDecodeExecute(st *st) {
	// Log the registers
	if flags.verbose {
		fmt.Printf("PC=0x%04X AF=0x%04X BC=0x%04X DE=0x%04X HL=0x%04X SP=0x%04X\n",
			PC.get(st), AF.get(st), BC.get(st), DE.get(st), HL.get(st), SP.get(st),
		)
	}

	// Fetch
	PC_0 := PC.get(st)
	opcode := st.readMem_u8(PC_0)

	// Decode
	var sizeOfOpcode u16
	var instr *instr
	if opcode != 0xCB {
		sizeOfOpcode = 1
		instr = jumpTable[opcode]
		if instr == nil {
			panic(fmt.Sprintf("Unknown opcode 0x%02X=0b%08b at 0x%04X.", opcode, opcode, PC_0))
		}
	} else {
		sizeOfOpcode = 2
		opcode = st.readMem_u8(PC_0 + 1)
		instr = extendedJumpTable[opcode]
		if instr == nil {
			panic(fmt.Sprintf("Unknown extended opcode 0xCB-0x%02X=0b%08b at 0x%04X.", opcode, opcode, PC_0))
		}
	}

	// Increment PC
	sizeOfInstr := sizeOfOpcode + instr.sizeOfOperands
	assert(1 <= sizeOfInstr && sizeOfInstr <= 3)
	PC.set(st, PC_0+sizeOfInstr)

	// Log the instruction
	if flags.verbose {
		r := func(offset u16) u8 { return st.readMem_u8(PC_0 + offset) }
		var instrBytes string
		switch sizeOfInstr {
		case 1:
			instrBytes = fmt.Sprintf("0x%02X          ", r(0))
		case 2:
			instrBytes = fmt.Sprintf("0x%02X 0x%02X     ", r(0), r(1))
		case 3:
			instrBytes = fmt.Sprintf("0x%02X 0x%02X 0x%02X", r(0), r(1), r(2))
		default:
			assert(false)
		}
		fmt.Println(" " + instrBytes + " | " + instr.toString(st))
	}

	// Execute
	instr.execute(st)
}

type instr struct {
	sizeOfOperands u16
	toString       func(*st) string
	execute        func(*st)
}

type operation struct {
	name string
	f    interface{}
}

type operand interface {
	sizeOf() u16
	toString(*st) string
}

func newInstr(operation operation, operands []operand) *instr {
	var sizeOfOperands u16 = 0
	for _, operand := range operands {
		sizeOfOperands += operand.sizeOf()
	}
	assert(0 <= sizeOfOperands && sizeOfOperands <= 2)

	toString := func(st *st) string {
		switch len(operands) {
		case 0:
			return operation.name
		case 1:
			x := operands[0]
			return operation.name + " " + x.toString(st)
		case 2:
			x := operands[0]
			y := operands[1]
			return operation.name + " " + x.toString(st) + ", " + y.toString(st)
		default:
			panic("Invalid number of operands.")
		}
	}

	// (This switch would not be necessary if Go had generics.)
	assert(reflect.TypeOf(operation.f).NumIn()-1 == len(operands))
	var execute func(*st)
	switch f := operation.f.(type) {
	case func(*st):
		execute = f
	case func(*st, r_i8):
		x := operands[0].(r_i8)
		execute = func(st *st) { f(st, x) }
	case func(*st, r_u16):
		x := operands[0].(r_u16)
		execute = func(st *st) { f(st, x) }
	case func(*st, r_u8):
		x := operands[0].(r_u8)
		execute = func(st *st) { f(st, x) }
	case func(*st, rw_u16):
		x := operands[0].(rw_u16)
		execute = func(st *st) { f(st, x) }
	case func(*st, rw_u8):
		x := operands[0].(rw_u8)
		execute = func(st *st) { f(st, x) }
	case func(*st, w_u16):
		x := operands[0].(w_u16)
		execute = func(st *st) { f(st, x) }
	case func(*st, r_bool, r_i8):
		x := operands[0].(r_bool)
		y := operands[1].(r_i8)
		execute = func(st *st) { f(st, x, y) }
	case func(*st, u3, r_u8):
		x := operands[0].(u3)
		y := operands[1].(r_u8)
		execute = func(st *st) { f(st, x, y) }
	case func(*st, w_u16, r_u16):
		x := operands[0].(w_u16)
		y := operands[1].(r_u16)
		execute = func(st *st) { f(st, x, y) }
	case func(*st, w_u8, r_u8):
		x := operands[0].(w_u8)
		y := operands[1].(r_u8)
		execute = func(st *st) { f(st, x, y) }
	default:
		panic(fmt.Sprintf("Unimplemented function type %T in %s.", operation.f, operation.name))
	}

	return &instr{sizeOfOperands, toString, execute}
}

type r_bool interface {
	operand
	get(*st) bool
}

type r_u8 interface {
	operand
	get(*st) u8
}

type w_u8 interface {
	operand
	set(*st, u8)
}

type rw_u8 interface {
	operand
	get(*st) u8
	set(*st, u8)
}

type r_i8 interface {
	operand
	get(*st) i8
}

type r_u16 interface {
	operand
	get(*st) u16
}

type w_u16 interface {
	operand
	set(*st, u16)
}

type rw_u16 interface {
	operand
	get(*st) u16
	set(*st, u16)
}

// ---------------------
// Constants as operands
// ---------------------

func (u3) sizeOf() u16 {
	return 0
}

func (x u3) toString(*st) string {
	return fmt.Sprint(x)
}

type FF00 struct {
	offset r_u8
}

func (ff00 FF00) sizeOf() u16 {
	return ff00.offset.sizeOf()
}

func (ff00 FF00) toString(st *st) string {
	return "0xFF00+" + ff00.offset.toString(st)
}

func (ff00 FF00) get(st *st) u16 {
	return u16(0xFF00) + u16(ff00.offset.get(st))
}
