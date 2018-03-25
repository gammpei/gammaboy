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

// The state of the emulator.
type state struct {
	regs [6]u16
	mem  [0xFFFF + 1]u8

	timing struct {
		// The number of elapsed clock cycles since powerup.
		// At 4.194304 MHz, a u64 is enough for 139 365 years...
		// Needless to say I'll let other people deal with that overflow bug...
		cycles      u64
		systemClock u16
		// The delayed system clock bit for the timer.
		delayedTimerBit bool
	}

	biosIsEnabled bool
	IME           bool // Interrupt Master Enable

	rom       []u8
	linkCable chan u8
}
type st = state

func newState(rom []u8, linkCable chan u8) *st {
	assert(len(rom) == 0x7FFF+1)

	return &st{
		timing: struct {
			cycles          u64
			systemClock     u16
			delayedTimerBit bool
		}{
			cycles:          0,
			systemClock:     0x0000,
			delayedTimerBit: false,
		},

		biosIsEnabled: true,
		IME:           false, // 0 at startup since the bios is mapped over the interrupt vector table.

		rom:       rom,
		linkCable: linkCable,
	}
}

func (st *st) addCycles(cycles int) {
	for i := 0; i < cycles; i++ {
		st.timing.cycles++

		// Update timer.
		st.timing.systemClock++
		TAC := st.readMem(0xFF07) // TAC: Timer control
		TAC_Freq := TAC & 0x03
		TAC_Enable := getBit(TAC, 2)
		systemClockBit := getBit_u16(st.timing.systemClock, [4]uint{9, 3, 5, 7}[TAC_Freq])
		timerBit := systemClockBit && TAC_Enable
		if st.timing.delayedTimerBit && !timerBit { // Falling edge.
			TIMA := st.readMem(0xFF05) // TIMA: Timer counter
			TIMA++
			if TIMA == 0x00 { // If the timer overflowed.
				st.requestInterrupt(2)    // Request timer interrupt.
				TMA := st.readMem(0xFF06) // TMA: Timer modulo
				TIMA = TMA
			}
			st.writeMem(0xFF05, TIMA)
		}
		st.timing.delayedTimerBit = timerBit
	}
}
