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
	"fmt"
)

func (st *st) readMem_u8(addr u16) u8 {
	var mask u8 = 0x00
	switch {
	case 0x0000 <= addr && addr <= 0x00FF:
		if st.biosIsEnabled {
			return bios[addr]
		} else {
			return st.rom[addr]
		}
	case 0x0100 <= addr && addr <= 0x7FFF:
		return st.rom[addr]
	case 0x8000 <= addr && addr <= 0x97FF: // Tile sets
	case 0x9800 <= addr && addr <= 0x9FFF: // BG tile maps
	case 0xC000 <= addr && addr <= 0xCFFF: // Work RAM Bank 0
	case 0xD000 <= addr && addr <= 0xDFFF: // Work RAM Bank 1
	case addr == 0xFF01: // SB: Serial transfer data
	case addr == 0xFF0F: // IF: Interrupt Flag
		mask = 0xE0
	case addr == 0xFF40: // LCDC: LCD Control
	case addr == 0xFF42: // SCY: Scroll Y
	case addr == 0xFF43: // SCX: Scroll X
	case addr == 0xFF44: // LY: LCDC Y-Coordinate
		return getScanline(st)
	case addr == 0xFF47: // BGP: BackGround Palette
	case 0xFF80 <= addr && addr <= 0xFFFE: // Zero Page
	case addr == 0xFFFF: // IE: Interrupt Enable
	default:
		panic(fmt.Sprintf("Unimplemented memory read at (0x%04X) and PC=0x%04X.",
			addr, PC.get(st),
		))
	}
	return st.mem[addr] | mask
}

func (st *st) writeMem_u8(addr u16, value u8) {
	switch {
	case 0x8000 <= addr && addr <= 0x97FF: // Tile sets
	case 0x9800 <= addr && addr <= 0x9FFF: // BG tile maps
	case 0xC000 <= addr && addr <= 0xCFFF: // Work RAM Bank 0
	case 0xD000 <= addr && addr <= 0xDFFF: // Work RAM Bank 1
	case addr == 0xFF01: // SB: Serial transfer data
	case addr == 0xFF02 && value == 0x81: // SC: Serial transfer Control
		s := string([]u8{st.readMem_u8(0xFF01)})
		fmt.Print(s)
	case addr == 0xFF05: // TIMA: Timer counter
		// TODO
	case addr == 0xFF07: // TAC: Timer control
		// TODO
	case addr == 0xFF0F: // IF: Interrupt Flag
	case 0xFF11 <= addr && addr <= 0xFF14: // TODO Audio
	case 0xFF24 <= addr && addr <= 0xFF26: // TODO Audio
	case addr == 0xFF40: // LCDC: LCD Control
	case addr == 0xFF42: // SCY: Scroll Y
	case addr == 0xFF43: // SCX: Scroll X
	case addr == 0xFF47: // BGP: BackGround Palette
	case addr == 0xFF50:
		st.biosIsEnabled = false
	case 0xFF80 <= addr && addr <= 0xFFFE: // Zero Page
	case addr == 0xFFFF: // IE: Interrupt Enable
		assert(!getBit(value, 0)) // V-Blank
		assert(!getBit(value, 1)) // LCD STAT
		// Timer
		assert(!getBit(value, 3)) // Serial
		assert(!getBit(value, 4)) // Joypad
	default:
		panic(fmt.Sprintf("Unimplemented memory write 0x%02X=0b%08b at (0x%04X) and PC=0x%04X.",
			value, value, addr, PC.get(st),
		))
	}
	st.mem[addr] = value
}

func (st *st) readMem_u16(addr u16) u16 {
	littleEnd := st.readMem_u8(addr)
	bigEnd := st.readMem_u8(addr + 1)
	return (u16(bigEnd) << 8) | u16(littleEnd)
}

func (st *st) writeMem_u16(addr u16, value u16) {
	littleEnd := u8(value)
	bigEnd := u8(value >> 8)
	st.writeMem_u8(addr, littleEnd)
	st.writeMem_u8(addr+1, bigEnd)
}

// ----------------------------
// Memory locations as operands
// ----------------------------

type mem struct {
	addr r_u16
}

func (mem mem) sizeOf() u16 {
	return mem.addr.sizeOf()
}

func (mem mem) toString(st *st) string {
	return "(" + mem.addr.toString(st) + ")"
}

func (mem mem) get(st *st) u8 {
	addr := mem.addr.get(st)
	return st.readMem_u8(addr)
}

func (mem mem) set(st *st, value u8) {
	addr := mem.addr.get(st)
	st.writeMem_u8(addr, value)
}

type mem_u16 mem

func (m mem_u16) sizeOf() u16 {
	return mem(m).sizeOf()
}

func (m mem_u16) toString(st *st) string {
	return mem(m).toString(st)
}

func (m mem_u16) set(st *st, value u16) {
	addr := m.addr.get(st)
	st.writeMem_u16(addr, value)
}

// ----------------
// Immediate values
// ----------------

type imm_u8_t struct{}

var imm_u8 imm_u8_t

func (imm_u8_t) sizeOf() u16 {
	return 1
}

func (imm_u8_t) toString(st *st) string {
	return fmt.Sprintf("0x%02X", imm_u8.get(st))
}

func (imm_u8_t) get(st *st) u8 {
	// When this is called, PC has already been incremented
	// so we need to read the previous byte.
	return st.readMem_u8(PC.get(st) - imm_u8.sizeOf())
}

type imm_i8_t struct{}

var imm_i8 imm_i8_t

func (imm_i8_t) sizeOf() u16 {
	return 1
}

func (imm_i8_t) toString(st *st) string {
	// Imm_i8 is only used for JR, and JR e is equivalent to JR imm_i8+2.
	// We add 2 to make the logged value consistent with the assembly.
	correctedValue := int(imm_i8.get(st)) + 2
	return fmt.Sprint(correctedValue)
}

func (imm_i8_t) get(st *st) i8 {
	return i8(imm_u8.get(st))
}

type imm_u16_t struct{}

var imm_u16 imm_u16_t

func (imm_u16_t) sizeOf() u16 {
	return 2
}

func (imm_u16_t) toString(st *st) string {
	return fmt.Sprintf("0x%04X", imm_u16.get(st))
}

func (imm_u16_t) get(st *st) u16 {
	// When this is called, PC has already been incremented
	// so we need to read the previous two bytes.
	return st.readMem_u16(PC.get(st) - imm_u16.sizeOf())
}
