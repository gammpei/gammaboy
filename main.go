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
	cmdLineFlag "flag"
	"fmt"
	"io/ioutil"
	"time"
)

var bios [256]u8
var jumpTable [256]*instr
var extendedJumpTable [256]*instr

var flags struct {
	green      bool
	record     bool
	scalingAlg string
	verbose    bool
}

func init() {
	loadBios()
	buildJumpTables()
}

func main() {
	cmdLineFlag.BoolVar(&flags.green, "green", false, "Use a green palette instead of grayscale.")
	cmdLineFlag.BoolVar(&flags.record, "record", false, "Create a video recording.")
	cmdLineFlag.StringVar(&flags.scalingAlg, "scaling-alg", "0",
		"Scaling algorithm: 0 or nearest, 1 or linear.")
	cmdLineFlag.BoolVar(&flags.verbose, "verbose", false, "Print every instruction (very slow).")
	cmdLineFlag.Parse()

	args := cmdLineFlag.Args()
	var rom []u8
	var title string
	switch len(args) {
	case 0:
		rom = make([]u8, 0x7FFF+1)
		for i, _ := range rom {
			rom[i] = 0xFF
		}

		title = "gammaboy"
	case 1:
		romPath := args[0]
		startLoadRom := time.Now()
		var err error
		rom, err = ioutil.ReadFile(romPath)
		check(err)
		fmt.Printf("SHA-256 hash of the rom: %s\n", sha256Hash(rom))
		stopWatch("load rom", startLoadRom)

		title = romPath + " < gammaboy"
	default:
		assert(false)
	}

	gb := newGameBoy(rom, true /*showGui*/, title)
	defer gb.close()

	defer stopWatch("main loop", time.Now())
	gb.run()
}

type gameBoy struct {
	st  *st
	gui *gui
}

func newGameBoy(rom []u8, showGui bool, title string) *gameBoy {
	var linkCable chan u8 = nil
	var gui *gui = nil
	if showGui {
		gui = newGui(title)
	} else {
		linkCable = make(chan u8, 80)
	}

	return &gameBoy{
		st:  newState(rom, linkCable),
		gui: gui,
	}
}

func newTestGameBoy(rom []u8) *gameBoy {
	return newGameBoy(rom, false /*showGui*/, "" /*title*/)
}

func (gb *gameBoy) run() {
	st := gb.st
	gui := gb.gui

	curScanline := getScanline(st)
	for {
		if gui != nil {
			// Draw the whole frame at once (good enough for now).
			gui.drawFrame(st)

			// Process the events once per frame (good enough for now).
			if !gui.processEvents() {
				break
			}
		}

		// Execute instructions until we need to draw a frame.
		for {
			prevScanline := curScanline

			// We assume that all instructions take 4 cycles to execute (good enough for now).
			fetchDecodeExecute(st)
			st.addCycles(4)

			// V-Blank.
			curScanline = getScanline(st)
			IF := st.readMem(0xFF0F) // IF: Interrupt Flag
			if prevScanline < 144 && curScanline >= 144 {
				// Request V-Blank interrupt.
				IF = st.requestInterrupt(0)
			}

			// Handle interrupts.
			IE := st.readMem(0xFFFF) // IE: Interrupt Enable
			for i := uint(0); i <= 4; i++ {
				if getBit(IF, i) && getBit(IE, i) {
					if st.IME {
						// Disable interrupts.
						st.IME = false

						// Acknowledge interrupt.
						IF = setBit(IF, i, false)
						st.writeMem(0xFF0F, IF)

						// Call interrupt handler.
						PUSH.f.(func(*state, r_u16))(st, PC)
						interruptVector := [5]u16{0x0040, 0x0048, 0x0050, 0x0058, 0x0060}
						PC.set(st, interruptVector[i])
					}
					break
				}
			}

			// If the scanline wraps around, we break and draw a frame.
			if curScanline < prevScanline {
				break
			}
		}
	}
}

func (gb *gameBoy) close() {
	if gb.gui != nil {
		gb.gui.close()
	}
}

func loadBios() {
	defer stopWatch("loadBios", time.Now())

	fis, err := ioutil.ReadDir(".")
	check(err)

	for _, fi := range fis {
		if fi.Size() == 256 {
			file, err := ioutil.ReadFile(fi.Name())
			check(err)

			assert(len(file) == 256)

			hash := sha256Hash(file)
			if hash == "cf053eccb4ccafff9e67339d4e78e98dce7d1ed59be819d2a1ba2232c6fce1c7" {
				copy(bios[:], file)
				return
			}
		}
	}

	panic("Could not find the bios file.")
}
