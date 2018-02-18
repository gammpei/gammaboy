package main

// The state of the emulator.
type state struct {
	regs [6]u16
	mem  [0xFFFF + 1]u8

	// The number of elapsed clock cycles since powerup.
	// At 4.19 MHz, a uint64 is enough for 139 508 years...
	// Needless to say I'll let other people deal with that overflow bug...
	cycles uint64
}
type st = state
