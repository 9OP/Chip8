package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	SCREEN_WIDTH  = 64    // px
	SCREEN_HEIGHT = 32    // px
	RAM_SIZE      = 4096  // kb
	NUM_REGS      = 16    // #registers
	STACK_SIZE    = 16    // stack depth
	NUM_KEYS      = 16    // #keys
	START_ADDR    = 0x200 // start address of program
	FONTSET_SIZE  = 80    // (0-9A-F) 16 * 5 bytes
	CARRY_FLAG    = 0xF   // last register
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

type Operands = []uint8
type Instruction = func(e *Emu, ops Operands)

var OPCODES_TABLE = map[uint16]Instruction{
	0x0000: (*Emu).nop,
	0x00E0: (*Emu).cls,
	0x00EE: (*Emu).ret,
	0x1000: (*Emu).jmp,
	0x2000: (*Emu).call,
	0x3000: (*Emu).se_d,
	0x4000: (*Emu).sne_d,
	0x5000: (*Emu).se,
	0x6000: (*Emu).ld_d,
	0x7000: (*Emu).add_d,
	0x8000: (*Emu).ld,
	0x8001: (*Emu).or,
	0x8002: (*Emu).and,
	0x8003: (*Emu).xor,
	0x8004: (*Emu).add,
	0x8005: (*Emu).sub,
	0x8006: (*Emu).shr,
	0x8007: (*Emu).subn,
	0x800E: (*Emu).shl,
	0x9000: (*Emu).sne,
	0xA000: (*Emu).ldI,
	0xB000: (*Emu).jmp_0,
	0xC000: (*Emu).rng,
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
	// 0xABCD -> 0xA, 0xB, 0xC, 0xD
	digit1 := uint8((opcode >> 12) & 0xF)
	digit2 := uint8((opcode >> 8) & 0xF)
	digit3 := uint8((opcode >> 4) & 0xF)
	digit4 := uint8(opcode & 0xF)

	switch digit1 {
	case 0x0:
		ops := []uint8{}
		getInstruction(opcode)(e, ops)

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
		getInstruction(opcode)(e, ops)

	case 0x5:
	case 0x8:
	case 0x9:
		opcode &= 0xF00F
		ops := []uint8{digit2, digit3}
		getInstruction(opcode)(e, ops)

	case 0xE:
	case 0xF:
		opcode &= 0xF0FF
		ops := []uint8{digit2}
		getInstruction(opcode)(e, ops)

	default:
		panic(fmt.Sprintf("opcode unimplemented %x", opcode))
	}
	// e.pc += 1
}

