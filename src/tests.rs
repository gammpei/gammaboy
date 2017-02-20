#![cfg(test)]

use state::*;
use state::Reg16::*;
use state::Reg8::*;

#[test]
fn test_registers() {
	let mut st = blank_state();

	let tests: [(Reg16, Reg8, Reg8, u16, u8, u8); 4] = [
		(AF, A, F, 0x1234, 0x12, 0x34),
		(BC, B, C, 0x2345, 0x23, 0x45),
		(DE, D, E, 0x3456, 0x34, 0x56),
		(HL, H, L, 0x4567, 0x45, 0x67),
	];
	for &(r16, r8_MSB, r8_LSB, hl, h, l) in tests.iter() {
		r16.set(&mut st, hl);
		assert!(r16.get(&st) == hl);
		assert!(r8_MSB.get(&st) == h);
		assert!(r8_LSB.get(&st) == l);

		// We clear the register.
		r16.set(&mut st, 0x0000);
		assert!(r16.get(&st) == 0x0000);
		assert!(r8_MSB.get(&st) == 0x00);
		assert!(r8_LSB.get(&st) == 0x00);

		r8_MSB.set(&mut st, h);
		r8_LSB.set(&mut st, l);
		assert!(r16.get(&st) == hl);
		assert!(r8_MSB.get(&st) == h);
		assert!(r8_LSB.get(&st) == l);
	}

	let tests: [(Reg16, u16); 2] = [
		(PC, 0x5678),
		(SP, 0x6789),
	];
	for &(r16, hl) in tests.iter() {
		r16.set(&mut st, hl);
		assert!(r16.get(&st) == hl);
	}
}

#[test]
fn test_flags() {
	let mut st = blank_state();

	let flags: [Flag; 4] = [
		Flag::Z,
		Flag::N,
		Flag::H,
		Flag::C,
	];
	for flag in flags.iter() {
		flag.set(&mut st, false);
		assert!(flag.get(&st) == false);

		flag.set(&mut st, true);
		assert!(flag.get(&st) == true);
	}
	assert!(F.get(&st) == 0b11110000);

	F.set(&mut st, 0b01010000);
	assert!(Flag::Z.get(&st) == false);
	assert!(Flag::N.get(&st) == true);
	assert!(Flag::H.get(&st) == false);
	assert!(Flag::C.get(&st) == true);
}

fn blank_state() -> State {
	let dummy_bios: [u8; 256] = [0x00; 256];
	State::new(dummy_bios)
}