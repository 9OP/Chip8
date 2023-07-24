# Chip8 emulator

Todo:
- document instruction set
- create a bin to assembly / assembly to bin program
- create a webassembly version

Compile: `GOOS=js GOARCH=wasm go build -o chip8.wasm`
Copy glue: `cp $(go env GOROOT)/misc/wasm/wasm_exec.js .`
start server: `python3 -m http.server`