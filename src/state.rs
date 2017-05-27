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
			Flag::Z => "F.Z",
			Flag::N => "F.N",
			Flag::H => "F.H",
			Flag::C => "F.C",
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

pub fn log_registers(st: &State) -> String {
	format!("PC=0x{:04X} AF=0x{:04X} BC=0x{:04X} DE=0x{:04X} HL=0x{:04X} SP=0x{:04X}",
		PC.get(st), AF.get(st), BC.get(st), DE.get(st), HL.get(st), SP.get(st)
	)
}

pub fn read_mem(st: &State, addr: u16) -> u8 {
	if addr <= 0x00FF { st.bios[addr as usize] }
	else { unimplemented!(); }
}