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
	// At 4.19 MHz, a uint64 is enough for 139 508 years...
	// Needless to say I'll let other people deal with that overflow bug...
	cycles uint64

	biosIsEnabled bool
	IME           bool // Interrupt Master Enable

	rom []u8
}
type st = state

func newState(romPath string) st {
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

		assert(len(rom) == 0x7FFF+1)

		testedRoms := map[string]struct{}{}
		for _, hash := range [...]string{
			"17ada54b0b9c1a33cd5429fce5b765e42392189ca36da96312222ffe309e7ed1", // Blargg's CPU test ROM #6
		} {
			testedRoms[hash] = struct{}{}
		}
		hash := sha256Hash(rom)
		_, ok := testedRoms[hash]
		if !ok {
			fmt.Printf("Untested rom %s\n", hash)
		}
	}

	return st{
		cycles:        0,
		biosIsEnabled: true,
		IME:           false, // 0 at startup since the bios is mapped over the interrupt vector table.
		rom:           rom,
	}
}
