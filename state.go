package main

// The state of the emulator.
type state struct {
	regs [6]u16
	mem  [0xFFFF + 1]u8
}
type st = state
