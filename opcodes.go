/*
 * gammaboy is a Game Boy emulator.
 * Copyright (C) 2018  gammpei
 *
 * This file is part of gammaboy.
 *
 * gammaboy is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * gammaboy is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with gammaboy.  If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"strconv"
)

func buildJumpTables() {
	buildJumpTable()
	buildExtendedJumpTable()
}

var R1 = map[string]rw_u16{
	"0": BC,
	"1": DE,
}

var R2 = map[string]rw_u16{
	"00": BC,
	"01": DE,
	"10": HL,
	"11": SP,
}

var R3 = map[string]rw_u16{
	"00": BC,
	"01": DE,
	"10": HL,
	"11": AF,
}

var D_operands = map[string]rw_u8{
	"000": B,
	"001": C,
	"010": D,
	"011": E,
	"100": H,
	"101": L,
	"110": mem{HL},
	"111": A,
}

var F_operands = map[string]r_bool{
	"00": F.NZ,
	"01": F.Z,
	"10": F.NC,
	"11": F.C,
}

var ALU = map[string]operation{
	"000": ADD1,
	"010": SUB,
	"101": XOR,
	"111": CP,
}

var N = map[string]u3{
	"000": 0,
	"001": 1,
	"010": 2,
	"011": 3,
	"100": 4,
	"101": 5,
	"110": 6,
	"111": 7,
}

func buildJumpTable() {
	add := func(strOpcode string, operation operation, operands ...operand) {
		addOpcode(&jumpTable, strOpcode, operation, operands)
	}

	// NOP
	add("00000000", NOP)

	// LD (N),SP

	// LD R,N
	for ii, x := range R2 {
		add("00"+ii+"0001", LD_u16, x, imm_u16)
	}

	// ADD HL,R

	// LD (R),A
	for i, x := range R1 {
		add("000"+i+"0010", LD_u8, mem{x}, A)
	}

	// LD A,(R)
	for i, x := range R1 {
		add("000"+i+"1010", LD_u8, A, mem{x})
	}

	// INC R
	for ii, x := range R2 {
		add("00"+ii+"0011", INC_u16, x)
	}

	// DEC R

	// INC D
	for iii, x := range D_operands {
		add("00"+iii+"100", INC_u8, x)
	}

	// DEC D
	for iii, x := range D_operands {
		add("00"+iii+"101", DEC_u8, x)
	}

	// LD D,N
	for iii, x := range D_operands {
		add("00"+iii+"110", LD_u8, x, imm_u8)
	}

	// RdCA

	// RdA
	add("00010111", RLA)
	add("00011111", RRA)

	// STOP

	// JR N
	add("00011000", JR1, imm_i8)

	// JR F,N
	for ii, x := range F_operands {
		add("001"+ii+"000", JR2, x, imm_i8)
	}

	// LDI (HL),A
	add("00100010", LDI, mem{HL}, A)

	// LDI A,(HL)
	add("00101010", LDI, A, mem{HL})

	// LDD (HL),A
	add("00110010", LDD, mem{HL}, A)

	// LDD A,(HL)
	add("00111010", LDD, A, mem{HL})

	// DAA
	// CPL
	// SCF
	// CCF

	// LD D,D
	for iii, x := range D_operands {
		for jjj, y := range D_operands {
			if !(iii == "110" && jjj == "110") {
				add("01"+iii+jjj, LD_u8, x, y)
			}
		}
	}

	// HALT

	// ALU A,D
	for iii, operation := range ALU {
		for jjj, x := range D_operands {
			add("10"+iii+jjj, operation, x)
		}
	}

	// ALU A,N
	for iii, operation := range ALU {
		add("11"+iii+"110", operation, imm_u8)
	}

	// POP R
	for ii, x := range R3 {
		add("11"+ii+"0001", POP, x)
	}

	// PUSH R
	for ii, x := range R3 {
		add("11"+ii+"0101", PUSH, x)
	}

	// RST N
	// RET F

	// RET
	add("11001001", RET0)

	// RETI
	// JP F,N
	// JP N
	// CALL F,N

	// CALL N
	add("11001101", CALL1, imm_u16)

	// ADD SP,N
	// LD HL,SP+N

	// LD (0xFF00+N),A
	add("11100000", LD_u8, mem{FF00{imm_u8}}, A)

	// LD A,(0xFF00+N)
	add("11110000", LD_u8, A, mem{FF00{imm_u8}})

	// LD (0xFF00+C),A
	add("11100010", LD_u8, mem{FF00{C}}, A)

	// LD A,(0xFF00+C)
	add("11110010", LD_u8, A, mem{FF00{C}})

	// LD (N),A
	add("11101010", LD_u8, mem{imm_u16}, A)

	// LD A,(N)
	add("11111010", LD_u8, A, mem{imm_u16})

	// ...
}

func buildExtendedJumpTable() {
	add := func(strOpcode string, operation operation, operands ...operand) {
		addOpcode(&extendedJumpTable, strOpcode, operation, operands)
	}

	// RdC

	// Rd D
	for jjj, x := range D_operands {
		add("00010"+jjj, RL, x)
	}
	for jjj, x := range D_operands {
		add("00011"+jjj, RR, x)
	}

	// SdA D
	// SWAP D
	// SRL D

	// BIT N,D
	for iii, x := range N {
		for jjj, y := range D_operands {
			add("01"+iii+jjj, BIT, x, y)
		}
	}

	// RES N,D
	// SET N,D
}

func addOpcode(jumpTable *[256]*instr, strOpcode string, operation operation, operands []operand) {
	assert(len(strOpcode) == 8)
	opcode64, err := strconv.ParseUint(strOpcode, 2 /*base*/, 8 /*bitsize*/)
	check(err)
	assert(0x00 <= opcode64 && opcode64 <= 0xFF)
	opcode := u8(opcode64)

	assert(jumpTable[opcode] == nil)
	jumpTable[opcode] = newInstr(operation, operands)
}
