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

// --------------------------
// Carry and borrow functions
// --------------------------

func carry_u8(xs ...u8) bool {
	assert(len(xs) == 2 || len(xs) == 3)
	result := 0
	for _, x := range xs {
		result += int(x)
	}
	return result > 0xFF
}

func halfCarry_u8(xs ...u8) bool {
	assert(len(xs) == 2 || len(xs) == 3)
	var result uint = 0
	for _, x := range xs {
		result += uint(x & 0x0F)
	}
	return result > 0x0F
}

func carry_u16(x, y u16) bool {
	return int(x)+int(y) > 0xFFFF
}

func halfCarry_u16(x, y u16) bool {
	return (x&0x0FFF)+(y&0x0FFF) > 0x0FFF
}

func borrow(x, y u8) bool {
	return x < y
}

func halfBorrow(x, y u8) bool {
	return (x & 0x0F) < (y & 0x0F)
}

// ------------------------------------
// CPU operations in alphabetical order
// ------------------------------------

// UM0080.pdf rev 11 p165 / 332
var ADC_A = operation{"ADC", func(st *st, x r_u8) {
	v1 := A.get(st)
	v2 := x.get(st)
	v3 := u8FromBool(F.C.get(st))
	result := v1 + v2 + v3
	A.set(st, result)

	F.Z.set(st, result == 0x00)
	F.N.set(st, false)
	F.H.set(st, halfCarry_u8(v1, v2, v3))
	F.C.set(st, carry_u8(v1, v2, v3))
}}

// UM0080.pdf rev 11 p159,161,162 / 332
var ADD_A = operation{"ADD", func(st *st, x r_u8) {
	v1 := A.get(st)
	v2 := x.get(st)
	result := v1 + v2
	A.set(st, result)

	F.Z.set(st, result == 0x00)
	F.N.set(st, false)
	F.H.set(st, halfCarry_u8(v1, v2))
	F.C.set(st, carry_u8(v1, v2))
}}

// UM0080.pdf rev 11 p202 / 332
var ADD_HL = operation{"ADD", func(st *st, x r_u16) {
	v1 := HL.get(st)
	v2 := x.get(st)
	HL.set(st, v1+v2)

	F.N.set(st, false)
	F.H.set(st, halfCarry_u16(v1, v2))
	F.C.set(st, carry_u16(v1, v2))
}}

// UM0080.pdf rev 11 p171 / 332
var AND = operation{"AND", func(st *st, x r_u8) {
	result := A.get(st) & x.get(st)
	A.set(st, result)

	F.Z.set(st, result == 0x00)
	F.N.set(st, false)
	F.H.set(st, true)
	F.C.set(st, false)
}}

// UM0080.pdf rev 11 p257,259 / 332
var BIT = operation{"BIT", func(st *st, x u3, y r_u8) {
	F.Z.set(st, getBit(y.get(st), uint(x)) == false)
	F.N.set(st, false)
	F.H.set(st, true)
}}

// UM0080.pdf rev 11 p177 / 332
var CP = operation{"CP", func(st *st, x r_u8) {
	v1 := A.get(st)
	v2 := x.get(st)

	F.Z.set(st, v1 == v2)
	F.N.set(st, true)
	F.H.set(st, halfBorrow(v1, v2))
	F.C.set(st, borrow(v1, v2))
}}

// UM0080.pdf rev 11 p295 / 332
var CALL1 = operation{"CALL", func(st *st, x r_u16) {
	PUSH.f.(func(*state, r_u16))(st, PC)
	PC.set(st, x.get(st))
}}

// UM0080.pdf rev 11 p297 / 332
var CALL2 = operation{"CALL", func(st *st, x r_bool, y r_u16) {
	if x.get(st) {
		CALL1.f.(func(*state, r_u16))(st, y)
	}
}}

// UM0080.pdf rev 11 p184 / 332
var DEC_u8 = operation{"DEC", func(st *st, x rw_u8) {
	oldValue := x.get(st)
	newValue := oldValue - 1
	x.set(st, newValue)

	F.Z.set(st, newValue == 0x00)
	F.N.set(st, true)
	F.H.set(st, halfBorrow(oldValue, 1))
}}

// UM0080.pdf rev 11 p196 / 332
var DI = operation{"DI", func(st *st) {
	st.IME = false
}}

// UM0080.pdf rev 11 p179,181 / 332
var INC_u8 = operation{"INC", func(st *st, x rw_u8) {
	v1 := x.get(st)
	var v2 u8 = 1
	result := v1 + v2
	x.set(st, result)

	F.Z.set(st, result == 0x00)
	F.N.set(st, false)
	F.H.set(st, halfCarry_u8(v1, v2))
}}

// UM0080.pdf rev 11 p212 / 332
var INC_u16 = operation{"INC", func(st *st, x rw_u16) {
	x.set(st, x.get(st)+1)
}}

// UM0080.pdf rev 11 p276 / 332
var JP1 = operation{"JP", func(st *st, x r_u16) {
	PC.set(st, x.get(st))
}}

// UM0080.pdf rev 11 p279 / 332
var JR1 = operation{"JR", func(st *st, x r_i8) {
	jumpLocation := int(PC.get(st)) + int(x.get(st))
	assert(0x0000 <= jumpLocation && jumpLocation <= 0xFFFF)
	PC.set(st, u16(jumpLocation))
}}

