package core

import (
	"fmt"
	"math/rand"
)

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
	upper_byte := e.ram[e.pc]
	lower_byte := e.ram[e.pc+1]
	var op uint16 = uint16(upper_byte)<<8 | uint16(lower_byte)
	e.pc += 2
	return op
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

	case 0x1, 0x2, 0x3, 0x4, 0x6, 0x7, 0xA, 0xB, 0xC, 0xD:
		opcode &= 0xF000
		ops := []uint8{digit2, digit3, digit4}
		getInstruction(opcode)(e, ops)

	case 0x5, 0x8, 0x9:
		opcode &= 0xF00F
		ops := []uint8{digit2, digit3}
		getInstruction(opcode)(e, ops)

	case 0xE, 0xF:
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
				var idx uint16 = uint16(x) + uint16(SCREEN_WIDTH)*uint16(y)
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
