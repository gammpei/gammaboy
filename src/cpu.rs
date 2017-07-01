use std;
use std::borrow::Borrow;

use logging;
use state::*;
use state::Reg16::*;
use state::Reg8::*;

pub fn fetch_decode_execute(cpu: &CPU, st: &mut State) {
	// Fetch
	let PC_0: u16 = PC.get(st);
	let opcode: u8 = read_mem(st, PC_0);
	let mut instr_length: i32 = 1; // bytes

	// Decode
	let instr: &IInstr =
		if opcode == 0xCB {
			let extended_opcode: u8 = read_mem(st, PC_0 + 1);
			instr_length += 1;
			match cpu.extended_jump_table[extended_opcode as usize] {
				Some(ref instr) => instr.borrow(),
				None => panic!("Unknown extended opcode 0xCB-0x{0:02X}=0b{0:08b} at PC=0x{1:04X}.", extended_opcode, PC_0),
			}
		} else {
			match cpu.jump_table[opcode as usize] {
				Some(ref instr) => instr.borrow(),
				None => panic!("Unknown opcode 0x{0:02X}=0b{0:08b} at PC=0x{1:04X}.", opcode, PC_0),
			}
		};
	instr_length += instr.operands_length();

	// Increment PC, log, and execute, in that order.
	PC.set(st, PC_0 + (instr_length as u16));
	logging::log_instr(st, instr, instr_length);
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
		Box::new(Instr { name, f, x })
	}

	fn execute(&self, st: &mut State) { (self.f)(st, self.x); }
}

