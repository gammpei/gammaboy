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
	"crypto/sha256"
	"fmt"
	"time"
)

type u3 uint
type u8 = uint8
type i8 = int8
type u16 = uint16
type u32 = uint32

func sha256Hash(x []u8) string {
	return fmt.Sprintf("%x", sha256.Sum256(x))
}

func stopWatch(s string, start time.Time) {
	elapsed := time.Since(start)
	fmt.Printf("%s: %.3fs\n", s, elapsed.Seconds())
}

func getBit(x u8, bit uint) bool {
	assert(0 <= bit && bit <= 7)
	return (x>>bit)&0x01 != 0x00
}

func u8FromBool(x bool) u8 {
	if x {
		return 0x01
	} else {
		return 0x00
	}
}

func assert(cond bool) {
	if !cond {
		panic("Assertion failed.")
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
