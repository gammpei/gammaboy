package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func main() {
	const FILENAME string = "src/cpu.rs"

	var bytes []byte
	var err error
	bytes, err = ioutil.ReadFile(FILENAME)
	check(err)
	var file string = string(bytes)

	file = replace(file, "jump_table", non_extended_opcodes())
	file = replace(file, "extended_jump_table", extended_opcodes())

	var fi os.FileInfo
	fi, err = os.Stat(FILENAME)
	check(err)
	err = ioutil.WriteFile(FILENAME, []byte(file), fi.Mode())
	check(err)
}

func replace(file string, keyword string, lines []string) string {
	var str string = fmt.Sprintf(`(?s)(// BEGIN %s\r?\n).*(\r?\n\t*// END %s)`, keyword, keyword)

	sort.Strings(lines)
	var payload string = strings.Join(lines, "\n")

	var re *regexp.Regexp = regexp.MustCompile(str)
	return re.ReplaceAllString(file, fmt.Sprintf("${1}%s${2}", payload))
}

var R1 = map[string]string{
	"0": "BC",
	"1": "DE",
}

var R2 = map[string]string{
	"00": "BC",
	"01": "DE",
	"10": "HL",
	"11": "SP",
}

var R3 = map[string]string{
	"00": "BC",
	"01": "DE",
	"10": "HL",
	"11": "AF",
}

var D = map[string]string{
	"000": "B",
	"001": "C",
	"010": "D",
	"011": "E",
	"100": "H",
	"101": "L",
	"110": "Mem(HL)",
	"111": "A",
}

var F = map[string]string{
	"00": "N(Flag::Z)",
	"01": "Flag::Z",
	"10": "N(Flag::C)",
	"11": "Flag::C",
}

var ALU = map[string]string{
	"000": "ADD1",
	"010": "SUB",
	"101": "XOR",
	"111": "CP",
}

var Direction = map[string]string{
	"0": "L",
	"1": "R",
}

var N = map[string]string{
	"000": "0",
	"001": "1",
	"010": "2",
	"011": "3",
	"100": "4",
	"101": "5",
	"110": "6",
	"111": "7",
}

func non_extended_opcodes() []string {
	var lines = []string{}
	var add = func(gb_specific_instr bool, opcode string, operation string, operands ...string) {
		lines = append(lines, entry(gb_specific_instr, opcode, operation, operands...))
	}

	// NOP
	add(false, "00000000", "NOP")

	// LD (N),SP

	// LD R,N
	for ii, x := range R2 {
		add(false, "00"+ii+"0001", "LD", x, "Imm_u16")
	}

	// ADD HL,R

	// LD (R),A
	for i, x := range R1 {
		add(false, "000"+i+"0010", "LD", "Mem("+x+")", "A")
	}

	// LD A,(R)
	for i, x := range R1 {
		add(false, "000"+i+"1010", "LD", "A", "Mem("+x+")")
	}

	// INC R
	for ii, x := range R2 {
		add(false, "00"+ii+"0011", "INC_u16", x)
	}

	// DEC R

	// INC D
	for iii, x := range D {
		add(false, "00"+iii+"100", "INC_u8", x)
	}

	// DEC D
	for iii, x := range D {
		add(false, "00"+iii+"101", "DEC_u8", x)
	}

	// LD D,N
	for iii, x := range D {
		add(false, "00"+iii+"110", "LD", x, "Imm_u8")
	}

	// RdCA

	// RdA
	for i, d := range Direction {
		add(false, "0001"+i+"111", "R"+d+"A")
	}

	// STOP

	// JR N
	add(false, "00011000", "JR1", "Imm_i8")

	// JR F,N
	for ii, x := range F {
		add(false, "001"+ii+"000", "JR2", x, "Imm_i8")
	}

	// LDI (HL),A
	add(true, "00100010", "LDI", "Mem(HL)", "A")

	// LDI A,(HL)
	add(true, "00101010", "LDI", "A", "Mem(HL)")

	// LDD (HL),A
	add(true, "00110010", "LDD", "Mem(HL)", "A")

	// LDD A,(HL)
	add(true, "00111010", "LDD", "A", "Mem(HL)")

	// DAA
	// CPL
	// SCF
	// CCF

	// LD D,D
	for iii, x := range D {
		for jjj, y := range D {
			if !(iii == "110" && jjj == "110") {
				add(false, "01"+iii+jjj, "LD", x, y)
			}
		}
	}

	// HALT

	// ALU A,D
	for iii, operation := range ALU {
		for jjj, x := range D {
			add(false, "10"+iii+jjj, operation, x)
		}
	}

	// ALU A,N
	for iii, operation := range ALU {
		add(false, "11"+iii+"110", operation, "Imm_u8")
	}

	// POP R
	for ii, x := range R3 {
		add(false, "11"+ii+"0001", "POP", x)
	}

	// PUSH R
	for ii, x := range R3 {
		add(false, "11"+ii+"0101", "PUSH", x)
	}

	// RST N
	// RET F

	// RET
	add(false, "11001001", "RET0")

	// RETI
	// JP F,N
	// JP N
	// CALL F,N

	// CALL N
	add(false, "11001101", "CALL1", "Imm_u16")

	// ADD SP,N
	// LD HL,SP+N

	// LD (0xFF00+N),A
	add(true, "11100000", "LD", "Mem(FF00(Imm_u8))", "A")

	// LD A,(0xFF00+N)
	add(true, "11110000", "LD", "A", "Mem(FF00(Imm_u8))")

	// LD (0xFF00+C),A
	add(true, "11100010", "LD", "Mem(FF00(C))", "A")

	// LD A,(0xFF00+C)
	add(true, "11110010", "LD", "A", "Mem(FF00(C))")

	// LD (N),A
	add(true, "11101010", "LD", "Mem(Imm_u16)", "A")

	// LD A,(N)
	add(true, "11111010", "LD", "A", "Mem(Imm_u16)")

	// ...

	return lines
}

func extended_opcodes() []string {
	var lines = []string{}
	var add = func(gb_specific_instr bool, opcode string, operation string, operands ...string) {
		lines = append(lines, entry(gb_specific_instr, opcode, operation, operands...))
	}

	// RdC

	// Rd D
	for i, d := range Direction {
		for jjj, x := range D {
			add(false, "0001"+i+jjj, "R"+d, x)
		}
	}

	// SdA D
	// SWAP D
	// SRL D

	// BIT N,D
	for iii, x := range N {
		for jjj, y := range D {
			add(false, "01"+iii+jjj, "BIT", x, y)
		}
	}

	// RES N,D
	// SET N,D

	return lines
}

func entry(gb_specific_instr bool, opcode string, operation string, operands ...string) string {
	if len(opcode) != 8 {
		panic("The opcode is not 8-bits long")
	}

	var int_opcode uint64
	var err error
	int_opcode, err = strconv.ParseUint(opcode, 2 /*base*/, 8 /*bitsize*/)
	check(err)

	var str_operands string
	switch len(operands) {
	case 1:
		str_operands = operands[0]
	case 0, 2:
		str_operands = "(" + strings.Join(operands, ", ") + ")"
	default:
		panic("Invalid number of operands")
	}

	return fmt.Sprintf(
		"\t\t\tadd(0x%02X, 0b%s, %t, Instr::new(%s_NAME, %s, %s));",
		int_opcode,
		opcode,
		gb_specific_instr,
		strings.ToUpper(operation),
		operation,
		str_operands,
	)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
