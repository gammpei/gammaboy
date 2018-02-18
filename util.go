package main

type u3 uint
type u8 = uint8
type i8 = int8
type u16 = uint16
type u32 = uint32

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