// UM0080.pdf rev 11 p281,283,285,287 / 332
var JR2 = operation{"JR", func(st *st, x r_bool, y r_i8) {
	if x.get(st) {
		JR1.f.(func(*state, r_i8))(st, y)
	}
}}

// UM0080.pdf rev 11 p85,86,88,93,99,102,103,105,106,113 / 332
var LD_u8 = operation{"LD", func(st *st, x w_u8, y r_u8) {
	x.set(st, y.get(st))
}}
var LD_u16 = operation{"LD", func(st *st, x w_u16, y r_u16) {
	x.set(st, y.get(st))
}}

// pandocs.htm
// ldd  (HL),A      32         8 ---- (HL)=A, HL=HL-1
// ldd  A,(HL)      3A         8 ---- A=(HL), HL=HL-1
var LDD = operation{"LDD", func(st *st, x w_u8, y r_u8) {
	LD_u8.f.(func(*state, w_u8, r_u8))(st, x, y)
	HL.set(st, HL.get(st)-1)
}}

// pandocs.htm
// ldi  (HL),A      22         8 ---- (HL)=A, HL=HL+1
// ldi  A,(HL)      2A         8 ---- A=(HL), HL=HL+1
var LDI = operation{"LDI", func(st *st, x w_u8, y r_u8) {
	LD_u8.f.(func(*state, w_u8, r_u8))(st, x, y)
	HL.set(st, HL.get(st)+1)
}}

// UM0080.pdf rev 11 p194 / 332
var NOP = operation{"NOP", func(*st) {}}

// UM0080.pdf rev 11 p173 / 332
var OR = operation{"OR", func(st *st, x r_u8) {
	result := A.get(st) | x.get(st)
	A.set(st, result)

	F.Z.set(st, result == 0x00)
	F.N.set(st, false)
	F.H.set(st, false)
	F.C.set(st, false)
}}

// UM0080.pdf rev 11 p133 / 332
var POP = operation{"POP", func(st *st, x w_u16) {
	top := SP.get(st)
	x.set(st, st.readMem_u16(top))
	SP.set(st, top+2)
}}

// UM0080.pdf rev 11 p129 / 332
var PUSH = operation{"PUSH", func(st *st, x r_u16) {
	top := SP.get(st) - 2
	SP.set(st, top)
	st.writeMem_u16(top, x.get(st))
}}

// UM0080.pdf rev 11 p299 / 332
var RET0 = operation{"RET", func(st *st) {
	POP.f.(func(*state, w_u16))(st, PC)
}}

// UM0080.pdf rev 11 p300 / 332
var RET1 = operation{"RET", func(st *st, x r_bool) {
	if x.get(st) {
		RET0.f.(func(*state))(st)
	}
}}

// UM0080.pdf rev 11 p235 / 332
var RL = operation{"RL", func(st *st, x rw_u8) {
	oldValue := x.get(st)
	oldBit7 := getBit(oldValue, 7)
	oldCarryFlag := F.C.get(st)

	result := (oldValue << 1) | u8FromBool(oldCarryFlag)
	x.set(st, result)

	F.Z.set(st, result == 0x00)
	F.N.set(st, false)
	F.H.set(st, false)
	F.C.set(st, oldBit7)
}}

// UM0080.pdf rev 11 p221 / 332
var RLA = operation{"RLA", func(st *st) {
	oldZeroFlag := F.Z.get(st)
	RL.f.(func(*state, rw_u8))(st, A)
	F.Z.set(st, oldZeroFlag)
}}

// UM0080.pdf rev 11 p241 / 332
var RR = operation{"RR", func(st *st, x rw_u8) {
	oldValue := x.get(st)
	oldBit0 := getBit(oldValue, 0)
	oldCarryFlag := F.C.get(st)

	result := (u8FromBool(oldCarryFlag) << 7) | (oldValue >> 1)
	x.set(st, result)

	F.Z.set(st, result == 0x00)
	F.N.set(st, false)
	F.H.set(st, false)
	F.C.set(st, oldBit0)
}}

// UM0080.pdf rev 11 p225 / 332
var RRA = operation{"RRA", func(st *st) {
	oldZeroFlag := F.Z.get(st)
	RR.f.(func(*state, rw_u8))(st, A)
	F.Z.set(st, oldZeroFlag)
}}

// UM0080.pdf rev 11 p250 / 332
var SRL = operation{"SRL", func(st *st, x rw_u8) {
	oldValue := x.get(st)
	oldBit0 := getBit(oldValue, 0)

	result := oldValue >> 1
	x.set(st, result)

	F.Z.set(st, result == 0x00)
	F.N.set(st, false)
	F.H.set(st, false)
	F.C.set(st, oldBit0)
}}

// UM0080.pdf rev 11 p167 / 332
var SUB = operation{"SUB", func(st *st, x r_u8) {
	CP.f.(func(*state, r_u8))(st, x)
	A.set(st, A.get(st)-x.get(st))
}}

// UM0080.pdf rev 11 p175 / 332
var XOR = operation{"XOR", func(st *st, x r_u8) {
	result := A.get(st) ^ x.get(st)
	A.set(st, result)

	F.Z.set(st, result == 0x00)
	F.N.set(st, false)
	F.H.set(st, false)
	F.C.set(st, false)
}}