/// Interface that all instructions must implement.
pub trait IInstr {
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

/// Negative flag
#[derive(Clone, Copy)]
struct N(Flag);

impl Operand for N {
	fn length(self) -> i32 { 0 }
	fn log(self, _: &State) -> String { format!("N{}", self.0.name()) }
}

impl R<bool> for N { fn get(self, st: &State) -> bool { !self.0.get(st) } }

// ----------------------------
// Memory locations as operands
// ----------------------------

#[derive(Clone, Copy)]
struct Mem<T>(T);

impl<T> Operand for Mem<T> where T: R<u16> {
	fn length(self) -> i32 { self.0.length() }
	fn log(self, st: &State) -> String { format!("({})", self.0.log(st)) }
}

impl<T> R<u8> for Mem<T> where T: R<u16> {
	fn get(self, st: &State) -> u8 {
		let addr: u16 = self.0.get(st);
		read_mem(st, addr)
	}
}

impl<T> W<u8> for Mem<T> where T: R<u16> {
	fn set(self, st: &mut State, new_value: u8) {
		let addr: u16 = self.0.get(st);
		set_mem(st, addr, new_value)
	}
}

impl<T> RW<u8> for Mem<T> where T: R<u16> { }

// ----------------
// Immediate values
// ----------------

#[derive(Clone, Copy)]
struct Imm_u8;

impl Operand for Imm_u8 {
	fn length(self) -> i32 { 1 }
	fn log(self, st: &State) -> String { format!("0x{:02X}", self.get(st)) }
}

impl R<u8> for Imm_u8 {
	fn get(self, st: &State) -> u8 {
		// When this is called, PC has already been incremented
		// so we need to read the previous byte.
		read_mem(st, PC.get(st) - 1)
	}
}

#[derive(Clone, Copy)]
struct Imm_i8;

impl Operand for Imm_i8 {
	fn length(self) -> i32 { 1 }
	fn log(self, st: &State) -> String {
		let e: i32 = self.get(st) as i32 + 2;
		format!("{}", e)
		// This looks crazy, right? Why the random "+ 2"?
		// Imm_i8 is only used for JR, and JR e is equivalent to JR Imm_i8+2.
	}
}

impl R<i8> for Imm_i8 {
	fn get(self, st: &State) -> i8 {
		// When this is called, PC has already been incremented
		// so we need to read the previous byte.
		read_mem(st, PC.get(st) - 1) as i8
	}
}

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

// ---------------------
// Constants as operands
// ---------------------

impl Operand for i32 {
	fn length(self) -> i32 { 0 }
	fn log(self, _: &State) -> String { format!("{}", self) }
}

/// 0xFF00 base address + u8 offset.
#[derive(Clone, Copy)]
struct FF00<T>(T);

impl<T> Operand for FF00<T> where T: R<u8> {
	fn length(self) -> i32 { self.0.length() }
	fn log(self, st: &State) -> String { format!("0xFF00+0x{:02X}", self.0.get(st)) }
}

impl<T> R<u16> for FF00<T> where T: R<u8> {
	fn get(self, st: &State) -> u16 {
		0xFF00 + (self.0.get(st) as u16)
	}
}

// --------------------------
// Carry and borrow functions
// --------------------------

fn carry(x: u8, y: u8) -> bool {
	(x as i32) + (y as i32) > 0xFF
}

fn half_carry(x: u8, y: u8) -> bool {
	(x & 0x0F) + (y & 0x0F) > 0x0F
}

fn borrow(x: u8, y: u8) -> bool {
	x < y
}

fn half_borrow(x: u8, y: u8) -> bool {
	(x & 0x0F) < (y & 0x0F)
}

// ------------------------------------
// CPU operations in alphabetical order
// ------------------------------------

// UM0080.pdf rev 11 p159,161,162 / 332
const ADD1_NAME: &'static str = "ADD";
fn ADD1<T>(st: &mut State, x: T) where T: R<u8> {
	let v1: u8 = A.get(st);
	let v2: u8 = x.get(st);
	let new_A: u8 = v1.wrapping_add(v2);
	A.set(st, new_A);

	Flag::Z.set(st, new_A == 0x00);
	Flag::N.set(st, false);
	Flag::H.set(st, half_carry(v1, v2));
	Flag::C.set(st, carry(v1, v2));
}

// UM0080.pdf rev 11 p257,259 / 332
const BIT_NAME: &'static str = "BIT";
fn BIT<T>(st: &mut State, (x, y): (i32, T)) where T: R<u8> {
	assert!(0 <= x && x <= 7);

	let bit_is_zero: bool = (y.get(st) >> x) & 0x01 == 0x00;
	Flag::Z.set(st, bit_is_zero);
	Flag::N.set(st, false);
	Flag::H.set(st, true);
}

// UM0080.pdf rev 11 p177 / 332
const CP_NAME: &'static str = "CP";
fn CP<T>(st: &mut State, x: T) where T: R<u8> {
	let v1: u8 = A.get(st);
	let v2: u8 = x.get(st);

	Flag::Z.set(st, v1 == v2);
	Flag::N.set(st, true);
	Flag::H.set(st, half_borrow(v1, v2));
	Flag::C.set(st, borrow(v1, v2));
}

// UM0080.pdf rev 11 p295 / 332
const CALL1_NAME: &'static str = "CALL";
fn CALL1<T>(st: &mut State, x: T) where T: R<u16> {
	PUSH(st, PC);

	let jump_addr: u16 = x.get(st);
	PC.set(st, jump_addr);
}

/// UM0080.pdf rev 11 p184 / 332
const DEC_U8_NAME: &'static str = "DEC";
fn DEC_u8<T>(st: &mut State, x: T) where T: RW<u8> {
	let old_value: u8 = x.get(st);
	let new_value: u8 = old_value.wrapping_sub(1);
	x.set(st, new_value);

	Flag::Z.set(st, new_value == 0x00);
	Flag::N.set(st, true);
	Flag::H.set(st, half_borrow(old_value, 1));
}

/// UM0080.pdf rev 11 p179,181 / 332
const INC_U8_NAME: &'static str = "INC";
fn INC_u8<T>(st: &mut State, x: T) where T: RW<u8> {
	let old_value: u8 = x.get(st);
	let new_value: u8 = old_value + 1;
	x.set(st, new_value);

	Flag::Z.set(st, new_value == 0x00);
	Flag::N.set(st, false);
	Flag::H.set(st, half_carry(old_value, 1));
}

/// UM0080.pdf rev 11 p212 / 332
const INC_U16_NAME: &'static str = "INC";
fn INC_u16<T>(st: &mut State, x: T) where T: RW<u16> {
	let old_value: u16 = x.get(st);
	x.set(st, old_value + 1);
}

/// UM0080.pdf rev 11 p279 / 332
const JR1_NAME: &'static str = "JR";
fn JR1<T>(st: &mut State, x: T) where T: R<i8> {
	let new_PC: u16 = ((PC.get(st) as i32) + (x.get(st) as i32)) as u16;
	PC.set(st, new_PC);
}

/// UM0080.pdf rev 11 p281,283,285,287 / 332
const JR2_NAME: &'static str = "JR";
fn JR2<T, U>(st: &mut State, (x, y): (T, U)) where T: R<bool>, U: R<i8> {
	if x.get(st) {
		JR1(st, y);
	}
}

/// UM0080.pdf rev 11 p85,86,88,93,99,102,103,105,106,113 / 332
const LD_NAME: &'static str = "LD";
fn LD<T, U, V>(st: &mut State, (x, y): (T, U)) where T: W<V>, U: R<V> {
	let new_value: V = y.get(st);
	x.set(st, new_value);
}

/// pandocs.htm
/// ldd  (HL),A      32         8 ---- (HL)=A, HL=HL-1
/// ldd  A,(HL)      3A         8 ---- A=(HL), HL=HL-1
const LDD_NAME: &'static str = "LDD";
fn LDD<T, U>(st: &mut State, (x, y): (T, U)) where T: W<u8>, U: R<u8> {
	LD(st, (x, y));

	let new_HL: u16 = HL.get(st) - 1;
	HL.set(st, new_HL);
}

/// pandocs.htm
/// ldi  (HL),A      22         8 ---- (HL)=A, HL=HL+1
/// ldi  A,(HL)      2A         8 ---- A=(HL), HL=HL+1
const LDI_NAME: &'static str = "LDI";
fn LDI<T, U>(st: &mut State, (x, y): (T, U)) where T: W<u8>, U: R<u8> {
	LD(st, (x, y));

	let new_HL: u16 = HL.get(st) + 1;
	HL.set(st, new_HL);
}

/// UM0080.pdf rev 11 p194 / 332
const NOP_NAME: &'static str = "NOP";
fn NOP(_: &mut State, _: ()) { }

/// UM0080.pdf rev 11 p133 / 332
const POP_NAME: &'static str = "POP";
fn POP<T>(st: &mut State, x: T) where T: W<u16> {
	let SP_value: u16 = SP.get(st);
	let little_end: u8 = read_mem(st, SP_value);
	let big_end: u8 = read_mem(st, SP_value + 1);
	SP.set(st, SP_value + 2);

	x.set(st, ((big_end as u16) << 8) | (little_end as u16));
}

/// UM0080.pdf rev 11 p129 / 332
const PUSH_NAME: &'static str = "PUSH";
fn PUSH<T>(st: &mut State, x: T) where T: R<u16> {
	let value: u16 = x.get(st);
	let little_end: u8 = value as u8;
	let big_end: u8 = (value >> 8) as u8;

	let SP_value: u16 = SP.get(st);
	set_mem(st, SP_value - 1, big_end);
	set_mem(st, SP_value - 2, little_end);
	SP.set(st, SP_value - 2);
}

/// UM0080.pdf rev 11 p299 / 332
const RET0_NAME: &'static str = "RET";
fn RET0(st: &mut State, _: ()) {
	POP(st, PC);
}

/// UM0080.pdf rev 11 p235 / 332
const RL_NAME: &'static str = "RL";
fn RL<T>(st: &mut State, x: T) where T: RW<u8> {
	let old_value: u8 = x.get(st);
	let old_bit_7: bool = (old_value >> 7) & 1 != 0;
	let old_carry_flag: bool = Flag::C.get(st);

	let new_value: u8 = (old_value << 1) | (old_carry_flag as u8);
	x.set(st, new_value);
	Flag::C.set(st, old_bit_7);

	Flag::Z.set(st, new_value == 0x00);
	Flag::N.set(st, false);
	Flag::H.set(st, false);
}

/// UM0080.pdf rev 11 p221 / 332
const RLA_NAME: &'static str = "RLA";
fn RLA(st: &mut State, _: ()) {
	let old_Z: bool = Flag::Z.get(st);
	RL(st, A);
	Flag::Z.set(st, old_Z);
}

/// UM0080.pdf rev 11 p241 / 332
const RR_NAME: &'static str = "RR";
fn RR<T>(st: &mut State, x: T) where T: RW<u8> {
	unimplemented!();
}

/// UM0080.pdf rev 11 p225 / 332
const RRA_NAME: &'static str = "RRA";
fn RRA(st: &mut State, _: ()) {
	unimplemented!();
}

/// UM0080.pdf rev 11 p167 / 332
const SUB_NAME: &'static str = "SUB";
fn SUB<T>(st: &mut State, x: T) where T: R<u8> {
	CP(st, x);
	let new_value: u8 = A.get(st) - x.get(st);
	A.set(st, new_value);
}

/// UM0080.pdf rev 11 p175 / 332
const XOR_NAME: &'static str = "XOR";
fn XOR<T>(st: &mut State, x: T) where T: R<u8> {
	let new_value: u8 = A.get(st) ^ x.get(st);
	A.set(st, new_value);

	let result_is_zero: bool = A.get(st) == 0x00;
	Flag::Z.set(st, result_is_zero);
	Flag::N.set(st, false);
	Flag::H.set(st, false);
	Flag::C.set(st, false);
}

// -----------------------
// Jump table construction
// -----------------------

pub struct CPU {
	jump_table: [Option<Box<IInstr>>; 256],
	extended_jump_table: [Option<Box<IInstr>>; 256],
}

impl CPU {
	pub fn new() -> CPU {
		let mut jump_table: [Option<Box<IInstr>>; 256] = init_jump_table();
		let mut extended_jump_table: [Option<Box<IInstr>>; 256] = init_jump_table();

		{
			/// Add an entry to a jump_table.
			fn add_entry(
				name: &'static str,
				jump_table: &mut [Option<Box<IInstr>>; 256],
				hex_opcode: u8,
				bin_opcode: u8,
				instr: Box<IInstr>
			) {
				assert!(hex_opcode == bin_opcode);
				let opcode: u8 = bin_opcode;
				match jump_table[opcode as usize] {
					None => jump_table[opcode as usize] = Some(instr),
					Some(_) => panic!("There is already an instruction in {}[0x{:02X}].", name, opcode),
				};
			}

			let mut add = |hex_opcode: u8, bin_opcode: u8, _: bool, instr: Box<IInstr>| {
				add_entry("jump_table", &mut jump_table, hex_opcode, bin_opcode, instr);
			};

			// BEGIN jump_table
			add(0x00, 0b00000000, false, Instr::new(NOP_NAME, NOP, ()));
			add(0x01, 0b00000001, false, Instr::new(LD_NAME, LD, (BC, Imm_u16)));
			add(0x02, 0b00000010, false, Instr::new(LD_NAME, LD, (Mem(BC), A)));
			add(0x03, 0b00000011, false, Instr::new(INC_U16_NAME, INC_u16, BC));
			add(0x04, 0b00000100, false, Instr::new(INC_U8_NAME, INC_u8, B));
			add(0x05, 0b00000101, false, Instr::new(DEC_U8_NAME, DEC_u8, B));
			add(0x06, 0b00000110, false, Instr::new(LD_NAME, LD, (B, Imm_u8)));
			add(0x0A, 0b00001010, false, Instr::new(LD_NAME, LD, (A, Mem(BC))));
			add(0x0C, 0b00001100, false, Instr::new(INC_U8_NAME, INC_u8, C));
			add(0x0D, 0b00001101, false, Instr::new(DEC_U8_NAME, DEC_u8, C));
			add(0x0E, 0b00001110, false, Instr::new(LD_NAME, LD, (C, Imm_u8)));
			add(0x11, 0b00010001, false, Instr::new(LD_NAME, LD, (DE, Imm_u16)));
			add(0x12, 0b00010010, false, Instr::new(LD_NAME, LD, (Mem(DE), A)));
			add(0x13, 0b00010011, false, Instr::new(INC_U16_NAME, INC_u16, DE));
			add(0x14, 0b00010100, false, Instr::new(INC_U8_NAME, INC_u8, D));
			add(0x15, 0b00010101, false, Instr::new(DEC_U8_NAME, DEC_u8, D));
			add(0x16, 0b00010110, false, Instr::new(LD_NAME, LD, (D, Imm_u8)));
			add(0x17, 0b00010111, false, Instr::new(RLA_NAME, RLA, ()));
			add(0x18, 0b00011000, false, Instr::new(JR1_NAME, JR1, Imm_i8));
			add(0x1A, 0b00011010, false, Instr::new(LD_NAME, LD, (A, Mem(DE))));
			add(0x1C, 0b00011100, false, Instr::new(INC_U8_NAME, INC_u8, E));
			add(0x1D, 0b00011101, false, Instr::new(DEC_U8_NAME, DEC_u8, E));
			add(0x1E, 0b00011110, false, Instr::new(LD_NAME, LD, (E, Imm_u8)));
			add(0x1F, 0b00011111, false, Instr::new(RRA_NAME, RRA, ()));
			add(0x20, 0b00100000, false, Instr::new(JR2_NAME, JR2, (N(Flag::Z), Imm_i8)));
			add(0x21, 0b00100001, false, Instr::new(LD_NAME, LD, (HL, Imm_u16)));
			add(0x22, 0b00100010, true, Instr::new(LDI_NAME, LDI, (Mem(HL), A)));
			add(0x23, 0b00100011, false, Instr::new(INC_U16_NAME, INC_u16, HL));
			add(0x24, 0b00100100, false, Instr::new(INC_U8_NAME, INC_u8, H));
			add(0x25, 0b00100101, false, Instr::new(DEC_U8_NAME, DEC_u8, H));
			add(0x26, 0b00100110, false, Instr::new(LD_NAME, LD, (H, Imm_u8)));
			add(0x28, 0b00101000, false, Instr::new(JR2_NAME, JR2, (Flag::Z, Imm_i8)));
			add(0x2A, 0b00101010, true, Instr::new(LDI_NAME, LDI, (A, Mem(HL))));
			add(0x2C, 0b00101100, false, Instr::new(INC_U8_NAME, INC_u8, L));
			add(0x2D, 0b00101101, false, Instr::new(DEC_U8_NAME, DEC_u8, L));
			add(0x2E, 0b00101110, false, Instr::new(LD_NAME, LD, (L, Imm_u8)));
			add(0x30, 0b00110000, false, Instr::new(JR2_NAME, JR2, (N(Flag::C), Imm_i8)));
			add(0x31, 0b00110001, false, Instr::new(LD_NAME, LD, (SP, Imm_u16)));
			add(0x32, 0b00110010, true, Instr::new(LDD_NAME, LDD, (Mem(HL), A)));
			add(0x33, 0b00110011, false, Instr::new(INC_U16_NAME, INC_u16, SP));
			add(0x34, 0b00110100, false, Instr::new(INC_U8_NAME, INC_u8, Mem(HL)));
			add(0x35, 0b00110101, false, Instr::new(DEC_U8_NAME, DEC_u8, Mem(HL)));
			add(0x36, 0b00110110, false, Instr::new(LD_NAME, LD, (Mem(HL), Imm_u8)));
			add(0x38, 0b00111000, false, Instr::new(JR2_NAME, JR2, (Flag::C, Imm_i8)));
			add(0x3A, 0b00111010, true, Instr::new(LDD_NAME, LDD, (A, Mem(HL))));
			add(0x3C, 0b00111100, false, Instr::new(INC_U8_NAME, INC_u8, A));
			add(0x3D, 0b00111101, false, Instr::new(DEC_U8_NAME, DEC_u8, A));
			add(0x3E, 0b00111110, false, Instr::new(LD_NAME, LD, (A, Imm_u8)));
			add(0x40, 0b01000000, false, Instr::new(LD_NAME, LD, (B, B)));
			add(0x41, 0b01000001, false, Instr::new(LD_NAME, LD, (B, C)));
			add(0x42, 0b01000010, false, Instr::new(LD_NAME, LD, (B, D)));
			add(0x43, 0b01000011, false, Instr::new(LD_NAME, LD, (B, E)));
			add(0x44, 0b01000100, false, Instr::new(LD_NAME, LD, (B, H)));
			add(0x45, 0b01000101, false, Instr::new(LD_NAME, LD, (B, L)));
			add(0x46, 0b01000110, false, Instr::new(LD_NAME, LD, (B, Mem(HL))));
			add(0x47, 0b01000111, false, Instr::new(LD_NAME, LD, (B, A)));
			add(0x48, 0b01001000, false, Instr::new(LD_NAME, LD, (C, B)));
			add(0x49, 0b01001001, false, Instr::new(LD_NAME, LD, (C, C)));
			add(0x4A, 0b01001010, false, Instr::new(LD_NAME, LD, (C, D)));
			add(0x4B, 0b01001011, false, Instr::new(LD_NAME, LD, (C, E)));
			add(0x4C, 0b01001100, false, Instr::new(LD_NAME, LD, (C, H)));
			add(0x4D, 0b01001101, false, Instr::new(LD_NAME, LD, (C, L)));
			add(0x4E, 0b01001110, false, Instr::new(LD_NAME, LD, (C, Mem(HL))));
			add(0x4F, 0b01001111, false, Instr::new(LD_NAME, LD, (C, A)));
			add(0x50, 0b01010000, false, Instr::new(LD_NAME, LD, (D, B)));
			add(0x51, 0b01010001, false, Instr::new(LD_NAME, LD, (D, C)));
			add(0x52, 0b01010010, false, Instr::new(LD_NAME, LD, (D, D)));
			add(0x53, 0b01010011, false, Instr::new(LD_NAME, LD, (D, E)));
			add(0x54, 0b01010100, false, Instr::new(LD_NAME, LD, (D, H)));
			add(0x55, 0b01010101, false, Instr::new(LD_NAME, LD, (D, L)));
			add(0x56, 0b01010110, false, Instr::new(LD_NAME, LD, (D, Mem(HL))));
			add(0x57, 0b01010111, false, Instr::new(LD_NAME, LD, (D, A)));
			add(0x58, 0b01011000, false, Instr::new(LD_NAME, LD, (E, B)));
			add(0x59, 0b01011001, false, Instr::new(LD_NAME, LD, (E, C)));
			add(0x5A, 0b01011010, false, Instr::new(LD_NAME, LD, (E, D)));
			add(0x5B, 0b01011011, false, Instr::new(LD_NAME, LD, (E, E)));
			add(0x5C, 0b01011100, false, Instr::new(LD_NAME, LD, (E, H)));
			add(0x5D, 0b01011101, false, Instr::new(LD_NAME, LD, (E, L)));
			add(0x5E, 0b01011110, false, Instr::new(LD_NAME, LD, (E, Mem(HL))));
			add(0x5F, 0b01011111, false, Instr::new(LD_NAME, LD, (E, A)));
			add(0x60, 0b01100000, false, Instr::new(LD_NAME, LD, (H, B)));
			add(0x61, 0b01100001, false, Instr::new(LD_NAME, LD, (H, C)));
			add(0x62, 0b01100010, false, Instr::new(LD_NAME, LD, (H, D)));
			add(0x63, 0b01100011, false, Instr::new(LD_NAME, LD, (H, E)));
			add(0x64, 0b01100100, false, Instr::new(LD_NAME, LD, (H, H)));
			add(0x65, 0b01100101, false, Instr::new(LD_NAME, LD, (H, L)));
			add(0x66, 0b01100110, false, Instr::new(LD_NAME, LD, (H, Mem(HL))));
			add(0x67, 0b01100111, false, Instr::new(LD_NAME, LD, (H, A)));
			add(0x68, 0b01101000, false, Instr::new(LD_NAME, LD, (L, B)));
			add(0x69, 0b01101001, false, Instr::new(LD_NAME, LD, (L, C)));
			add(0x6A, 0b01101010, false, Instr::new(LD_NAME, LD, (L, D)));
			add(0x6B, 0b01101011, false, Instr::new(LD_NAME, LD, (L, E)));
			add(0x6C, 0b01101100, false, Instr::new(LD_NAME, LD, (L, H)));
			add(0x6D, 0b01101101, false, Instr::new(LD_NAME, LD, (L, L)));
			add(0x6E, 0b01101110, false, Instr::new(LD_NAME, LD, (L, Mem(HL))));
			add(0x6F, 0b01101111, false, Instr::new(LD_NAME, LD, (L, A)));
			add(0x70, 0b01110000, false, Instr::new(LD_NAME, LD, (Mem(HL), B)));
			add(0x71, 0b01110001, false, Instr::new(LD_NAME, LD, (Mem(HL), C)));
			add(0x72, 0b01110010, false, Instr::new(LD_NAME, LD, (Mem(HL), D)));
			add(0x73, 0b01110011, false, Instr::new(LD_NAME, LD, (Mem(HL), E)));
			add(0x74, 0b01110100, false, Instr::new(LD_NAME, LD, (Mem(HL), H)));
			add(0x75, 0b01110101, false, Instr::new(LD_NAME, LD, (Mem(HL), L)));
			add(0x77, 0b01110111, false, Instr::new(LD_NAME, LD, (Mem(HL), A)));
			add(0x78, 0b01111000, false, Instr::new(LD_NAME, LD, (A, B)));
			add(0x79, 0b01111001, false, Instr::new(LD_NAME, LD, (A, C)));
			add(0x7A, 0b01111010, false, Instr::new(LD_NAME, LD, (A, D)));
			add(0x7B, 0b01111011, false, Instr::new(LD_NAME, LD, (A, E)));
			add(0x7C, 0b01111100, false, Instr::new(LD_NAME, LD, (A, H)));
			add(0x7D, 0b01111101, false, Instr::new(LD_NAME, LD, (A, L)));
			add(0x7E, 0b01111110, false, Instr::new(LD_NAME, LD, (A, Mem(HL))));
			add(0x7F, 0b01111111, false, Instr::new(LD_NAME, LD, (A, A)));
			add(0x80, 0b10000000, false, Instr::new(ADD1_NAME, ADD1, B));
			add(0x81, 0b10000001, false, Instr::new(ADD1_NAME, ADD1, C));
			add(0x82, 0b10000010, false, Instr::new(ADD1_NAME, ADD1, D));
			add(0x83, 0b10000011, false, Instr::new(ADD1_NAME, ADD1, E));
			add(0x84, 0b10000100, false, Instr::new(ADD1_NAME, ADD1, H));
			add(0x85, 0b10000101, false, Instr::new(ADD1_NAME, ADD1, L));
			add(0x86, 0b10000110, false, Instr::new(ADD1_NAME, ADD1, Mem(HL)));
			add(0x87, 0b10000111, false, Instr::new(ADD1_NAME, ADD1, A));
			add(0x90, 0b10010000, false, Instr::new(SUB_NAME, SUB, B));
			add(0x91, 0b10010001, false, Instr::new(SUB_NAME, SUB, C));
			add(0x92, 0b10010010, false, Instr::new(SUB_NAME, SUB, D));
			add(0x93, 0b10010011, false, Instr::new(SUB_NAME, SUB, E));
			add(0x94, 0b10010100, false, Instr::new(SUB_NAME, SUB, H));
			add(0x95, 0b10010101, false, Instr::new(SUB_NAME, SUB, L));
			add(0x96, 0b10010110, false, Instr::new(SUB_NAME, SUB, Mem(HL)));
			add(0x97, 0b10010111, false, Instr::new(SUB_NAME, SUB, A));
			add(0xA8, 0b10101000, false, Instr::new(XOR_NAME, XOR, B));
			add(0xA9, 0b10101001, false, Instr::new(XOR_NAME, XOR, C));
			add(0xAA, 0b10101010, false, Instr::new(XOR_NAME, XOR, D));
			add(0xAB, 0b10101011, false, Instr::new(XOR_NAME, XOR, E));
			add(0xAC, 0b10101100, false, Instr::new(XOR_NAME, XOR, H));
			add(0xAD, 0b10101101, false, Instr::new(XOR_NAME, XOR, L));
			add(0xAE, 0b10101110, false, Instr::new(XOR_NAME, XOR, Mem(HL)));
			add(0xAF, 0b10101111, false, Instr::new(XOR_NAME, XOR, A));
			add(0xB8, 0b10111000, false, Instr::new(CP_NAME, CP, B));
			add(0xB9, 0b10111001, false, Instr::new(CP_NAME, CP, C));
			add(0xBA, 0b10111010, false, Instr::new(CP_NAME, CP, D));
			add(0xBB, 0b10111011, false, Instr::new(CP_NAME, CP, E));
			add(0xBC, 0b10111100, false, Instr::new(CP_NAME, CP, H));
			add(0xBD, 0b10111101, false, Instr::new(CP_NAME, CP, L));
			add(0xBE, 0b10111110, false, Instr::new(CP_NAME, CP, Mem(HL)));
			add(0xBF, 0b10111111, false, Instr::new(CP_NAME, CP, A));
			add(0xC1, 0b11000001, false, Instr::new(POP_NAME, POP, BC));
			add(0xC5, 0b11000101, false, Instr::new(PUSH_NAME, PUSH, BC));
			add(0xC6, 0b11000110, false, Instr::new(ADD1_NAME, ADD1, Imm_u8));
			add(0xC9, 0b11001001, false, Instr::new(RET0_NAME, RET0, ()));
			add(0xCD, 0b11001101, false, Instr::new(CALL1_NAME, CALL1, Imm_u16));
			add(0xD1, 0b11010001, false, Instr::new(POP_NAME, POP, DE));
			add(0xD5, 0b11010101, false, Instr::new(PUSH_NAME, PUSH, DE));
			add(0xD6, 0b11010110, false, Instr::new(SUB_NAME, SUB, Imm_u8));
			add(0xE0, 0b11100000, true, Instr::new(LD_NAME, LD, (Mem(FF00(Imm_u8)), A)));
			add(0xE1, 0b11100001, false, Instr::new(POP_NAME, POP, HL));
			add(0xE2, 0b11100010, true, Instr::new(LD_NAME, LD, (Mem(FF00(C)), A)));
			add(0xE5, 0b11100101, false, Instr::new(PUSH_NAME, PUSH, HL));
			add(0xEA, 0b11101010, true, Instr::new(LD_NAME, LD, (Mem(Imm_u16), A)));
			add(0xEE, 0b11101110, false, Instr::new(XOR_NAME, XOR, Imm_u8));
			add(0xF0, 0b11110000, true, Instr::new(LD_NAME, LD, (A, Mem(FF00(Imm_u8)))));
			add(0xF1, 0b11110001, false, Instr::new(POP_NAME, POP, AF));
			add(0xF2, 0b11110010, true, Instr::new(LD_NAME, LD, (A, Mem(FF00(C)))));
			add(0xF5, 0b11110101, false, Instr::new(PUSH_NAME, PUSH, AF));
			add(0xFA, 0b11111010, true, Instr::new(LD_NAME, LD, (A, Mem(Imm_u16))));
			add(0xFE, 0b11111110, false, Instr::new(CP_NAME, CP, Imm_u8));
			// END jump_table

			let mut add = |hex_opcode: u8, bin_opcode: u8, _: bool, instr: Box<IInstr>| {
				add_entry("extended_jump_table", &mut extended_jump_table, hex_opcode, bin_opcode, instr);
			};

			// BEGIN extended_jump_table
			add(0x10, 0b00010000, false, Instr::new(RL_NAME, RL, B));
			add(0x11, 0b00010001, false, Instr::new(RL_NAME, RL, C));
			add(0x12, 0b00010010, false, Instr::new(RL_NAME, RL, D));
			add(0x13, 0b00010011, false, Instr::new(RL_NAME, RL, E));
			add(0x14, 0b00010100, false, Instr::new(RL_NAME, RL, H));
			add(0x15, 0b00010101, false, Instr::new(RL_NAME, RL, L));
			add(0x16, 0b00010110, false, Instr::new(RL_NAME, RL, Mem(HL)));
			add(0x17, 0b00010111, false, Instr::new(RL_NAME, RL, A));
			add(0x18, 0b00011000, false, Instr::new(RR_NAME, RR, B));
			add(0x19, 0b00011001, false, Instr::new(RR_NAME, RR, C));
			add(0x1A, 0b00011010, false, Instr::new(RR_NAME, RR, D));
			add(0x1B, 0b00011011, false, Instr::new(RR_NAME, RR, E));
			add(0x1C, 0b00011100, false, Instr::new(RR_NAME, RR, H));
			add(0x1D, 0b00011101, false, Instr::new(RR_NAME, RR, L));
			add(0x1E, 0b00011110, false, Instr::new(RR_NAME, RR, Mem(HL)));
			add(0x1F, 0b00011111, false, Instr::new(RR_NAME, RR, A));
			add(0x40, 0b01000000, false, Instr::new(BIT_NAME, BIT, (0, B)));
			add(0x41, 0b01000001, false, Instr::new(BIT_NAME, BIT, (0, C)));
			add(0x42, 0b01000010, false, Instr::new(BIT_NAME, BIT, (0, D)));
			add(0x43, 0b01000011, false, Instr::new(BIT_NAME, BIT, (0, E)));
			add(0x44, 0b01000100, false, Instr::new(BIT_NAME, BIT, (0, H)));
			add(0x45, 0b01000101, false, Instr::new(BIT_NAME, BIT, (0, L)));
			add(0x46, 0b01000110, false, Instr::new(BIT_NAME, BIT, (0, Mem(HL))));
			add(0x47, 0b01000111, false, Instr::new(BIT_NAME, BIT, (0, A)));
			add(0x48, 0b01001000, false, Instr::new(BIT_NAME, BIT, (1, B)));
			add(0x49, 0b01001001, false, Instr::new(BIT_NAME, BIT, (1, C)));
			add(0x4A, 0b01001010, false, Instr::new(BIT_NAME, BIT, (1, D)));
			add(0x4B, 0b01001011, false, Instr::new(BIT_NAME, BIT, (1, E)));
			add(0x4C, 0b01001100, false, Instr::new(BIT_NAME, BIT, (1, H)));
			add(0x4D, 0b01001101, false, Instr::new(BIT_NAME, BIT, (1, L)));
			add(0x4E, 0b01001110, false, Instr::new(BIT_NAME, BIT, (1, Mem(HL))));
			add(0x4F, 0b01001111, false, Instr::new(BIT_NAME, BIT, (1, A)));
			add(0x50, 0b01010000, false, Instr::new(BIT_NAME, BIT, (2, B)));
			add(0x51, 0b01010001, false, Instr::new(BIT_NAME, BIT, (2, C)));
			add(0x52, 0b01010010, false, Instr::new(BIT_NAME, BIT, (2, D)));
			add(0x53, 0b01010011, false, Instr::new(BIT_NAME, BIT, (2, E)));
			add(0x54, 0b01010100, false, Instr::new(BIT_NAME, BIT, (2, H)));
			add(0x55, 0b01010101, false, Instr::new(BIT_NAME, BIT, (2, L)));
			add(0x56, 0b01010110, false, Instr::new(BIT_NAME, BIT, (2, Mem(HL))));
			add(0x57, 0b01010111, false, Instr::new(BIT_NAME, BIT, (2, A)));
			add(0x58, 0b01011000, false, Instr::new(BIT_NAME, BIT, (3, B)));
			add(0x59, 0b01011001, false, Instr::new(BIT_NAME, BIT, (3, C)));
			add(0x5A, 0b01011010, false, Instr::new(BIT_NAME, BIT, (3, D)));
			add(0x5B, 0b01011011, false, Instr::new(BIT_NAME, BIT, (3, E)));
			add(0x5C, 0b01011100, false, Instr::new(BIT_NAME, BIT, (3, H)));
			add(0x5D, 0b01011101, false, Instr::new(BIT_NAME, BIT, (3, L)));
			add(0x5E, 0b01011110, false, Instr::new(BIT_NAME, BIT, (3, Mem(HL))));
			add(0x5F, 0b01011111, false, Instr::new(BIT_NAME, BIT, (3, A)));
			add(0x60, 0b01100000, false, Instr::new(BIT_NAME, BIT, (4, B)));
			add(0x61, 0b01100001, false, Instr::new(BIT_NAME, BIT, (4, C)));
			add(0x62, 0b01100010, false, Instr::new(BIT_NAME, BIT, (4, D)));
			add(0x63, 0b01100011, false, Instr::new(BIT_NAME, BIT, (4, E)));
			add(0x64, 0b01100100, false, Instr::new(BIT_NAME, BIT, (4, H)));
			add(0x65, 0b01100101, false, Instr::new(BIT_NAME, BIT, (4, L)));
			add(0x66, 0b01100110, false, Instr::new(BIT_NAME, BIT, (4, Mem(HL))));
			add(0x67, 0b01100111, false, Instr::new(BIT_NAME, BIT, (4, A)));
			add(0x68, 0b01101000, false, Instr::new(BIT_NAME, BIT, (5, B)));
			add(0x69, 0b01101001, false, Instr::new(BIT_NAME, BIT, (5, C)));
			add(0x6A, 0b01101010, false, Instr::new(BIT_NAME, BIT, (5, D)));
			add(0x6B, 0b01101011, false, Instr::new(BIT_NAME, BIT, (5, E)));
			add(0x6C, 0b01101100, false, Instr::new(BIT_NAME, BIT, (5, H)));
			add(0x6D, 0b01101101, false, Instr::new(BIT_NAME, BIT, (5, L)));
			add(0x6E, 0b01101110, false, Instr::new(BIT_NAME, BIT, (5, Mem(HL))));
			add(0x6F, 0b01101111, false, Instr::new(BIT_NAME, BIT, (5, A)));
			add(0x70, 0b01110000, false, Instr::new(BIT_NAME, BIT, (6, B)));
			add(0x71, 0b01110001, false, Instr::new(BIT_NAME, BIT, (6, C)));
			add(0x72, 0b01110010, false, Instr::new(BIT_NAME, BIT, (6, D)));
			add(0x73, 0b01110011, false, Instr::new(BIT_NAME, BIT, (6, E)));
			add(0x74, 0b01110100, false, Instr::new(BIT_NAME, BIT, (6, H)));
			add(0x75, 0b01110101, false, Instr::new(BIT_NAME, BIT, (6, L)));
			add(0x76, 0b01110110, false, Instr::new(BIT_NAME, BIT, (6, Mem(HL))));
			add(0x77, 0b01110111, false, Instr::new(BIT_NAME, BIT, (6, A)));
			add(0x78, 0b01111000, false, Instr::new(BIT_NAME, BIT, (7, B)));
			add(0x79, 0b01111001, false, Instr::new(BIT_NAME, BIT, (7, C)));
			add(0x7A, 0b01111010, false, Instr::new(BIT_NAME, BIT, (7, D)));
			add(0x7B, 0b01111011, false, Instr::new(BIT_NAME, BIT, (7, E)));
			add(0x7C, 0b01111100, false, Instr::new(BIT_NAME, BIT, (7, H)));
			add(0x7D, 0b01111101, false, Instr::new(BIT_NAME, BIT, (7, L)));
			add(0x7E, 0b01111110, false, Instr::new(BIT_NAME, BIT, (7, Mem(HL))));
			add(0x7F, 0b01111111, false, Instr::new(BIT_NAME, BIT, (7, A)));
			// END extended_jump_table
		}

		CPU { jump_table, extended_jump_table }
	}
}

fn init_jump_table() -> [Option<Box<IInstr>>; 256] {
	let mut jump_table: [Option<Box<IInstr>>; 256];
	unsafe {
		jump_table = std::mem::uninitialized();
		for x in &mut jump_table[..] {
			std::ptr::write(x, None);
		}
	}
	jump_table
}