package main

import "fmt"

const (
	SCREEN_WIDTH  = 64    // px
	SCREEN_HEIGHT = 32    // px
	RAM_SIZE      = 4096  // kb
	NUM_REGS      = 16    // #registers
	STACK_SIZE    = 16    // stack depth
	NUM_KEYS      = 16    // #keys
	START_ADDR    = 0x200 // start address of program
	FONTSET_SIZE  = 80    // (0-9A-F) 16 * 5 bytes
)

var FONTSET = [FONTSET_SIZE]uint8{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

var OPCODES = map[uint16]interface{}{
	0x0000: 0,
	0x00E0: 0,
	0x00EE: 0,
	0x1000: 0,
	0x2000: 0,
	0x3000: 0,
	0x4000: 0,
	0x5000: 0,
	0x6:    0,
	0x7:    0,
	0x8:    0,
	0x9:    0,
	0xA:    0,
	0xB:    0,
	0xC:    0,
	0xD:    0,
	0xE:    0,
	0xF:    0,
}

type Emu struct {
	pc     uint16
	ram    [RAM_SIZE]uint8
	screen [SCREEN_WIDTH * SCREEN_HEIGHT]bool
	v_reg  [NUM_REGS]uint8
	i_reg  uint16
	sp     uint16
	stack  [STACK_SIZE]uint16
	keys   [NUM_KEYS]bool
	dt     uint8
	st     uint8
}

func NewEmu() Emu {
	newEmu := Emu{
		pc:     START_ADDR,
		ram:    [RAM_SIZE]uint8{},
		screen: [SCREEN_WIDTH * SCREEN_HEIGHT]bool{},
		v_reg:  [NUM_REGS]uint8{},
		i_reg:  0,
		sp:     0,
		stack:  [STACK_SIZE]uint16{},
		keys:   [NUM_KEYS]bool{},
		dt:     0,
		st:     0,
	}

	// copy fontset to ram
	copy(newEmu.ram[:FONTSET_SIZE], FONTSET[:])

	return newEmu
}

func (e *Emu) push(val uint16) {
	e.stack[e.sp] = val
	e.sp += 1
}

func (e *Emu) pop() uint16 {
	e.sp -= 1
	return e.stack[e.sp]
}

func (e *Emu) reset() {
	e.pc = START_ADDR
	e.ram = [RAM_SIZE]uint8{}
	e.screen = [SCREEN_WIDTH * SCREEN_HEIGHT]bool{}
	e.v_reg = [NUM_REGS]uint8{}
	e.i_reg = 0
	e.sp = 0
	e.stack = [STACK_SIZE]uint16{}
	e.keys = [NUM_KEYS]bool{}
	e.dt = 0
	e.st = 0
	copy(e.ram[:FONTSET_SIZE], FONTSET[:])
}

func (e *Emu) tick() {
	// fetch
	op := e.fetch()
	// decode & execute
	e.execute(op)
}

func (e *Emu) fetch() uint16 {
	// opcodes are 2 bytes (including operands)
	lower_byte := e.ram[e.pc]
	upper_byte := e.ram[e.pc+1]
	// combine as big endian
	var op uint16 = uint16(upper_byte)<<8 | uint16(lower_byte)
	e.pc += 2
	return op
}

func (e *Emu) tick_timers() {
	if e.dt > 0 {
		e.dt -= 1
	}

	if e.st > 0 {
		if e.st == 1 {
			// beep
		}
		e.st -= 1
	}
}

func (e *Emu) execute(op uint16) {
	// 0xabcd -> 0xa / 0xb / 0xc / 0xd
	digit1 := (op >> 12) & 0xF
	// digit2 := (op >> 8) & 0xF
	// digit3 := (op >> 4) & 0xF
	// digit4 := op & 0xF

	switch {
	case digit1 == 0x1:
		fmt.Println("1")
	default:
		panic(fmt.Sprintf("opcode unimplemented %x", op))
	}
}

func main() {
	var emu Emu = NewEmu()
	fmt.Println(emu.pc)

	op := 0xabcd
	v1 := (op >> 12) & 0xF
	v2 := (op >> 8) & 0xF
	v3 := (op >> 4) & 0xF
	v4 := op & 0xF

	fmt.Printf("%x-%x-%x-%x", v1, v2, v3, v4)
}
