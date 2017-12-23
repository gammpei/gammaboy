package main

import (
	"fmt"
)

func (st *st) readMem_u8(addr u16) u8 {
	if 0x0000 <= addr && addr <= 0x00FF {
		return bios[addr]
	} else if 0x0104 <= addr && addr <= 0x0133 {
		// The bios already contains the Nintendo logo data at 0x00A8,
		// so we cheat and we make 0x0104...0x0133 return that data.
		// We do this so that the bios doesn't lock up.
		return bios[0x00A8+addr-0x0104]
	} else if 0x0134 <= addr && addr <= 0x014D {
		// Dummy values to make the bios checksum work.
		// If we don't, the bios locks up.
		if addr == 0x0134 {
			return 0xE7
		} else {
			return 0x00
		}
	} else if addr == 0xFF42 {
		// SCY: Scroll Y
		return st.mem[addr]
	} else if addr == 0xFF44 {
		// LY: LCDC Y-Coordinate
		// We return 144 because the bios will wait forever for that value.
		return 144
	} else if 0xFF80 <= addr && addr <= 0xFFFE {
		// Zero Page
		return st.mem[addr]
	} else {
		panic(fmt.Sprintf("Unimplemented memory read at 0x%04X.", addr))
	}
}

func (st *st) writeMem_u8(addr u16, value u8) {
	if 0x8000 <= addr && addr <= 0x97FF {
		// Character RAM
	} else if 0x9800 <= addr && addr <= 0x9BFF {
		// BG Map Data 1
	} else if 0x9C00 <= addr && addr <= 0x9FFF {
		// BG Map Data 2
	} else if 0xFF11 <= addr && addr <= 0xFF14 {
		// TODO Audio
	} else if 0xFF24 <= addr && addr <= 0xFF26 {
		// TODO Audio
	} else if addr == 0xFF40 {
		// LCDC: LCD Control
	} else if addr == 0xFF42 {
		// SCY: Scroll Y
	} else if addr == 0xFF47 {
		// BGP: BackGround Palette
	} else if 0xFF80 <= addr && addr <= 0xFFFE {
		// Zero Page
	} else {
		panic(fmt.Sprintf("Unimplemented memory write at 0x%04X.", addr))
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
