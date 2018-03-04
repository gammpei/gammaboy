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

var AF = &reg16{"AF", 0}
var BC = &reg16{"BC", 1}
var DE = &reg16{"DE", 2}
var HL = &reg16{"HL", 3}
var PC = &reg16{"PC", 4}
var SP = &reg16{"SP", 5}

var A = newHiReg8(AF)
var B = newHiReg8(BC)
var D = newHiReg8(DE)
var H = newHiReg8(HL)

var F = &flagReg{
	Z:  &flag{"Z", 7},
	N:  &flag{"N", 6},
	H:  &flag{"H", 5},
	C:  &flag{"C", 4},
	NZ: &negFlag{"NZ", 7},
	NC: &negFlag{"NC", 4},
}
var C = newLoReg8(BC)
var E = newLoReg8(DE)
var L = newLoReg8(HL)

type reg16 struct {
	name string
	i    int
}

func (*reg16) sizeOf() u16 {
	return 0
}

func (r *reg16) toString(*st) string {
	return r.name
}

func (r *reg16) get(st *st) u16 {
	result := st.regs[r.i]
	if r.i == AF.i {
		// The lowest 4 bits of F are always 0.
		result &= 0xFFF0
	}
	return result
}

func (r *reg16) set(st *st, value u16) {
	st.regs[r.i] = value
}

type hiReg8 struct {
	name   string
	parent *reg16
}

func newHiReg8(parent *reg16) *hiReg8 {
	return &hiReg8{parent.name[:1], parent}
}

func (*hiReg8) sizeOf() u16 {
	return 0
}

func (r *hiReg8) toString(*st) string {
	return r.name
}

func (r *hiReg8) get(st *st) u8 {
	return u8(r.parent.get(st) >> 8)
}

func (r *hiReg8) set(st *st, value u8) {
	payload := r.parent.get(st)
	payload &= 0x00FF
	payload |= u16(value) << 8
	r.parent.set(st, payload)
}

type loReg8 struct {
	name   string
	parent *reg16
}

func newLoReg8(parent *reg16) *loReg8 {
	return &loReg8{parent.name[1:], parent}
}

func (*loReg8) sizeOf() u16 {
	return 0
}

func (r *loReg8) toString(*st) string {
	return r.name
}

func (r *loReg8) get(st *st) u8 {
	return u8(r.parent.get(st))
}

func (r *loReg8) set(st *st, value u8) {
	payload := r.parent.get(st)
	payload &= 0xFF00
	payload |= u16(value)
	r.parent.set(st, payload)
}

type flagReg struct {
	Z  *flag
	N  *flag
	H  *flag
	C  *flag
	NZ *negFlag
	NC *negFlag
}

func (*flagReg) sizeOf() u16 {
	return 0
}

func (*flagReg) toString(*st) string {
	return "F"
}

func (*flagReg) get(st *st) u8 {
	return u8(AF.get(st))
}

func (*flagReg) set(st *st, value u8) {
	payload := AF.get(st)
	payload &= 0xFF00
	payload |= u16(value)
	AF.set(st, payload)
}

type flag struct {
	name string
	bit  uint
}

func (*flag) sizeOf() u16 {
	return 0
}

func (flag *flag) toString(*st) string {
	return flag.name
}

func (flag *flag) get(st *st) bool {
	return getBit(F.get(st), flag.bit)
}

func (flag *flag) set(st *st, value bool) {
	payload := setBit(F.get(st), flag.bit, value)
	F.set(st, payload)
}

// Negative flag
type negFlag struct {
	name string
	bit  uint
}

func (*negFlag) sizeOf() u16 {
	return 0
}

func (flag *negFlag) toString(*st) string {
	return flag.name
}

func (flag *negFlag) get(st *st) bool {
	return !getBit(F.get(st), flag.bit)
}