func getInstruction(op uint16) Instruction {
	if instruction, found := OPCODES_TABLE[op]; found {
		return instruction
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
func (e *Emu) se_d(ops []uint8) {
	x := ops[0]
	nn := ops[1]<<4 | ops[2]
	if e.v_reg[x] == nn {
		e.pc += 2
	}
}
func (e *Emu) sne_d(ops []uint8) {
	x := ops[0]
	nn := ops[1]<<4 | ops[2]
	if e.v_reg[x] != nn {
		e.pc += 2
	}
}
func (e *Emu) se(ops []uint8) {
	vx := e.v_reg[ops[0]]
	vy := e.v_reg[ops[1]]
	if vx == vy {
		e.pc += 2
	}
}
func (e *Emu) ld_d(ops []uint8) {
	x := ops[0]
	nn := ops[1]<<4 | ops[2]
	e.v_reg[x] = nn
}
func (e *Emu) add_d(ops []uint8) {
	x := ops[0]
	nn := ops[1]<<4 | ops[2]
	e.v_reg[x] += nn
}
func (e *Emu) ld(ops []uint8) {
	x := ops[0]
	vy := e.v_reg[ops[1]]
	e.v_reg[x] = vy
}
func (e *Emu) or(ops []uint8) {
	x := ops[0]
	vy := e.v_reg[ops[1]]
	e.v_reg[x] |= vy
}
func (e *Emu) and(ops []uint8) {
	x := ops[0]
	vy := e.v_reg[ops[1]]
	e.v_reg[x] &= vy
}
func (e *Emu) xor(ops []uint8) {
	x := ops[0]
	vy := e.v_reg[ops[1]]
	e.v_reg[x] ^= vy
}
func (e *Emu) add(ops []uint8) {
	x, vx, vy := ops[0], e.v_reg[ops[0]], e.v_reg[ops[1]]
	sum := vx + vy
	// set carry flag on overflow
	if sum < vx || sum < vy {
		e.v_reg[CARRY_FLAG] = 0x01
	}
	e.v_reg[x] = sum
}
func (e *Emu) sub(ops []uint8) {
	x, vx, vy := ops[0], e.v_reg[ops[0]], e.v_reg[ops[1]]
	sub := vx - vy
	// unset carry flag on underflow
	if vx < vy {
		e.v_reg[CARRY_FLAG] = 0x00
	}
	e.v_reg[x] = sub
}
func (e *Emu) shr(ops []uint8) {
	x, vx := ops[0], e.v_reg[ops[0]]
	var mask uint8 = 0x01    // 0000 0001
	dropped_lsb := vx & mask // dropped bit
	e.v_reg[CARRY_FLAG] = dropped_lsb
	e.v_reg[x] >>= 1
}
func (e *Emu) subn(ops []uint8) {
	x, vx, vy := ops[0], e.v_reg[ops[0]], e.v_reg[ops[1]]
	sub := vy - vx
	// unset carry flag on underflow
	if vy < vx {
		e.v_reg[CARRY_FLAG] = 0x00
	}
	e.v_reg[x] = sub
}
func (e *Emu) shl(ops []uint8) {
	x, vx := ops[0], e.v_reg[ops[0]]
	var mask uint8 = 0x80           // 1000 0000
	dropped_msb := (vx & mask) >> 7 // dropped bit
	e.v_reg[CARRY_FLAG] = dropped_msb
	e.v_reg[x] <<= 1
}
func (e *Emu) sne(ops []uint8) {
	vx, vy := e.v_reg[ops[0]], e.v_reg[ops[1]]
	if vx != vy {
		e.pc += 2
	}
}
func (e *Emu) ldI(ops []uint8) {
	nnn := uint16(ops[0])<<8 | uint16(ops[1])<<4 | uint16(ops[2])
	e.i_reg = nnn
}
func (e *Emu) jmp_0(ops []uint8) {
	nnn := uint16(ops[0])<<8 | uint16(ops[1])<<4 | uint16(ops[2])
	e.pc = uint16(e.v_reg[0x00]) + nnn
}
func (e *Emu) rng(ops []uint8) {
	x := ops[0]
	nn := ops[1]<<4 | ops[2]
	rand.Seed(time.Now().UnixNano())
	random := uint8(rand.Uint32() % 256)
	e.v_reg[x] = random & nn
}
func (e *Emu) draw(ops []uint8) {
	/*
		Display n-byte sprite starting at memory location I at (Vx, Vy), set VF = collision.
		The interpreter reads n bytes from memory, starting at the address stored in I.
		These bytes are then displayed as sprites on screen at coordinates (Vx, Vy).
		Sprites are XORed onto the existing screen. If this causes any pixels to be erased,
		VF is set to 1, otherwise it is set to 0. If the sprite is positioned so part of
		it is outside the coordinates of the display, it wraps around to the opposite side
		of the screen.
	*/
	// x, y, n := e.v_reg[ops[0]], e.v_reg[ops[1]], ops[2]
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

	// ops := [3]uint8{0xA, 0xB, 0xCC}
	// fmt.Printf("%x", uint16(ops[0])<<8|uint16(ops[1])<<4|uint16(ops[2]))

	// // overflow
	// var op1 uint8 = 0xAA
	// var op2 uint8 = 0x56

	// var sum = op1 + op2
	// var overflow = sum < op1 || sum < op2
	// fmt.Printf("%x, %v\n", sum, overflow)

	// shift
	var val uint8 = 0x70 // 1111 0000
	fmt.Printf("%x\n", val>>1)
	fmt.Printf("%x\n", val<<1)
	var droppedBit = (val & 0x80) >> 7 // 1000 0000
	fmt.Printf("%x\n", droppedBit)
}
