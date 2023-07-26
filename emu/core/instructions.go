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

/*
 * OPCODE INSTRUCTIONS
 */
func (e *Emu) nop(ops []uint8) {
	// NOP
	e.pc += 1
}
func (e *Emu) cls(ops []uint8) {
	// CLS
	e.screen = [SCREEN_WIDTH * SCREEN_HEIGHT]bool{}
}
func (e *Emu) ret(ops []uint8) {
	// RET
	ret_addr := e.pop()
	e.pc = ret_addr
}
func (e *Emu) jp_a(ops []uint8) {
	// JUMP NNN
	// ex: {0xA,0xB,0xC} -> 0xABC
	nnn := uint16(ops[0])<<8 | uint16(ops[1])<<4 | uint16(ops[2])
	nnn &= 0xFFF
	e.pc = nnn
}
func (e *Emu) call_a(ops []uint8) {
	// CALL NNN
	nnn := uint16(ops[0])<<8 | uint16(ops[1])<<4 | uint16(ops[2])
	nnn &= 0xFFF
	e.push(e.pc)
	e.pc = nnn
}
func (e *Emu) se_x_b(ops []uint8) {
	// SKIP VX == NN
	x := ops[0]
	nn := ops[1]<<4 | ops[2]
	if e.v_reg[x] == nn {
		e.pc += 2
	}
}
func (e *Emu) sne_x_b(ops []uint8) {
	// SKIP VX != NN
	x := ops[0]
	nn := ops[1]<<4 | ops[2]
	if e.v_reg[x] != nn {
		e.pc += 2
	}
}
func (e *Emu) se_x_y(ops []uint8) {
	// SKIP VX == XY
	vx := e.v_reg[ops[0]]
	vy := e.v_reg[ops[1]]
	if vx == vy {
		e.pc += 2
	}
}
func (e *Emu) ld_x_b(ops []uint8) {
	// LOAD VX == NN
	x := ops[0]
	nn := ops[1]<<4 | ops[2]
	e.v_reg[x] = nn
}
func (e *Emu) add_x_b(ops []uint8) {
	// ADD VX += NN
	x := ops[0]
	nn := ops[1]<<4 | ops[2]
	e.v_reg[x] += nn
}
func (e *Emu) ld_x_y(ops []uint8) {
	// LOAD VX == VY
	e.v_reg[ops[0]] = e.v_reg[ops[1]]
}
func (e *Emu) or_x_y(ops []uint8) {
	// OR VX |= VY
	e.v_reg[ops[0]] |= e.v_reg[ops[1]]
}
func (e *Emu) and_x_y(ops []uint8) {
	// AND VX & VY
	e.v_reg[ops[0]] &= e.v_reg[ops[1]]
}
func (e *Emu) xor_x_y(ops []uint8) {
	// XOR VX ^= VY
	e.v_reg[ops[0]] ^= e.v_reg[ops[1]]
}
func (e *Emu) add_x_y(ops []uint8) {
	// ADD VX += VY
	x, vx, vy := ops[0], uint16(e.v_reg[ops[0]]), uint16(e.v_reg[ops[1]])
	sum := vx + vy
	if sum > 0xFF {
		e.v_reg[CARRY_FLAG] = 0x01
	}
	e.v_reg[x] = uint8(sum)
}
func (e *Emu) sub_x_y(ops []uint8) {
	// SUB VX -= VY
	x, vx, vy := ops[0], e.v_reg[ops[0]], e.v_reg[ops[1]]
	sub := vx - vy
	if vx > vy {
		e.v_reg[CARRY_FLAG] = 0x01
	} else {
		e.v_reg[CARRY_FLAG] = 0x00
	}
	e.v_reg[x] = sub
}
func (e *Emu) shr_x(ops []uint8) {
	// SHIFT RIGHT VX >>= 1
	x, vx := ops[0], e.v_reg[ops[0]]
	var mask uint8 = 0b0000_0001
	lsb := vx & mask
	e.v_reg[x] >>= 1
	e.v_reg[CARRY_FLAG] = lsb
}
func (e *Emu) subn_x_y(ops []uint8) {
	// SUBN VX = VY - VX
	x, vx, vy := ops[0], e.v_reg[ops[0]], e.v_reg[ops[1]]
	sub := vy - vx
	if vx > vy {
		e.v_reg[CARRY_FLAG] = 0x01
	} else {
		e.v_reg[CARRY_FLAG] = 0x00
	}
	e.v_reg[x] = sub
}
func (e *Emu) shl_x(ops []uint8) {
	// SHIFT LEFT VX <<= 1
	x, vx := ops[0], e.v_reg[ops[0]]
	var mask uint8 = 0b1000_0000
	msb := (vx >> 7) & mask // dropped bit
	e.v_reg[x] <<= 1
	e.v_reg[CARRY_FLAG] = msb
}
func (e *Emu) sne_x_y(ops []uint8) {
	// SKIP VX != VY
	vx, vy := e.v_reg[ops[0]], e.v_reg[ops[1]]
	if vx != vy {
		e.pc += 2
	}
}
func (e *Emu) ld_i_a(ops []uint8) {
	// LOAD I = NNN
	nnn := uint16(ops[0])<<8 | uint16(ops[1])<<4 | uint16(ops[2])
	nnn &= 0xFFF
	e.i_reg = nnn
}
func (e *Emu) jp_0_a(ops []uint8) {
	// JUMP V0 + NNN
	nnn := uint16(ops[0])<<8 | uint16(ops[1])<<4 | uint16(ops[2])
	nnn &= 0xFFF
	e.pc = uint16(e.v_reg[0x00]) + nnn
}
func (e *Emu) rnd_x_b(ops []uint8) {
	// RND VX = rnd & NN
	x := ops[0]
	nn := ops[1]<<4 | ops[2]
	random := uint8(rand.Uint32() % 256)
	e.v_reg[x] = random & nn
}
func (e *Emu) draw_x_y_n(ops []uint8) {
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
func (e *Emu) skp_x(ops []uint8) {
	// SKIP KEY PRESS
	vx := e.v_reg[ops[0]]
	if e.keys[vx] {
		e.pc += 2
	}
}
func (e *Emu) sknp_x(ops []uint8) {
	// SKIP KEY RELEASE
	vx := e.v_reg[ops[0]]
	if !e.keys[vx] {
		e.pc += 2
	}
}
func (e *Emu) ld_x_dt(ops []uint8) {
	// LOAD VX = DT
	x := ops[0]
	e.v_reg[x] = e.dt
}
func (e *Emu) ld_x_k(ops []uint8) {
	// WAIT KEY
	x := ops[0]
	pressed := false
	for i, is_pressed := range e.keys {
		if is_pressed {
			e.v_reg[x] = uint8(i)
			pressed = true
			break
		}
	}
	if !pressed {
		// Redo opcode
		e.pc -= 2
	}
}
func (e *Emu) ld_dt_x(ops []uint8) {
	// LOAD DT = VX
	vx := e.v_reg[ops[0]]
	e.dt = vx
}
func (e *Emu) ld_st_x(ops []uint8) {
	// LOAD ST = VX
	vx := e.v_reg[ops[0]]
	e.st = vx
}
func (e *Emu) add_i_x(ops []uint8) {
	// ADD I += VX
	vx := e.v_reg[ops[0]]
	e.i_reg += uint16(vx)
}
func (e *Emu) ld_i_f(ops []uint8) {
	// LOAD I = FONT ADDRESS
	vx := uint16(e.v_reg[ops[0]])
	addr := (vx * 5) % FONTSET_SIZE // font sprites take 5bytes each
	e.i_reg = addr
}
func (e *Emu) ld_bcd(ops []uint8) {
	// BCD
	vx, I := e.v_reg[ops[0]], e.i_reg
	e.ram[I] = uint8(vx / 100)         // hundreds
	e.ram[I+1] = uint8((vx / 10) % 10) // tens
	e.ram[I+2] = uint8(vx % 10)        // ones
}
func (e *Emu) ld_i_x(ops []uint8) {
	// STORE V0 - VX
	reg := e.v_reg[:]
	start := e.i_reg
	end := e.i_reg + uint16(len(reg))
	copy(e.ram[start:end], reg)
}
func (e *Emu) ld_x_i(ops []uint8) {
	// LOAD V0 - VX
	start := e.i_reg
	end := e.i_reg + uint16(len(e.v_reg))
	copy(e.v_reg[:], e.ram[start:end])
}
