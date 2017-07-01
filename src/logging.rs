use cpu::IInstr;
use state::Reg16::*;
use state::{read_mem, State};

const LOGGING_IS_ON: bool = true;

pub fn log_registers(st: &State) {
	if !LOGGING_IS_ON { return; }

	println!("PC=0x{:04X} AF=0x{:04X} BC=0x{:04X} DE=0x{:04X} HL=0x{:04X} SP=0x{:04X}",
		PC.get(st), AF.get(st), BC.get(st), DE.get(st), HL.get(st), SP.get(st)
	);
}

/// Logs an instruction. This must be called after PC was incremented,
/// but before the instruction is executed.
pub fn log_instr(st: &State, instr: &IInstr, instr_length: i32) {
	if !LOGGING_IS_ON { return; }

	let PC_0: u16 = PC.get(st) - (instr_length as u16);
	let r = |offset: u16| -> u8 {
		let addr: u16 = PC_0 + offset;
		read_mem(st, addr)
	};

	let instr_bytes: String = match instr_length {
		1 => format!("0x{:02X}          ", r(0)),
		2 => format!("0x{:02X} 0x{:02X}     ", r(0), r(1)),
		3 => format!("0x{:02X} 0x{:02X} 0x{:02X}", r(0), r(1), r(2)),
		_ => unreachable!(),
	};
	println!(" {} | {}", instr_bytes, instr.log(st));
}