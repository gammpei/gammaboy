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
	var romPath string
	var title string
	switch len(args) {
	case 0:
		romPath = ""
		title = "gammaboy"
	case 1:
		romPath = args[0]
		title = romPath + " < gammaboy"
	default:
		assert(false)
	}

	st := newState(romPath)

	gui := newGui(title)
	defer gui.close()

	defer stopWatch("main loop", time.Now())

	curScanline := getScanline(&st)
	for {
		// Draw the whole frame at once (good enough for now).
		gui.drawFrame(&st)

		// Process the events once per frame (good enough for now).
		if !gui.processEvents() {
			break
		}

		// Execute instructions until we need to draw a frame.
		for {
			prevScanline := curScanline

			// We assume that all instructions take 4 cycles to execute (good enough for now).
			fetchDecodeExecute(&st)
			st.cycles += 4

			// If the scanline wraps around, we break and draw a frame.
			curScanline = getScanline(&st)
			if curScanline < prevScanline {
				break
			}
		}
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
