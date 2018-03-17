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

import (
	"fmt"
	"io/ioutil"
	"time"
)

// The state of the emulator.
type state struct {
	regs [6]u16
	mem  [0xFFFF + 1]u8

	// The number of elapsed clock cycles since powerup.
	// At 4.194304 MHz, a uint64 is enough for 139 365 years...
	// Needless to say I'll let other people deal with that overflow bug...
	cycles uint64

	biosIsEnabled bool
	IME           bool // Interrupt Master Enable

	rom []u8
}
type st = state

func newState(romPath string) *st {
	var rom []u8
	if romPath == "" {
		rom = make([]u8, 0x7FFF+1)
		for i, _ := range rom {
			rom[i] = 0xFF
		}
	} else {
		defer stopWatch("load rom", time.Now())

		var err error
		rom, err = ioutil.ReadFile(romPath)
		check(err)

		fmt.Printf("SHA-256 hash of the rom: %s\n", sha256Hash(rom))

		assert(len(rom) == 0x7FFF+1)
	}

	return &st{
		cycles:        0,
		biosIsEnabled: true,
		IME:           false, // 0 at startup since the bios is mapped over the interrupt vector table.
		rom:           rom,
	}
}
