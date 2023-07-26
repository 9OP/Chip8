package core

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

var KEYS = map[string]uint8{
	"1": 0x1,
	"2": 0x2,
	"3": 0x3,
	"4": 0xC,
	"q": 0x4,
	"w": 0x5,
	"e": 0x6,
	"r": 0xD,
	"a": 0x7,
	"s": 0x8,
	"d": 0x9,
	"f": 0xE,
	"z": 0xA,
	"x": 0x0,
	"c": 0xB,
	"v": 0xF,
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

func NewEmu() *Emu {
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
	copy(newEmu.ram[0:FONTSET_SIZE], FONTSET[:])
	return &newEmu
}

func (e *Emu) GetDisplay() [SCREEN_WIDTH * SCREEN_HEIGHT]bool {
	return e.screen
}

func (e *Emu) Keypress(idx uint8, pressed bool) {
	e.keys[idx] = pressed
}

func (e *Emu) Load(data []uint8) {
	start := START_ADDR
	end := START_ADDR + len(data)
	copy(e.ram[start:end], data)
}

func (e *Emu) Tick() {
	op := e.fetch()
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
