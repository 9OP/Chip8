//go:build js && wasm
// +build js,wasm

package wasm

import (
	"syscall/js"

	"github.com/9op/Chip8/core"
)

type EmuWasm struct {
	emu *core.Emu
}

func (e *EmuWasm) Tick(this js.Value, args []js.Value) any {
	e.emu.Tick()
	return nil
}
func (e *EmuWasm) Reset(this js.Value, args []js.Value) any {
	e.emu.Reset()
	return nil
}
func (e *EmuWasm) Load(this js.Value, args []js.Value) any {
	length := args[0].Get("length").Int()
	rom := make([]byte, length)
	js.CopyBytesToGo(rom, args[0])
	e.emu.Load(rom)
	return nil
}
func (e *EmuWasm) Keypress(this js.Value, args []js.Value) any {
	var key string = args[0].String()
	var pressed bool = args[1].Bool()
	if idx, ok := core.KEYS[key]; ok {
		e.emu.Keypress(idx, pressed)
	}
	return nil
}
func (e *EmuWasm) DrawScreen(this js.Value, args []js.Value) any {
	scale := args[0].Int()
	display := e.emu.GetDisplay()
	doc := js.Global().Get("document")
	canvas := doc.Call("getElementById", "canvas")
	context := canvas.Call("getContext", "2d")

	for i := 0; i < core.SCREEN_WIDTH*core.SCREEN_HEIGHT; i++ {
		if display[i] {
			x := i % core.SCREEN_WIDTH
			y := i / core.SCREEN_WIDTH
			context.Call("fillRect", x*scale, y*scale, scale, scale)
		}
	}

	return nil
}

func main() {
	var emuWasm = EmuWasm{emu: core.NewEmu()}

	js.Global().Set("EmuTick", js.FuncOf(emuWasm.Tick))
	js.Global().Set("EmuReset", js.FuncOf(emuWasm.Reset))
	js.Global().Set("EmuKeypress", js.FuncOf(emuWasm.Keypress))
	js.Global().Set("EmuLoad", js.FuncOf(emuWasm.Load))
	js.Global().Set("EmuDrawScreen", js.FuncOf(emuWasm.DrawScreen))

	select {} // wait
}
