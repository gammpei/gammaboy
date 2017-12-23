package main

import (
	"crypto/sha256"
	cmdLineFlag "flag"
	"fmt"
	"io/ioutil"
)

var bios [256]u8
var jumpTable [256]*instr
var extendedJumpTable [256]*instr
var doLog bool

func init() {
	loadBios()
	buildJumpTables()
}

func main() {
	cmdLineFlag.BoolVar(&doLog, "log", false, "Output what's happening.")
	cmdLineFlag.Parse()

	st := st{}
	for {
		fetchDecodeExecute(&st)
	}
}

func loadBios() {
	file, err := ioutil.ReadFile("DMG_ROM.gb")
	check(err)

	assert(len(file) == 256)

	hash := fmt.Sprintf("%x", sha256.Sum256(file))
	assert(hash == "cf053eccb4ccafff9e67339d4e78e98dce7d1ed59be819d2a1ba2232c6fce1c7")

	copy(bios[:], file)
}
