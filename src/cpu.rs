use state::*;
use state::Reg16::*;
use state::Reg8::*;

pub fn fetch_decode_execute(st: &mut State) {
	// Fetch
	let PC_0: u16 = PC.get(st);
	let opcode: u8 = read_mem(st, PC_0);

	// Decode
	let instr: Box<IInstr> = match opcode {
		0x00 => Instr::new("NOP", NOP, ()),
		0x21 => Instr::new("LD", LD, (HL, Imm_u16)),
		0x31 => Instr::new("LD", LD, (SP, Imm_u16)),
		0xAF => Instr::new("XOR", XOR, A),
		_ => panic!("Unknown opcode 0x{0:02X}=0b{0:08b} at PC=0x{1:04X}.", opcode, PC_0),
	};

	// Increment PC
	let instr_length: i32 = 1 + instr.operands_length(); // bytes
	PC.set(st, PC_0 + (instr_length as u16));

	// Execute
	let bytes: String = match instr_length {
		1 => format!("0x{:02X}          ", opcode),
		2 => format!("0x{:02X} 0x{:02X}     ", opcode, read_mem(st, PC_0 + 1)),
		3 => format!("0x{:02X} 0x{:02X} 0x{:02X}", opcode, read_mem(st, PC_0 + 1), read_mem(st, PC_0 + 2)),
		_ => unreachable!(),
	};
	println!(" {} | {}", bytes, instr.log(st));
	instr.execute(st);
}

struct Instr<T> {
	name: &'static str,
	/// Operation
	f: fn(&mut State, T),
	/// Operand(s)
	x: T,
}

impl<T> Instr<T> where T: Copy {
	fn new(name: &'static str, f: fn(&mut State, T), x: T) -> Box<Instr<T>> {
		Box::new(Instr { name: name, f: f, x: x })
	}

	fn execute(&self, st: &mut State) { (self.f)(st, self.x); }
}

/// Interface that all instructions must implement.
trait IInstr {
	/// Length of all the operands in bytes, for the purpose of incrementing PC.
	fn operands_length(&self) -> i32;
	fn log(&self, st: &State) -> String;
	fn execute(&self, st: &mut State);
}

/// 0-operand instruction.
impl IInstr for Instr<()> {
	fn operands_length(&self) -> i32 { 0 }
	fn log(&self, _: &State) -> String { String::from(self.name) }
	fn execute(&self, st: &mut State) { self.execute(st); }
}

/// 1-operand instruction.
impl<T> IInstr for Instr<T> where T: Operand {
	fn operands_length(&self) -> i32 { self.x.length() }
	fn log(&self, st: &State) -> String { format!("{} {}", self.name, self.x.log(st)) }
	fn execute(&self, st: &mut State) { self.execute(st); }
}

/// 2-operand instruction.
impl<T, U> IInstr for Instr<(T, U)> where T: Operand, U: Operand {
	fn operands_length(&self) -> i32 { self.x.0.length() + self.x.1.length() }
	fn log(&self, st: &State) -> String { format!("{} {}, {}", self.name, self.x.0.log(st), self.x.1.log(st)) }
	fn execute(&self, st: &mut State) { self.execute(st); }
}

trait Operand: Copy {
	/// Length of the operand in bytes, for the purpose of incrementing PC.
	fn length(self) -> i32;
	fn log(self, st: &State) -> String;
}
trait R<T>: Operand { fn get(self, st: &State) -> T; }
trait W<T>: Operand { fn set(self, st: &mut State, new_value: T); }
trait RW<T>: R<T> + W<T> { }

// ---------------------
// Registers as operands
// ---------------------

impl Operand for Reg16 {
	fn length(self) -> i32 { 0 }
	fn log(self, _: &State) -> String { String::from(self.name()) }
}

impl Operand for Reg8 {
	fn length(self) -> i32 { 0 }
	fn log(self, _: &State) -> String { String::from(self.name()) }
}

impl Operand for Flag {
	fn length(self) -> i32 { 0 }
	fn log(self, _: &State) -> String { String::from(self.name()) }
}

impl R<u16> for Reg16 { fn get(self, st: &State) -> u16 { self.get(st) } }
impl W<u16> for Reg16 { fn set(self, st: &mut State, new_value: u16) { self.set(st, new_value) } }
impl RW<u16> for Reg16 { }

impl R<u8> for Reg8 { fn get(self, st: &State) -> u8 { self.get(st) } }
impl W<u8> for Reg8 { fn set(self, st: &mut State, new_value: u8) { self.set(st, new_value) } }
impl RW<u8> for Reg8 { }

impl R<bool> for Flag { fn get(self, st: &State) -> bool { self.get(st) } }
impl W<bool> for Flag { fn set(self, st: &mut State, new_value: bool) { self.set(st, new_value) } }
impl RW<bool> for Flag { }

// ----------------
// Immediate values
// ----------------

#[derive(Clone, Copy)]
struct Imm_u16;

impl Operand for Imm_u16 {
	fn length(self) -> i32 { 2 }
	fn log(self, st: &State) -> String { format!("0x{:04X}", self.get(st)) }
}

impl R<u16> for Imm_u16 {
	fn get(self, st: &State) -> u16 {
		// When this is called, PC has already been incremented
		// so we need to read the previous two bytes.
		let PC_v: u16 = PC.get(st);
		let little_end: u8 = read_mem(st, PC_v - 2);
		let big_end: u8 = read_mem(st, PC_v - 1);
		((big_end as u16) << 8) | (little_end as u16)
	}
}

// ------------------------------------
// CPU operations in alphabetical order
// ------------------------------------

/// UM0080.pdf rev 11 p113 / 332
fn LD<T, U, V>(st: &mut State, (x, y): (T, U)) where T: W<V>, U: R<V> {
	let new_value: V = y.get(st);
	x.set(st, new_value);
}

/// UM0080.pdf rev 11 p194 / 332
fn NOP(_: &mut State, _: ()) { }

/// UM0080.pdf rev 11 p175 / 332
fn XOR<T>(st: &mut State, x: T) where T: R<u8> {
	let new_value: u8 = A.get(st) ^ x.get(st);
	A.set(st, new_value);

	let result_is_zero: bool = A.get(st) == 0x00;
	Flag::Z.set(st, result_is_zero);
	Flag::N.set(st, false);
	Flag::H.set(st, false);
	Flag::C.set(st, false);
}