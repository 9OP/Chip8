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

type OpcodeInstruction = func(e *Emu, operands []uint8)

var OPCODES = map[uint16]OpcodeInstruction{
	0x0000: (*Emu).nop,
	0x00E0: (*Emu).cls,
	0x00EE: (*Emu).ret,
	0x1000: (*Emu).jmp,
	0x2000: (*Emu).call,
	0x3000: (*Emu).nextIf,
	0x4000: (*Emu).nextIfNot,
	0x5000: (*Emu).nextIfReg,
	0x6000: (*Emu).setReg,
	0x7000: (*Emu).incrReg,
	0x8000: (*Emu).copyReg,
	// 0x8001: 0,
	// 0x8002: 0,
	// 0x8003: 0,
	// 0x8004: 0,
	// 0x8005: 0,
	// 0x8006: 0,
	// 0x8007: 0,
	// 0x800E: 0,
	// 0x9000: 0,
	// 0xA000: 0,
	// 0xB000: 0,
	// 0xC000: 0,
	// 0xD000: 0,
	// 0xE09E: 0,
	// 0xE0A1: 0,
	// 0xF007: 0,
	// 0xF00A: 0,
	// 0xF015: 0,
	// 0xF018: 0,
	// 0xF01E: 0,
	// 0xF029: 0,
	// 0xF033: 0,
	// 0xF055: 0,
	// 0xF065: 0,
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

func (e *Emu) Tick() {
	// fetch
	op := e.fetch()
	// decode & execute
	e.execute(op)
}

func (e *Emu) Reset() {
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

func (e *Emu) push(val uint16) {
	e.stack[e.sp] = val
	e.sp += 1
}
func (e *Emu) pop() uint16 {
	e.sp -= 1
	return e.stack[e.sp]
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

func (e *Emu) execute(opcode uint16) {
	// 0xabcd -> 0xa / 0xb / 0xc / 0xd
	digit1 := uint8((opcode >> 12) & 0xF)
	digit2 := uint8((opcode >> 8) & 0xF)
	digit3 := uint8((opcode >> 4) & 0xF)
	digit4 := uint8(opcode & 0xF)

	switch digit1 {
	case 0x0:
		ops := []uint8{}
		getHandler(opcode)(e, ops)

	case 0x1:
	case 0x2:
	case 0x3:
	case 0x4:
	case 0x6:
	case 0x7:
	case 0xA:
	case 0xB:
	case 0xC:
	case 0xD:
		opcode &= 0xF000
		ops := []uint8{digit2, digit3, digit4}
		getHandler(opcode)(e, ops)

	case 0x5:
	case 0x8:
	case 0x9:
		opcode &= 0xF00F
		ops := []uint8{digit2, digit3}
		getHandler(opcode)(e, ops)

	case 0xE:
	case 0xF:
		opcode &= 0xF0FF
		ops := []uint8{digit2}
		getHandler(opcode)(e, ops)

	default:
		panic(fmt.Sprintf("opcode unimplemented %x", opcode))
	}

	// e.pc += 1
}

func getHandler(op uint16) OpcodeInstruction {
	if handler, found := OPCODES[op]; found {
		return handler
	}
	panic(fmt.Sprintf("opcode unimplemented %x", op))
}

/* OPCODE INSTRUCTION
 */
func (e *Emu) nop(ops []uint8) {
	e.pc += 1
}
func (e *Emu) cls(ops []uint8) {
	e.screen = [SCREEN_WIDTH * SCREEN_HEIGHT]bool{}
}
func (e *Emu) ret(ops []uint8) {
	ret_addr := e.pop()
	e.pc = ret_addr
}
func (e *Emu) jmp(ops []uint8) {
	// ex: {0xA,0xB,0xC} -> 0xABC
	nnn := uint16(ops[0])<<8 | uint16(ops[1])<<4 | uint16(ops[2])
	e.pc = nnn
}
func (e *Emu) call(ops []uint8) {
	nnn := uint16(ops[0])<<8 | uint16(ops[1])<<4 | uint16(ops[2])
	e.push(e.pc)
	e.pc = nnn
}
func (e *Emu) nextIf(ops []uint8) {
	reg := ops[0]
	nn := ops[1]<<4 | ops[2]
	if e.v_reg[reg] == nn {
		e.pc += 2
	}
}
func (e *Emu) nextIfNot(ops []uint8) {
	reg := ops[0]
	nn := ops[1]<<4 | ops[2]
	if e.v_reg[reg] != nn {
		e.pc += 2
	}
}
func (e *Emu) nextIfReg(ops []uint8) {
	reg1 := ops[0]
	reg2 := ops[1]
	if e.v_reg[reg1] == e.v_reg[reg2] {
		e.pc += 2
	}
}
func (e *Emu) setReg(ops []uint8) {
	reg := ops[0]
	val := ops[1]<<4 | ops[2]
	e.v_reg[reg] = val
}
func (e *Emu) incrReg(ops []uint8) {
	reg := ops[0]
	val := ops[1]<<4 | ops[2]
	e.v_reg[reg] += val
}
func (e *Emu) copyReg(ops []uint8) {
	reg1 := ops[0]
	reg2 := ops[1]
	e.v_reg[reg2] = e.v_reg[reg1]
}

func main() {
	var emu Emu = NewEmu()
	fmt.Println(emu.pc)

	// op := 0xabcd
	// v1 := (op >> 12) & 0xF
	// v2 := (op >> 8) & 0xF
	// v3 := (op >> 4) & 0xF
	// v4 := op & 0xF
	// fmt.Printf("%x-%x-%x-%x", v1, v2, v3, v4)

	// op := 0x1abc
	// op &= 0xF000
	// fmt.Printf("%x, %d, %d", op, 0xFFF, 0x0FFF)

	ops := [3]uint8{0xA, 0xB, 0xCC}
	fmt.Printf("%x", uint16(ops[0])<<8|uint16(ops[1])<<4|uint16(ops[2]))
}
