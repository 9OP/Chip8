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

// Chip-8 original instruction set
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
	0xD000: (*Emu).draw,
	0xE09E: (*Emu).skp,
	0xE0A1: (*Emu).sknp,
	0xF007: (*Emu).ldvdt,
	0xF00A: (*Emu).ldk,
	0xF015: (*Emu).lddtv,
	0xF018: (*Emu).ldstv,
	0xF01E: (*Emu).addiv,
	0xF029: (*Emu).ldfv,
	0xF033: (*Emu).ldbv,
	0xF055: (*Emu).ldiv,
	0xF065: (*Emu).ldvi,
	// Extension set
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
	var mask uint8 = 0b0000_0001
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
	var mask uint8 = 0b1000_0000
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
	vx, vy, n, I := e.v_reg[ops[0]], e.v_reg[ops[1]], ops[2], e.i_reg
	sprite := e.ram[I : I+uint16(n)]
	flipped := false

	for row, pixel_row := range sprite {
		for pixel_idx := 0; pixel_idx < 8; pixel_idx++ {
			var mask uint8 = 0b1000_0000 >> pixel_idx // select individual pixel
			if (pixel_row & mask) != 0x0 {
				// wraps around the opposite side of the screen
				x := (vx + uint8(pixel_idx)) % SCREEN_WIDTH
				y := (vy + uint8(row)) % SCREEN_HEIGHT

				idx := x + SCREEN_WIDTH*y
				flipped = flipped || e.screen[idx] // erasion cause flip
				e.screen[idx] = !e.screen[idx]     // XOR pixels
			}
		}
	}

	if flipped {
		e.v_reg[CARRY_FLAG] = 0x01
	} else {
		e.v_reg[CARRY_FLAG] = 0x00
	}
}
func (e *Emu) skp(ops []uint8) {
	vx := e.v_reg[ops[0]]
	if e.keys[vx] {
		e.pc += 2
	}
}
func (e *Emu) sknp(ops []uint8) {
	vx := e.v_reg[ops[0]]
	if !e.keys[vx] {
		e.pc += 2
	}
}
func (e *Emu) ldvdt(ops []uint8) {
	x := ops[0]
	e.v_reg[x] = e.dt
}
func (e *Emu) ldk(ops []uint8) {
	// Blocking operation - wait for key press
	x := ops[0]
	pressed := false
	for i, k := range e.keys {
		if k {
			pressed = true
			e.v_reg[x] = uint8(i)
			break
		}
	}
	if !pressed {
		e.pc -= 2 // redo
	}
}
func (e *Emu) lddtv(ops []uint8) {
	vx := e.v_reg[ops[0]]
	e.dt = vx
}
func (e *Emu) ldstv(ops []uint8) {
	vx := e.v_reg[ops[0]]
	e.st = vx
}
func (e *Emu) addiv(ops []uint8) {
	vx := e.v_reg[ops[0]]
	e.i_reg += uint16(vx)
}
func (e *Emu) ldfv(ops []uint8) {
	// font sprites take 5bytes each
	vx := e.v_reg[ops[0]]
	addr := (vx * 5) % FONTSET_SIZE
	e.i_reg = uint16(addr)
}
func (e *Emu) ldbv(ops []uint8) {
	vx, I := e.v_reg[ops[0]], e.i_reg
	// BCD representation
	hundreds := uint8(vx / 100)
	tens := uint8((vx / 10) % 10)
	ones := uint8(vx % 10)
	e.ram[I] = hundreds
	e.ram[I+1] = tens
	e.ram[I+2] = ones
}
func (e *Emu) ldiv(ops []uint8) {
	copy(e.ram[e.i_reg:e.i_reg+uint16(len(e.v_reg))], e.v_reg[:])
}
func (e *Emu) ldvi(ops []uint8) {
	copy(e.v_reg[:], e.ram[e.i_reg:e.i_reg+uint16(len(e.v_reg))])
}

func main() {
	var emu Emu = NewEmu()
	fmt.Println(emu.pc)
}
