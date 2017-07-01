use state::Reg16::*;
use state::Reg8::*;

pub struct State {
	AF: u16,
	BC: u16,
	DE: u16,
	HL: u16,
	PC: u16,
	SP: u16,
	bios: [u8; 256],
	mem: [u8; 0xFFFF+1],
}

impl State {
	pub fn new(bios: [u8; 256]) -> State {
		State {
			AF: 0x00,
			BC: 0x00,
			DE: 0x00,
			HL: 0x00,
			PC: 0x00,
			SP: 0x00,
			bios: bios,
			mem: [0x00; 0xFFFF+1],
		}
	}
}

#[allow(dead_code)]
#[derive(Clone, Copy)]
pub enum Reg16 { AF, BC, DE, HL, PC, SP }

#[allow(dead_code)]
#[derive(Clone, Copy)]
pub enum Reg8 { A, F, B, C, D, E, H, L }

#[allow(dead_code)]
#[derive(Clone, Copy)]
pub enum Flag { Z, N, H, C }

impl Reg16 {
	pub fn name(self) -> &'static str {
		match self {
			AF => "AF",
			BC => "BC",
			DE => "DE",
			HL => "HL",
			PC => "PC",
			SP => "SP",
		}
	}

	pub fn get(self, st: &State) -> u16 {
		match self {
			AF => st.AF,
			BC => st.BC,
			DE => st.DE,
			HL => st.HL,
			PC => st.PC,
			SP => st.SP,
		}
	}

	pub fn set(self, st: &mut State, new_value: u16) {
		let x: &mut u16 = match self {
			AF => &mut st.AF,
			BC => &mut st.BC,
			DE => &mut st.DE,
			HL => &mut st.HL,
			PC => &mut st.PC,
			SP => &mut st.SP,
		};
		*x = new_value;
	}
}

impl Reg8 {
	pub fn name(self) -> &'static str {
		match self {
			A => "A", F => "F",
			B => "B", C => "C",
			D => "D", E => "E",
			H => "H", L => "L",
		}
	}

	pub fn get(self, st: &State) -> u8 {
		fn MSB(x: u16) -> u8 { (x >> 8) as u8 }
		fn LSB(x: u16) -> u8 { x as u8 }
		match self {
			A => MSB(st.AF), F => LSB(st.AF),
			B => MSB(st.BC), C => LSB(st.BC),
			D => MSB(st.DE), E => LSB(st.DE),
			H => MSB(st.HL), L => LSB(st.HL),
		}
	}

	pub fn set(self, st: &mut State, new_value: u8) {
		let MSB = |x: &mut u16| { *x = (*x & 0x00FF) | ((new_value as u16) << 8) };
		let LSB = |x: &mut u16| { *x = (*x & 0xFF00) | (new_value as u16) };
		match self {
			A => MSB(&mut st.AF), F => LSB(&mut st.AF),
			B => MSB(&mut st.BC), C => LSB(&mut st.BC),
			D => MSB(&mut st.DE), E => LSB(&mut st.DE),
			H => MSB(&mut st.HL), L => LSB(&mut st.HL),
		}
	}
}

impl Flag {
	fn bit(self) -> usize {
		match self {
			Flag::Z => 7,
			Flag::N => 6,
			Flag::H => 5,
			Flag::C => 4,
		}
	}

	pub fn name(self) -> &'static str {
		match self {
			Flag::Z => "Z",
			Flag::N => "N",
			Flag::H => "H",
			Flag::C => "C",
		}
	}

	pub fn get(self, st: &State) -> bool {
		(st.AF >> self.bit()) & 1 != 0
	}

	pub fn set(self, st: &mut State, new_value: bool) {
		match new_value {
			true => st.AF |= 1 << self.bit(),
			false => st.AF &= !(1 << self.bit()),
		};
	}
}

pub fn read_mem(st: &State, addr: u16) -> u8 {
	match addr {
		0x0000...0x00FF => st.bios[addr as usize],
		
		// The bios already contains the Nintendo logo data at 0x00A8,
		// so we cheat and we make 0x0104...0x0133 return that data.
		// We do this so that the bios doesn't lock up.
		0x0104...0x0133 => st.bios[(0x00A8 + addr - 0x0104) as usize],

		// Dummy values to make the bios checksum work.
		// If we don't, the bios locks up.
		0x0134...0x014D => if addr == 0x0134 { 0xE7 } else { 0x00 },

		0xFF42 => st.mem[addr as usize], // SCY: Scroll Y

		// We return 144 because otherwise the bios will wait forever
		// on that value to come.
		0xFF44 => 144, // LY: LCDC Y-Coordinate

		0xFF80...0xFFFE => st.mem[addr as usize], // Zero Page

		_ => panic!("Unimplemented memory read at 0x{:04X}.", addr),
	}
}

pub fn set_mem(st: &mut State, addr: u16, new_value: u8) {
	let set = |st: &mut State| { st.mem[addr as usize] = new_value; };
	match addr {
		0x8000...0x97FF => set(st), // Character RAM
		0x9800...0x9BFF => set(st), // BG Map Data 1
		0x9C00...0x9FFF => set(st), // BG Map Data 2
		0xFF11 => (), // TODO Audio
		0xFF12 => (), // TODO Audio
		0xFF13 => (), // TODO Audio
		0xFF14 => (), // TODO Audio
		0xFF24 => (), // TODO Audio
		0xFF25 => (), // TODO Audio
		0xFF26 => (), // TODO Audio
		0xFF40 => set(st), // LCDC: LCD Control
		0xFF42 => set(st), // SCY: Scroll Y
		0xFF47 => set(st), // BGP: BackGround Palette
		0xFF80...0xFFFE => set(st), // Zero Page
		_ => panic!("Unimplemented memory write at 0x{:04X}.", addr),
	}
}